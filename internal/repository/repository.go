package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/alexPavlikov/auth-service/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	DB *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{
		DB: db,
	}
}

func (r *Repository) FindUserByUUID(ctx context.Context, userUUID uuid.UUID) (user models.UserStore, err error) {
	query := `
	SELECT uuid, email, ip_address FROM "users" WHERE uuid = $1
	`

	row := r.DB.QueryRow(ctx, query, userUUID)

	if err = row.Scan(&user.UUID, &user.Email, &user.IPAddress); err != nil {
		return models.UserStore{}, fmt.Errorf("find user by uuid scan err: %w", err)
	}

	return user, nil
}

func (r *Repository) UpdateUserAuth(ctx context.Context, user models.UserStore) error {
	query := `
	UPDATE "users" SET id_access_token = $1, hash_refresh_token = $2, ip_address = $3 WHERE uuid = $4
	`

	r.DB.QueryRow(ctx, query, user.AccessTokenID, user.RefreshTokenHash, user.IPAddress, user.UUID)

	return nil
}

func (r *Repository) SelectRefreshHashByUUID(ctx context.Context, uuid uuid.UUID) (string, error) {
	query := `
	SELECT hash_refresh_token FROM "users" WHERE uuid = $4
	`

	row := r.DB.QueryRow(ctx, query, uuid)

	var ref string

	if err := row.Scan(&ref); err != nil {
		return "", fmt.Errorf("failed scan refresh token: %w", err)
	}

	return ref, nil
}

func (r *Repository) FindAccessTokenByID(ctx context.Context, id string) error {
	query := `
	SELECT id_access_token FROM "users" WHERE id_access_token = $1
	`

	row := r.DB.QueryRow(ctx, query, id)

	if err := row.Scan(&id); err != nil {
		return err
	}

	if id == "" {
		return errors.New("not found access token id")
	}

	return nil
}

func (r *Repository) UpdateUserAccessTokenID(ctx context.Context, tokenID string, id uuid.UUID) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	query := `UPDATE "users" SET id_access_token = $1 WHERE uuid = $4 RETURNING id_access_token`

	row := tx.QueryRow(ctx, query, tokenID, id)

	var access string

	if err := row.Scan(&access); err != nil {
		return err
	}

	return nil
}
