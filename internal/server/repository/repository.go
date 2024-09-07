package repository

import (
	"context"

	"github.com/alexPavlikov/auth-service/internal/models"
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

func (r *Repository) Authentication(ctx context.Context, user models.User) (models.User, error) {
	query := `
	SELECT id, firstname, lastname, email, ip_address FROM "users" WHERE login = $1 AND pass_hash = $2 AND deleted = false
	`

	row := r.DB.QueryRow(ctx, query, user.Login, user.PassHash)

	if err := row.Scan(&user.ID, &user.Firstname, &user.Lastname, &user.Email, &user.LastIPAddress); err != nil {
		return models.User{}, err
	}

	return user, nil
}
