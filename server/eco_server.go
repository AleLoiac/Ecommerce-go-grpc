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

func (s *Server) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.CreateUserResponse, error) {

	fmt.Printf("CreateUser function is invoked with: %v\n", req)

	user := &userpb.User{
		Id:       uuid.New().String(),
		Username: req.GetUsername(),
		Email:    req.GetEmail(),
		Password: req.GetPassword(),
	}

	err := s.db.Update(func(txn *badger.Txn) error {
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

func (s *Server) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {

	fmt.Printf("GetUser function is invoked with: %v\n", req)

	var user userpb.User

	err := s.db.View(func(txn *badger.Txn) error {
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

func (s *Server) AddProduct(ctx context.Context, req *productpb.CreateProductRequest) (*productpb.Product, error) {

	fmt.Printf("AddProduct function is invoked with: %v\n", req)

	product := &productpb.Product{
		Id:          uuid.New().String(),
		Name:        req.GetName(),
		Description: req.GetDescription(),
		Price:       req.GetPrice(),
	}

	err := s.db.Update(func(txn *badger.Txn) error {
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

func (s *Server) GetProduct(ctx context.Context, req *productpb.GetProductRequest) (*productpb.Product, error) {

	fmt.Printf("GetProduct function is invoked with: %v\n", req)

	var product productpb.Product

	err := s.db.View(func(txn *badger.Txn) error {
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

func (s *Server) ListProducts(req *productpb.Empty, stream productpb.ProductService_ListProductsServer) error {

	fmt.Println("ListProducts function is invoked with an empty request")

	opts := badger.DefaultIteratorOptions
	opts.PrefetchSize = 10
	opts.PrefetchValues = false
	prefix := []byte("product_")
	opts.Prefix = prefix
	err := s.db.View(func(txn *badger.Txn) error {
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

func (s *Server) DeleteProduct(ctx context.Context, req *productpb.DeleteProductRequest) (*productpb.Empty, error) {

	fmt.Printf("DeleteProduct function is invoked with %v\n", req)

	err := s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(req.GetProductId()))
	})
	if err != nil {
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Cannot find product with id: %v", req.GetProductId()))
	}
	return &productpb.Empty{}, nil
}

func (s *Server) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.Order, error) {

	fmt.Printf("CreateOrder function is invoked with %v\n", req)

	var items []*orderpb.OrderItem
	var totalPrice float32

	var user userpb.User

	err := s.db.View(func(txn *badger.Txn) error {
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

	err = s.db.Update(func(txn *badger.Txn) error {
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

func (s *Server) GetOrder(ctx context.Context, req *orderpb.GetOrderRequest) (*orderpb.Order, error) {

	fmt.Printf("GetOrder function is invoked with: %v\n", req)

	var order orderpb.Order

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(req.GetOrderId()))
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			err = proto.Unmarshal(val, &order)
			if err != nil {
				fmt.Printf("Error unmarshaling order data: %v\n", err)
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
				fmt.Sprintf("order with ID '%s' not found", req.OrderId),
			)
		} else {
			return nil, status.Errorf(
				codes.Internal,
				fmt.Sprintf("Failed to get order: %v", err),
			)
		}
	}

	return &orderpb.Order{
		Id:         order.GetId(),
		UserId:     order.GetUserId(),
		Items:      order.GetItems(),
		TotalPrice: order.GetTotalPrice(),
	}, nil
}

type Server struct {
	userpb.UserServiceServer
	productpb.ProductServiceServer
	orderpb.OrderServiceServer
	db *badger.DB
}

func NewServer(db *badger.DB) *Server {
	return &Server{db: db}
}

//var DB *badger.DB

func main() {
	fmt.Println("Server started...")

	db, _ := badger.Open(badger.DefaultOptions("/Users/aless/Desktop/Go/Ecommerce/DB"))
	defer db.Close()

	server := NewServer(db)

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	userpb.RegisterUserServiceServer(s, server)
	productpb.RegisterProductServiceServer(s, server)
	orderpb.RegisterOrderServiceServer(s, server)

	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
