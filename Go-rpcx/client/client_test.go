package client

import (
	"Go-rpcx/server"
	"context"
	"fmt"
	"testing"
	"time"
)

type Args struct {
	A int
	B int
}
type Reply struct {
	C int
}

type Arith int

func (t *Arith) Mul(ctx context.Context, args *Args, reply *Reply) error {
	reply.C = args.A * args.B
	return nil
}

func TestClient_IT(t *testing.T) {
	server.UsePool = false

	s := server.NewServer()
	_ = s.RegisterName("Arith", new(Arith), "")
	go func() {
		_ = s.Serve("tcp", "127.0.0.1:8080")
	}()
	defer s.Close()
	time.Sleep(500 * time.Millisecond)

	addr := s.Address().String()
	fmt.Printf(addr)

	client := &Client{
		option: DefaultOption,
	}

	err := client.Connect("tcp", addr)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer client.Close()

	args := &Args{
		A: 10,
		B: 20,
	}

	reply := &Reply{}
	err = client.Call(context.Background(), "Arith", "Mul", args, reply)
	if err != nil {
		t.Fatalf("failed to call: %v", err)
	}
	if reply.C != 200 {
		t.Fatalf("expect 200 but got %d", reply.C)
	}

}
