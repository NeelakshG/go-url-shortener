package db

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// migrate runs the schema migration against the test database.
func migrate(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	sql, err := os.ReadFile("../../migrations/001_create_links.sql")
	require.NoError(t, err)
	_, err = pool.Exec(context.Background(), string(sql))
	require.NoError(t, err)
}

// setupPostgres starts a throwaway Postgres container and returns the DSN + cleanup func.
func setupPostgres(t *testing.T) (string, func()) {
	t.Helper()
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:16",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	dsn := fmt.Sprintf(
		"postgres://test:test@%s:%s/testdb?sslmode=disable",
		host, port.Port(),
	)

	cleanup := func() {
		_ = container.Terminate(ctx)
	}

	return dsn, cleanup
}

func TestCreateLink_Integration(t *testing.T) {
	dsn, cleanup := setupPostgres(t)
	defer cleanup()

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err)
	defer pool.Close()

	require.NoError(t, pool.Ping(ctx))

	migrate(t, pool)

	store := NewStore(pool)

	longURL := "https://google.com"

	link, err := store.CreateLink(ctx, longURL)
	require.NoError(t, err)
	require.NotNil(t, link)
	require.Equal(t, longURL, link.LongURL)
	require.Len(t, link.ShortCode, 6)

	// confirm the link actually exists in the DB
	fetched, err := store.GetLink(ctx, link.ShortCode)
	require.NoError(t, err)
	require.Equal(t, longURL, fetched.LongURL)
	require.Equal(t, link.ShortCode, fetched.ShortCode)
}
