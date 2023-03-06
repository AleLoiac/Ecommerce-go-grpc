package main

import (
	"Ecommerce/user/userpb"
	"bufio"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"strings"
)

func unaryCreate(c userpb.UserServiceClient, reader *bufio.Reader) {

	fmt.Println("Starting Unary RPC...")

	var username, email, password string

	fmt.Println("Create new user, username:")
	username, _ = reader.ReadString('\n')
	username = strings.TrimSpace(username)

	fmt.Println("Create new user, email:")
	email, _ = reader.ReadString('\n')
	email = strings.TrimSpace(email)

	fmt.Println("Create new user, password:")
	password, _ = reader.ReadString('\n')
	password = strings.TrimSpace(password)

	req := &userpb.CreateUserRequest{
		Username: username,
		Email:    email,
		Password: password,
	}

	res, err := c.CreateUser(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling CreateUser RPC: %v", err)
	}
	log.Printf("Response from CreateUser: %v", res.GetUserId())
}

func main() {

	cc, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials())) //needs to be secured
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}
	defer cc.Close()

	reader := bufio.NewReader(os.Stdin)

	c := userpb.NewUserServiceClient(cc)

	unaryCreate(c, reader)
}
