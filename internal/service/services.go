package service

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"time"

	"github.com/alexPavlikov/auth-service/internal/config"
	"github.com/alexPavlikov/auth-service/internal/models"
	"github.com/alexPavlikov/auth-service/internal/repository"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	gomail "gopkg.in/mail.v2"
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
		if err := s.SendWarningToEmail(usr.Email, user.IP); err != nil {
			return "", "", fmt.Errorf("send email err: %w", err)
		}
		return "", "", errors.New("another ip address")
	}

	accessTokenID := uuid.New()

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"jti": accessTokenID,
		"sub": user.UUID,
		"ip":  user.IP,
	})

	verify := jwt.SigningMethodHS512.Hash.New().Sum(user.UUID.NodeID())

	tokenString, err := token.SignedString(verify)
	if err != nil {
		return "", "", fmt.Errorf("failed access token signing string: %w", err)
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"sub": user.UUID,
	})

	verifyRef := jwt.SigningMethodHS512.Hash.New().Sum(user.UUID.NodeID())

	refreshTokenString, err := refreshToken.SignedString(verifyRef)
	if err != nil {
		return "", "", fmt.Errorf("failed refresh token signing string: %w", err)
	}

	hashRefreshToken, err := bcrypt.GenerateFromPassword([]byte(user.UUID.String()), 2)
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

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
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

func (s *Service) RefreshUserAuthToken(ref models.Refresh) (string, string, error) {

	verify := jwt.SigningMethodHS512.Hash.New().Sum(ref.User.NodeID())

	keyFunc := func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("refresh auth keyFunc error")
		}
		return verify, nil
	}

	claims := &jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(ref.Access, claims, keyFunc)
	if err != nil {
		return "", "", fmt.Errorf("failed parse acces token: %w", err)
	}

	var tokenID string
	var UUID uuid.UUID

	for key, val := range *claims {
		switch key {
		case "jti":
			tokenID = val.(string)
		case "sub":
			UUID = uuid.MustParse(val.(string))
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	if err := s.Repo.FindAccessTokenByID(ctx, tokenID); err != nil {
		return "", "", fmt.Errorf("failed find access token: %w", err)
	}

	ctx, cancel = context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	hashRef, err := s.Repo.SelectRefreshHashByUUID(ctx, UUID)
	if err != nil {
		return "", "", fmt.Errorf("failed select refresh token: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashRef), []byte(UUID.String())); err != nil {
		return "", "", fmt.Errorf("another refresh token: %w", err)
	}

	accessTokenID := uuid.New()

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"jti": accessTokenID,
		"sub": UUID,
		"ip":  ref.IP,
	})

	verifyNew := jwt.SigningMethodHS512.Hash.New().Sum(ref.User.NodeID())

	tokenString, err := token.SignedString(verifyNew)
	if err != nil {
		return "", "", fmt.Errorf("failed access token signing string: %w", err)
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"sub": UUID,
	})

	verifyRef := jwt.SigningMethodHS512.Hash.New().Sum(UUID.NodeID())

	refreshTokenString, err := refreshToken.SignedString(verifyRef)
	if err != nil {
		return "", "", fmt.Errorf("failed refresh token signing string: %w", err)
	}

	hashRefreshToken, err := bcrypt.GenerateFromPassword([]byte(UUID.String()), 2)
	if err != nil {
		return "", "", fmt.Errorf("hash refresh token err: %w", err)
	}

	if err := s.Repo.UpdateUserTokens(ctx, tokenString, string(hashRefreshToken), UUID); err != nil {
		return "", "", fmt.Errorf("failed update access token: %w", err)
	}

	return tokenString, refreshTokenString, nil
}

// sent to email warning
func (s *Service) SendWarningToEmail(email string, ip string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.Cfg.Email)
	m.SetHeader("To", email)

	m.SetHeader("Subject", "Go-notes auth warning")

	message := fmt.Sprintf(`Hello, an attempt was made to log in to your account from another ip address - %s if it's not you, contact support`, ip)
	m.SetBody("text/plain", message)
	d := gomail.NewDialer("smtp.gmail.com", 587, s.Cfg.Email, "secret code")

	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("send email warning err: %w", err)
	}
	return nil
}
