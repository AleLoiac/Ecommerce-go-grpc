# E-commerce Platform backend simulation with Go, gRPC and Badger
This is an e-commerce platform built with Go and gRPC. It allows users to view and purchase products through a gRPC API.

## Getting Started

To get started, follow these steps:

Clone the repository:

`git clone https://github.com/AleLoiac/Ecommerce-go-grpc.git`

Install the dependencies:

`go mod tidy`

Create a directory to save data and direct Badger to its path, for example:

`db, _ := badger.Open(badger.DefaultOptions("/Users/aless/Desktop/Go/Ecommerce/DB"))`

Run the server:

`go run eco_server.go`

Select and run a client:

`go run "..._client"/main.go`

## Usage

The e-commerce platform supports the following API methods:

* CreateUser
* GetUser
* AddProduct
* GetProduct
* ListProducts
* DeleteProduct
* CreateOrder
* GetOrder

To use the API, start the server and run Evans:
`evans -p 50051 -r`.

Then, chose the desired API package, service and method. For example:

`package product`

`service ProductService`

`call ListProducts`

You can also use Postman to try the APIs, click on "NEW" and then on "gRPC Request".

Remember to use the Example Message feature to check how the requests are structured.