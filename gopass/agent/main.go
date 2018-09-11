package main

import (
	"fmt"
	"net"
)

func log(v ...interface{}) {
	fmt.Println(v...)
}

type agent struct {
	addr string

	src net.Conn
}

func (ag *agent) dial() {

	conn, err := net.Dial("tcp", ag.addr)
	if err != nil {
		log(err)
	}
	defer ag.shutdown()
	ag.src = conn
}

func (ag *agent) shutdown() {
	if ag.src != nil {
		err := ag.src.Close()
		log(err)
	}
}

func main() {
	fmt.Println("vim-go")
	fmt.Println("go agent test")
	ag := &agent{addr: ":6666"}
	ag.dial()
}
