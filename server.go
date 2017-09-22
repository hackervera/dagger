package main

import (
	"bufio"
	"encoding/json"
	"log"
	"net"
	"strconv"
	"time"
)

// ClientManager holds state for connected clients.
type ClientManager struct {
	clients  []Client
	messages chan []byte
}

// Client represents a network connection.
type Client struct {
	conn net.Conn
}

// Response represents a response being sent back over the wire to client.
type Response struct {
	Result interface{} `json:"result"`
	Error  interface{} `json:"error"`
	Id     string      `json:"id"`
}

// Notification represents a notification being broadcasted to clients.
type Notification struct {
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}

func grabTime(t <-chan time.Time, cm *ClientManager, dag *Dag) {

	var data map[string]interface{}
	data = make(map[string]interface{})
	for i := range t {
		data["time"] = i.String()
		p("It's time: " + i.String())
		cm.messages <- dataToNode(dag, data)
	}
}

// registerConnection adds client to the client manager.
func registerConnection(conns chan net.Conn, clientManager *ClientManager, dag *Dag) {
	for conn := range conns {
		clientManager.clients = append(clientManager.clients, Client{conn: conn})
		go readConn(conn, dag)
		p("New connection")
	}
}

func dataToNode(dag *Dag, data map[string]interface{}) []byte {
	nodes := []Node{dag.AddNode(data)}
	notification := Notification{Method: "addNode", Params: nodes}
	nodeSlice, err := json.Marshal(notification)
	if err != nil {
		log.Fatal(err)
	}
	return nodeSlice
}

// notificationMontior broadcasts incoming []byte to all client connections.
func notificationMontior(cm *ClientManager) {
	for m := range cm.messages {
		p("Got message")
		for _, client := range cm.clients {
			p("Sending to clients")
			client.conn.Write(append(m, '\n'))
		}
	}
}

func startServer(dag *Dag) *ClientManager {
	var clientManager ClientManager
	clientManager.messages = make(chan []byte, 10)
	server, err := net.Listen("tcp", ":"+strconv.Itoa(1234))
	if server == nil {
		log.Fatal(err)
	}
	var cycle time.Duration = 3
	// Send the current time every cycle seconds onto channel t.
	t := time.Tick(cycle * time.Second)
	go grabTime(t, &clientManager, dag)
	conns := clientConns(server)
	go registerConnection(conns, &clientManager, dag)
	// go sendNodeNotification(&clientManager, dag)
	go notificationMontior(&clientManager)
	return &clientManager
}

func logConnection(listener net.Listener, ch chan net.Conn) {
	for {
		client, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		ch <- client

	}

}

func clientConns(listener net.Listener) chan net.Conn {
	ch := make(chan net.Conn, 10)
	go logConnection(listener, ch)
	return ch
}

// Ping returns a byte slice representing a ping notification.
func Ping() []byte {
	params := make([]interface{}, 0)
	notification, err := json.Marshal(Notification{Method: "ping", Params: params})
	if err != nil {
		log.Fatal(err)
	}
	return notification
}

// readConn listens to client connection and sends response based on rpc commands.
func readConn(client net.Conn, dag *Dag) {
	defer client.Close()
	bufReader := bufio.NewReader(client)
	for {
		buffer, err := bufReader.ReadBytes('\n')
		if err != nil {
			p("Client disconnect")
			return
		}
		var request map[string]interface{}
		json.Unmarshal(buffer, &request)
		switch request["method"] {
		case "getNode":
			p("triggered getnode")
			params := request["params"].([]interface{})
			hashes := []string{}
			for _, param := range params {
				hashes = append(hashes, param.(string))
			}
			node, err := dag.GetNode(hashes[0])
			var response Response
			if err != nil {
				response = Response{Result: nil, Error: err.Error(), Id: request["id"].(string)}
			} else {
				response = Response{Result: []Node{node}, Error: nil, Id: request["id"].(string)}
			}
			nodeJson, err := json.Marshal(response)
			if err != nil {
				log.Fatal(err)
			}
			client.Write(append(nodeJson, '\n'))
		case "allNodes":
			p("triggered allnodes")
			response := Response{Result: dag.GetNodes(), Error: nil, Id: request["id"].(string)}
			nodeJson, err := json.Marshal(response)
			if err != nil {
				log.Fatal(err)
			}
			client.Write(append(nodeJson, '\n'))
		default:
			p("method not recognized")
			p(request)
		}
	}
}
