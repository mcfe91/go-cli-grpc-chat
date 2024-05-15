package server

import (
	"fmt"
	"log"
	"net"

	"github.com/mcfe91/go-cli-grpc-chat/api"
	"google.golang.org/grpc"
)

type Connection struct {
	stream       api.Broadcast_CreateStreamServer
	id           string
	displayName  string
	chattingWith []string
	active       bool
	error        chan error
}

type Server struct {
	api.UnimplementedBroadcastServer
	Connection map[string]*Connection
}

func StartServer(port string) {
	server := &Server{Connection: make(map[string]*Connection)}

	grpcServer := grpc.NewServer()
	ln, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", port))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("starting gRPC server at %s", ln.Addr().String())

	api.RegisterBroadcastServer(grpcServer, server)
	err = grpcServer.Serve(ln)
	if err != nil {
		log.Fatalf("error starting gRPC server at %s: %v", ln.Addr().String(), err)
	}
}
