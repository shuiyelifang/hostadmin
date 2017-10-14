package main

import (
	"net"
	"net/http"
	"os"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/sirupsen/logrus"
	"github.com/xuebing1110/hostadmin/cluster-admin/server"
	"github.com/xuebing1110/hostadmin/log"
	pb "github.com/xuebing1110/hostadmin/proto/HostManager"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var GRPC_PORT string
var RESTFUL_FLAG string
var logger *logrus.Logger

func init() {
	GRPC_PORT = os.Getenv("GRPC_PORT")
	if GRPC_PORT == "" {
		GRPC_PORT = "50051"
	}

	RESTFUL_FLAG = os.Getenv("RESTFUL_FLAG")

	log.InitLogger("", false)
	logger = log.GetLogger()
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var exit = make(chan bool, 1)

	// grpc server
	go func() {
		defer func() {
			exit <- true
		}()

		logger.Info("start grpc server...")
		err := grpcServer()
		if err != nil {
			logger.Fatalf("failed to start grpc server: %v", err)
		}
	}()

	// rest server
	if RESTFUL_FLAG == "1" || RESTFUL_FLAG == "true" {
		go func() {
			defer func() {
				exit <- true
			}()

			logger.Info("start rest server...")
			err := restServer(ctx)
			if err != nil {
				logger.Fatalf("failed to start rest server: %v", err)
			}
		}()
	}

	// wait
	<-exit
}

func grpcServer() error {
	lis, err := net.Listen("tcp", ":"+GRPC_PORT)
	if err != nil {
		return err
	}

	s := grpc.NewServer()
	pb.RegisterHostManagerServer(s, &server.HostManagerServer{})

	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		return err
	}

	return nil
}

func restServer(ctx context.Context) error {
	// mux := http.NewServeMux()
	gwmux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := pb.RegisterHostManagerHandlerFromEndpoint(ctx, gwmux, "127.0.0.1:"+GRPC_PORT, opts)
	if err != nil {
		return err
	}

	// mux.Handle("/", gwmux)
	// serveSwagger(mux)
	err = http.ListenAndServe(":8080", gwmux)
	if err != nil {
		return err
	}

	return nil
}
