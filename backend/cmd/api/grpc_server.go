package main

import (
	"log"
	"net"

	"guiltmachine/internal/auth"
	grpchandlers "guiltmachine/internal/transport/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func StartGRPCServer(register func(*grpc.Server)) {
	l, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	register(s)
	reflection.Register(s)

	log.Println("gRPC server listening on :9090")

	if err := s.Serve(l); err != nil {
		log.Fatalf("failed to serve gRPC: %v", err)
	}
}

// StartGRPCServerWithAuth starts the gRPC server with JWT authentication interceptor
func StartGRPCServerWithAuth(jwtManager *auth.JWTManager, register func(*grpc.Server)) {
	l, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Create server with auth interceptors
	s := grpc.NewServer(
		grpc.UnaryInterceptor(grpchandlers.AuthInterceptor(jwtManager)),
		grpc.StreamInterceptor(grpchandlers.AuthStreamInterceptor(jwtManager)),
	)
	register(s)
	reflection.Register(s)

	log.Println("gRPC server listening on :9090 (with JWT auth)")

	if err := s.Serve(l); err != nil {
		log.Fatalf("failed to serve gRPC: %v", err)
	}
}
