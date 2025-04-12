package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Payload struct {
	ID        int32     `json:"id"`
	Username  string    `json:"username"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expired_at"`
}

func NewPayload(id int32, username string, duration time.Duration) (*Payload, error) {

	return &Payload{
		ID:        id,
		Username:  username,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(duration),
	}, nil

}

func (p *Payload) Valid() error {
	if time.Now().After(p.ExpiresAt) {
		return ErrTokenExpired
	}

	return nil
}

// GetAudience implements jwt.Claims.
func (p *Payload) GetAudience() (jwt.ClaimStrings, error) {
	return nil, nil
}

// GetExpirationTime implements jwt.Claims.
func (p *Payload) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(p.ExpiresAt), nil
}

// GetIssuedAt implements jwt.Claims.
func (p *Payload) GetIssuedAt() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(p.IssuedAt), nil
}

// GetIssuer implements jwt.Claims.
func (p *Payload) GetIssuer() (string, error) {
	return "banker", nil
}

// GetNotBefore implements jwt.Claims.
func (p *Payload) GetNotBefore() (*jwt.NumericDate, error) {
	return nil, nil
}

// GetSubject implements jwt.Claims.
func (p *Payload) GetSubject() (string, error) {
	return p.Username, nil
}
