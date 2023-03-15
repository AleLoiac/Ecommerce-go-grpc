package main

import (
	"Ecommerce/product/productpb"
	"bufio"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"log"
	"strconv"
	"strings"
)

func productCreate(c productpb.ProductServiceClient, reader *bufio.Reader) {

	fmt.Println("Starting Unary RPC...")

	var name, description, price string

	fmt.Println("Create new product, name:")
	name, _ = reader.ReadString('\n')
	name = strings.TrimSpace(name)

	fmt.Println("Create new product, description:")
	description, _ = reader.ReadString('\n')
	description = strings.TrimSpace(description)

	fmt.Println("Create new product, price:")
	price, _ = reader.ReadString('\n')
	price = strings.TrimSpace(price)

	fPrice, err := strconv.ParseFloat(price, 32)

	req := &productpb.CreateProductRequest{
		Name:        name,
		Description: description,
		Price:       float32(fPrice),
	}

	res, err := c.AddProduct(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling AddProduct RPC: %v", err)
	}
	log.Printf("Response from AddProduct: %v", res.GetId())
}

func productGet(c productpb.ProductServiceClient, reader *bufio.Reader) {

	fmt.Println("Starting Unary RPC...")

	var id string

	fmt.Println("Get a product, id:")
	id, _ = reader.ReadString('\n')
	id = strings.TrimSpace(id)

	req := &productpb.GetProductRequest{ProductId: id}

	res, err := c.GetProduct(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling GetProduct RPC: %v", err)
	}
	log.Printf("Name: %v", res.Name)
	log.Printf("Description: %v", res.Description)
	log.Printf("Price: %v", res.Price)
}

func productList(c productpb.ProductServiceClient) {

	fmt.Println("Starting Server Streaming RPC...")

	req := &productpb.Empty{}

	resStream, err := c.ListProducts(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling ListProducts RPC: %v", err)
	}
	for {
		msg, err := resStream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error while reading the stream: %v", err)
		}
		log.Printf("Listing products: %v", msg)
	}
}

func productDelete(c productpb.ProductServiceClient, reader *bufio.Reader) {

	fmt.Println("Starting Unary RPC...")

	var id string

	fmt.Println("Delete a product, id:")
	id, _ = reader.ReadString('\n')
	id = strings.TrimSpace(id)

	req := &productpb.DeleteProductRequest{ProductId: id}

	_, err := c.DeleteProduct(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling DeleteProduct RPC: %v", err)
	}
	fmt.Printf("Product with id %v successfully deleted\n", id)
}

func main() {

	cc, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}
	defer cc.Close()

	//reader := bufio.NewReader(os.Stdin)
	//c := productpb.NewProductServiceClient(cc)
}
