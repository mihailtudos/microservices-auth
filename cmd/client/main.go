// package main is the starting point when runing a service client
package main

import (
	"context"
	"log"
	"time"

	desc "github.com/mihailtudos/microservices/auth/pkg/auth_v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	address = "localhost:50051"
	userID  = 12
)

func main() {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("failed to connect to server: %s", err)
	}

	defer func() {
		err := conn.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	c := desc.NewAuthV1Client(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	r, err := c.Get(ctx, &desc.GetRequest{Id: userID})
	if err != nil {
		log.Fatal("failed to get user by id: ", err)
	}

	log.Printf("user info: %+v", r.GetUser())
}
