package main

import (
	// "math/rand"
	// "encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strconv"
	// "reflect"
	// "errors"
	"bufio"
	// "os"
	// "bytes"
)

type ClientManager struct {
	clients  []Client
	messages chan []byte
}

type Client struct {
	conn net.Conn
}

type Response struct {
	Result interface{} `json:"result"`
	Error  interface{} `json:"error"`
	Id     string      `json:"id"`
}

type Notification struct {
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}

func startServer(dag *Dag) (clientManager ClientManager) {
	server, err := net.Listen("tcp", ":"+strconv.Itoa(1234))
	if server == nil {
		log.Fatal(err)
	}
	clientManager.messages = make(chan []byte, 10)
	go IrcConn(clientManager.messages)
	conns := clientConns(server)
	go func() {
		for {
			conn := <-conns
			clientManager.clients = append(clientManager.clients, Client{conn: conn})
			go readConn(conn, dag)
			p("New connection")
		}
	}()
	// Broadcast message loop
	// message is []byte
	go func() {
		for {
			message := <-clientManager.messages
			var messageMap map[string]interface{}
			json.Unmarshal(message, &messageMap)
			nodes := []Node{dag.AddNode(messageMap)}
			notification := Notification{Method: "addNode", Params: nodes}
			nodeSlice, err := json.Marshal(notification)
			if err != nil {
				log.Fatal(err)
			}
			for _, client := range clientManager.clients {
				client.conn.Write(append(nodeSlice, '\n'))
			}
		}
	}()
	return clientManager
}

func clientConns(listener net.Listener) chan net.Conn {
	ch := make(chan net.Conn, 10)
	i := 0
	go func() {
		for {
			p("in server startup")
			client, err := listener.Accept()
			if err != nil {
				log.Fatal(err)
			}
			i++
			fmt.Printf("%d: %v <-> %v\n", i, client.LocalAddr(), client.RemoteAddr())
			ch <- client
		}
	}()
	return ch
}

func Ping() []byte {
	params := make([]interface{}, 0)
	notification, err := json.Marshal(Notification{Method: "ping", Params: params})
	if err != nil {
		log.Fatal(err)
	}
	return append(notification, '\n')
}

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
