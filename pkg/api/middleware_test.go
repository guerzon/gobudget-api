package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guerzon/gobudget-api/pkg/token"
	"github.com/guerzon/gobudget-api/pkg/util"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware(t *testing.T) {

	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Builder)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Builder) {
				token, _, err := tokenMaker.CreateToken(util.RandomUsername(), time.Duration(time.Minute*15))
				require.NoError(t, err)
				authzHeader := "Bearer " + token
				request.Header.Set("Authorization", authzHeader)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "NoAuthZHeader",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Builder) {
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InvalidHeader",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Builder) {
				token, _, err := tokenMaker.CreateToken(util.RandomUsername(), time.Duration(time.Minute*15))
				require.NoError(t, err)
				authzHeader := token // i.e. "Bearer " is missing
				request.Header.Set("Authorization", authzHeader)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "UnsupportedScheme",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Builder) {
				authzHeader := "Basic cGFzc3dvcmQ=" // Basic authentication with base64 of a password
				request.Header.Set("Authorization", authzHeader)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InvalidToken",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Builder) {
				token, _, err := tokenMaker.CreateToken(util.RandomUsername(), -1) // expired token
				require.NoError(t, err)
				authzHeader := "Bearer " + token
				request.Header.Set("Authorization", authzHeader)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {

			server := NewTestServer(t, nil, nil)
			authPath := "/auth" // dummy path
			server.Router.GET(authPath, AuthMiddleware(server.tokenBuilder), func(ctx *gin.Context) {
				ctx.JSON(http.StatusOK, gin.H{})
			})

			recorder := httptest.NewRecorder()
			request, err := http.NewRequest(http.MethodGet, authPath, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenBuilder)
			server.Router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}
