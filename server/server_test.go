package main

import (
	"Ecommerce/product/productpb"
	"context"
	"github.com/dgraph-io/badger/v3"
	"google.golang.org/protobuf/proto"
	"testing"
)

func TestAddProduct(t *testing.T) {
	// Mock database
	db, err := badger.Open(badger.DefaultOptions("").WithInMemory(true))
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	// Server instance
	server := &Server{db: db}

	req := &productpb.CreateProductRequest{
		Name:        "Test Product",
		Description: "Sample product for testing",
		Price:       9.99,
	}

	got, err := server.AddProduct(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to create product: %v", err)
	}

	if got.Name != req.Name {
		t.Errorf("Expected name %q, got %q", req.Name, got.Name)
	}
	if got.Description != req.Description {
		t.Errorf("Expected description %q, got %q", req.Description, got.Description)
	}
	if got.Price != req.Price {
		t.Errorf("Expected price %v, got %v", req.Price, got.Price)
	}
	if got.Id == "" {
		t.Error("Expected non-empty ID")
	}

	//if got.GetName() != req.GetName() || got.GetDescription() != req.GetDescription() || got.GetPrice() != req.GetPrice() {
	//	t.Errorf("AddProduct() = %v, want %v", got, req)
	//}
}

func TestGetProduct(t *testing.T) {

	db, err := badger.Open(badger.DefaultOptions("").WithInMemory(true))
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}
	defer db.Close()

	server := &Server{db: db}

	product := &productpb.Product{
		Id:          "test-id",
		Name:        "Test Product",
		Description: "A sample product for testing",
		Price:       9.99,
	}

	productData, err := proto.Marshal(product)
	if err != nil {
		t.Fatalf("Failed to marshal product: %v", err)
	}
	err = db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte("product_"+product.Id), productData)
	})
	if err != nil {
		t.Fatalf("Failed to store product: %v", err)
	}

	req := &productpb.GetProductRequest{ProductId: "product_" + product.Id}
	res, err := server.GetProduct(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to get product: %v", err)
	}

	if res.Name != product.Name {
		t.Errorf("Expected name %q, got %q", product.Name, res.Name)
	}
	if res.Description != product.Description {
		t.Errorf("Expected description %q, got %q", product.Description, res.Description)
	}
	if res.Price != product.Price {
		t.Errorf("Expected price %v, got %v", product.Price, res.Price)
	}
}
