package main

import (
	"net"
	// "bufio"
	"encoding/json"
	"fmt"
	"github.com/thoj/go-ircevent"
	"log"
	"strconv"
)

type Message struct {
	Nick    string
	Channel string
	Body    string
}

type Client struct {
	messages chan Message
}

type ClientManager struct {
	clients []Client
}

const PORT = 3540
const channel = "#pdxbots"
const server = "chat.freenode.net:6667"
const nick = "kodobot"

func main() {
	irccon := irc.IRC(nick, "whatisthis")
	clientManager := ClientManager{}
	irccon.AddCallback("001", func(e *irc.Event) {
		fmt.Println("Connected?")
		irccon.Join(channel)
	})
	err := irccon.Connect(server)
	go irccon.Loop()
	server, err := net.Listen("tcp", ":"+strconv.Itoa(PORT))
	if server == nil {
		fmt.Println(err)
		panic("couldn't start listening: ")
	}
	conns := clientConns(server)
	irccon.AddCallback("PRIVMSG", func(event *irc.Event) {
		fmt.Println(event)
		message := Message{Nick: event.Nick, Channel: event.Arguments[0], Body: event.Arguments[1]}
		for _, client := range clientManager.clients {
			client.messages <- message
		}
	})
	for {
		conn := <-conns
		messages := make(chan Message)
		clientManager.clients = append(clientManager.clients, Client{messages: messages})
		go handleConn(conn, messages)
	}
}

func clientConns(listener net.Listener) chan net.Conn {
	ch := make(chan net.Conn)
	i := 0
	go func() {
		for {
			client, err := listener.Accept()
			if client == nil {
				fmt.Println(err)
				fmt.Printf("couldn't accept: ")
				continue
			}
			i++
			fmt.Printf("%d: %v <-> %v\n", i, client.LocalAddr(), client.RemoteAddr())
			ch <- client
		}
	}()
	return ch
}

func handleConn(client net.Conn, messages <-chan Message) {
	for {
		message := <-messages
		// request := map[string]interface{}{
		//     "method": "addNode",
		//     "params": []Message{message},
		// }

		eventJson, err := json.Marshal(message)
		if err != nil {
			log.Fatal(err)
		}
		client.Write(append(eventJson, '\n'))
	}
}
