package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	mock "github.com/guerzon/gobudget-api/pkg/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreateUserAPI(t *testing.T) {

	user, _ := buildTestUser(t)

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mock.MockStore, dist *mock.MockTaskDistributor)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"username": user.Username,
				"email":    user.Email,
				"password": user.Password,
			},
			buildStubs: func(store *mock.MockStore, dist *mock.MockTaskDistributor) {
				store.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusCreated)
			},
		},
		{
			name: "InvalidEmail",
			body: gin.H{
				"username": "validuser",
				"email":    "charlesatgmaildotcom",
				"password": user.Password,
			},
			buildStubs: func(store *mock.MockStore, dist *mock.MockTaskDistributor) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, recorder.Code, http.StatusBadRequest)
			},
		},
	}

	for i := range testCases {

		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mock.NewMockStore(ctrl)

			dist := mock.NewMockTaskDistributor(ctrl)
			tc.buildStubs(store, dist)

			server := NewTestServer(t, store, dist)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/beta/user"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.Router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}
