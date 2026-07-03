package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Devlaner/devlane/api/internal/handler"
	"github.com/Devlaner/devlane/api/internal/testutil"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestHealth(t *testing.T) {
	ts := testutil.NewTestServer(t)

	rr := ts.GET("/health", "")
	require.Equal(t, http.StatusOK, rr.Code, "body=%s", rr.Body.String())
	body := testutil.MustJSONMap(t, rr)
	assert.Equal(t, "ok", body["status"])
}

func TestReadiness(t *testing.T) {
	ts := testutil.NewTestServer(t)

	rr := ts.GET("/ready", "")
	require.Equal(t, http.StatusOK, rr.Code, "body=%s", rr.Body.String())
	body := testutil.MustJSONMap(t, rr)
	assert.Equal(t, "ready", body["status"])
}

// unusableConnPool implements gorm.ConnPool but is not a *sql.DB, *sql.Tx, or
// gorm.GetDBConnector, so gorm.DB.DB() cannot extract a usable connection
// from it — exactly what happens when the underlying database connection is
// unusable. See gorm.io/gorm@.../gorm.go's DB() method.
type unusableConnPool struct{ gorm.ConnPool }

// TestReadiness_DBUnreachable exercises the readiness handler directly
// against a *gorm.DB whose connection cannot be resolved to a real *sql.DB
// (rather than against the shared test container, which other tests in this
// package depend on staying up), proving the handler's db.DB() error branch
// returns 503 instead of the previous unconditional 200.
func TestReadiness_DBUnreachable(t *testing.T) {
	gdb := &gorm.DB{Config: &gorm.Config{ConnPool: unusableConnPool{}}}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/ready", handler.NewReadinessHandler(gdb))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	r.ServeHTTP(rr, req)

	require.Equal(t, http.StatusServiceUnavailable, rr.Code, "body=%s", rr.Body.String())
	body := testutil.MustJSONMap(t, rr)
	assert.Equal(t, "not_ready", body["status"])
}

func TestUnknownRouteReturns404(t *testing.T) {
	ts := testutil.NewTestServer(t)

	rr := ts.GET("/no/such/path", "")
	require.Equal(t, http.StatusNotFound, rr.Code)
}
