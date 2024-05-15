package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"

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

func (s *Server) CreateStream(pconn *api.Connect, stream api.Broadcast_CreateStreamServer) error {
	log.Printf("handling connection request for %s", pconn.User.GetDisplayName())

	conn := &Connection{
		stream:       stream,
		id:           pconn.User.GetId(),
		displayName:  pconn.User.GetDisplayName(),
		chattingWith: pconn.GetChattingWith(),
		active:       true,
		error:        make(chan error),
	}

	if _, ok := s.Connection[conn.displayName]; ok {
		return errors.New("try using a different user name")
	} else {
		s.Connection[conn.displayName] = conn
	}

	return <-conn.error
}

func canChatWith(senderChattingWith, currentUserChattingWith []string, senderUserDisplayName, currentUserDisplayName string) bool {
	if senderUserDisplayName == currentUserDisplayName {
		return true
	}

	if len(senderChattingWith) == 0 {
		return false
	}

	if len(senderChattingWith) == 1 && strings.EqualFold(senderChattingWith[0], "all") {
		log.Printf("sender %s, current user %s, current user chatting with %v", senderUserDisplayName, currentUserDisplayName, currentUserChattingWith)
		if len(currentUserChattingWith) == 0 {
			return false
		} else if len(currentUserChattingWith) == 1 && strings.EqualFold(currentUserChattingWith[0], "all") {
			return true
		}

		for _, username := range currentUserChattingWith {
			if senderUserDisplayName == username {
				return true
			}
		}

		return false
	}

	for _, username := range senderChattingWith {
		if currentUserDisplayName == username {
			return true
		}
	}

	return false
}

func (s *Server) BroadcastMessage(ctx context.Context, msg *api.Message) (*api.Close, error) {
	wait := sync.WaitGroup{}
	done := make(chan int)

	for sendingTo, conn := range s.Connection {
		wait.Add(1)

		go func(sendingTo string, msg *api.Message, conn *Connection) {
			defer wait.Done()

			senderConn := s.Connection[msg.User.GetDisplayName()]

			if conn.active && canChatWith(senderConn.chattingWith, conn.chattingWith, senderConn.displayName, conn.displayName) {
				fmt.Printf("sending message to %s: %v", sendingTo, conn.stream)
				err := conn.stream.Send(msg)
				if err != nil {
					fmt.Printf("error with stream: %v - error: %v; try re-connecting...", conn.stream, err)
					conn.active = false
					conn.error <- err
				}
			}
		}(sendingTo, msg, conn)
	}

	go func() {
		wait.Wait()
		close(done)
	}()

	<-done
	return &api.Close{}, nil
}
