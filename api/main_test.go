package api

import (
	"os"
	"testing"
	"time"

	db "github.com/bagashiz/Simple-Bank/db/sqlc"
	"github.com/bagashiz/Simple-Bank/util"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func NewTestServer(t *testing.T, store db.Store) *Server {
  config := util.Config{
    TokenSymmetricKey: util.RandomString(32),
    AccessTokenDuration: time.Minute,
  }

  server, err := NewServer(config, store)
  require.NoError(t, err)

  return server
}

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
