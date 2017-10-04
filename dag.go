package dagger

import (
	"dagger/rpc"
	"errors"

	"reflect"

	"github.com/deckarep/golang-set"
	"golang.org/x/crypto/ed25519"
)

// Dag contains Nodes and Edges.
type Dag struct {
	Edges   []Edge
	Nodes   []rpc.Node
	PrivKey ed25519.PrivateKey
	PubKey  string
}

func (dag *Dag) AddEdge(edge Edge) {
	dag.Edges = append(dag.Edges, edge)
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

func (dag *Dag) Attach(node *rpc.Node) {
	leafNodes := dag.LeafNodes()
	for _, leafNode := range leafNodes {
		edge := Edge{From: &leafNode, To: node}
		if leafNode.Hash != node.Hash {
			dag.AddEdge(edge)
		}
	}
}

func (dag *Dag) LeafNodes() (nodes []rpc.Node) {
	leafHashes := dag.Leaves()
	for _, hash := range leafHashes {
		var foundNode rpc.Node
		for _, node := range dag.Nodes {
			if node.Hash == hash {
				foundNode = node
			}
		}
		nodes = append(nodes, foundNode)
	}
	return nodes
}

func (dag *Dag) GetNodes() (nodes []rpc.Node) {
	for _, node := range dag.Nodes {
		nodes = append(nodes, node)
	}
	if nodes == nil {
		nodes = []rpc.Node{}
	}
	return nodes
}

func (dag *Dag) GetNode(hash string) (node rpc.Node, err error) {
	for _, nodeIterator := range dag.Nodes {
		if nodeIterator.Hash == hash {
			node = nodeIterator
		}
	}
	if reflect.DeepEqual(node, rpc.Node{}) {
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
