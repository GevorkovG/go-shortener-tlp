package grpc

import (
	"context"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetOriginalURL обрабатывает запрос на получение оригинального URL по сокращенной версии
func (s *Server) GetOriginalURL(ctx context.Context, req *ShortURLRequest) (*OriginalURLResponse, error) {

	if ctx.Err() != nil {
		return nil, status.Error(codes.Canceled, "request canceled")
	}

	short := req.GetUrl()

	if len(short) == 0 {
		return nil, status.Error(codes.InvalidArgument, "empty URL")
	}

	link, err := s.Storage.GetOriginal(short)
	if err != nil {
		log.Printf("Failed to get original URL: %v", err)
		return nil, status.Error(codes.NotFound, "URL not found")
	}

	return &OriginalURLResponse{
		Original: link.Original,
	}, nil
}
