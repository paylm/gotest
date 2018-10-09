package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

func main() {

	addr := flag.String("addr", "localhost:12345", "http service address")
	gp := flag.String("group", "default", "websocket group")
	flag.Parse()
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	fmt.Println(u)
	var dialer *websocket.Dialer
	header := make(http.Header)
	header.Add("group", *gp)
	conn, _, err := dialer.Dial(u.String(), header)
	if err != nil {
		fmt.Println(err)
		return
	}

	go timeWriter(conn)
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("read:", err)
			return
		}

		fmt.Printf("received: %s\n", message)
	}

}

func timeWriter(conn *websocket.Conn) {
	for {
		time.Sleep(time.Second * 5)
		conn.WriteMessage(websocket.TextMessage, []byte(time.Now().Format("2006-01-02 15:04:05")))
	}
}
