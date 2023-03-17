package main

import (
	"Ecommerce/product/productpb"
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
