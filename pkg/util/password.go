package util

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Returns the bcrypt hash of plain
func HashPassword(plain string) (string, error) {

	hashBytes, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password %s: %s", plain, err)
	}

	return string(hashBytes), nil
}

// Compares a password hash with a plain string. Returns nil on success, or an error on failure.
func CheckPassword(hashedPassword string, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plain))
}
