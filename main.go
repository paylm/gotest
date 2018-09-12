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

type tsChannel struct {
	remoteAddr string
	remotePort int
	ctype      int    //连接类型 0 受连接端， 1 主动连接端
	ids        string //连接ID
	sec        string //连接密钥
	conn       net.Conn
	status     int       //连接状态,0 空闲，1 连接中
	pskConn    string    //对端连接的连接名
	quit       chan bool //关闭信号
}

//连接处理中心处理单元
type tsCbd struct {
	TsMaps map[string]*tsChannel
	addr   string
	msg    chan int
}

func (td *tsCbd) startServer() {
	nl, err := net.Listen("tcp", td.addr)
	if err != nil {
		log("listen " + td.addr + " fail !!!!")
	}
	defer func() {
		err := nl.Close()
		log(err)
	}()
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

func (td *tsCbd) docking(b []byte, rtids string) bool {

	for _, v := range td.TsMaps {
		if v.status == 0 && v.ctype == 0 {
			if v.ids == rtids {
				//只能与其它连接进行绑定
				continue
			}
			s := "ack:" + v.ids + ":" + v.sec
			if strings.Compare(s, string(b[:14])) == 0 {
				log("dockingAndBind success :", s, string(b[:14]))
				v.pskConn = rtids
				v.status = 1
				td.TsMaps[rtids].pskConn = v.ids
				td.TsMaps[rtids].status = 1
				return true
			}
			log("docking fail -->> ", s, string(b[:14]))
		}
	}
	log("docking --->>> ", string(b[:14]))
	return false
}

func (td *tsCbd) working() {

}

func (td *tsCbd) register(conn net.Conn) string {

	ids := string(krand(6, 3))
	if _, ok := td.TsMaps[ids]; ok {
		ids = string(krand(6, 3)) //如果已有则重新生成
	}
	var t *tsChannel
	t = &tsChannel{}
	t.remoteAddr = conn.RemoteAddr().String()
	t.ctype = 0
	t.status = 0
	t.sec = string(krand(3, 3))
	t.ids = ids
	t.conn = conn
	t.quit = make(chan bool)
	td.TsMaps[t.ids] = t
	log("register new conn :", t.ids, " , sec :", t.sec)

	return ids
}

func (td *tsCbd) runloop(conn net.Conn, ids string) {
	defer func() {
		td.unbind(ids)
		err := conn.Close()
		log("funloop close :", ids, err)
	}()

	buffer := make([]byte, 1024)
	myChannel := td.TsMaps[ids]

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			log(conn.RemoteAddr().String(), ":", ids, "runloop()", err)
			break
		}
		log(conn.RemoteAddr().String(), ":", ids, ":", string(buffer[:n]))

		if myChannel.status == 0 {
			if ok := td.docking(buffer, ids); ok {
				log("dockingAndBind ok")
				rtChannel := td.TsMaps[myChannel.pskConn]
				connectChannel(myChannel, rtChannel)
			}
		} else {
			rtChannel := td.TsMaps[myChannel.pskConn]
			connectChannel(myChannel, rtChannel)
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

/***
 连接两个单元
**/
func connectChannel(fromChannel, toChannel *tsChannel) {

	defer func() {
		//处理关闭连接时的painc 处理
		if err := recover(); err != nil {
			log(err)
			if fromChannel != nil {
				fromChannel.status = 0
				fromChannel.ctype = 0
				fromChannel.pskConn = ""
			}
			if toChannel != nil {
				toChannel.status = 0
				toChannel.ctype = 0
				toChannel.pskConn = ""
			}
		}
	}()

	go io.Copy(toChannel.conn, fromChannel.conn)
	io.Copy(fromChannel.conn, toChannel.conn)
	fromChannel.status = 1
	toChannel.status = 1
	fromChannel.ctype = 1
	toChannel.ctype = 1
	toChannel.pskConn = fromChannel.ids
}

func main() {
	fmt.Println("run go pass server success")

	var tt tcptransfer = &tsCbd{addr: ":6666", TsMaps: make(map[string]*tsChannel)}
	defer func() {
		tt.shutdown()
	}()
	tt.startServer()
	tt.working()
}
