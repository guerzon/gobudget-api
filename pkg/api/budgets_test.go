package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guerzon/gobudget-api/pkg/db"
	mock "github.com/guerzon/gobudget-api/pkg/mock"
	"github.com/guerzon/gobudget-api/pkg/util"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreateBudgetAPI(t *testing.T) {

	peso_budget_name := "PH Budget"
	usd_budget_name := "US Investments Budget"

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mock.MockStore, dist *mock.MockTaskDistributor)
		checkResponse func(recoder *httptest.ResponseRecorder)
	}{
		{
			name: "OKPeso",
			body: gin.H{
				"name":          peso_budget_name,
				"currency_code": "PHP",
			},
			buildStubs: func(store *mock.MockStore, dist *mock.MockTaskDistributor) {
				store.EXPECT().
					GetBudgetDetails(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Budget{}, nil)
				store.EXPECT().
					CreateBudget(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Budget{}, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "OKDollar",
			body: gin.H{
				"name":          usd_budget_name,
				"currency_code": "USD",
			},
			buildStubs: func(store *mock.MockStore, dist *mock.MockTaskDistributor) {
				store.EXPECT().
					GetBudgetDetails(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Budget{}, nil)
				store.EXPECT().
					CreateBudget(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Budget{}, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "InvalidCurrency",
			body: gin.H{
				"name":          "My Budget",
				"currency_code": "Dollars",
			},
			buildStubs: func(store *mock.MockStore, dist *mock.MockTaskDistributor) {
				store.EXPECT().
					GetBudgetDetails(gomock.Any(), gomock.Any()).
					Times(0).
					Return(db.Budget{}, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "MissingCurrency",
			body: gin.H{
				"name": "My Budget",
			},
			buildStubs: func(store *mock.MockStore, dist *mock.MockTaskDistributor) {
				store.EXPECT().
					GetBudgetDetails(gomock.Any(), gomock.Any()).
					Times(0).
					Return(db.Budget{}, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
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

			server := NewTestServer(t, store, dist)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/beta/budgets"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)
			// add the header
			token, _, err := server.tokenBuilder.CreateToken(util.RandomUsername(), time.Duration(time.Minute*15))
			require.NoError(t, err)
			authzHeader := "Bearer " + token
			request.Header.Set("Authorization", authzHeader)

			tc.buildStubs(store, dist)
			server.Router.ServeHTTP(recorder, request)
			tc.checkResponse(recorder)
		})
	}
}
