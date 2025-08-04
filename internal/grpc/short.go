package grpc

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/GevorkovG/go-shortener-tlp/internal/app"
	"github.com/GevorkovG/go-shortener-tlp/internal/objects"
	"github.com/GevorkovG/go-shortener-tlp/internal/storage"
)

// CreateShortURL создает сокращенную версию URL
func (s *Server) CreateShortURL(ctx context.Context, req *ShortURLRequest) (*ShortURLResponse, error) {
	if ctx.Err() != nil {
		return nil, status.Error(codes.Canceled, "request canceled")
	}

	url := req.GetUrl()

	if url == "" {
		return nil, status.Error(codes.InvalidArgument, "URL is required")
	}

	userID := req.GetUserId()
	if userID == "" {
		userID = uuid.New().String()
	}

	link := &objects.Link{
		Short:    app.GenerateID(),
		Original: url,
		UserID:   userID,
	}

	err := s.Storage.Insert(ctx, link)
	if err != nil {
		if errors.Is(err, storage.ErrConflict) {
			link, err = s.Storage.GetShort(link.Original)
			if err != nil {
				zap.L().Error("Failed to get existing URL",
					zap.String("url", url),
					zap.Error(err))
				return nil, status.Error(codes.Internal, "failed to process URL")
			}
		} else {
			zap.L().Error("Failed to insert URL",
				zap.String("url", url),
				zap.Error(err))
			return nil, status.Error(codes.Internal, "failed to create short URL")
		}
	}

	return &ShortURLResponse{
		ShortUrl: fmt.Sprintf("%s/%s", s.App.GetConfig().ResultURL, link.Short),
	}, nil
}
