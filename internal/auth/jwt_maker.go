package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTMaker struct {
	secretKey string
}

func NewJWTMaker(secretKey string) Maker {
	if len(secretKey) < MinSecretKeySize {
		panic("secret key must be at least 32 characters")
	}
	return &JWTMaker{secretKey}
}

func (maker *JWTMaker) CreateToken(id int32, username string, duration time.Duration) (string, error) {

	payload, err := NewPayload(id, username, duration)
	if err != nil {
		return "", err
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, payload).SignedString([]byte(maker.secretKey))
	if err != nil {
		return "", err
	}
	return token, nil

}
func (maker *JWTMaker) VerifyToken(token string) (*Payload, error) {
	parsedToken, err := jwt.ParseWithClaims(token, &Payload{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(maker.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	payload, ok := parsedToken.Claims.(*Payload)
	if !ok {
		return nil, ErrInvalidToken
	}

	if err := payload.Valid(); err != nil {
		return nil, err
	}

	return payload, nil
}
