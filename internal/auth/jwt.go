package auth

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

type JWTAuth interface {
	GenerateToken(jwt.Claims) (string, error)
	ValidateToken(string) (*jwt.Token, error)
}

type Authenticator struct {
	secret string
	aud    string
	iss    string
}

func NewAuthenticator(secret, aud, iss string) *Authenticator {
	return &Authenticator{
		secret: secret,
		aud:    aud,
		iss:    iss,
	}
}

func (auth *Authenticator) GenerateToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(auth.secret))

	if err != nil {
		return "", nil
	}
	return tokenString, nil
}

func (auth *Authenticator) ValidateToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}

		return []byte(auth.secret), nil
	},
		jwt.WithAudience(auth.aud),
		jwt.WithExpirationRequired(),
		jwt.WithIssuer(auth.iss),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)
}
