package client

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mcfe91/go-cli-grpc-chat/api"
	"google.golang.org/grpc"
)

var client api.BroadcastClient
var wait *sync.WaitGroup

func init() {
	wait = &sync.WaitGroup{}
}

func connect(user *api.User, receivers []string) error {
	var streamerror error

	stream, err := client.CreateStream(context.Background(), &api.Connect{
		User:         user,
		Active:       true,
		ChattingWith: receivers,
	})
	if err != nil {
		return fmt.Errorf("connection failed: %v", err)
	}

	wait.Add(1)
	go func(str api.Broadcast_CreateStreamClient) {
		defer wait.Done()
		for {
			msg, err := str.Recv()
			if err != nil {
				streamerror = fmt.Errorf("error reading message %v", err)
				break
			}
			log.Printf("%s : %s\n", msg.User.DisplayName, msg.Message)
		}
	}(stream)

	return streamerror
}

func StartClient(name, receivers, remoteServerHost string) {
	log.Printf("starting client for %s", name)

	timestamp := time.Now()

	done := make(chan int)

	conn, err := grpc.NewClient(fmt.Sprintf("%s", remoteServerHost), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("error connecting to host %s: %v", remoteServerHost, err)
	}

	client = api.NewBroadcastClient(conn)

	id := sha256.Sum256([]byte(timestamp.String() + name))
	user := &api.User{
		Id:          hex.EncodeToString(id[:]),
		DisplayName: name,
	}

	err = connect(user, strings.Split(receivers, ","))
	if err != nil {
		log.Fatalf("error while creating user stream %v", err)
	}

	wait.Add(1)
	go func() {
		defer wait.Done()

		scanner := bufio.NewScanner(os.Stdin)

		for scanner.Scan() {
			msg := &api.Message{
				Id:        user.Id,
				User:      user,
				Message:   scanner.Text(),
				Timestamp: timestamp.String(),
			}

			_, err := client.BroadcastMessage(context.Background(), msg)
			if err != nil {
				log.Printf("error sending message: %v", err)
				break
			}
		}
	}()

	go func() {
		wait.Wait()
		close(done)
	}()

	log.Printf("started client for %s", name)

	<-done
}
