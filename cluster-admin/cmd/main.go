package main

import (
	"log"
	"net"

	"github.com/xuebing1110/hostadmin/cluster-admin/server"
	pb "github.com/xuebing1110/hostadmin/proto/HostManager"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterHostManagerServer(s, &server.HostManagerServer{})

	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
