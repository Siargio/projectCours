package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Siargio/shortener/internal/domain"
)

// Mock для репозитория
type mockLinkRepo struct {
	saveFunc            func(ctx context.Context, link *domain.Link) error
	findByShortCodeFunc func(ctx context.Context, shortCode string) (*domain.Link, error)
	updateClicksFunc    func(ctx context.Context, shortCode string) error
}

func (m *mockLinkRepo) Save(ctx context.Context, link *domain.Link) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, link)
	}
	return nil
}

func (m *mockLinkRepo) FindByShortCode(ctx context.Context, shortCode string) (*domain.Link, error) {
	if m.findByShortCodeFunc != nil {
		return m.findByShortCodeFunc(ctx, shortCode)
	}
	return nil, domain.ErrLinkNotFound
}

func (m *mockLinkRepo) UpdateClicks(ctx context.Context, shortCode string) error {
	if m.updateClicksFunc != nil {
		return m.updateClicksFunc(ctx, shortCode)
	}
	return nil
}

// Mock для кеша
type mockCache struct {
	getFunc func(ctx context.Context, key string) (string, error)
	setFunc func(ctx context.Context, key, value string, ttl time.Duration) error
}

func (m *mockCache) Get(ctx context.Context, key string) (string, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, key)
	}
	return "", errors.New("cache miss")
}

func (m *mockCache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	if m.setFunc != nil {
		return m.setFunc(ctx, key, value, ttl)
	}
	return nil
}

// ==================== ТЕСТЫ ====================

func TestLinkService_Shorten(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockLinkRepo{
			saveFunc: func(ctx context.Context, link *domain.Link) error {
				link.ID = 1
				link.CreatedAt = time.Now()
				return nil
			},
		}
		cache := &mockCache{}
		service := NewLinkService(repo, cache, "http://localhost:8080")

		shortURL, err := service.Shorten(context.Background(), "https://example.com")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if shortURL == "" {
			t.Error("expected non-empty short URL")
		}
	})

	t.Run("save error", func(t *testing.T) {
		repo := &mockLinkRepo{
			saveFunc: func(ctx context.Context, link *domain.Link) error {
				return errors.New("db error")
			},
		}
		cache := &mockCache{}
		service := NewLinkService(repo, cache, "http://localhost:8080")

		_, err := service.Shorten(context.Background(), "https://example.com")

		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestLinkService_GetLongURL(t *testing.T) {
	t.Run("cache hit", func(t *testing.T) {
		cache := &mockCache{
			getFunc: func(ctx context.Context, key string) (string, error) {
				return "https://cached.com", nil
			},
		}
		repo := &mockLinkRepo{}
		service := NewLinkService(repo, cache, "http://localhost:8080")

		longURL, err := service.GetLongURL(context.Background(), "abc123")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if longURL != "https://cached.com" {
			t.Errorf("expected cached url, got %s", longURL)
		}
	})

	t.Run("cache miss, db hit", func(t *testing.T) {
		cache := &mockCache{
			getFunc: func(ctx context.Context, key string) (string, error) {
				return "", errors.New("cache miss")
			},
		}
		repo := &mockLinkRepo{
			findByShortCodeFunc: func(ctx context.Context, shortCode string) (*domain.Link, error) {
				return &domain.Link{
					ShortCode: shortCode,
					LongURL:   "https://db.com",
				}, nil
			},
		}
		service := NewLinkService(repo, cache, "http://localhost:8080")

		longURL, err := service.GetLongURL(context.Background(), "abc123")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if longURL != "https://db.com" {
			t.Errorf("expected db url, got %s", longURL)
		}
	})

	t.Run("not found", func(t *testing.T) {
		cache := &mockCache{
			getFunc: func(ctx context.Context, key string) (string, error) {
				return "", errors.New("cache miss")
			},
		}
		repo := &mockLinkRepo{
			findByShortCodeFunc: func(ctx context.Context, shortCode string) (*domain.Link, error) {
				return nil, domain.ErrLinkNotFound
			},
		}
		service := NewLinkService(repo, cache, "http://localhost:8080")

		_, err := service.GetLongURL(context.Background(), "notexist")

		if err != domain.ErrLinkNotFound {
			t.Errorf("expected ErrLinkNotFound, got %v", err)
		}
	})
}

func TestLinkService_GetStats(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo := &mockLinkRepo{
			findByShortCodeFunc: func(ctx context.Context, shortCode string) (*domain.Link, error) {
				return &domain.Link{
					ID:        1,
					ShortCode: shortCode,
					LongURL:   "https://stats.com",
					Clicks:    42,
				}, nil
			},
		}
		cache := &mockCache{}
		service := NewLinkService(repo, cache, "http://localhost:8080")

		link, err := service.GetStats(context.Background(), "abc123")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if link.Clicks != 42 {
			t.Errorf("expected clicks 42, got %d", link.Clicks)
		}
	})
}
