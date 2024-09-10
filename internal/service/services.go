package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/alexPavlikov/auth-service/internal/config"
	"github.com/alexPavlikov/auth-service/internal/models"
	"github.com/alexPavlikov/auth-service/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	Repo *repository.Repository
	Cfg  *config.Config
}

func New(repo *repository.Repository) *Service {
	return &Service{
		Repo: repo,
	}
}

func (s *Service) Auth(ctx context.Context, user models.User) (string, string, error) {

	usr, err := s.FindUserByUUID(user.UUID)
	if err != nil {
		return "", "", fmt.Errorf("failed find user: %w", err)
	}

	if usr.IPAddress != user.IP {
		// send to email message
		return "", "", errors.New("another ip address")
	}

	accessTokenID := uuid.New()

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"jti": accessTokenID,
		"sub": user.UUID,
		"ip":  user.IP,
	})

	tokenString, err := token.SignedString(s.Cfg.Secret)
	if err != nil {
		return "", "", fmt.Errorf("failed access token signing string: %w", err)
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"sub": user.UUID,
	})

	refreshTokenString, err := refreshToken.SignedString(s.Cfg.Secret)
	if err != nil {
		return "", "", fmt.Errorf("failed refresh token signing string: %w", err)
	}

	hashRefreshToken, err := bcrypt.GenerateFromPassword(refreshToken.Signature, 4)
	if err != nil {
		return "", "", fmt.Errorf("hash refresh token err: %w", err)
	}

	var userStorage = models.UserStore{
		UUID:             user.UUID,
		AccessTokenID:    accessTokenID.String(),
		RefreshTokenHash: string(hashRefreshToken),
		IPAddress:        user.IP,
	}

	if err := s.UpdateAuthUser(userStorage); err != nil {
		return "", "", fmt.Errorf("update user err: %w", err)
	}

	return tokenString, refreshTokenString, nil
}

func (s *Service) FindUserByUUID(uuid uuid.UUID) (user models.UserStore, err error) {

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	user, err = s.Repo.FindUserByUUID(ctx, uuid)
	if err != nil {
		return models.UserStore{}, fmt.Errorf("failed to find user by uuid: %w", err)
	}

	return user, nil
}

func (s *Service) UpdateAuthUser(user models.UserStore) error {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	if err := s.Repo.UpdateUserAuth(ctx, user); err != nil {
		return fmt.Errorf("failed update user auth: %w", err)
	}

	return nil
}

func (s *Service) RefreshUserAuthToken(ip string, access string, refresh string) (string, error) {

	keyFunc := func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Неожиданный метод подписи: %v", t.Header["alg"])
		}
		return s.Cfg.Secret, nil
	}

	claims := &jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(access, claims, keyFunc)
	if err != nil {
		return "", fmt.Errorf("failed parse acces token: %w", err)
	}

	var tokenID, UUID string

	for key, val := range *claims {
		switch key {
		case "jti":
			tokenID = val.(string)
		case "sub":
			UUID = val.(string)
		case "ip":
			if val != ip {
				//send to email
				return "", errors.New("another ip address check your email")
			}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	if err := s.Repo.FindAccessTokenByID(ctx, tokenID); err != nil {
		return "", fmt.Errorf("failed find access token: %w", err)
	}

	hashRefreshToken, err := bcrypt.GenerateFromPassword([]byte(refresh), 4)
	if err != nil {
		return "", fmt.Errorf("hash refresh token err: %w", err)
	}

	ctx, cancel = context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	hashRef, err := s.Repo.SelectRefreshHashByUUID(ctx, uuid.MustParse(UUID))
	if err != nil {
		return "", fmt.Errorf("failed select refresh token: %w", err)
	}

	if string(hashRefreshToken) != hashRef {
		return "", errors.New("another refresh token")
	}

	accessTokenID := uuid.New()

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"jti": accessTokenID,
		"sub": UUID,
		"ip":  ip,
	})

	tokenString, err := token.SignedString(s.Cfg.Secret)
	if err != nil {
		return "", fmt.Errorf("failed access token signing string: %w", err)
	}

	if err := s.Repo.UpdateUserAccessTokenID(ctx, tokenString, uuid.MustParse(UUID)); err != nil {
		return "", fmt.Errorf("failed update access token: %w", err)
	}

	return tokenString, nil
}
