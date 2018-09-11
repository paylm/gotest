package main

import (
	"fmt"
	"net"
)

func log(v ...interface{}) {
	fmt.Println(v...)
}

type tcptransfer interface {
	startServer()
	working()
	register()
}

type transMeta struct {
	remoteAddr string
	remotePort int
	ctype      int    //连接类型 0 受连接端， 1 主动连接端
	ids        string //连接ID
	sec        string //连接密钥
	conn       net.Conn
	status     int    //连接状态,0 空闲，1 连接中
	pskConn    string //对端连接的连接名
}

//连接处理中心处理单元
type tsCbd struct {
	TsMaps map[string]*transMeta
	addr   string
	msg    chan int
}

func (td *tsCbd) startServer() {
	nl, err := net.Listen("tcp", td.addr)
	if err != nil {
		log("listen " + td.addr + " fail !!!!")
	}
	defer nl.Close()

	for {
		conn, er := nl.Accept()
		if er != nil {
			log(er)
		}

		td.register(conn)
		go runloop(conn)
	}
}

func (td *tsCbd) working() {

}

func (td *tsCbd) register(conn net.Conn) {

	var t *transMeta
	t = &transMeta{}
	t.remoteAddr = conn.RemoteAddr().String()
	t.ctype = 0
	t.status = 0
	t.sec = "sec-6666"
	t.ids = conn.RemoteAddr().String()

	td.TsMaps[t.ids] = t

}

func runloop(conn net.Conn) {
	defer func() {
		conn.Close()
	}()
	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			log(conn.RemoteAddr().String(), "runloop()", err)
		}
		log(conn.RemoteAddr().String(), ":", string(buffer[:n]))

	}
}

func main() {
	fmt.Println("vim-go")

	var tt tcptransfer = &tsCbd{addr: ":6666"}
	tt.startServer()
	tt.working()
	fmt.Println("dd")
}
