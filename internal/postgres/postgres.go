package postgres

import (
	"context"
	"fmt"

	"github.com/alexPavlikov/auth-service/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Connect(ctx context.Context, cfg config.Config) (*pgxpool.Pool, error) {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.Path, cfg.Postgres.Port, cfg.Postgres.DB)
	db, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("connect to postgres err: %w", err)
	}

	return db, nil
}
