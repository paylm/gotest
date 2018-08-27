// server1
package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"
)

type Sver struct {
	connect net.Conn
}

func log(v ...interface{}) {
	fmt.Println(v...)
}

type chats interface {
	handleConnection(conn net.Conn)
}

func Log(v ...interface{}) {
	fmt.Println(v...)
}

func handleConnection(conn net.Conn) {
	defer func() {
		//关闭连接
		Log(conn.RemoteAddr().String(), "close connect:")
		conn.Close()
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
			log(rebytemsg)
			addr, method := pairse(buffer)

			//获得了请求的host和port，就开始拨号吧
			client, err := net.Dial("tcp", addr)
			if err != nil {
				log(err)
				return
			}
			if method == "CONNECT" {
				fmt.Fprint(client, "HTTP/1.1 200 Connection established\r\n")
			} else {
				client.Write(buffer[:n])
			}
			go io.Copy(conn, client)
			io.Copy(client, conn)

		}

	}
}

func pairse(b []byte) (string, string) {
	var method, host, address string
	fmt.Sscanf(string(b[:bytes.IndexByte(b[:], '\n')]), "%s%s", &method, &host)
	hostPortURL, err := url.Parse(host)
	if err != nil {
		log(err)
		return "", "CONNECT"
	}

	if hostPortURL.Opaque == "443" { //https访问
		address = hostPortURL.Scheme + ":443"
	} else { //http访问
		if strings.Index(hostPortURL.Host, ":") == -1 { //host不带端口， 默认80
			address = hostPortURL.Host + ":80"
		} else {
			address = hostPortURL.Host
		}
	}
	return address, method
}

func main() {

	ln, err := net.Listen("tcp", ":8888")
	if err != nil {
		// handle error
	}
	log("listen at :8888 ")
	//做回收处理
	defer closeAll(&ln)
	for {
		fmt.Println("ready for recive....")
		conn, err := ln.Accept()
		if err != nil {
			// handle error
		}
		//go handleConnection(conn,wg)
		go handleConnection(conn)
	}

}

func closeAll(ln *net.Listener) {
	fmt.Println("defer->closeAll()")
	if r := recover(); r != nil {
		fmt.Println("Recovered in f", r)
	}
}
