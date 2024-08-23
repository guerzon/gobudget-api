package token

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/guerzon/gobudget-api/pkg/util"
	"github.com/stretchr/testify/require"
)

func TestJWTBuilder(t *testing.T) {

	maker, err := NewJWTBuilder(util.RandomString(32, ""))
	require.NoError(t, err)

	// these are for verification purposes against the created token:
	issuedAt := time.Now()
	expiredAt := issuedAt.Add(time.Hour * 24)

	username := "tst" + util.RandomUsername()
	token, payload, err := maker.CreateToken(username, time.Duration(time.Hour*24))
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, payload)

	payload, err = maker.VerifyToken(token)
	require.NoError(t, err)
	require.NotEmpty(t, payload)
	// inspect the fields
	require.NotZero(t, payload.ID)
	require.Equal(t, username, payload.Username)
	require.WithinDuration(t, issuedAt, payload.IssuedAt.Time, time.Second)   // issuance should happen within 1 second
	require.WithinDuration(t, expiredAt, payload.ExpiresAt.Time, time.Second) // expiration should be within 1 second of diff
}

func TestExpiredToken(t *testing.T) {

	maker, err := NewJWTBuilder(util.RandomString(32, ""))
	require.NoError(t, err)

	username := "tst" + util.RandomUsername()
	token, payload, err := maker.CreateToken(username, -1)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, payload)

	payload, err = maker.VerifyToken(token)
	require.Error(t, err)
	require.ErrorIs(t, err, jwt.ErrTokenExpired)
	require.Nil(t, payload)
}

func TestInvalidTokenAlgNone(t *testing.T) {

	// we build our token
	claims, _ := NewTokenPayload(util.RandomString(32, ""), time.Hour*24)
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	signedToken, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	maker, err := NewJWTBuilder(util.RandomString(32, ""))
	require.NoError(t, err)

	// now, we verify the payload
	claims, err = maker.VerifyToken(signedToken)
	require.Error(t, err)
	require.ErrorIs(t, err, jwt.ErrTokenUnverifiable)
	require.Nil(t, claims)
}
