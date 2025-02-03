package db

import (
	"context"
	"main/util"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func createTestUser(t *testing.T) User {
	password, err := util.HashPassword(util.RandomStr(12))
	require.NoError(t, err)
	arg := CreateUserParams{
		HashedPassword: password,
		FullName:       util.RandomStr(5),
		Email:          util.RandomEmail(),
		Role:           "user",
	}
	user, err := testQueries.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.FullName, user.FullName)
	require.Equal(t, arg.Email, user.Email)
	require.Equal(t, arg.HashedPassword, user.HashedPassword)
	require.NotZero(t, user.UserID)
	require.NotZero(t, user.CreatedAt)
	require.True(t, user.PasswordChangedAt.IsZero())

	return user
}
func TestCreateUser(t *testing.T) {
	createTestUser(t)
}
func TestGetUser(t *testing.T) {
	user := createTestUser(t)

	user2, err := testQueries.GetUser(context.Background(), user.UserID)
	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user.UserID, user2.UserID)
	require.Equal(t, user.FullName, user2.FullName)
	require.Equal(t, user.Email, user2.Email)
	require.Equal(t, user.HashedPassword, user2.HashedPassword)
	require.WithinDuration(t, user.CreatedAt, user2.CreatedAt, time.Second)
	require.WithinDuration(t, user.PasswordChangedAt, user2.PasswordChangedAt, time.Second)
}
