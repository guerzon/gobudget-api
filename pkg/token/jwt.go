package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const minSecretKeySize = 32

// JWTBuilder is a JSON web token maker which implements the Maker interface.
type JWTBuilder struct {
	secretKey string
}

// Creates a new JWTBuilder.
func NewJWTBuilder(secretKey string) (Builder, error) {

	if len(secretKey) < minSecretKeySize {
		return nil, fmt.Errorf("invalid secret key size: must be at least %d characters", minSecretKeySize)
	}

	return &JWTBuilder{
		secretKey: secretKey,
	}, nil
}

// Create a token using the symmetric signing algorithm HS256 (HMAC + SHA256). Returns the signed token string, the payload used, and possibly an error.
func (j *JWTBuilder) CreateToken(username string, duration time.Duration) (string, *TokenPayload, error) {

	// Create the payload to in include in the token
	claims, err := NewTokenPayload(username, duration)
	if err != nil {
		return "", nil, fmt.Errorf("cannot create a token: %s", err)
	}

	// Create the token
	unsignedToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with our secret key
	signedToken, err := unsignedToken.SignedString([]byte(j.secretKey))
	if err != nil {
		return "", nil, fmt.Errorf("cannot sign token: %s", err)
	}

	return signedToken, claims, nil

}

// VerifyToken checks if the token is valid or not
func (j *JWTBuilder) VerifyToken(token string) (*TokenPayload, error) {

	// keyfunc is a function that receives a parsed but unverified token
	// should verify the token's header to make sure that the signing alg
	// matches the alg used to sign tokens. If so, return the key so can be used
	// to verify the token
	keyFunc := func(t *jwt.Token) (interface{}, error) {
		// try to convert it to a specific implementation (jwt.SigningMethodHMAC)
		// bc we're using SigningMethodHS256, which is an instance of SigningMethodHMAC struct
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenUnverifiable
		}
		// then return the secret key
		return []byte(j.secretKey), nil
	}

	parsedToken, err := jwt.ParseWithClaims(token, &TokenPayload{}, keyFunc)
	if err != nil {
		return nil, err
	}

	// convert the token into a Payload struct
	payload, ok := parsedToken.Claims.(*TokenPayload)
	if !ok {
		return nil, fmt.Errorf("invalid token: cannot convert payload")
	}

	return payload, nil
}
