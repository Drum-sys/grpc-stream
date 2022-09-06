package main

import (
	"google.golang.org/grpc"
	"log"
	pb "grpcDemo/productInfo/client/proto"
	"context"
	"time"
)

const (
	address = "localhost:50051"
)

func main()  {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewProductInfoClient(conn)

	name := "Apple iPhone 14"
	description := `Meet Apple iPhone 11. All-new dual-camera system with Ultra Wide and Night mode.`


	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second)
	defer cancelFunc()
	productID, err := client.AddProduct(ctx, &pb.Product{
		Name:        name,
		Description: description,
	})
	if err != nil {
		log.Fatalf("Could not add product:%v", err)
	}
	log.Printf("Product ID: %s added successfully", productID.Value)
	product, err := client.GetProduct(ctx, &pb.ProductID{Value: productID.Value})
	if err != nil {
		log.Fatalf("Could not get product:%v", err)
	}
	log.Printf("Product:%v", product.String())
}