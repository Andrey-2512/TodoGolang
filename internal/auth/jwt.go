package auth

import (
	"errors"
	"time"
	"todo/domain/apperrors"
	"todo/domain/entity"

	"github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	UserId    int    `json:"user_id"`
	Username  string `json:"username"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

type JWTManager interface {
	CreateAccessToken(user entity.UserPayload) (string, error)
	ParseAccessToken(jwtToken string) (*UserClaims, error)
	VerifyAccessToken(jwtToken string) (bool, error)
	VerifyRefreshToken(jwtToken string) (bool, error)
	CreateRefreshToken(user entity.UserPayload) (string, error)
	ParseRefreshToken(jwtToken string) (*UserClaims, error)
}

type jwtManager struct {
	secretKey  string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewJWTManager(secretKey string, accessTTL time.Duration, refreshTTL time.Duration) JWTManager {
	return &jwtManager{secretKey: secretKey, accessTTL: accessTTL, refreshTTL: refreshTTL}
}

func (jwtManager *jwtManager) createToken(user entity.UserPayload, duration time.Duration, tokenType string) (string, error) {
	claims := UserClaims{Username: user.Username, UserId: user.UserID, TokenType: tokenType, RegisteredClaims: jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
	}}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(jwtManager.secretKey))
}

func (jwtManager *jwtManager) parseToken(jwtToken string, exceptedType string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(jwtToken, &UserClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, apperrors.ErrInvalidToken
		}

		return []byte(jwtManager.secretKey), nil
	}, jwt.WithLeeway(3*time.Second), jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, apperrors.ErrSessionExpired
		}
		return nil, apperrors.ErrInvalidToken
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		return nil, apperrors.ErrInvalidToken
	}

	if claims.TokenType != exceptedType {
		return nil, apperrors.ErrInvalidTokenType
	}

	return claims, nil
}

func (jwtManager *jwtManager) VerifyAccessToken(jwtToken string) (bool, error) {
	_, err := jwtManager.parseToken(jwtToken, "access")
	if err != nil {
		return false, err
	}
	return true, nil
}

func (jwtManager *jwtManager) VerifyRefreshToken(jwtToken string) (bool, error) {
	_, err := jwtManager.parseToken(jwtToken, "refresh")
	if err != nil {
		return false, err
	}
	return true, nil
}

func (jwtManager *jwtManager) CreateAccessToken(user entity.UserPayload) (string, error) {

	return jwtManager.createToken(user, jwtManager.accessTTL, "access")
}

func (jwtManager *jwtManager) ParseAccessToken(jwtToken string) (*UserClaims, error) {
	return jwtManager.parseToken(jwtToken, "access")
}

func (jwtManager *jwtManager) CreateRefreshToken(user entity.UserPayload) (string, error) {
	return jwtManager.createToken(user, jwtManager.refreshTTL, "refresh")
}

func (jwtManager *jwtManager) ParseRefreshToken(jwtToken string) (*UserClaims, error) {
	return jwtManager.parseToken(jwtToken, "refresh")
}
