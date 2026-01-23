package token

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/Testzyler/go-microservice/global/pkg/auth"
)

type Service interface {
	IssueAccessToken(ctx context.Context, claims auth.Claims) (string, error)
	ParseAccessToken(ctx context.Context, token string) (auth.Claims, error)
}

type JWTService struct {
	secret   []byte
	issuer   string
	audience string
	ttl      time.Duration
}

func NewJWTService(secret, issuer, audience string, ttl time.Duration) *JWTService {
	return &JWTService{
		secret:   []byte(secret),
		issuer:   issuer,
		audience: audience,
		ttl:      ttl,
	}
}

type accessClaims struct {
	Roles       []string `json:"roles,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
	jwt.RegisteredClaims
}

func (s *JWTService) IssueAccessToken(_ context.Context, claims auth.Claims) (string, error) {
	if claims.Subject == uuid.Nil {
		return "", errors.New("missing subject")
	}
	now := time.Now().UTC()
	registered := jwt.RegisteredClaims{
		Subject:   claims.Subject.String(),
		Issuer:    s.issuer,
		Audience:  []string{s.audience},
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
	}
	issued := accessClaims{
		Roles:            append([]string(nil), claims.Roles...),
		Permissions:      append([]string(nil), claims.Permissions...),
		RegisteredClaims: registered,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, issued)
	return token.SignedString(s.secret)
}

func (s *JWTService) ParseAccessToken(_ context.Context, tokenString string) (auth.Claims, error) {
	parsed, err := jwt.ParseWithClaims(tokenString, &accessClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return s.secret, nil
	}, jwt.WithAudience(s.audience), jwt.WithIssuer(s.issuer))
	if err != nil {
		return auth.Claims{}, err
	}
	claims, ok := parsed.Claims.(*accessClaims)
	if !ok || !parsed.Valid {
		return auth.Claims{}, errors.New("invalid token")
	}
	subject, err := uuid.Parse(claims.Subject)
	if err != nil {
		return auth.Claims{}, err
	}
	return auth.Claims{
		Subject:     subject,
		Roles:       append([]string(nil), claims.Roles...),
		Permissions: append([]string(nil), claims.Permissions...),
	}, nil
}
