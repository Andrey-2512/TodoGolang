package security

import (
	"errors"
	"fmt"
	"time"
	"todo/domain/apperrors"
	"todo/domain/entity"
	"uuid"

	"github.com/golang-jwt/jwt/v5"
)

type userClaims struct {
	UserId    int    `json:"user_id"`
	Username  string `json:"username"`
	TokenType string `json:"token_type"`
	JTI       string `json:"jti"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secretKey  string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewJWTManager(secretKey string, accessTTL, refreshTTL time.Duration) *JWTManager {
	return &JWTManager{secretKey: secretKey, accessTTL: accessTTL, refreshTTL: refreshTTL}
}

func (m *JWTManager) createToken(userId int, username string, duration time.Duration, tokenType string) (string, error) {
	const op = "security.JWTManager.createToken"
	claims := userClaims{Username: username, UserId: userId, TokenType: tokenType, JTI: uuid.New().String(), ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration))}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(m.secretKey))
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return token, nil
}

func (m *JWTManager) parseToken(jwtToken string, expectedType string) (*entity.UserClaims, error) {
	const op = "security.JWTManager.parseToken"
	token, err := jwt.ParseWithClaims(jwtToken, &userClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%s: %w", op, apperrors.ErrInvalidToken)
		}

		return []byte(m.secretKey), nil
	}, jwt.WithLeeway(3*time.Second), jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, fmt.Errorf("%s: %w", op, apperrors.ErrSessionExpired)
		}
		return nil, fmt.Errorf("%s: %w", op, apperrors.ErrInvalidToken)
	}

	claims, ok := token.Claims.(*userClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("%s: %w", op, apperrors.ErrInvalidToken)
	}

	if claims.TokenType != expectedType {
		return nil, fmt.Errorf("%s: %w", op, apperrors.ErrInvalidTokenType)
	}

	return &entity.UserClaims{UserId: claims.UserId, Username: claims.Username, TokenType: claims.TokenType, ExpiresAt: claims.ExpiresAt.Time, JTI: claims.JTI}, nil
}

func (m *JWTManager) CreateAccessToken(userId int, username string) (string, error) {
	const op = "security.JWTManager.CreateAccessToken"
	access, err := m.createToken(userId, username, m.accessTTL, "access")
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return access, nil
}

func (m *JWTManager) ParseAccessToken(jwtToken string) (*entity.UserClaims, error) {
	const op = "security.JWTManager.ParseAccessToken"
	claims, err := m.parseToken(jwtToken, "access")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return claims, nil
}

func (m *JWTManager) CreateRefreshToken(userId int, username string) (string, error) {
	const op = "security.JWTManager.CreateRefreshToken"
	refresh, err := m.createToken(userId, username, m.refreshTTL, "refresh")
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return refresh, nil
}

func (m *JWTManager) ParseRefreshToken(jwtToken string) (*entity.UserClaims, error) {
	const op = "security.JWTManager.ParseRefreshToken"
	claims, err := m.parseToken(jwtToken, "refresh")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return claims, nil
}
