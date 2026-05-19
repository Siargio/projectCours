package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Siargio/shortener/internal/domain"
)

// Mock для сервиса
type mockLinkService struct {
	shortenFunc    func(ctx context.Context, longURL string) (string, error)
	getLongURLFunc func(ctx context.Context, shortCode string) (string, error)
	getStatsFunc   func(ctx context.Context, shortCode string) (*domain.Link, error)
}

func (m *mockLinkService) Shorten(ctx context.Context, longURL string) (string, error) {
	if m.shortenFunc != nil {
		return m.shortenFunc(ctx, longURL)
	}
	return "", nil
}

func (m *mockLinkService) GetLongURL(ctx context.Context, shortCode string) (string, error) {
	if m.getLongURLFunc != nil {
		return m.getLongURLFunc(ctx, shortCode)
	}
	return "", domain.ErrLinkNotFound
}

func (m *mockLinkService) GetStats(ctx context.Context, shortCode string) (*domain.Link, error) {
	if m.getStatsFunc != nil {
		return m.getStatsFunc(ctx, shortCode)
	}
	return nil, domain.ErrLinkNotFound
}

// ==================== ТЕСТЫ ====================

func TestLinkHandler_Shorten(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		service := &mockLinkService{
			shortenFunc: func(ctx context.Context, longURL string) (string, error) {
				return "http://localhost:8080/abc123", nil
			},
		}
		handler := NewLinkHandler(service)

		body := bytes.NewBufferString(`{"url": "https://example.com"}`)
		req := httptest.NewRequest("POST", "/shorten", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Shorten(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d", w.Code)
		}

		var resp ShortenResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.ShortURL == "" {
			t.Error("expected non-empty short_url")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		service := &mockLinkService{}
		handler := NewLinkHandler(service)

		req := httptest.NewRequest("POST", "/shorten", bytes.NewBufferString(`{invalid}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Shorten(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("empty url", func(t *testing.T) {
		service := &mockLinkService{}
		handler := NewLinkHandler(service)

		body := bytes.NewBufferString(`{"url": ""}`)
		req := httptest.NewRequest("POST", "/shorten", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Shorten(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("service error", func(t *testing.T) {
		service := &mockLinkService{
			shortenFunc: func(ctx context.Context, longURL string) (string, error) {
				return "", domain.ErrLinkNotFound
			},
		}
		handler := NewLinkHandler(service)

		body := bytes.NewBufferString(`{"url": "https://example.com"}`)
		req := httptest.NewRequest("POST", "/shorten", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Shorten(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

func TestLinkHandler_Redirect(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		service := &mockLinkService{
			getLongURLFunc: func(ctx context.Context, shortCode string) (string, error) {
				return "https://example.com", nil
			},
		}
		handler := NewLinkHandler(service)

		req := httptest.NewRequest("GET", "/abc123", nil)
		req.SetPathValue("code", "abc123") // Важно для Go 1.22+ роутинга
		w := httptest.NewRecorder()

		handler.Redirect(w, req)

		if w.Code != http.StatusMovedPermanently {
			t.Errorf("expected status 301, got %d", w.Code)
		}

		location := w.Header().Get("Location")
		if location != "https://example.com" {
			t.Errorf("expected Location https://example.com, got %s", location)
		}
	})

	t.Run("not found", func(t *testing.T) {
		service := &mockLinkService{
			getLongURLFunc: func(ctx context.Context, shortCode string) (string, error) {
				return "", domain.ErrLinkNotFound
			},
		}
		handler := NewLinkHandler(service)

		req := httptest.NewRequest("GET", "/notexist", nil)
		req.SetPathValue("code", "notexist")
		w := httptest.NewRecorder()

		handler.Redirect(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})
}

func TestLinkHandler_Stats(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		service := &mockLinkService{
			getStatsFunc: func(ctx context.Context, shortCode string) (*domain.Link, error) {
				return &domain.Link{
					ShortCode: shortCode,
					LongURL:   "https://example.com",
					Clicks:    42,
				}, nil
			},
		}
		handler := NewLinkHandler(service)

		req := httptest.NewRequest("GET", "/stats/abc123", nil)
		req.SetPathValue("code", "abc123")
		w := httptest.NewRecorder()

		handler.Stats(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var resp StatsResponse
		if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if resp.Clicks != 42 {
			t.Errorf("expected clicks 42, got %d", resp.Clicks)
		}
	})

	t.Run("not found", func(t *testing.T) {
		service := &mockLinkService{
			getStatsFunc: func(ctx context.Context, shortCode string) (*domain.Link, error) {
				return nil, domain.ErrLinkNotFound
			},
		}
		handler := NewLinkHandler(service)

		req := httptest.NewRequest("GET", "/stats/notexist", nil)
		req.SetPathValue("code", "notexist")
		w := httptest.NewRecorder()

		handler.Stats(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})
}
