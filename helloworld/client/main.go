package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"time"

	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

const (
	defaultName = "world"
	n           = 1000
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
	name = flag.String("name", defaultName, "Name to greet")
)

func main() {
	flag.Parse()
	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// Simple RPC
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	for i := 0; i < n; i++ {
		r, err := c.SimpleRPC(ctx, &pb.HelloRequest{Name: *name})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		log.Printf("Greeting: %s", r.GetMessage())
	}

	elapsed := time.Since(start)

	// Bidictional streaming RPC
	start1 := time.Now()
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	stream, err := c.Bidirectional_StreamingRPC(ctx)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			if err := stream.Send(&pb.HelloRequest{Name: defaultName}); err != nil {
				log.Fatal(err)
			}
			time.Sleep(time.Nanosecond)
		}
	}()

	for i := 0; i < n; i++ {
		reply, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}
		fmt.Println("[BidirectionalRPC] Greeting ", i, ":", reply.GetMessage())
	}

	elapsed1 := time.Since(start1)

	// Server-side streaming RPC
	start2 := time.Now()
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	serverStream, err := c.Ser_StreamingRPC(ctx, &pb.HelloRequest{Name: defaultName})
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < n; i++ {
		reply, err := serverStream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}
		fmt.Println("[ServerStreamingRPC] Greeting ", i, ":", reply.GetMessage())
	}

	elapsed2 := time.Since(start2)

	// Client-side streaming RPC
	start3 := time.Now()
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	clientStream, err := c.Cli_StreamingRPC(ctx)
	if err != nil {
		log.Fatal(err)
	}

	if err := clientStream.Send(&pb.HelloRequest{Name: defaultName}); err != nil {
		log.Fatal(err)
	}

	reply, err := clientStream.CloseAndRecv()
	if err != nil {
		log.Fatalf("%v.CloseAndRecv() got error %v, want %v", stream, err, nil)
	}
	for i := 0; i < n; i++ {
		fmt.Println("[ClientStreamingRPC] Greeting :", i, ":", reply.GetMessage())
	}
	elapsed3 := time.Since(start3)

	fmt.Println("Execution time of SimpleRPC :", elapsed)
	fmt.Println("Execution time of BidirectionalRPC :", elapsed1)
	fmt.Println("Execution time of ServerStreamingRPC :", elapsed2)
	fmt.Println("Execution time of ClientStreamingRPC :", elapsed3)

}
