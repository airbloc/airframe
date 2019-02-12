package rpcserver

import (
	"fmt"
	"github.com/airbloc/airframe/database"
	"github.com/airbloc/logger"
	"google.golang.org/grpc"
	"net"
)

var (
	log = logger.New("rpcserver")
)

type Server struct {
	srv  *grpc.Server
	port string
}

func New(backend database.Database, port int, debug bool) *Server {
	srv := grpc.NewServer()
	RegisterV1API(srv, backend)
	return &Server{
		srv:  srv,
		port: fmt.Sprintf(":%d", port),
	}
}

func (s *Server) Start() error {
	lis, err := net.Listen("tcp", s.port)
	if err != nil {
		return err
	}
	return s.srv.Serve(lis)
}

func (s *Server) Stop() {
	s.srv.GracefulStop()
}
