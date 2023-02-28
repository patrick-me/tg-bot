package main

import (
	"flag"
	pb "github.com/patrick-me/tg-bot/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
)

var (
	addr = flag.String("addr",
		getEnvOrDefault("SERVER_ADDR", "localhost:50051"),
		"the address to connect to")
)

func getEnvOrDefault(key string, def string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return def
}

func CreateProxyClient() pb.ProxyClient {
	flag.Parse()
	// Set up a connection to the server.
	log.Printf("Connecting to addr: %v", *addr)
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	//defer conn.Close()
	cl := pb.NewProxyClient(conn)
	return cl
}
