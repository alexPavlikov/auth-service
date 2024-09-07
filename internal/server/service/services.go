package service

import (
	"context"
	"fmt"

	"github.com/alexPavlikov/auth-service/internal/models"
	"github.com/alexPavlikov/auth-service/internal/server/repository"
)

type Service struct {
	Repo *repository.Repository
}

func New(repo *repository.Repository) *Service {
	return &Service{
		Repo: repo,
	}
}

func (s *Service) Auth(ctx context.Context, user models.UserPayLoad) (models.User, error) {
	passHash := user.Password // зашифровать пароль
	var usr = models.User{
		Login:    user.Login,
		PassHash: passHash,
	}

	usr, err := s.Repo.Authentication(ctx, usr)
	if err != nil {
		return models.User{}, fmt.Errorf("service auth error: %w", err)
	}

	// check token or create token

	return usr, nil
}
