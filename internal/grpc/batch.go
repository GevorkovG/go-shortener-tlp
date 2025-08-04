package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/GevorkovG/go-shortener-tlp/internal/app"
	"github.com/GevorkovG/go-shortener-tlp/internal/objects"
)

// ShortURLBatch обрабатывает пакетный запрос на сокращение URL
func (s *Server) ShortURLBatch(ctx context.Context, req *ShortURLBatchRequest) (*ShortURLBatchResponse, error) {

	if ctx.Err() != nil {
		return nil, status.Error(codes.Canceled, "request ShortURLBatch canceled")
	}

	var (
		response []*BatchResponse
		links    []*objects.Link
		userID   string = req.UserId
	)

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

	if err := s.Storage.InsertLinks(ctx, links); err != nil {
		return nil, status.Error(codes.Internal, "failed to insert URLs")
	}

	return &ShortURLBatchResponse{
		Urls: response,
	}, nil
}
