package main

import (
	pb "grpcDemo/productInfo/service/proto"
	"context"
	"github.com/gofrs/uuid"
	"errors"
	"log"
	"net"
	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

type server struct {
	productMap map[string]*pb.Product
}

func (s *server) AddProduct(ctx context.Context, in *pb.Product) (*pb.ProductID, error) {
	out, err := uuid.NewV4()
	if err != nil {
		return nil, errors.New("generate productID err")
	}
	in.Id = out.String()
	if s.productMap == nil {
		s.productMap = make(map[string]*pb.Product)
	}
	s.productMap[in.Id] = in
	log.Printf(" Add Product %v : %v", in.Id, in.Name)
	return &pb.ProductID{Value: in.Id}, nil
}

func (s *server) GetProduct(ctx context.Context, in *pb.ProductID) (*pb.Product, error) {

	if product, ok := s.productMap[in.Value]; ok && product != nil {
		log.Printf(" Get Product %v : %v", product.Id, product.Name)
		return product, nil
	}
	return nil, errors.New("don't find product by ID")
}

func main() {
	listen, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("Server start listening....")
	s := grpc.NewServer()
	pb.RegisterProductInfoServer(s, &server{})
	if err := s.Serve(listen); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}


