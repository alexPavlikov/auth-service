package models

import (
	"time"

	"github.com/google/uuid"
)

type UserPayLoad struct {
	ID       uuid.UUID `mapstructure:"id"`
	Login    string    `mapstructure:"login"`
	Password string    `mapstructure:"password"`
	IP       string    `mapstructure:"ip_address"`
}

type User struct {
	ID            uuid.UUID
	Login         string
	PassHash      string
	Email         string
	Firstname     string
	Lastname      string
	LastIPAddress string

	Create   time.Time
	LastAuth time.Time
}

type Token struct{}
