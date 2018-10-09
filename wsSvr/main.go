package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

type ClientManager struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

type Client struct {
	id       string
	socket   *websocket.Conn
	send     chan []byte
	group    string //default
	nickName string
}

type Message struct {
	Sender    string `json:"sender,omitempty"`
	Recipient string `json:"recipient,omitempty"`
	Content   string `json:"content,omitempty"`
	//Group     string `json:"content,omitempty"`
	NikeName string `json:"NikeName,omitempty"`
}

var manager = ClientManager{
	broadcast:  make(chan []byte),
	register:   make(chan *Client),
	unregister: make(chan *Client),
	clients:    make(map[*Client]bool),
}

func (manager *ClientManager) start() {
	for {
		select {
		case conn := <-manager.register:
			manager.clients[conn] = true
			jsonMessage, _ := json.Marshal(&Message{Content: fmt.Sprintf("/A new socket(%s) has connected.", conn.id), NikeName: conn.nickName})
			manager.send(jsonMessage, conn, conn.group)
		case conn := <-manager.unregister:
			if _, ok := manager.clients[conn]; ok {
				close(conn.send)
				delete(manager.clients, conn)
				jsonMessage, _ := json.Marshal(&Message{Content: fmt.Sprintf("/A socket(%s) has disconnected.", conn.id), NikeName: conn.nickName})
				manager.send(jsonMessage, conn, conn.group)
			}
		case message := <-manager.broadcast:
			for conn := range manager.clients {
				select {
				case conn.send <- message:
				default:
					close(conn.send)
					delete(manager.clients, conn)
				}
			}
		}
	}
}

func (manager *ClientManager) send(message []byte, ignore *Client, togroup string) {
	for conn := range manager.clients {
		if conn != ignore && conn.group == togroup {
			conn.send <- message
		}
	}
}

func (c *Client) read() {
	defer func() {
		manager.unregister <- c
		c.socket.Close()
	}()

	for {
		_, message, err := c.socket.ReadMessage()
		if err != nil {
			manager.unregister <- c
			c.socket.Close()
			break
		}
		jsonMessage, _ := json.Marshal(&Message{Sender: c.id, Content: string(message), NikeName: c.nickName})
		manager.broadcast <- jsonMessage
	}
}

func (c *Client) write() {
	defer func() {
		c.socket.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.socket.WriteMessage(websocket.TextMessage, message)
		}
	}
}

func main() {
	addr := flag.String("addr", ":12345", "listen at addr(default :12345)")
	flag.Parse()
	fmt.Printf("Starting application at %s... \n", *addr)
	go manager.start()
	http.HandleFunc("/ws", wsPage)
	http.HandleFunc("/", index)
	http.HandleFunc("/allws", allws)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		fmt.Println("start fail by error :", err)
	}
}

func wsPage(res http.ResponseWriter, req *http.Request) {
	conn, error := (&websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}).Upgrade(res, req, nil)
	if error != nil {
		http.NotFound(res, req)
		return
	}
	header := req.Header
	gp := header.Get("group")

	//fmt.Println("rec wsPage, gp = ", gp)
	if gp == "" {
		gp = "default"
	}
	//client := &Client{id: Randuuid(), socket: conn, send: make(chan []byte), group: gp}
	client := &Client{id: Randuuid(), socket: conn, send: make(chan []byte), group: gp, nickName: createName()}

	fmt.Println(client)
	manager.register <- client

	go client.read()
	go client.write()
}

func index(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte("test"))
}

func allws(rs http.ResponseWriter, req *http.Request) {
	for c, _ := range manager.clients {
		fmt.Println(c.id, c.group)
		rs.Write([]byte(fmt.Sprintf("group:%s,id:%s <br/>", c.group, c.id)))
	}
}
