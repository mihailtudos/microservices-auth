// package main is the starting point when runing a service client
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	desc "github.com/mihailtudos/microservices/auth/pkg/auth_v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	address = "localhost:50051"
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

	key := make([]byte, 16)
	rand.Read(key)
	randonPassword := hex.EncodeToString(key)

	u, err := c.Create(ctx, &desc.CreateRequest{
		Info: &desc.UserRegistration{
			Name:            "Mihail Tudos",
			Email:           "mihail@example.com",
			Password:        randonPassword,
			PasswordConfirm: randonPassword,
			Role:            desc.Role_ADMIN,
		},
	})

	if err != nil {
		log.Fatal("failed to create new user: ", err)
	}

	fmt.Printf("user created success: %s\n", u.Id)

	r, err := c.Get(ctx, &desc.GetRequest{Id: u.Id})
	if err != nil {
		log.Fatal("failed to get user by id: ", err)
	}

	log.Printf("user info: %+v", r.GetUser())
}
