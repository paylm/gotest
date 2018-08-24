// server1
package main

import (
	"fmt"
	"net"
	"runtime"
	"strings"
	"sync"
)

type Sver struct {
	connect net.Conn
}

type chats interface {
	handleConnection(conn net.Conn, wg sync.WaitGroup)
	handleConnectiontoAll(conn net.Conn, wg sync.WaitGroup, conns map[string]net.Conn)
}

func Log(v ...interface{}) {
	fmt.Println(v...)
}

func handleConnection(conn net.Conn, wg sync.WaitGroup) {
	defer func() {
		//关闭连接
		Log(conn.RemoteAddr().String(), "close connect:")
		conn.Close()
		wg.Done()
	}()
	buffer := make([]byte, 512)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			Log(conn.RemoteAddr().String(), " connection error: ", err)
			return
		}
		Log(conn.RemoteAddr().String(), "receive data length:", n)
		if n > 1 {
			Log(conn.RemoteAddr().String(), "receive data:", buffer[:n])
			Log(conn.RemoteAddr().String(), "receive data string:", string(buffer[:n]))
			msg := string(buffer[:n])
			repmsg := strings.ToUpper(msg)
			rebytemsg := []byte(repmsg)
			conn.Write(rebytemsg)
		}

	}
}

func handleConnectiontoAll(conn net.Conn, wg sync.WaitGroup, conns map[string]net.Conn) {
	defer func() {
		//关闭连接
		Log(conn.RemoteAddr().String(), "close connect:")
		msg := []byte(conn.RemoteAddr().String() + " has exit chat group")
		sendBroadcast(conn.RemoteAddr().String(), conns, msg)
		delete(conns, conn.RemoteAddr().String())
		conn.Close()
		wg.Done()
	}()

	buffer := make([]byte, 512)

	rebytemsg := []byte(conn.RemoteAddr().String() + " join chat group")
	sendBroadcast(conn.RemoteAddr().String(), conns, rebytemsg)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			Log(conn.RemoteAddr().String(), " connection error: ", err)
			return
		}
		Log(conn.RemoteAddr().String(), "receive data length:", n)
		if n > 1 {
			Log(conn.RemoteAddr().String(), "receive data:", buffer[:n])
			Log(conn.RemoteAddr().String(), "receive data string:", string(buffer[:n]))
			msg := string(buffer[:n])
			repmsg := strings.ToUpper(msg)
			rebytemsg := []byte(conn.RemoteAddr().String() + " say: " + repmsg)
			//conn.Write(rebytemsg)
			sendBroadcast(conn.RemoteAddr().String(), conns, rebytemsg)
		}
	}
}

func sendBroadcast(myconn string, conns map[string]net.Conn, msg []byte) {
	for k, cc := range conns {
		if k != myconn {
			cc.Write(msg)
		}
	}
}

func main() {
	//做回收处理
	defer closeAll()

	wg := sync.WaitGroup{}
	wg.Add(runtime.NumCPU())
	allconn := make(map[string]net.Conn)

	ln, err := net.Listen("tcp", ":7777")
	if err != nil {
		// handle error
	}
	for {
		fmt.Println("ready for recive....")
		conn, err := ln.Accept()
		allconn[conn.RemoteAddr().String()] = conn
		if err != nil {
			// handle error
		}
		//go handleConnection(conn,wg)
		go handleConnectiontoAll(conn, wg, allconn)
	}
	wg.Wait()
}

func closeAll() {
	fmt.Println("defer->closeAll()")
	if r := recover(); r != nil {
		fmt.Println("Recovered in f", r)
	}
}
