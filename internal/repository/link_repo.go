package repository

import (
	"context"

	"github.com/Siargio/shortener/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LinkRepository interface {
	Save(ctx context.Context, link *domain.Link) error
	FindByShortCode(ctx context.Context, shortCode string) (*domain.Link, error)
	UpdateClicks(ctx context.Context, shortCode string) error
}

type PostgresLinkRepo struct {
	pool *pgxpool.Pool
}

func NewPostgresLinkRepo(pool *pgxpool.Pool) *PostgresLinkRepo {
	return &PostgresLinkRepo{pool: pool}
}

func (r *PostgresLinkRepo) Save(ctx context.Context, link *domain.Link) error {
	query := `INSERT INTO links (short_code, long_url) VALUES ($1, $2) RETURNING id, created_at`
	return r.pool.QueryRow(ctx, query, link.ShortCode, link.LongURL).Scan(&link.ID, &link.CreatedAt)
}

func (r *PostgresLinkRepo) FindByShortCode(ctx context.Context, shortCode string) (*domain.Link, error) {
	query := `SELECT id, short_code, long_url, clicks, created_at FROM links WHERE short_code = $1`
	var link domain.Link
	err := r.pool.QueryRow(ctx, query, shortCode).Scan(
		&link.ID, &link.ShortCode, &link.LongURL, &link.Clicks, &link.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, domain.ErrLinkNotFound
	}
	if err != nil {
		return nil, err
	}
	return &link, nil
}

func (r *PostgresLinkRepo) UpdateClicks(ctx context.Context, shortCode string) error {
	query := `UPDATE links SET clicks = clicks + 1 WHERE short_code = $1`
	_, err := r.pool.Exec(ctx, query, shortCode)
	return err
}
