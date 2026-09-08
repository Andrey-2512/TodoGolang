package entity

import "time"

type UserClaims struct {
	UserId    int
	Username  string
	TokenType string
	JTI       string
	ExpiresAt time.Time
}
