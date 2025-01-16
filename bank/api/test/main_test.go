package api

import (
	"main/api"
	db "main/db/sqlc"
	"main/util"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
func newTestServer(t *testing.T, store db.Store) *api.Server {
	config := util.Config{
		TokenSymmetricKey: util.RandomStr(32),
		TokenDuration:     time.Minute,
	}
	server, err := api.NewServer(config, store)

	require.NoError(t, err)
	return server
}
