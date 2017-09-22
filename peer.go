package main

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"log"
	"math/rand"
	"net"
)

type Request struct {
	Method string   `json:"method"`
	Params []string `json:"params"`
	Id     string   `json:"id"`
}

type ResponseHandler struct {
	Function func(interface{}, *Dag, *RequestManager)
}

type RequestManager struct {
	Requests       map[string]ResponseHandler
	RequestChannel chan []byte
	Dag            *Dag
}

func RandomId() string {
	id := make([]byte, 8)
	rand.Read(id)
	return hex.EncodeToString(id)
}

func (rm *RequestManager) NodeRequest(hash string) {
	id := RandomId()
	rm.Requests[id] = ResponseHandler{Function: ProcessNode}
	request := Request{Method: "getNode", Params: []string{hash}, Id: id}
	output, err := json.Marshal(request)
	if err != nil {
		log.Fatal(err)
	}

	rm.RequestChannel <- output
}

func (rm *RequestManager) AllNodes() {
	id := RandomId()
	// rm.Requests[id] = ResponseHandler{Function: rm.Dag.AddNodes}
	request := Request{Method: "allNodes", Params: []string{}, Id: id}
	output, err := json.Marshal(request)
	if err != nil {
		log.Fatal(err)
	}

	rm.RequestChannel <- output

}

func connectPeer(requests <-chan []byte, responses chan<- []byte) {
	servAddr := "localhost:1234"
	tcpAddr, err := net.ResolveTCPAddr("tcp", servAddr)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		for {
			request := <-requests
			request = append(request, '\n')
			p("Sending: " + string(request))
			conn.Write(request)
		}
	}()
	go func() {
		bufReader := bufio.NewReader(conn)
		for {
			response, err := bufReader.ReadBytes('\n')
			if err != nil {
				log.Fatal(err)
			}
			p("Response: " + string(response))
			responses <- response
		}
	}()
}
