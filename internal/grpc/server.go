package grpc

import (
	"fmt"
	"net"

	"google.golang.org/grpc"

	"github.com/GevorkovG/go-shortener-tlp/internal/app"
	"github.com/GevorkovG/go-shortener-tlp/internal/objects"
)

type Server struct {
	UnimplementedShortenerServer
	Storage objects.Storage
	Server  *grpc.Server
	App     *app.App
}

func NewServer(storage objects.Storage, app1 *app.App) *Server {
	return &Server{
		Storage: storage,
		Server:  grpc.NewServer(),
		App:     app1,
	}
}

func Run(s *Server) error {

	listen, err := net.Listen("tcp", ":3200")

	if err != nil {
		return err
	}

	RegisterShortenerServer(s.Server, s.UnimplementedShortenerServer)
	fmt.Println("Сервер gRPC начал работу")
	if err := s.Server.Serve(listen); err != nil {
		return err
	}

	return nil
}

func (s *Server) Stop() {
	s.Server.GracefulStop()
}
