package main

import (
	"context"
	"flag"
	"fmt"
	pb "github.com/patrick-me/tg-bot/proto"
	"google.golang.org/grpc"
	"log"
	"net"
)

var (
	port = flag.Int("port", 50051, "Server port")
)

type server struct {
	pb.UnimplementedProxyServer
}

func (s *server) Process(ctx context.Context, in *pb.ProxyRequest) (*pb.ProxyResponse, error) {
	log.Printf("Received: %v", in.GetMessage())
	return &pb.ProxyResponse{Message: "Got msg: *" + in.Message + "*", ApplyMarkdownV2: true}, nil
}

func main() {
	flag.Parse()

	for {
		saveRunner()
	}
}

func saveRunner() {
	defer func() {
		if err := recover(); err != nil {
			log.Println("panic occurred:", err)
		}
	}()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))

	if err != nil {
		fmt.Printf("failed to listen: %v\n", err)
	}

	s := grpc.NewServer()
	pb.RegisterProxyServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())

	if err = s.Serve(lis); err != nil {
		fmt.Printf("failed to serve: %v\n", err)
	}
}
