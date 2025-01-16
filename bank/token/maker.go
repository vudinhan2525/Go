package token

import "time"

type Maker interface {
	CreateToken(userID, email string, duration time.Duration) (string, *Payload, error)
	VerifyToken(token string) (*Payload, error)
}
