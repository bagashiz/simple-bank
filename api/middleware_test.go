package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bagashiz/Simple-Bank/token"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func addAuthorization(
  t *testing.T,
  request *http.Request,
  tokenMaker token.Maker,
  authorizationType string,
  username string,
  duration time.Duration,
) {
  token, err := tokenMaker.CreateToken(username, duration)
  require.NoError(t, err)

  authorizationHeader := fmt.Sprintf("%s %s", authorizationType, token)
  request.Header.Set(authorizationHeaderKey, authorizationHeader)
}

func TestAuthMiddleware(t *testing.T) {
  testCases := []struct{
    name string
    setupAuth func(t *testing.T, request *http.Request, tokenMaker token.Maker)
    checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
  }{
    // case 1: authorized successfully
    {
      name: "OK",
      setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
        addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "user", time.Minute)
      },
      checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
        require.Equal(t, http.StatusOK, recorder.Code)
      },
    },

    // case 2: no authorization header
    {
      name: "NoAuthorization",
      setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
      },
      checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
        require.Equal(t, http.StatusUnauthorized, recorder.Code)
      },
    },
    
    // case 3: unsupported authorization
    {
      name: "UnsupportedAuthorization",
      setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
          addAuthorization(t, request, tokenMaker, "unsupported", "user", time.Minute)
      },
      checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
        require.Equal(t, http.StatusUnauthorized, recorder.Code)
      },
    },
      
    // case 4: invalid authorization format
    {
      name: "InvalidAuthorizationFormat",
      setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
          addAuthorization(t, request, tokenMaker, "", "user", time.Minute)
      },
      checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
        require.Equal(t, http.StatusUnauthorized, recorder.Code)
      },
    },
      
    // case 5: expired access token
    {
      name: "ExpiredToken",
      setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
          addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "user", -time.Minute)
      },
      checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
        require.Equal(t, http.StatusUnauthorized, recorder.Code)
      },
    },
  }

  for i := range testCases {
    tc := testCases[i]

    t.Run(tc.name, func(t *testing.T){
      server := NewTestServer(t, nil)
      
      authPath := "/auth"
      server.router.GET(
        authPath,
        authMiddleware(server.tokenMaker),
        func(ctx *gin.Context) {
          ctx.JSON(http.StatusOK, gin.H{})
        },
      )

      recorder := httptest.NewRecorder()
      request, err := http.NewRequest(http.MethodGet, authPath, nil)
      require.NoError(t, err)

      tc.setupAuth(t, request, server.tokenMaker)
      server.router.ServeHTTP(recorder, request)
      tc.checkResponse(t, recorder)
    })
  }
}