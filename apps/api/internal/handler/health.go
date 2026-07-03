package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// readinessPingTimeout bounds the DB ping so a hung connection can't block
// the probe indefinitely regardless of the caller's own context deadline.
const readinessPingTimeout = 3 * time.Second

// Health is the liveness probe.
// GET /health
func Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// NewReadinessHandler returns the readiness probe. It pings the database (the
// one hard dependency) and returns 503 if it's unreachable, so an
// orchestrator holds the instance out of rotation instead of routing traffic
// to it. Optional infra (Redis/RabbitMQ/MinIO) is intentionally excluded per
// the graceful-degradation convention.
// GET /ready
func NewReadinessHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), readinessPingTimeout)
		defer cancel()

		sqlDB, err := db.DB()
		if err != nil || sqlDB.PingContext(ctx) != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	}
}
