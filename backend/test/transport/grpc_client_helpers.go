package transport

import (
	"net"
	"testing"

	"google.golang.org/grpc"
)

type testGRPCServer struct {
	grpcServer *grpc.Server
	lis        net.Listener
	addr       string
}

func startTestGRPC(t *testing.T, register func(*grpc.Server)) *testGRPCServer {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}

	s := grpc.NewServer()
	register(s)

	go func() {
		_ = s.Serve(l)
	}()

	return &testGRPCServer{
		grpcServer: s,
		lis:        l,
		addr:       l.Addr().String(),
	}
}

func (s *testGRPCServer) stop() {
	s.grpcServer.Stop()
}

func (s *testGRPCServer) getAddr() string {
	return s.addr
}
