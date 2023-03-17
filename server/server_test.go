package main

import (
	"Ecommerce/product/productpb"
	"context"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"log"
	"testing"
)

func TestAddProduct(t *testing.T) {

	cc, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}
	defer cc.Close()

	c := productpb.NewProductServiceClient(cc)

	req := &productpb.CreateProductRequest{
		Name:        "Test Product",
		Description: "Test Product Description",
		Price:       99.99,
	}

	got, err := c.AddProduct(context.Background(), req)

	if err != nil {
		t.Fatalf("AddProduct() error = %v, wantErr nil", err)
	}

	if got.GetName() != req.GetName() || got.GetDescription() != req.GetDescription() || got.GetPrice() != req.GetPrice() {
		t.Errorf("AddProduct() = %v, want %v", got, req)
	}

}

// not working
func TestGetProduct(t *testing.T) {

	cc, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}
	defer cc.Close()

	product := &productpb.Product{
		Id:          "Product Id",
		Name:        "Product Name",
		Description: "Product Description",
		Price:       99.99,
	}

	err = DB.Update(func(txn *badger.Txn) error {
		productData, err := proto.Marshal(product)
		if err != nil {
			return err
		}
		return txn.Set([]byte(product.GetId()), productData)
	})

	if err != nil {
		err = status.Errorf(
			codes.Internal,
			fmt.Sprintf("Failed to create test product: %v", err),
		)
		if err != nil {
			return
		}
		return
	}

	c := productpb.NewProductServiceClient(cc)

	req := &productpb.GetProductRequest{ProductId: "Product Id"}

	got, err := c.GetProduct(context.Background(), req)
	if err != nil {
		t.Fatalf("GetProduct() error = %v, wantErr nil", err)
	}

	if got.Id != "Product Id" {
		t.Errorf("GetProduct returned incorrect product ID: got %v, want %v", got.Id, "Product Id")
	}
	if got.Name != "Product Name" {
		t.Errorf("GetProduct returned incorrect product name: got %v, want %v", got.Name, "Product Name")
	}
	if got.Description != "Product Description" {
		t.Errorf("GetProduct returned incorrect product description: got %v, want %v", got.Description, "Product Description")
	}
	if got.Price != 99.99 {
		t.Errorf("GetProduct returned incorrect product price: got %v, want %v", got.Price, 99.99)
	}

	err = DB.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(req.GetProductId()))
	})
	if err != nil {
		err := status.Errorf(codes.NotFound, fmt.Sprintf("Cannot find product with id: %v", req.GetProductId()))
		if err != nil {
			return
		}
		return
	}

}
