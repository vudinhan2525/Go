package token

import (
	"errors"
	db "main/db/sqlc"
	"strconv"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("token is invalid")
	ErrExpiredToken = errors.New("token has expired")
)

type Payload struct {
	ID        uuid.UUID   `json:"id"`
	UserID    int         `json:"user_id"`
	Email     string      `json:"email"`
	Role      db.UserRole `json:"role"`
	IssuedAt  time.Time   `json:"issued_at"`
	ExpiredAt time.Time   `json:"expired_at"`
}

func NewPayload(userID, email string, role db.UserRole, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	id, err := strconv.Atoi(userID)
	if err != nil {
		panic(err)
	}

	payload := &Payload{
		ID:        tokenID,
		UserID:    id,
		Email:     email,
		Role:      role,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}
	return payload, nil
}
func (payload *Payload) Valid() error {
	if time.Now().After(payload.ExpiredAt) {
		return ErrExpiredToken
	}
	return nil
}
