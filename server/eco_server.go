package main

import (
	"Ecommerce/user/userpb"
	"context"
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"log"
	"net"
)

func (s *server) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.CreateUserResponse, error) {

	fmt.Printf("CreateUser function is invoked with %v\n", req)

	user := &userpb.User{
		Id:       uuid.New().String(),
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	}

	err := db.Update(func(txn *badger.Txn) error {
		userBytes, err := json.Marshal(user)
		if err != nil {
			return err
		}

		return txn.Set([]byte(user.Id), userBytes)
	})

	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Failed to create user: %v", err),
		)
	}

	return &userpb.CreateUserResponse{
		UserId: user.Id,
	}, nil
}

type server struct {
	userpb.UserServiceServer
}

var db *badger.DB

func main() {
	fmt.Println("Server started...")

	db, _ = badger.Open(badger.DefaultOptions("/Users/aless/Desktop/Go/Ecommerce/db"))

	defer db.Close()

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	userpb.RegisterUserServiceServer(s, &server{})

	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
