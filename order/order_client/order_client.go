package main

import (
	"Ecommerce/order/orderpb"
	"bufio"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"strings"
)

func orderCreate(c orderpb.OrderServiceClient, reader *bufio.Reader) {

	fmt.Println("Starting Unary RPC...")

	var userId string

	items := []*orderpb.OrderItem{
		&orderpb.OrderItem{
			ProductId: "1",
			Quantity:  2,
			Price:     3.2,
		},
		&orderpb.OrderItem{
			ProductId: "2",
			Quantity:  1,
			Price:     2.1,
		},
	}

	fmt.Println("Create new product, name:")
	userId, _ = reader.ReadString('\n')
	userId = strings.TrimSpace(userId)

	req := &orderpb.CreateOrderRequest{
		UserId: userId,
		Items:  items,
	}

	res, err := c.CreateOrder(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling CreateOrder RPC: %v", err)
	}
	log.Printf("Response from CreateOrder: %v", res)

}

func main() {

	cc, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}
	defer cc.Close()

	reader := bufio.NewReader(os.Stdin)
	c := orderpb.NewOrderServiceClient(cc)

	orderCreate(c, reader)
}
