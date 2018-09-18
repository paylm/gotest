package main

import (
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net"
)

type ProxyManger interface {
	startServer() error
	working(*ProxySvr)
	register(net.Conn)
	shutdown()
	show()
}

type ProxySvr struct {
	port        string
	acpConn     net.Conn
	fwdConn     net.Conn
	fwdListener net.Listener
	sec         string
	msg         chan bool
}

type ProxyMng struct {
	addr  string
	Pxmap map[string]*ProxySvr
}

func getPort(start, end int) int {
	rand.Int()
	return 555
}

func newProxyListener() (net.Listener, string) {
	port := fmt.Sprintf(":%d", RandIntAt(20000, 30000))
	ln, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Printf("listen port %d fail \n", port)
	}
	fmt.Println("listen port ok :", port)
	return ln, port
}

func newProxySvr(acpConn net.Conn) *ProxySvr {

	//var pxln net.Listener
	pxs := &ProxySvr{acpConn: acpConn, msg: make(chan bool), sec: "sec6666"}
	pxln, port := newProxyListener()
	pxs.fwdListener = pxln
	pxs.port = port
	pxs.sec = "sec" + port
	return pxs
}

func (pm *ProxyMng) startServer() error {
	ln, err := net.Listen("tcp", pm.addr)
	if err != nil {
		// handle error
		return err
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
			fmt.Println("startServer fail", err)
		}
		pxs := newProxySvr(conn)
		go pm.working(pxs)
	}
	return nil
}

func (pm *ProxyMng) newProxySvr(acpConn net.Conn) {

}

func handelProxyConn(targetConn, remoteConn net.Conn) {

	defer func() {
		if targetConn != nil {
			targetConn.Close()
		}
		if remoteConn != nil {
			remoteConn.Close()
		}
	}()
	buffer := make([]byte, 1024)
Loop:
	for {
		n, err := targetConn.Read(buffer)
		if err != nil {
			fmt.Println("handelProxyConn err on read")
			break Loop
		}
		fmt.Println(targetConn.RemoteAddr().String(), "read => ", string(buffer[:n]))

		remoteConn.Write(buffer[:n])
		//io.Copy(targetConn, remoteConn)
		go io.Copy(targetConn, remoteConn)
		io.Copy(remoteConn, targetConn)
	}
}

func Pmrecover() {
	if err := recover(); err != nil {
		fmt.Println("panic info is :", err)
	}
}

func (pm *ProxyMng) working(ps *ProxySvr) {
	defer Pmrecover() //错误处理

	pm.Pxmap[ps.sec] = ps

	for {
		conn, err := ps.fwdListener.Accept()
		if err != nil {
			fmt.Printf("working %s ac fail \n ", ps.fwdListener.Addr().String)
		}

		go handelProxyConn(conn, ps.acpConn)
	}
}

func (pm *ProxyMng) register(acpConn net.Conn) {

}

func (pm *ProxyMng) show() {
	//
	for k, v := range pm.Pxmap {
		fmt.Printf("%s => %s", k, v.port)
	}
}

func (pm *ProxyMng) shutdown() {
}

func newProxyMng(laddr string) ProxyManger {

	var ipm ProxyManger
	pm := ProxyMng{addr: laddr}
	pm.Pxmap = make(map[string]*ProxySvr)
	ipm = &pm
	return ipm
}

func main() {
	var ipm ProxyManger
	port := flag.String("port", ":8080", "listent at port")
	flag.Parse()
	ipm = newProxyMng(*port)
	fmt.Println("start server at ", *port)
	err := ipm.startServer()
	if err != nil {
		fmt.Println("start proxy server fail")
	}
	defer ipm.shutdown()
}
