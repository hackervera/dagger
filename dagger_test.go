package dagger

import (
	"context"
	"fmt"
	"log"

	"github.com/tjgillies/dagger/rpc"
)

func Example() {
	go StartServer()
	//time.Sleep(time.Second)
	client, _, err := Client()
	if err != nil {
		log.Fatal(err)
	}
	node := rpc.Node{Hash: "1234"}
	fmt.Println(client.GetNode(context.Background(), &node))
	// Output: Hash:"1234" Data:"Sup"  <nil>
}
