package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/NeelakshG/snippy/internal/model"
	"github.com/NeelakshG/snippy/internal/shortcode"
)

// Store implements handler.Store backed by Postgres.
type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// CreateLink generates a short code, inserts the link, and retries on collisions.
// Collision handling: if Postgres returns a unique-violation (23505), we generate
// a new code and retry — the loop exits only on success or a non-collision error.
func (s *Store) CreateLink(ctx context.Context, longURL string) (*model.Link, error) {
	for {
		// Generate a fresh 6-char base62 code each attempt
		code, err := shortcode.Generate()
		if err != nil {
			return nil, fmt.Errorf("generate short code: %w", err)
		}

		link := &model.Link{}
		err = s.pool.QueryRow(ctx,
			`INSERT INTO links (short_code, long_url)
			 VALUES ($1, $2)
			 RETURNING id, short_code, long_url, created_at`,
			code, longURL,
		).Scan(&link.ID, &link.ShortCode, &link.LongURL, &link.CreatedAt)

		if err == nil {
			// Insert succeeded — return the new link
			return link, nil
		}

		// Check if Postgres rejected the insert due to a duplicate short_code
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			// Collision — loop back and try a new code
			continue
		}

		// Any other error is unexpected — wrap and return it
		return nil, fmt.Errorf("insert link: %w", err)
	}
}

func (s *Store) GetLink(ctx context.Context, shortCode string) (*model.Link, error) {
	link := &model.Link{}
	err := s.pool.QueryRow(ctx,
		`SELECT id, short_code, long_url, created_at
		 FROM links WHERE short_code = $1`,
		shortCode,
	).Scan(&link.ID, &link.ShortCode, &link.LongURL, &link.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get link: %w", err)
	}
	return link, nil
}

func (s *Store) IncrementClicks(ctx context.Context, shortCode string) error {
	_, err := s.pool.Exec(ctx,
		`UPDATE links SET click_count = click_count + 1 WHERE short_code = $1`,
		shortCode,
	)
	return err
}

func (s *Store) GetClickCount(ctx context.Context, shortCode string) (int64, error) {
	var count int64
	err := s.pool.QueryRow(ctx,
		`SELECT click_count FROM links WHERE short_code = $1`,
		shortCode,
	).Scan(&count)
	return count, err
}
