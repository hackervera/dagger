package dagger

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net"

	"github.com/tjgillies/dagger/rpc"

	bencode "github.com/jackpal/bencode-go"
	"golang.org/x/crypto/ed25519"
	"google.golang.org/grpc"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

// Edge represents a connection between two Nodes.
type Edge struct {
	From *rpc.Node
	To   *rpc.Node
}

// Message represents a message being sent over json-rpc.
type Message struct {
	Author string
	Type   string
	Body   map[string]interface{}
}

func canonicalSign(data interface{}, priv ed25519.PrivateKey) []byte {
	var buf bytes.Buffer
	bencode.Marshal(&buf, data)
	return ed25519.Sign(priv, buf.Bytes())
}

type rpcServer struct {
	Dag *Dag
}

func (src *rpcServer) GetNode(ctx context.Context, node *rpc.Node) (*rpc.Node, error) {
	return &src.Dag.Nodes[0], nil
}

func StartServer() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 1234))
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()
	dag := &Dag{Nodes: []rpc.Node{rpc.Node{Hash: "1234", Data: "Sup"}}}
	rpc.RegisterRpcServer(grpcServer, &rpcServer{Dag: dag})
	grpcServer.Serve(lis)
	return nil
}

func Client() (rpc.RpcClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial("127.0.0.1:1234", grpc.WithInsecure())
	if err != nil {
		return nil, nil, err
	}
	client := rpc.NewRpcClient(conn)

	return client, conn, nil
}
