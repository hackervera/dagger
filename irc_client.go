package main

import (
	"net"
	"os"
	// "bytes"
	// "strings"
	"bufio"
	"log"
)

func IrcConn(ch chan<- []byte) {
	servAddr := "localhost:3540"
	tcpAddr, err := net.ResolveTCPAddr("tcp", servAddr)
	if err != nil {
		println("ResolveTCPAddr failed:", err.Error())
		os.Exit(1)
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		println("Dial failed:", err.Error())
		os.Exit(1)
	}
	for {
		bufReader := bufio.NewReader(conn)
		reply, err := bufReader.ReadBytes('\n')
		if err != nil {
			log.Fatal(err)
		}
		ch <- reply
	}
}
