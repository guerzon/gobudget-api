package util

import (
	"math/rand"
	"strings"
	"time"
)

const letters = "abcdefghijklmnopqrstuvwzy"
const specialCharacters = `!@#$%^&*();.`
const numbers = "1234567890"

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

// Return a random number between max and min
func RandomNumber(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

// Return a random string
func RandomString(n int, src string) string {

	if src == "" {
		src = letters + numbers
	}

	var sb strings.Builder
	l := len(src)

	for i := 0; i < n; i++ {
		c := src[rand.Intn(l)]
		sb.WriteByte(c)
	}

	return sb.String()
}

// Generate a random username with length 10
func RandomUsername() string {
	return RandomString(10, letters)
}

// Generate a random password with length 20
func RandomPassword() string {
	return RandomString(20, letters+specialCharacters+numbers)
}
