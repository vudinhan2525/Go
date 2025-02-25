package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	mockdb "main/db/mock"
	db "main/db/sqlc"
	"main/pkg/middlewares"
	"main/token"
	"main/util"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreateUserAPI(t *testing.T) {
	user, password := RandomUser(t)
	testCases := []struct {
		name          string
		body          gin.H
		setupAuth     func(t *testing.T, req *http.Request, tokenMaker token.Maker)
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "Happy Case",
			body: gin.H{
				"email":    user.Email,
				"fullName": user.FullName,
				"password": password,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				AddTokenHeader(t, req, user.UserID, user.Email, time.Minute, tokenMaker, middlewares.AuthorizationType)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Times(1).Return(user, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchUser(t, recorder.Body, user)
			},
		},
		{
			name: "Bad Request",
			body: gin.H{
				"email":    user.Email,
				"fullName": user.FullName,
				"password": "0",
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				AddTokenHeader(t, req, user.UserID, user.Email, time.Minute, tokenMaker, middlewares.AuthorizationType)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)

			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"password": password,
				"fullName": user.FullName,
				"email":    user.Email,
			},
			setupAuth: func(t *testing.T, req *http.Request, tokenMaker token.Maker) {
				AddTokenHeader(t, req, user.UserID, user.Email, time.Minute, tokenMaker, middlewares.AuthorizationType)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := mockdb.NewMockStore(ctrl)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.buildStubs(store)
			server := newTestServer(t, store)
			recorder := httptest.NewRecorder()
			// Marshal body data to JSON
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, "/users", bytes.NewReader(data))

			require.NoError(t, err)
			tc.setupAuth(t, request, server.TokenMaker)
			server.Router.ServeHTTP(recorder, request)

			tc.checkResponse(t, recorder)
		})

	}
}
func RandomUser(t *testing.T) (db.User, string) {
	password := util.RandomStr(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)
	return db.User{
		UserID:         util.RandomInt(0, 100000),
		Email:          util.RandomEmail(),
		FullName:       util.RandomStr(6),
		HashedPassword: hashedPassword,
	}, password
}
func requireBodyMatchUser(t *testing.T, body *bytes.Buffer, user db.User) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	type response struct {
		Status string  `json:"status"`
		Data   db.User `json:"data"`
	}
	var gotResponse response
	err = json.Unmarshal(data, &gotResponse)
	require.NoError(t, err)

	gotUser := gotResponse.Data

	require.Equal(t, user, gotUser)
}
