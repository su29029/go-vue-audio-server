package main

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
)

// ClientManager ...
// manage all connections and clients
type ClientManager struct {
	Clients     map[*Client]bool   //all connections  Im.the clients' ips
	Users       map[string]*Client //login users(only login)  Im.the users' ids
	ClientsLock sync.RWMutex
	UsersLock   sync.RWMutex
	Connect     chan *Client //start recording process
	Disconnect  chan *Client //stop recording or occur errors process
	Broadcast   chan Message //sending channel
}

// Client ...
// manage a connected client
type Client struct {
	UserID string
	Addr   string
	Socket *websocket.Conn
	Send   chan Message
}

// Message ...
// manage a message
type Message struct {
	Cmd     string `json:"cmd"`
	UserID  string `json:"userID"`
	Message string `json:"msg"`
}

var clientManager = NewClientManager()

var userCh = make(chan string, 100)

var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// NewClientManager ...
// Create a new clientmanager to manage all clients
func NewClientManager() (clientManager *ClientManager) {
	clientManager = &ClientManager{
		Clients:    make(map[*Client]bool),
		Users:      make(map[string]*Client),
		Connect:    make(chan *Client, 100),
		Disconnect: make(chan *Client, 100),
		Broadcast:  make(chan Message, 1000),
	}
	return
}

// NewClient ...
// Create a new client map when a new client connected
func NewClient(addr string, userID string, socket *websocket.Conn) (client *Client) {
	for clients := range clientManager.Clients {
		if clients.UserID == userID && clients.Addr == addr {
			return nil
		}
	}
	client = &Client{
		UserID: userID,
		Addr:   addr,
		Socket: socket,
		Send:   make(chan Message, 100),
	}
	return client
}

// NewConnection ...
// create a new ws connection when build a new connection from client.
// handle function of '/ws/audio/start/:userID'.
func NewConnection(ctx *gin.Context) {
	ws, err := upGrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"cmd":    "connect",
			"userID": ctx.Param("userID"),
			"msg":    "connection update failed",
		})
		return
	}
	userID := <-userCh
	client := NewClient(ws.RemoteAddr().String(), userID, ws)
	if client != nil {
		fmt.Println("A new connection,userID is", userID)
		go client.read()
		go client.write()
		clientManager.Connect <- client
	} else {
		// ws.Close()
		msg := Message{Cmd: "start", UserID: client.UserID, Message: "repeated"}
		ws.WriteJSON(&msg)
		ws.Close()
	}

}

// AddClients ...
// Add a new client when there is a new connection
func (manager *ClientManager) AddClients(client *Client) {
	manager.ClientsLock.Lock()
	defer manager.ClientsLock.Unlock()

	clientManager.Clients[client] = true
}

// DeleteClients ...
// delete the disconnected client when a client is disconnected
func (manager *ClientManager) DeleteClients(client *Client) {
	manager.ClientsLock.Lock()
	defer manager.ClientsLock.Unlock()
	fmt.Println("before delete:")
	for key, value := range clientManager.Clients {
		fmt.Println("key:", key, "value:", value)
	}
	delete(clientManager.Clients, client)
	fmt.Println("after delete:")
	for key, value := range clientManager.Clients {
		fmt.Println("key:", key, "value:", value)
	}
	// if err != nil {
	// 	fmt.Println(err)
	// }
}

// AddUsers ...
// Add a user when a new client is connected
// abandoned.
func (manager *ClientManager) AddUsers(userID string, client *Client) {
	manager.UsersLock.Lock()
	defer manager.UsersLock.Unlock()

	clientManager.Users[userID] = client
}

// DeleteUsers ...
// delete the disconnected user
// abandoned.
func (manager *ClientManager) DeleteUsers(userID string, client *Client) (result bool) {
	manager.UsersLock.Lock()
	defer manager.UsersLock.Unlock()

	if value, ok := manager.Users[userID]; ok { //value is ("manager.Users[userID]")'s value
		if value.Addr != client.Addr {
			return
		}
		delete(manager.Users, userID)
		result = true
	}
	return
}

