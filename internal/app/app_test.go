package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/GevorkovG/go-shortener-tlp/config"
	"github.com/GevorkovG/go-shortener-tlp/internal/objects"
	"github.com/GevorkovG/go-shortener-tlp/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetStats(t *testing.T) {
	// Настройка тестового сервера
	cfg := &config.AppConfig{
		TrustedSubnet: "192.168.1.0/24",
	}
	storage := storage.NewInMemoryStorage()
	app := &App{cfg: cfg, Storage: storage}

	ts := httptest.NewServer(http.HandlerFunc(app.GetStats))
	defer ts.Close()

	tests := []struct {
		name       string
		ip         string
		wantStatus int
	}{
		{
			name:       "Valid IP",
			ip:         "192.168.1.100",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Invalid IP",
			ip:         "10.0.0.1",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "No IP header",
			ip:         "",
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", ts.URL, nil)
			if tt.ip != "" {
				req.Header.Set("X-Real-IP", tt.ip)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, resp.StatusCode)
			}
		})
	}
}

func TestGetStatsWithFileStorage(t *testing.T) {
	type want struct {
		contentType string
		code        int
		urls        int
		users       int
		ErrMsg      string
	}

	tests := []struct {
		name          string
		method        string
		ip            string
		prepopData    []*objects.Link
		trustedSubnet string
		want          want
	}{
		{
			name:          "successful request with stats",
			method:        http.MethodGet,
			ip:            "192.168.1.100",
			trustedSubnet: "192.168.1.0/24",
			prepopData: []*objects.Link{
				{Short: "short1", Original: "http://example.com/1", UserID: "user1"},
				{Short: "short2", Original: "http://example.com/2", UserID: "user1"},
				{Short: "short3", Original: "http://example.com/3", UserID: "user2"},
			},
			want: want{
				code:        http.StatusOK,
				contentType: "application/json",
				urls:        3,
				users:       2,
			},
		},
		{
			name:          "no access - IP not in trusted subnet",
			method:        http.MethodGet,
			ip:            "10.0.0.1",
			trustedSubnet: "192.168.1.0/24",
			prepopData: []*objects.Link{
				{Short: "short1", Original: "http://example.com/1", UserID: "user1"},
			},
			want: want{
				code:        http.StatusForbidden,
				contentType: "text/plain; charset=utf-8",
				ErrMsg:      "forbidden",
			},
		},
		{
			name:          "no access - missing X-Real-IP header",
			method:        http.MethodGet,
			ip:            "",
			trustedSubnet: "192.168.1.0/24",
			prepopData: []*objects.Link{
				{Short: "short1", Original: "http://example.com/1", UserID: "user1"},
			},
			want: want{
				code:        http.StatusForbidden,
				contentType: "text/plain; charset=utf-8",
				ErrMsg:      "required",
			},
		},
		{
			name:          "no access - empty trusted subnet",
			method:        http.MethodGet,
			ip:            "192.168.1.100",
			trustedSubnet: "",
			prepopData: []*objects.Link{
				{Short: "short1", Original: "http://example.com/1", UserID: "user1"},
			},
			want: want{
				code:        http.StatusForbidden,
				contentType: "text/plain; charset=utf-8",
				ErrMsg:      "forbidden",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile, err := os.CreateTemp("", "test-stats-*.json")
			require.NoError(t, err)
			tmpFileName := tmpFile.Name()
			tmpFile.Close()
			defer os.Remove(tmpFileName)

			conf := &config.AppConfig{
				Host:          "localhost:8080",
				ResultURL:     "http://localhost:8080",
				FilePATH:      tmpFileName,
				TrustedSubnet: tt.trustedSubnet,
			}

			storage := storage.NewFileStorage(conf.FilePATH)
			app := &App{
				Storage: storage,
				cfg:     conf,
			}

			// Вставляем тестовые данные
			for _, link := range tt.prepopData {
				err := app.Storage.Insert(context.Background(), link)
				require.NoError(t, err)
			}

			r := httptest.NewRequest(tt.method, "/api/internal/stats", nil)
			if tt.ip != "" {
				r.Header.Set("X-Real-IP", tt.ip)
			}

			w := httptest.NewRecorder()
			app.GetStats(w, r)

			assert.Equal(t, tt.want.code, w.Code, "Unexpected status code")
			assert.Equal(t, tt.want.contentType, w.Header().Get("Content-Type"))

			if tt.want.code == http.StatusOK {
				var stats struct {
					URLs  int `json:"urls"`
					Users int `json:"users"`
				}
				require.NoError(t, json.NewDecoder(w.Body).Decode(&stats))
				assert.Equal(t, tt.want.urls, stats.URLs)
				assert.Equal(t, tt.want.users, stats.Users)
			} else if tt.want.code == http.StatusForbidden {
				assert.Contains(t, w.Body.String(), tt.want.ErrMsg, "Error message should contain 'forbidden'")
			}
		})
	}
}
