package grpc

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GetUserURLs возвращает все URL пользователя
func (s *Server) GetUserURLs(ctx context.Context, req *UserURLsRequest) (*UserURLsResponse, error) {
	if ctx.Err() != nil {
		return nil, status.Error(codes.Canceled, "request canceled")
	}

	userID := req.GetUserId()

	if userID == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	userURLs, err := s.Storage.GetAllByUserID(userID)
	if err != nil {
		zap.L().Error("Failed to get user URLs",
			zap.String("userID", userID),
			zap.Error(err))
		return nil, status.Error(codes.NotFound, "failed to get user URLs")
	}

	var links []*URLsResponse
	for _, url := range userURLs {
		links = append(links, &URLsResponse{
			Short:    strings.TrimSpace(fmt.Sprintf(s.App.GetConfig().ResultURL+"/%s", url.Short)),
			Original: strings.TrimSpace(url.Original),
		})
	}

	return &UserURLsResponse{
		Urls: links,
	}, nil
}
