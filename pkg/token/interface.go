package token

import "time"

// This is the Token maker interface, to make it easier to switch
// between JWT and PASETO if I decide to use it in the future
type Builder interface {
	CreateToken(username string, duration time.Duration) (string, *TokenPayload, error)
	VerifyToken(token string) (*TokenPayload, error)
}
