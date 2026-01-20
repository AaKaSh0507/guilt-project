package main

import (
	"log"
	"net"

	"google.golang.org/grpc"
)

func StartGRPCServer(register func(*grpc.Server)) {
	l, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	register(s)

	log.Println("gRPC server listening on :9090")

	if err := s.Serve(l); err != nil {
		log.Fatalf("failed to serve gRPC: %v", err)
	}
}
