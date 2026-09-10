package server

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestReadinessReportsDependencyFailureWithoutSecrets(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, healthy := range []bool{true, false} {
		r := gin.New()
		r.GET("/ready", readinessHandler(map[string]func(context.Context) error{
			"database": func(ctx context.Context) error {
				_, ok := ctx.Deadline()
				require.True(t, ok)
				if !healthy {
					return errors.New("private DSN")
				}
				return nil
			},
		}))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", "/ready", nil))
		if healthy {
			require.Equal(t, 200, w.Code)
		} else {
			require.Equal(t, 503, w.Code)
		}
		require.NotContains(t, w.Body.String(), "private DSN")
	}
}