// ProcessData ...
// check the command type and deal with the data
func ProcessData(client *Client, message Message) {
	// check the message's cmd.

	// record command, send the messages to all clients.
	// stop command,close the client, delete the user and send messages to all clients that xxx client has disconnected.
	// heartbeat command
	switch message.Cmd {
	case "record":
		clientManager.Broadcast <- message
	case "close":
		message.Message = "stop"
		client.SendAll(message)
		time.Sleep(1e7)
		clientManager.Disconnect <- client
	case "heartbeat":
		message.Message = "ok"
		client.SendMsg(message)
	}

}

// SendMsg ...
// use channel to send the message
func (c *Client) SendMsg(msg Message) {
	c.Send <- msg
	fmt.Println("Send:", msg.UserID)
}

// SendAll ...
// use the broadcast channel to send message to all clients
func (c *Client) SendAll(msg Message) {
	clientManager.Broadcast <- msg
}

// read ...
// read the message from websocket
func (c *Client) read() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("read stop", string(debug.Stack()), r)
		}
	}()

	defer func() {
		clientManager.Disconnect <- c
	}()

	for {
		var message Message
		err := c.Socket.ReadJSON(&message)
		if err != nil {
			fmt.Println("读取数据错误", c.Addr, err)
			return
		}

		ProcessData(c, message)
	}
}

// write ...
// write the message by websocket
func (c *Client) write() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("write stop", string(debug.Stack()), r)
		}
	}()
	defer func() {
		clientManager.Disconnect <- c
		// c.Socket.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				fmt.Println("发送数据错误:", c.UserID)

				return
			}
			if message.Message == "stop" {
				fmt.Println("stop message:", message)
			}
			c.Socket.WriteJSON(message)
		}
	}
}

// EventConnect ...
// Deal with a connect event
func (manager *ClientManager) EventConnect(client *Client) {
	//add client and add user.
	clientManager.AddClients(client)
	msg := Message{Cmd: "start", UserID: client.UserID, Message: "OK"}
	client.SendAll(msg)
}

// EventDisconnect ...
// Deal with a disconnect event
func (manager *ClientManager) EventDisconnect(client *Client) {
	//delete client and delete user.
	clientManager.DeleteClients(client)
	fmt.Println("断开连接:", client.UserID)
	// client.Socket.Close()
}

// GetClients ...
// Get all connected clients
func (manager *ClientManager) GetClients() (clients map[*Client]bool) {
	clients = make(map[*Client]bool)
	manager.ClientsLock.RLock()
	defer manager.ClientsLock.RUnlock()

	manager.ClientsRange(func(client *Client, value bool) (result bool) {
		clients[client] = value

		return true
	})

	return
}

// ClientsRange ...
// range all clients
func (manager *ClientManager) ClientsRange(f func(client *Client, value bool) (result bool)) {
	manager.ClientsLock.RLock()
	defer manager.ClientsLock.RUnlock()

	for key, value := range clientManager.Clients {
		result := f(key, value)
		if result == false {
			return
		}
	}
}

// start ...
// main dealing and task distribute function
func (manager *ClientManager) start() {
	for {
		select {
		case conn := <-manager.Connect:
			manager.EventConnect(conn)
		case conn := <-manager.Disconnect:
			manager.EventDisconnect(conn)
		case message := <-manager.Broadcast: // first,get all clients,second,use channel to send messages.
			clients := manager.GetClients()
			for conn := range clients { // use channel, in order to ensure thread safety.
				//TODO:use channel to send messages.
				select {
				case conn.Send <- message:
				default:
					close(conn.Send)
				}
			}
		}
	}
}

func main() {
	engine := gin.Default()
	engine.Use(cors.Default())
	audio := engine.Group("/ws")
	{
		audio.POST("/saveuser", SaveUser)
		audio.GET("/audio/start/", NewConnection)
	}
	go clientManager.start()
	engine.RunTLS(":8560", "fullchain.pem", "privkey.pem")

}

// SaveUser ...
// get the userID and send to function 'NewConnection' to create a new client.
// handle function of '/saveuser'.
func SaveUser(ctx *gin.Context) {
	var msg Message
	fmt.Println(ctx)
	err := ctx.BindJSON(&msg)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"msg": "failed",
		})
	} else {
		userCh <- msg.UserID
		ctx.JSON(http.StatusOK, gin.H{
			"msg":    "success",
			"userID": msg.UserID,
		})
	}
}
