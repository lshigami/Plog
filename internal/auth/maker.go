package auth

import (
	"errors"
	"time"
)

const (
	MinSecretKeySize = 32
)

var (
	ErrInvalidToken   = errors.New("invalid token")
	ErrTokenExpired   = errors.New("token expired")
	ErrInvalidKeySize = errors.New("invalid key size")
)

type Maker interface {
	CreateToken(id int32, username string, duration time.Duration) (string, error)
	VerifyToken(token string) (*Payload, error)
}
