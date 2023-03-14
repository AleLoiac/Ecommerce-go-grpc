package main

import (
	"Ecommerce/order/orderpb"
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

	err := DB.Update(func(txn *badger.Txn) error {
		userBytes, err := proto.Marshal(user)
		if err != nil {
			return err
		}
		return txn.Set([]byte("user_"+user.Id), userBytes)
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

	err := DB.View(func(txn *badger.Txn) error {
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

	err := DB.Update(func(txn *badger.Txn) error {
		productData, err := proto.Marshal(product)
		if err != nil {
			return err
		}
		return txn.Set([]byte("product_"+product.Id), productData)
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

	fmt.Printf("GetProduct function is invoked with: %v\n", req)

	var product productpb.Product

	err := DB.View(func(txn *badger.Txn) error {
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

func (s *server) ListProducts(req *productpb.Empty, stream productpb.ProductService_ListProductsServer) error {

	fmt.Println("ListProducts function is invoked with an empty request")

	opts := badger.DefaultIteratorOptions
	opts.PrefetchSize = 10
	opts.PrefetchValues = false
	prefix := []byte("product_")
	opts.Prefix = prefix
	err := DB.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			key := item.Key()
			product, err := s.GetProduct(stream.Context(), &productpb.GetProductRequest{ProductId: string(key)})
			if err != nil {
				return err
			}
			if err = stream.Send(product); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil

}

func (s *server) DeleteProduct(ctx context.Context, req *productpb.DeleteProductRequest) (*productpb.Empty, error) {

	fmt.Printf("DeleteProduct function is invoked with %v\n", req)

	err := DB.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(req.GetProductId()))
	})
	if err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Cannot find product with id: %v", req.GetProductId()))
	}
	return &productpb.Empty{}, nil
}

func (s *server) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.Order, error) {

	fmt.Printf("CreateOrder function is invoked with %v\n", req)

	var items []*orderpb.OrderItem
	var totalPrice float32

	var user userpb.User

	err := DB.View(func(txn *badger.Txn) error {
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

	for _, item := range req.GetItems() {
		items = append(items, &orderpb.OrderItem{
			ProductId: item.GetProductId(),
			Quantity:  item.GetQuantity(),
			Price:     item.GetPrice(),
		})
		totalPrice += float32(item.GetQuantity()) * item.GetPrice()
	}

	order := &orderpb.Order{
		Id:         uuid.New().String(),
		UserId:     req.GetUserId(),
		Items:      items,
		TotalPrice: totalPrice,
	}

	err = DB.Update(func(txn *badger.Txn) error {
		orderBytes, err := proto.Marshal(order)
		if err != nil {
			return err
		}
		return txn.Set([]byte("order_"+order.Id), orderBytes)
	})

	if err != nil {
		return nil, status.Errorf(
			codes.Internal,
			fmt.Sprintf("Failed to create order: %v", err),
		)
	}

	return order, nil

}

type server struct {
	userpb.UserServiceServer
	productpb.ProductServiceServer
	orderpb.OrderServiceServer
}

var DB *badger.DB

func main() {
	fmt.Println("Server started...")

	DB, _ = badger.Open(badger.DefaultOptions("/Users/aless/Desktop/Go/Ecommerce/DB"))

	defer DB.Close()

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	userpb.RegisterUserServiceServer(s, &server{})
	productpb.RegisterProductServiceServer(s, &server{})
	orderpb.RegisterOrderServiceServer(s, &server{})

	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
