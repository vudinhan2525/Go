package token

import (
	db "main/db/sqlc"
	"time"
)

type Maker interface {
	CreateToken(userID, email string, role db.UserRole, duration time.Duration) (string, *Payload, error)
	VerifyToken(token string) (*Payload, error)
}
