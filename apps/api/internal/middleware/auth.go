package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/Devlaner/devlane/api/internal/auth"
	"github.com/Devlaner/devlane/api/internal/model"
	"github.com/Devlaner/devlane/api/internal/store"
	"github.com/gin-gonic/gin"
)

const (
	SessionCookieName = "session_id"
	UserContextKey    = "user"
)

// SessionKeyFromCookieOrBearer returns the session id from the session cookie or Authorization: Bearer.
// Must stay in sync with how authenticated clients send the session (including OAuth SPA fragment flow).
func SessionKeyFromCookieOrBearer(c *gin.Context) string {
	sessionKey, _ := c.Cookie(SessionCookieName)
	if sessionKey == "" {
		if authHeader := c.GetHeader("Authorization"); len(authHeader) > 7 && strings.EqualFold(authHeader[:7], "bearer ") {
			sessionKey = strings.TrimSpace(authHeader[7:])
		}
	}
	return sessionKey
}

// RequireAuth loads the user from a session cookie or Authorization: Bearer
// header, and returns 401 if not authenticated. Bearer values are tried
// first as an API token (hashed + looked up in api_tokens); if that doesn't
// match, they fall back to being treated as a raw session key — kept for
// the cross-origin OAuth SPA fragment flow (see SessionKeyFromCookieOrBearer).
func RequireAuth(authSvc *auth.Service, apiTokens *store.ApiTokenStore, log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		if cookieKey, _ := c.Cookie(SessionCookieName); cookieKey != "" {
			if user, err := authSvc.UserFromSession(ctx, cookieKey); err == nil && user != nil {
				c.Set(UserContextKey, user)
				c.Next()
				return
			}
		} else if authHeader := c.GetHeader("Authorization"); len(authHeader) > 7 && strings.EqualFold(authHeader[:7], "bearer ") {
			bearer := strings.TrimSpace(authHeader[7:])
			if apiTokens != nil {
				if tok, err := apiTokens.GetActiveByHash(ctx, store.HashToken(bearer)); err == nil && tok != nil {
					if user, err := authSvc.ActiveUserByID(ctx, tok.UserID); err == nil && user != nil {
						_ = apiTokens.UpdateLastUsed(ctx, tok.ID)
						c.Set(UserContextKey, user)
						c.Next()
						return
					}
				}
			}
			if user, err := authSvc.UserFromSession(ctx, bearer); err == nil && user != nil {
				c.Set(UserContextKey, user)
				c.Next()
				return
			}
		}

		if log != nil {
			log.Debug("auth required", "has_session_key", SessionKeyFromCookieOrBearer(c) != "")
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
	}
}

// GetUser returns the authenticated user from context (must be used after RequireAuth).
func GetUser(c *gin.Context) *model.User {
	v, ok := c.Get(UserContextKey)
	if !ok {
		return nil
	}
	u, _ := v.(*model.User)
	return u
}
