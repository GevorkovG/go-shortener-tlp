package grpc

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/GevorkovG/go-shortener-tlp/internal/app"
	"github.com/GevorkovG/go-shortener-tlp/internal/objects"
)

// ShortURLBatch -
func (s *Server) ShortURLBatch(_, req *ShortURLBatchRequest) (*ShortURLBatchResponse, error) {

	var (
		response []*BatchResponse
		links    []*objects.Link
		userID   string
	)

	userID = req.UserId

	if userID == "" {
		userID = uuid.New().String()
	}

	for _, url := range req.GetUrls() {

		key := app.GenerateID()
		resp := &BatchResponse{
			CorrelationId: url.CorrelationId,
			ResultUrl:     fmt.Sprintf(s.App.GetConfig().ResultURL+"/%s", key),
		}
		link := &objects.Link{
			Short:    key,
			Original: url.OriginalUrl,
			UserID:   userID,
		}

		response = append(response, resp)
		links = append(links, link)

	}

	if err := s.Storage.InsertLinks(context.Background(), links); err != nil {
		return nil, status.Error(codes.InvalidArgument, `Don't insert URLs`)
	}

	return &ShortURLBatchResponse{
		Urls: response,
	}, nil
}
