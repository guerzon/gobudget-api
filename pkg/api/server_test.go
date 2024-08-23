package api

import (
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guerzon/gobudget-api/pkg/db"
	"github.com/guerzon/gobudget-api/pkg/util"
	"github.com/guerzon/gobudget-api/pkg/worker"
	"github.com/stretchr/testify/require"
)

func NewTestServer(t *testing.T, store db.Store, taskDistributor worker.TaskDistributor) *Server {
	config := util.Config{
		SecretKey:           util.RandomString(32, ""),
		AccessTokenDuration: time.Minute * 15,
	}

	testServer, err := NewServer(config, store, taskDistributor)
	require.NoError(t, err)

	return testServer
}

// entrypoint for all unit tests in a Golang package
func TestMain(m *testing.M) {

	gin.SetMode(gin.TestMode)

	os.Exit(m.Run())
}

func buildTestUser(t *testing.T) (db.User, string) {

	plainPassword := "Ch@ngem333Pleaseeeee"
	hashedPassword, err := util.HashPassword(plainPassword)
	require.NoError(t, err)

	charles := db.User{
		Username:      "randomusername",
		Email:         "charlesleclerc@gmail.com",
		EmailVerified: true,
		Password:      hashedPassword,
	}

	return charles, plainPassword
}
