package main

import (
	"Ecommerce/product/productpb"
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

	err := usersDB.Update(func(txn *badger.Txn) error {
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

	err := usersDB.View(func(txn *badger.Txn) error {
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

func (s *server) AddProduct(ctx context.Context, req *productpb.CreateProductRequest) (*productpb.Product, error) {

	fmt.Printf("AddProduct function is invoked with: %v\n", req)

	product := &productpb.Product{
		Id:          uuid.New().String(),
		Name:        req.GetName(),
		Description: req.GetDescription(),
		Price:       req.GetPrice(),
	}

	err := productsDB.Update(func(txn *badger.Txn) error {
		productData, err := proto.Marshal(product)
		if err != nil {
			return err
		}
		return txn.Set([]byte(product.Id), productData)
	})

	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Failed to create product: %v", err),
		)
	}

	return &productpb.Product{
		Id:          product.GetId(),
		Name:        product.GetName(),
		Description: product.GetDescription(),
		Price:       product.GetPrice(),
	}, nil
}

func (s *server) GetProduct(ctx context.Context, req *productpb.GetProductRequest) (*productpb.Product, error) {

	fmt.Printf("GetUser function is invoked with: %v\n", req)

	var product productpb.Product

	err := productsDB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(req.GetProductId()))
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			err = proto.Unmarshal(val, &product)
			if err != nil {
				fmt.Printf("Error unmarshaling product data: %v\n", err)
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
				fmt.Sprintf("Product with ID '%s' not found", req.ProductId),
			)
		} else {
			return nil, status.Errorf(
				codes.Internal,
				fmt.Sprintf("Failed to get product: %v", err),
			)
		}
	}

	return &productpb.Product{
		Id:          product.GetId(),
		Name:        product.GetName(),
		Description: product.GetDescription(),
		Price:       product.GetPrice(),
	}, nil
}

type server struct {
	userpb.UserServiceServer
	productpb.ProductServiceServer
}

var usersDB *badger.DB
var productsDB *badger.DB

func main() {
	fmt.Println("Server started...")

	usersDB, _ = badger.Open(badger.DefaultOptions("/Users/aless/Desktop/Go/Ecommerce/db/Users_DB"))
	productsDB, _ = badger.Open(badger.DefaultOptions("/Users/aless/Desktop/Go/Ecommerce/db/Products_DB"))

	defer usersDB.Close()
	defer productsDB.Close()

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	userpb.RegisterUserServiceServer(s, &server{})
	productpb.RegisterProductServiceServer(s, &server{})

	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
