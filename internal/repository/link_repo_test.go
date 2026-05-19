package repository

import (
	"context"
	"os"
	"testing"

	"github.com/Siargio/shortener/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

//Проверка
//Проверяе что бд запущена
//docker compose up -d db
//Создаем тестовую БД
//docker exec -i projectcours-db-1 psql -U postgres -c "CREATE DATABASE shortener_test" 2>/dev/null || true
//Миграция для тестовой БД
//docker exec -i projectcours-db-1 psql -U postgres -d shortener_test < migrations/001_create_links.sql
//Запуск тестов
//go test -coverprofile=coverage.out ./internal/repository/...
//go tool cover -func=coverage.out | grep total

// setupTestDB поднимает тестовую БД и возвращает pool + cleanup функцию
func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	// Можно использовать ту же БД, но отдельную таблицу или транзакцию
	dbURL := os.Getenv("TEST_DB_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5433/shortener_test"
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Skipf("Cannot connect to test DB: %v. Run 'docker compose up -d db' first", err)
		return nil, nil
	}

	// Очищаем таблицу перед тестами
	_, err = pool.Exec(context.Background(), "TRUNCATE links RESTART IDENTITY")
	if err != nil {
		t.Fatalf("Failed to truncate table: %v", err)
	}

	cleanup := func() {
		pool.Close()
	}

	return pool, cleanup
}

// TestPostgresLinkRepo_Save тестирует сохранение ссылки
func TestPostgresLinkRepo_Save(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	if pool == nil {
		return
	}
	defer cleanup()

	repo := NewPostgresLinkRepo(pool)
	ctx := context.Background()

	link := &domain.Link{
		ShortCode: "save123",
		LongURL:   "https://example.com/save",
	}

	err := repo.Save(ctx, link)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if link.ID == 0 {
		t.Error("expected ID to be set, got 0")
	}

	if link.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set, got zero")
	}
}

// TestPostgresLinkRepo_FindByShortCode тестирует поиск по коду
func TestPostgresLinkRepo_FindByShortCode(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	if pool == nil {
		return
	}
	defer cleanup()

	repo := NewPostgresLinkRepo(pool)
	ctx := context.Background()

	// Сначала сохраняем тестовую ссылку
	original := &domain.Link{
		ShortCode: "find123",
		LongURL:   "https://example.com/find",
	}
	err := repo.Save(ctx, original)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Теперь ищем
	found, err := repo.FindByShortCode(ctx, "find123")
	if err != nil {
		t.Fatalf("FindByShortCode failed: %v", err)
	}

	if found.ShortCode != original.ShortCode {
		t.Errorf("expected ShortCode=%s, got %s", original.ShortCode, found.ShortCode)
	}
	if found.LongURL != original.LongURL {
		t.Errorf("expected LongURL=%s, got %s", original.LongURL, found.LongURL)
	}
	if found.ID != original.ID {
		t.Errorf("expected ID=%d, got %d", original.ID, found.ID)
	}
}

// TestPostgresLinkRepo_FindByShortCode_NotFound тестирует поиск несуществующей ссылки
func TestPostgresLinkRepo_FindByShortCode_NotFound(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	if pool == nil {
		return
	}
	defer cleanup()

	repo := NewPostgresLinkRepo(pool)
	ctx := context.Background()

	_, err := repo.FindByShortCode(ctx, "doesnotexist")
	if err != domain.ErrLinkNotFound {
		t.Errorf("expected ErrLinkNotFound, got %v", err)
	}
}

// TestPostgresLinkRepo_UpdateClicks тестирует обновление счётчика
func TestPostgresLinkRepo_UpdateClicks(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	if pool == nil {
		return
	}
	defer cleanup()

	repo := NewPostgresLinkRepo(pool)
	ctx := context.Background()

	// Сохраняем ссылку
	link := &domain.Link{
		ShortCode: "click123",
		LongURL:   "https://example.com/click",
	}
	err := repo.Save(ctx, link)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Обновляем счётчик
	err = repo.UpdateClicks(ctx, "click123")
	if err != nil {
		t.Fatalf("UpdateClicks failed: %v", err)
	}

	// Проверяем
	found, _ := repo.FindByShortCode(ctx, "click123")
	if found.Clicks != 1 {
		t.Errorf("expected Clicks=1, got %d", found.Clicks)
	}

	// Ещё раз обновляем
	repo.UpdateClicks(ctx, "click123")
	found, _ = repo.FindByShortCode(ctx, "click123")
	if found.Clicks != 2 {
		t.Errorf("expected Clicks=2, got %d", found.Clicks)
	}
}

// TestPostgresLinkRepo_UpdateClicks_NotFound обновление несуществующей ссылки не должно падать
func TestPostgresLinkRepo_UpdateClicks_NotFound(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	if pool == nil {
		return
	}
	defer cleanup()

	repo := NewPostgresLinkRepo(pool)
	ctx := context.Background()

	// Обновление несуществующей ссылки не возвращает ошибку,
	// просто ничего не обновляет (0 rows affected)
	err := repo.UpdateClicks(ctx, "nonexistent")
	if err != nil {
		t.Errorf("UpdateClicks for nonexistent returned error: %v", err)
	}
}
