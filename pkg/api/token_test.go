package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/guerzon/gobudget-api/pkg/db"
	mockdb "github.com/guerzon/gobudget-api/pkg/mock"
	"github.com/guerzon/gobudget-api/pkg/token"
	"github.com/guerzon/gobudget-api/pkg/util"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestToken(t *testing.T) {

	testCases := []struct {
		name          string
		buildStubs    func(store *mockdb.MockStore, refreshTokenClaimsID uuid.UUID)
		setupAuth     func(tokenMaker token.Builder) (string, *token.TokenPayload)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			buildStubs: func(store *mockdb.MockStore, refreshTokenClaimsID uuid.UUID) {
				store.EXPECT().
					GetSession(gomock.Any(), refreshTokenClaimsID).
					Times(1).
					Return(db.Session{}, nil)
			},
			setupAuth: func(tokenMaker token.Builder) (string, *token.TokenPayload) {
				refreshToken, refreshTokenClaims, err := tokenMaker.CreateToken(util.RandomUsername(), time.Duration(time.Minute*15))
				require.NoError(t, err)
				return refreshToken, refreshTokenClaims
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "ExpiredToken",
			buildStubs: func(store *mockdb.MockStore, refreshTokenClaimsID uuid.UUID) {
				store.EXPECT().
					GetSession(gomock.Any(), refreshTokenClaimsID).
					Times(0).
					Return(db.Session{}, nil)
			},
			setupAuth: func(tokenMaker token.Builder) (string, *token.TokenPayload) {
				refreshToken, refreshTokenClaims, err := tokenMaker.CreateToken(util.RandomUsername(), -1)
				require.NoError(t, err)
				return refreshToken, refreshTokenClaims
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for i := range testCases {

		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store := mockdb.NewMockStore(ctrl)

			server := NewTestServer(t, store, nil)

			// get a refresh token
			refreshToken, refreshTokenClaims := tc.setupAuth(server.tokenBuilder)
			body := gin.H{
				"refresh_token": refreshToken,
			}
			data, err := json.Marshal(body)
			require.NoError(t, err)

			tc.buildStubs(store, refreshTokenClaims.ID)

			recorder := httptest.NewRecorder()
			request, err := http.NewRequest(http.MethodPost, "/beta/renew_token", bytes.NewReader(data))
			require.NoError(t, err)

			server.Router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}
