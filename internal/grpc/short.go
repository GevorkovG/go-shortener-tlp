package grpc

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/GevorkovG/go-shortener-tlp/internal/app"
	"github.com/GevorkovG/go-shortener-tlp/internal/objects"
	"github.com/GevorkovG/go-shortener-tlp/internal/storage"
)

// CreateShortURL -
func (s *Server) CreateShortURL(_, req *ShortURLRequest) (*ShortURLResponse, error) {

	url := req.Url

	if len(url) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Empty URL")
	}

	userID := req.UserId

	if len(userID) == 0 {
		userID = uuid.New().String()
	}

	link := &objects.Link{
		Short:    app.GenerateID(),
		Original: url,
		UserID:   userID,
	}

	if err := s.Storage.Insert(context.Background(), link); err != nil {
		if errors.Is(err, storage.ErrConflict) {
			link, err = s.Storage.GetShort(link.Original)
			if err != nil {
				zap.L().Error("Don't get short URL", zap.Error(err))
				return nil, status.Errorf(codes.Internal, "Internal error")
			}
		} else {
			zap.L().Error("Don't insert URL", zap.Error(err))
			return nil, status.Errorf(codes.Internal, "Internal error")
		}
	}
	return &ShortURLResponse{
		ShortUrl: link.Short,
	}, nil
}
