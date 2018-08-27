// server1
package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
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
			//Log(conn.RemoteAddr().String(), "receive data:", buffer[:n])
			//Log(conn.RemoteAddr().String(), "receive data string:", string(buffer[:n]))
			msg := string(buffer[:n])
			repmsg := strings.ToUpper(msg)
			rebytemsg := []byte(repmsg)
			//log(rebytemsg)

			if buffer[0] == 0x05 {
				//只处理Socket5协议
				socket5Proxy(conn, buffer, n)
			} else {
				//支持默认连接
				proxyAll(conn, buffer, n)
			}
		}

	}
}

/**

 */
func proxyAll(conn net.Conn, b []byte, n int) {
	addr, method := pairse(b)

	//获得了请求的host和port，就开始拨号吧
	client, err := net.Dial("tcp", addr)
	if err != nil {
		log(err)
		return
	}
	log(addr)
	if method == "CONNECT" {
		fmt.Fprint(client, "HTTP/1.1 200 Connection established\r\n")
	} else {
		client.Write(b[:n])
	}
	go io.Copy(conn, client)
	io.Copy(client, conn)
}

/***
只处理Socket5协议
*/
func socket5Proxy(client net.Conn, b []byte, n int) {
	//客户端回应：Socket服务端不需要验证方式
	client.Write([]byte{0x05, 0x00})
	n, err := client.Read(b[:])
	var host, port string
	switch b[3] {
	case 0x01: //IP V4
		host = net.IPv4(b[4], b[5], b[6], b[7]).String()
	case 0x03: //域名
		host = string(b[5 : n-2]) //b[4]表示域名的长度
	case 0x04: //IP V6
		host = net.IP{b[4], b[5], b[6], b[7], b[8], b[9], b[10], b[11], b[12], b[13], b[14], b[15], b[16], b[17], b[18], b[19]}.String()
	}
	port = strconv.Itoa(int(b[n-2])<<8 | int(b[n-1]))

	server, err := net.Dial("tcp", net.JoinHostPort(host, port))
	if err != nil {
		log(err)
		return
	}
	defer server.Close()
	client.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) //响应客户端连接成功
	//进行转发
	go io.Copy(server, client)
	io.Copy(client, server)
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
