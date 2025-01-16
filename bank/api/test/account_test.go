package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"time"

	mockdb "main/db/mock"
	db "main/db/sqlc"
	"main/pkg/middlewares"
	"main/token"
	"main/util"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func AddTokenHeader(t *testing.T, req *http.Request, userID int64, email string, duration time.Duration, tokenMaker token.Maker, authorizationType string) {
	userId := fmt.Sprintf("%v", userID)
	token, _, err := tokenMaker.CreateToken(userId, email, duration)
	require.NoError(t, err)

	authorizationHeader := fmt.Sprintf("%s %s", authorizationType, token)
	req.Header.Set(middlewares.AuthorizationHeaderKey, authorizationHeader)
}
func TestGetAccountAPI(t *testing.T) {
	user, _ := RandomUser(t)
	account := RandomAccount(user.UserID)
	testCases := []struct {
		name          string
		accountId     int64
		setupAuth     func(t *testing.T, req *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "Happy Case",
			accountId: account.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				AddTokenHeader(t, req, user.UserID, user.Email, time.Minute, tokenMaker, middlewares.AuthorizationType)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), account.ID).Times(1).Return(account, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				checkBodyResponse(t, recorder.Body, account)
			},
		},
		{
			name:      "Bad Request",
			accountId: 0,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				AddTokenHeader(t, req, user.UserID, user.Email, time.Minute, tokenMaker, middlewares.AuthorizationType)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), 0).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)

			},
		},
		{
			name:      "Not Found",
			accountId: account.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				AddTokenHeader(t, req, user.UserID, user.Email, time.Minute, tokenMaker, middlewares.AuthorizationType)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), account.ID).Times(1).Return(db.Account{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "Internal Server Error",
			accountId: account.ID,
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				AddTokenHeader(t, req, user.UserID, user.Email, time.Minute, tokenMaker, middlewares.AuthorizationType)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), account.ID).Times(1).Return(db.Account{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)
			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d", tc.accountId)
			request, err := http.NewRequest(http.MethodGet, url, nil)

			require.NoError(t, err)

			tc.setupAuth(t, request, server.TokenMaker)

			server.Router.ServeHTTP(recorder, request)

			tc.checkResponse(t, recorder)
		})

	}
}

func RandomAccount(ownerId int64) db.Account {
	return db.Account{
		ID:       util.RandomInt(0, 1000),
		Owner:    ownerId,
		Balance:  util.RandomMoney(),
		Currency: util.RandomCurrency(),
	}
}

func checkBodyResponse(t *testing.T, body *bytes.Buffer, account db.Account) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	type response struct {
		Status string     `json:"status"`
		Data   db.Account `json:"data"`
	}
	var gotResponse response
	err = json.Unmarshal(data, &gotResponse)
	require.NoError(t, err)

	gotAccount := gotResponse.Data

	require.Equal(t, account, gotAccount)
}
