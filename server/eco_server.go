package main

import (
	"Ecommerce/user/userpb"
	"context"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"log"
	"net"
)

func (s *server) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.CreateUserResponse, error) {

	fmt.Printf("CreateUser function is invoked with: %v\n", req)

	user := &userpb.User{
		Id:       uuid.New().String(),
		Username: req.GetUsername(),
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}

	err := db.Update(func(txn *badger.Txn) error {
		userBytes, err := proto.Marshal(user)
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
		UserId: user.GetId(),
	}, nil
}

func (s *server) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {

	fmt.Printf("GetUser function is invoked with: %v\n", req)

	var user userpb.User

	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(req.GetUserId()))
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			err = proto.Unmarshal(val, &user)
			if err != nil {
				fmt.Printf("Error unmarshaling user data: %v\n", err)
				return err
			}
			return nil
		})
		return err
	})

	if err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, status.Errorf(
				codes.NotFound,
				fmt.Sprintf("User with ID '%s' not found", req.UserId),
			)
		} else {
			return nil, status.Errorf(
				codes.Internal,
				fmt.Sprintf("Failed to get user: %v", err),
			)
		}
	}

	return &userpb.GetUserResponse{
		Username: user.Username,
		Email:    user.Email,
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
