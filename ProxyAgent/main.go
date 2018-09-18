package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"time"
)

func handleConnection(conn net.Conn, toport string, quit chan bool) {
	var targetconn net.Conn
	var er error
	defer func() {
		if conn != nil {
			conn.Close()
		}
		if targetconn != nil {
			targetconn.Close()
		}
		fmt.Println("handleConnection defer close")
		quit <- true
	}()
	buffer := make([]byte, 1024)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			fmt.Println("handleConnection read err:", conn.RemoteAddr().String(), err)
			goto Loop
		}
		//fmt.Println(string(buffer[:n]))
		if targetconn == nil {
			targetconn, er = net.Dial("tcp", toport)
			if er != nil {
				fmt.Println("handleConnection", toport, er)
				conn.Write([]byte("proxyPort conn fail !!!"))
				continue
			}
		}
		n, err = targetconn.Write(buffer[:n])
		if err != nil {
			fmt.Printf("Unable to write to output, error: %s\n", err.Error())
			//conn.Close()
			targetconn.Close()
			targetconn = nil
			continue
		}
		//go proxyRequest(conn, targetconn)
		//go proxyRequest(targetconn, conn)
		//conn.SetReadDeadline(time.Time{}.Add(time.Second * 3))
		go io.Copy(targetconn, conn)
		io.Copy(conn, targetconn)
		//go proxyRequest(conn, targetconn)
		fmt.Println("after io.Copy")
	}
Loop:
	fmt.Println("handleConnection conn reset")
}

// Forward all requests from r to w
func proxyRequest(r net.Conn, w net.Conn) {
	defer r.Close()
	defer w.Close()

	var buffer = make([]byte, 40960)
	for {
		n, err := r.Read(buffer)
		if err != nil {
			fmt.Printf("Unable to read from input, error: %s\n", err.Error())
			break
		}

		n, err = w.Write(buffer[:n])
		if err != nil {
			fmt.Printf("Unable to write to output, error: %s\n", err.Error())
			break
		}
	}
}

func startProxy(remoteAddr string, targetPort string, exitMsg chan bool) {

	quitConn := make(chan bool)
	/*
		ln, err := net.Listen("tcp", remoteAddr)
		if err != nil {
			// handle error
		}
		for {
			conn, err := ln.Accept()
			if err != nil {
				// handle error
				fmt.Println(err)
			}
			go handleConnection(conn, targetPort, quitConn)
		}
	*/
	var conn net.Conn
	var err error
	retry, delay := 1, 5 //失败连接次数>retry ，增加重试间隔
	for {
		if conn == nil {
			conn, err = net.Dial("tcp", remoteAddr)
			if err != nil {
				fmt.Println(err)
				fmt.Printf("try to connect %d s later , fail conn count : %d \n", delay, retry)
				time.Sleep(time.Duration(delay) * time.Second)
				if retry > 100 {
					delay = 30
				}
				retry++
				continue
			}
			retry = 1
			delay = 5
		}
		go handleConnection(conn, targetPort, quitConn)
		<-quitConn
		conn = nil
	}
	exitMsg <- true
}

func main() {

	targetPort := flag.String("fp", ":6060", "tcp Forward port")
	remoteAddr := flag.String("dp", ":8080", "tcp  remote port")
	flag.Parse()
	fmt.Println("run server at :", *remoteAddr)
	fmt.Println("target port", *targetPort)
	quit := make(chan bool)
	startProxy(*remoteAddr, *targetPort, quit)
	<-quit
}
