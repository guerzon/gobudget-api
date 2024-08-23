package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/guerzon/gobudget-api/pkg/db"
	mockdb "github.com/guerzon/gobudget-api/pkg/mock"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestLoginAPI(t *testing.T) {

	user, plainPassword := buildTestUser(t)
	user2 := user
	user2.EmailVerified = false

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore, dist *mockdb.MockTaskDistributor)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"username": user.Username,
				"password": plainPassword,
			},
			buildStubs: func(store *mockdb.MockStore, dist *mockdb.MockTaskDistributor) {
				store.EXPECT().
					GetUserByUsername(gomock.Any(), user.Username).
					Times(1).
					Return(user, nil)
				store.EXPECT().
					CreateSession(gomock.Any(), gomock.Any()).
					Times(1)

			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "EmailNotFound",
			body: gin.H{
				"username": "randomusername",
				"password": user.Password,
			},
			buildStubs: func(store *mockdb.MockStore, dist *mockdb.MockTaskDistributor) {
				store.EXPECT().
					GetUserByUsername(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, pgx.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "IncorrectPassword",
			body: gin.H{
				"username": user.Username,
				"password": "incorrectPassword",
			},
			buildStubs: func(store *mockdb.MockStore, dist *mockdb.MockTaskDistributor) {
				store.EXPECT().
					GetUserByUsername(gomock.Any(), gomock.Eq(user.Username)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "InvalidUsername",
			body: gin.H{
				"username": "1",
				"password": user.Password,
			},
			buildStubs: func(store *mockdb.MockStore, dist *mockdb.MockTaskDistributor) {
				store.EXPECT().
					GetUserByUsername(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "EmailNotVerified",
			body: gin.H{
				"username": user.Username,
				"password": plainPassword,
			},
			buildStubs: func(store *mockdb.MockStore, dist *mockdb.MockTaskDistributor) {
				store.EXPECT().
					GetUserByUsername(gomock.Any(), user2.Username).
					Times(1).
					Return(user2, nil)
				store.EXPECT().
					GetPendingVerifyEmails(gomock.Any(), gomock.Any()).
					Times(1).
					Return([]db.VerifyEmail{}, nil)
				dist.EXPECT().
					DistributeSendEmail(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1)
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
			dist := mockdb.NewMockTaskDistributor(ctrl)
			tc.buildStubs(store, dist)

			server := NewTestServer(t, store, dist)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/beta/login"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.Router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}
