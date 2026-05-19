package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Siargio/shortener/internal/domain"
	"github.com/Siargio/shortener/internal/repository"
)

type LinkCache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
}

type LinkService struct {
	repo    repository.LinkRepository
	cache   LinkCache
	baseURL string
}

func NewLinkService(repo repository.LinkRepository, cache LinkCache, baseURL string) *LinkService {
	return &LinkService{
		repo:    repo,
		cache:   cache,
		baseURL: baseURL,
	}
}

func (s *LinkService) Shorten(ctx context.Context, longURL string) (string, error) {
	// Генерируем короткий код
	shortCode, err := domain.GenerateShortCode(8)
	if err != nil {
		return "", err
	}

	link := &domain.Link{
		ShortCode: shortCode,
		LongURL:   longURL,
	}

	if err := s.repo.Save(ctx, link); err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", s.baseURL, shortCode), nil
}

func (s *LinkService) GetLongURL(ctx context.Context, shortCode string) (string, error) {
	// Сначала проверяем кеш
	if longURL, err := s.cache.Get(ctx, shortCode); err == nil {
		// Обновляем счётчик асинхронно (можно фоном)
		go s.repo.UpdateClicks(context.Background(), shortCode)
		return longURL, nil
	}

	// Ищем в БД
	link, err := s.repo.FindByShortCode(ctx, shortCode)
	if err != nil {
		return "", err
	}

	// Сохраняем в кеш (на 5 минут)
	s.cache.Set(ctx, shortCode, link.LongURL, 5*time.Minute)

	// Обновляем счётчик
	s.repo.UpdateClicks(ctx, shortCode)

	return link.LongURL, nil
}

func (s *LinkService) GetStats(ctx context.Context, shortCode string) (*domain.Link, error) {
	return s.repo.FindByShortCode(ctx, shortCode)
}
