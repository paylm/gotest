package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

type Dialer interface {
	Dial() (conn net.Conn, err error)
}

type tcpDial struct {
	addr string
}

func (d *tcpDial) Dial() (conn net.Conn, err error) {
	conn, err = net.DialTimeout("tcp", d.addr, time.Second*5)
	return
}

type Agent struct {
	upConn   net.Conn
	downConn net.Conn
	timeout  time.Duration
	Dialer
}

func NewAgent(conn net.Conn, timeout time.Duration, dialer Dialer) *Agent {
	ag := new(Agent)
	ag.timeout = timeout
	ag.downConn = &timeoutConn{conn, ag.timeout}
	ag.Dialer = dialer
	return ag
}

func (agent *Agent) Server() {
	conn, err := agent.Dial()
	if err != nil {
		log.Fatal("fail to dailer ", err)
	}
	agent.upConn = &timeoutConn{conn, agent.timeout}

	defer agent.close()
	agent.transport()
}

func (agent *Agent) close() {
	var eu, ed error
	if agent.downConn != nil {
		ed = agent.downConn.Close()
	}
	if agent.upConn != nil {
		eu = agent.upConn.Close()
	}
	if eu != nil || ed != nil {
		log.Printf("agent.close() error. (eu: %v, ed: %v)\n", eu, ed)
	}
	return
}

func (agent *Agent) transport() {
	ch := make(chan error)
	defer close(ch)
	buf := make([]byte, 1024)
loop:
	for {
		n, err := agent.downConn.Read(buf)
		if err != nil {
			log.Println("transport err :", err)
			break loop
		}
		agent.upConn.Write(buf[:n])

		go func() {
			_, e := io.Copy(agent.downConn, agent.upConn)
			ch <- e
		}()
		_, ed := io.Copy(agent.upConn, agent.downConn)
		eu := <-ch
		log.Printf("transport has error (ed:%s,eu:%s)\n", ed, eu)
	}
}

type Server struct {
	dp string `listen port when in listen mode`
	fp string `forward port`
	rp string `revers to remoteAddr`
}

func (s *Server) listenAndServer() {

	ln, err := net.Listen("tcp", s.dp)
	if err != nil {
		log.Fatal("listent fail", err)
		return
	}
	for {
		conn, er := ln.Accept()
		if er != nil {
			log.Fatal("accept conn fail ", er)
		}

		dial := &tcpDial{s.fp}
		//fmt.Println("%v dial at %v", conn.RemoteAddr().String(), dial)
		ag := NewAgent(conn, time.Second*5, dial)
		go ag.Server()
	}
}

func (s *Server) reverAndProxy() {

	fd := &tcpDial{s.rp}
	upstream, err := fd.Dial()
	if err != nil {
		log.Fatal("reverAndProxy fail", err)
	}
	dial := &tcpDial{s.fp}
	ag := NewAgent(upstream, time.Second*5, dial)
	ag.Server()
}

func main() {
	fmt.Println("start client ")
	fp := flag.String("fp", ":6060", "forward src port")
	dp := flag.String("dp", ":8080", "forwaro dst port")
	rp := flag.String("rp", "", "forwaro dst port")
	flag.Parse()
	fmt.Printf("start agent at fp:%v , dp:%v,rp:%v", *fp, *dp, *rp)
	var s Server
	if *rp == "" {
		s = Server{fp: *fp, dp: *dp}
		s.listenAndServer()
	} else {
		s = Server{fp: *fp, rp: *rp}
		s.reverAndProxy()
	}
}
