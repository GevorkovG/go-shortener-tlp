package grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Ping проверяет доступность хранилища
func (s *Server) Ping(ctx context.Context, req *PingRequest) (*PingResponse, error) {
	if ctx.Err() != nil {
		return nil, status.Error(codes.Canceled, "request canceled")
	}

	err := s.Storage.Ping()
	if err != nil {
		return nil, status.Error(codes.Internal, "database unavailable")
	}

	return &PingResponse{
		Result: "OK",
	}, nil
}
