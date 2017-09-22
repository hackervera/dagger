package main

import (
	//  "crypto/rand"
	//  "fmt"
	//  "golang.org/x/crypto/ed25519"
	//  "log"
	//  "bytes"
	//  bencode "github.com/jackpal/bencode-go"
	//  "encoding/json"
	//  "github.com/funkygao/golib/dag"
	"encoding/hex"
	"errors"
	"github.com/deckarep/golang-set"
	//   "flag"
	"reflect"
	// "github.com/davecgh/go-spew/spew"
	"github.com/mitchellh/mapstructure"
)

func (dag *Dag) AddNodes(response Response) {
	var nodes []Node
	mapstructure.Decode(response.Result, &nodes)
	for _, node := range nodes {
		index := SliceIndex(len(dag.Nodes), func(i int) bool { return dag.Nodes[i].Hash == node.Hash })
		if index == -1 {
			dag.Nodes = append(dag.Nodes, node)
		}
	}
}

func (dag *Dag) AddEdge(edge Edge) {
	dag.Edges = append(dag.Edges, edge)
}

func (dag *Dag) CreateRoom(name string) {
	body := make(map[string]string)
	body["name"] = name
	message := map[string]interface{}{
		"author": dag.PubKey,
		"type":   "room",
		"body":   body,
	}
	dag.AddNode(message)
}

func (dag *Dag) CreateMessage(text string) {
	body := make(map[string]string)
	body["text"] = text
	message := map[string]interface{}{
		"author": dag.PubKey,
		"type":   "message",
		"body":   body,
	}
	dag.AddNode(message)
}

func (dag *Dag) RootNode() (node *Node) {
	for i := range dag.Edges {
		edge := dag.Edges[i]
		data := edge.From.Data
		// fmt.Println(data["Type"])
		if data["Type"] == "room" {
			node = edge.From
		}
	}
	return node
}

func (dag *Dag) FindEdge(hash string) (edge Edge, err error) {
	for i := range dag.Edges {
		if dag.Edges[i].From.Hash == hash {
			edge = dag.Edges[i]
		}
	}
	// fmt.Println(edge)
	if edge == (Edge{}) {
		err = errors.New("No Edge found")
	}
	// fmt.Println(edge.From.Hash)
	return edge, err
}

func (dag *Dag) AddNode(obj map[string]interface{}) Node {
	objSig := canonicalSign(obj, dag.PrivKey)
	node := Node{Hash: hex.EncodeToString(objSig), Data: obj, Parents: dag.Leaves()}
	dag.Nodes = append(dag.Nodes, node)
	dag.Attach(&node)
	return node
}

func (dag *Dag) AddRawNode(node Node) {
	dag.Nodes = append(dag.Nodes, node)
	dag.Attach(&node)
}

func (dag *Dag) Attach(node *Node) {
	leafNodes := dag.LeafNodes()
	for _, leafNode := range leafNodes {
		edge := Edge{From: &leafNode, To: node}
		if leafNode.Hash != node.Hash {
			dag.AddEdge(edge)
		}
	}
}

func (dag *Dag) LeafNodes() (nodes []Node) {
	leafHashes := dag.Leaves()
	for _, hash := range leafHashes {
		var foundNode Node
		for _, node := range dag.Nodes {
			if node.Hash == hash {
				foundNode = node
			}
		}
		nodes = append(nodes, foundNode)
	}
	return nodes
}

func (dag *Dag) GetNodes() (nodes []Node) {
	for _, node := range dag.Nodes {
		nodes = append(nodes, node)
	}
	if nodes == nil {
		nodes = []Node{}
	}
	return nodes
}

func (dag *Dag) GetNode(hash string) (node Node, err error) {
	for _, nodeIterator := range dag.Nodes {
		if nodeIterator.Hash == hash {
			node = nodeIterator
		}
	}
	if reflect.DeepEqual(node, Node{}) {
		err = errors.New("Node not found")
	}
	return node, err
}

func (dag *Dag) Leaves() (leaves []string) {
	toNodeHashes := mapset.NewSet()
	fromNodeHashes := mapset.NewSet()
	for _, edge := range dag.Edges {
		toNodeHashes.Add(edge.To.Hash)
		fromNodeHashes.Add(edge.From.Hash)
	}

	slice := toNodeHashes.Difference(fromNodeHashes).ToSlice()
	for _, e := range slice {
		leaves = append(leaves, e.(string))
	}

	if dag.Edges == nil && dag.Nodes != nil {
		leaves = []string{dag.Nodes[0].Hash}
	}

	if dag.Edges == nil && dag.Nodes == nil {
		leaves = []string{}
	}
	return leaves
}
