package main

import (
	"fmt"
	"io"
	"net"
	"strings"
)

func log(v ...interface{}) {
	fmt.Println(v...)
}

type tcptransfer interface {
	startServer()
	working()
	register(conn net.Conn) string
	unbind(string) //去除连接绑定
	shutdown()
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
	log("start server at :", td.addr)
	for {
		conn, er := nl.Accept()
		if er != nil {
			log(er)
		}
		log("Accept connect from ", conn.RemoteAddr().String())
		ids := td.register(conn)
		go td.runloop(conn, ids)
	}
}

func (t *tsCbd) docking(b []byte, rtids string) bool {

	for _, v := range t.TsMaps {
		if v.status == 0 && v.ctype == 0 {
			s := "ack:" + v.ids + ":" + v.sec
			if strings.Compare(s, string(b[:len(b)])) == 0 {
				log("dockingAndBind success :", s, string(b[:len(b)]))
				v.pskConn = rtids
				v.status = 1
				return true
			}
			log("docking fail -->> ", s, string(b[:len(b)]))
		}
	}
	log("docking --->>> ", string(b[:len(b)]))
	return false
}

func (td *tsCbd) working() {

}

func (td *tsCbd) register(conn net.Conn) string {

	ids := string(krand(6, 3))
	if _, ok := td.TsMaps[ids]; ok {
		ids = string(krand(6, 3)) //如果已有则重新生成
	}
	var t *transMeta
	t = &transMeta{}
	t.remoteAddr = conn.RemoteAddr().String()
	t.ctype = 0
	t.status = 0
	t.sec = string(krand(3, 3))
	t.ids = ids

	td.TsMaps[t.ids] = t

	log("register new conn :", t.ids, " , sec :", t.sec)

	return ids
}

func (td *tsCbd) runloop(conn net.Conn, ids string) {
	defer func() {
		td.unbind(ids)
		conn.Close()
	}()

	buffer := make([]byte, 1024)
	stat := 0
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			log(conn.RemoteAddr().String(), ":", ids, "runloop()", err)
			break
		}
		log(conn.RemoteAddr().String(), ":", ids, ":", string(buffer[:n]))

		if stat == 0 {
			if ok := td.docking(buffer, ids); ok {
				log("dockingAndBind ok")
				myMeta := td.TsMaps[ids]
				rtMeta := td.TsMaps[myMeta.pskConn]
				go io.Coy(rtMeta.conn, myMeta.conn)
				io.Copy(myMeta.conn, rtMeta.conn)
				stat = 1
			}

		}

	}
}

func (td *tsCbd) unbind(ids string) {
	delete(td.TsMaps, ids) //去除已有的key
	for _, v := range td.TsMaps {
		if v.pskConn == ids {
			v.pskConn = ""
			v.ctype = 0
			v.status = 0
			log("unbind ok to :", v.ids)
		}
	}
	log("unbind ok by :", ids)
}

func (td *tsCbd) shutdown() {
	for _, v := range td.TsMaps {
		err := v.conn.Close()
		log(err)
		log("shutdown :", v.ids)
		delete(td.TsMaps, v.ids)
	}
}

func main() {
	fmt.Println("vim-go")

	var tt tcptransfer = &tsCbd{addr: ":6666", TsMaps: make(map[string]*transMeta)}
	defer func() {
		tt.shutdown()
	}()
	tt.startServer()
	tt.working()
}
