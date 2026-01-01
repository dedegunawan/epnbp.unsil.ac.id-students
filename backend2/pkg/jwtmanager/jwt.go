package jwtmanager

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Manager struct {
	secret    []byte
	issuer    string
	expires   time.Duration
	clockSkew time.Duration
}

type Claims struct {
	UserID uint64 `json:"uid"`
	Email  string `json:"email,omitempty"`
	Name   string `json:"name,omitempty"`
	jwt.RegisteredClaims
}

func New(secret, issuer string, expires time.Duration) *Manager {
	return &Manager{
		secret:    []byte(secret),
		issuer:    issuer,
		expires:   expires,
		clockSkew: 30 * time.Second, // leeway aman
	}
}

func (m *Manager) Generate(userID uint64, email, name string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID: userID,
		Email:  email,
		Name:   name,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   strconv.FormatUint(userID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now.Add(-m.clockSkew)),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.expires)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *Manager) Validate(tokenString string) (*Claims, error) {
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
		jwt.WithIssuedAt(),
		jwt.WithLeeway(m.clockSkew),
	)
	token, err := parser.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid claims type")
	}
	// issuer check (opsional tapi disarankan)
	if claims.Issuer != m.issuer {
		return nil, errors.New("invalid issuer")
	}
	return claims, nil
}
