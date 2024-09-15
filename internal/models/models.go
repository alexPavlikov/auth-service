package models

import (
	"github.com/google/uuid"
)

type UserPayLoad struct {
	UUID uuid.UUID `json:"uuid"`
}

type User struct {
	UUID uuid.UUID
	IP   string
}

type UserStore struct {
	UUID             uuid.UUID
	Email            string
	AccessTokenID    string
	RefreshTokenHash string
	IPAddress        string
}

type Refresh struct {
	IP      string
	Access  string
	Refresh string
	User    uuid.UUID
}
