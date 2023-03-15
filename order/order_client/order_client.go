package main

import (
	"Ecommerce/order/orderpb"
	"bufio"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"strings"
)

func orderCreate(c orderpb.OrderServiceClient, reader *bufio.Reader) {

	fmt.Println("Starting Unary RPC...")

	var userId string

	items := []*orderpb.OrderItem{
		{
			ProductId: "1",
			Quantity:  2,
			Price:     3.2,
		},
		{
			ProductId: "2",
			Quantity:  1,
			Price:     2.1,
		},
	}

	fmt.Println("Create new order, user id:")
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

func orderGet(c orderpb.OrderServiceClient, reader *bufio.Reader) {

	fmt.Println("Starting Unary RPC...")

	var id string

	fmt.Println("Get an order, id:")
	id, _ = reader.ReadString('\n')
	id = strings.TrimSpace(id)

	req := &orderpb.GetOrderRequest{OrderId: id}

	res, err := c.GetOrder(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling GetOrder RPC: %v", err)
	}
	log.Printf("User: %v", res.UserId)
	log.Printf("Items: %v", res.Items)
	log.Printf("Price: %v", res.TotalPrice)
}

func main() {

	cc, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Could not connect: %v", err)
	}
	defer cc.Close()

	//reader := bufio.NewReader(os.Stdin)
	//c := orderpb.NewOrderServiceClient(cc)
}
