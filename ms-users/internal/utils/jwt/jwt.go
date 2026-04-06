package jwt

import (
	"time"
	"users/internal/utils/config"

	"github.com/golang-jwt/jwt/v5"
)

type Client struct {
	secret          []byte
	expirationHours int
}

func New(cfg config.JwtConfig) *Client {
	return &Client{
		secret:          []byte(cfg.Secret),
		expirationHours: cfg.ExpirationHours,
	}
}

func (c *Client) GenerateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Duration(c.expirationHours) * time.Hour).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(c.secret)
}

func (c *Client) VerifyToken(tokenStr string) (map[string]any, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.MapClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return c.secret, nil
	}, jwt.WithExpirationRequired())
	if err != nil || !token.Valid {
		return nil, err
	}

	claims, ok := token.Claims.(*jwt.MapClaims)
	if !ok {
		return nil, jwt.ErrTokenInvalidClaims
	}

	return map[string]any(*claims), nil
}
