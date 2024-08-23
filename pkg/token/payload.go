package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Payload contains the payload data in the token. The payload ID is used to invalidate a token in case it gets leaked
// Note: Implementes jwt.Claims.
type TokenPayload struct {
	// Unique token identifier
	ID uuid.UUID `json:"id"`
	// Username
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// NewTokenPayload is used to build a claim, adding any useful information defined in the TokenPayload struct. It takes in a username, the token duration, and a slice of roles and adds them to the claim.
func NewTokenPayload(username string, duration time.Duration) (*TokenPayload, error) {

	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	tokenPayload := &TokenPayload{
		tokenID,
		username,
		jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			Audience:  []string{"user"},
		},
	}

	return tokenPayload, nil
}
