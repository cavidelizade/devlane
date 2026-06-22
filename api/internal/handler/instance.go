package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Devlaner/devlane/api/internal/auth"
	"github.com/Devlaner/devlane/api/internal/crypto"
	"github.com/Devlaner/devlane/api/internal/middleware"
	"github.com/Devlaner/devlane/api/internal/model"
	"github.com/Devlaner/devlane/api/internal/store"
	"github.com/gin-gonic/gin"
)

// Allowed instance setting section keys (must match migration seed).
var allowedSettingKeys = map[string]bool{
	"general": true, "email": true, "auth": true, "oauth": true, "ai": true, "image": true,
	"github_app": true,
}

// InstanceHandler serves instance setup (first-run); no auth required.
type InstanceHandler struct {
	Auth     *auth.Service
	Users    *store.UserStore
	Settings *store.InstanceSettingStore
}

// InstanceSettingsHandler serves instance settings (GET/PATCH); requires auth.
type InstanceSettingsHandler struct {
	Settings *store.InstanceSettingStore
	// OnSectionUpdated, if set, is invoked after a successful update with the
	// section key. Used for hot-reload of integration clients (e.g. github_app)
	// so the new credentials take effect without an API restart.
	OnSectionUpdated func(ctx context.Context, key string)
}

// SetupStatusResponse for GET /api/instance/setup-status/
type SetupStatusResponse struct {
	SetupRequired bool `json:"setup_required"`
}

// SetupStatus reports whether the instance requires initial setup.
// GET /api/instance/setup-status/
func (h *InstanceHandler) SetupStatus(c *gin.Context) {
	count, err := h.Users.Count(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check setup status"})
		return
	}
	c.JSON(http.StatusOK, SetupStatusResponse{SetupRequired: count == 0})
}

// InstanceSetupRequest for POST /api/instance/setup/
type InstanceSetupRequest struct {
	FirstName   string `json:"first_name" binding:"required"`
	LastName    string `json:"last_name" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	CompanyName string `json:"company_name"`
}

// InstanceSetup performs first-run setup (creates first user and session).
// POST /api/instance/setup/
func (h *InstanceHandler) InstanceSetup(c *gin.Context) {
	count, err := h.Users.Count(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Setup check failed"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Instance is already set up"})
		return
	}
	var req InstanceSetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "detail": err.Error()})
		return
	}
	sessionKey, user, err := h.Auth.SignUp(c.Request.Context(), auth.SignUpRequest{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		if err == auth.ErrEmailTaken {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
			return
		}
		if err == auth.ErrUsernameTaken {
			c.JSON(http.StatusConflict, gin.H{"error": "Username already taken"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Sign up failed"})
		return
	}

	// Seed general instance settings: generated instance_id, admin email from setup, instance name from company name
	if h.Settings != nil {
		instanceID := generateInstanceID()
		instanceName := strings.TrimSpace(req.CompanyName)
		if instanceName == "" {
			instanceName = user.DisplayName
		}
		_ = h.Settings.Upsert(c.Request.Context(), "general", model.JSONMap{
			"instance_id":                     instanceID,
			"admin_email":                     req.Email,
			"instance_name":                   instanceName,
			"only_admin_can_create_workspace": false,
		})
	}

	setSessionCookie(c, sessionKey)
	c.JSON(http.StatusCreated, userResponse(user))
}

// generateInstanceID returns a 24-character hex string for the instance.
func generateInstanceID() string {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return hex.EncodeToString([]byte("default"))
	}
	return hex.EncodeToString(b)
}

// isInstanceAdmin reports whether the authenticated user is the instance
// administrator, identified by general.admin_email (seeded at first-run setup).
// Returns false (fail closed) when no admin email is configured.
func (h *InstanceSettingsHandler) isInstanceAdmin(c *gin.Context) bool {
	user := middleware.GetUser(c)
	if user == nil || user.Email == nil {
		return false
	}
	row, err := h.Settings.Get(c.Request.Context(), "general")
	if err != nil || row == nil {
		return false
	}
	adminEmail, _ := row.Value["admin_email"].(string)
	adminEmail = strings.TrimSpace(adminEmail)
	if adminEmail == "" {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(*user.Email), adminEmail)
}

// requireInstanceAdmin writes a 403 and returns false if the caller is not the
// instance administrator.
func (h *InstanceSettingsHandler) requireInstanceAdmin(c *gin.Context) bool {
	if !h.isInstanceAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Instance admin access required"})
		return false
	}
	return true
}

// GetSettings returns all instance settings sections. Secret values are never
// returned (only <secret>_set booleans); admin access is required.
// GET /api/instance/settings/
func (h *InstanceSettingsHandler) GetSettings(c *gin.Context) {
	if !h.requireInstanceAdmin(c) {
		return
	}
	all, err := h.Settings.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load settings"})
		return
	}
	out := make(map[string]model.JSONMap)
	for k, row := range all {
		out[k] = sanitizeSectionForResponse(k, row.Value)
	}
	// Ensure all sections exist with defaults (migration seed may not have run if DB was created before seed)
	for _, key := range []string{"general", "email", "auth", "oauth", "ai", "image", "github_app"} {
		if _, ok := out[key]; !ok {
			out[key] = defaultSettingValue(key)
		}
	}
	c.JSON(http.StatusOK, out)
}

// secretKeysBySection lists the encrypted secret fields stored in each settings section.
var secretKeysBySection = map[string][]string{
	"email":      {"password"},
	"oauth":      {"google_client_secret", "github_client_secret", "gitlab_client_secret"},
	"ai":         {"api_key"},
	"image":      {"unsplash_access_key"},
	"github_app": {"private_key", "client_secret", "webhook_secret"},
}

// decryptSectionSecretsInternal returns a copy of m with secret fields decrypted
// for SERVER-SIDE use only (building OAuth providers, proxying Unsplash). The
// result MUST NOT be returned to a client — use sanitizeSectionForResponse for
// anything sent over the wire.
func decryptSectionSecretsInternal(sectionKey string, m model.JSONMap) model.JSONMap {
	if m == nil {
		return nil
	}
	out := make(model.JSONMap, len(m))
	for k, v := range m {
		out[k] = v
	}
	for _, sk := range secretKeysBySection[sectionKey] {
		if v, ok := out[sk].(string); ok {
			out[sk] = crypto.DecryptOrPlain(v)
		}
	}
	return out
}

// sanitizeSectionForResponse returns a copy of m with every secret value removed
// and a corresponding <secret>_set boolean reflecting whether a value is stored.
// Secrets are NEVER returned to any client — the admin re-enters them to change
// them (UpdateSetting preserves existing secrets when a field is sent empty).
func sanitizeSectionForResponse(sectionKey string, m model.JSONMap) model.JSONMap {
	if m == nil {
		return nil
	}
	out := make(model.JSONMap, len(m))
	for k, v := range m {
		out[k] = v
	}
	for _, sk := range secretKeysBySection[sectionKey] {
		stored, _ := out[sk].(string)
		out[sk] = ""
		out[sk+"_set"] = strings.TrimSpace(stored) != ""
	}
	return out
}

func defaultSettingValue(key string) model.JSONMap {
	switch key {
	case "general":
		return model.JSONMap{"instance_name": "", "admin_email": "", "instance_id": "", "only_admin_can_create_workspace": false}
	case "email":
		return model.JSONMap{"host": "", "port": "587", "sender_email": "", "security": "TLS", "username": "", "password_set": false}
	case "auth":
		return model.JSONMap{"allow_public_signup": true, "magic_code": true, "password": true, "google": false, "github": false, "gitlab": false}
	case "oauth":
		return model.JSONMap{
			"google_client_id": "", "google_client_secret_set": false,
			"github_client_id": "", "github_client_secret_set": false,
			"gitlab_client_id": "", "gitlab_client_secret_set": false, "gitlab_host": "",
		}
	case "ai":
		return model.JSONMap{"model": "gpt-4o-mini", "api_key_set": false}
	case "image":
		return model.JSONMap{"unsplash_access_key_set": false}
	case "github_app":
		return model.JSONMap{
			"app_id": "", "app_name": "", "client_id": "",
			"client_secret_set": false, "private_key_set": false, "webhook_secret_set": false,
		}
	default:
		return model.JSONMap{}
	}
}

// UpdateSettingRequest for PATCH /api/instance/settings/:key
type UpdateSettingRequest struct {
	Value model.JSONMap `json:"value" binding:"required"`
}

// UpdateSetting updates one instance settings section by key.
// PATCH /api/instance/settings/:key
func (h *InstanceSettingsHandler) UpdateSetting(c *gin.Context) {
	if !h.requireInstanceAdmin(c) {
		return
	}
	key := strings.TrimSpace(strings.ToLower(c.Param("key")))
	if key == "" || !allowedSettingKeys[key] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid settings key"})
		return
	}
	var req UpdateSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "detail": err.Error()})
		return
	}
	value := req.Value
	if key == "general" {
		// instance_name and only_admin_can_create_workspace are editable; preserve admin_email and instance_id
		existing, _ := h.Settings.Get(c.Request.Context(), "general")
		merged := model.JSONMap{}
		if existing != nil {
			for k, v := range existing.Value {
				merged[k] = v
			}
		}
		if name, ok := req.Value["instance_name"]; ok {
			merged["instance_name"] = name
		}
		if onlyAdmin, ok := req.Value["only_admin_can_create_workspace"]; ok {
			merged["only_admin_can_create_workspace"] = onlyAdmin
		}
		value = merged
	}
	if key == "email" {
		// Merge with existing; store password encrypted (preserve existing if not sent)
		existing, _ := h.Settings.Get(c.Request.Context(), "email")
		merged := model.JSONMap{}
		if existing != nil {
			for k, v := range existing.Value {
				merged[k] = v
			}
		}
		for k, v := range req.Value {
			if k == "password" {
				if s, ok := v.(string); ok && s != "" {
					merged["password"] = crypto.EncryptOrPlain(s)
					merged["password_set"] = true
				}
				continue
			}
			merged[k] = v
		}
		value = merged
	}
	if key == "ai" {
		// Store api_key encrypted (preserve existing if not sent)
		existing, _ := h.Settings.Get(c.Request.Context(), "ai")
		merged := model.JSONMap{}
		if existing != nil {
			for k, v := range existing.Value {
				merged[k] = v
			}
		}
		for k, v := range req.Value {
			if k == "api_key" {
				if s, ok := v.(string); ok && s != "" {
					merged["api_key"] = crypto.EncryptOrPlain(s)
					merged["api_key_set"] = true
				}
				continue
			}
			merged[k] = v
		}
		value = merged
	}
	if key == "image" {
		// Store unsplash_access_key encrypted (preserve existing if not sent)
		existing, _ := h.Settings.Get(c.Request.Context(), "image")
		merged := model.JSONMap{}
		if existing != nil {
			for k, v := range existing.Value {
				merged[k] = v
			}
		}
		for k, v := range req.Value {
			if k == "unsplash_access_key" {
				if s, ok := v.(string); ok && s != "" {
					merged["unsplash_access_key"] = crypto.EncryptOrPlain(s)
					merged["unsplash_access_key_set"] = true
				}
				continue
			}
			merged[k] = v
		}
		value = merged
	}
	if key == "auth" {
		// Merge with existing so per-provider pages (Google/GitHub/GitLab) do not wipe other flags.
		existing, _ := h.Settings.Get(c.Request.Context(), "auth")
		merged := model.JSONMap{}
		if existing != nil {
			for k, v := range existing.Value {
				merged[k] = v
			}
		} else {
			for k, v := range defaultSettingValue("auth") {
				merged[k] = v
			}
		}
		for k, v := range req.Value {
			merged[k] = v
		}
		value = merged
	}
	if key == "oauth" {
		existing, _ := h.Settings.Get(c.Request.Context(), "oauth")
		merged := model.JSONMap{}
		if existing != nil {
			for k, v := range existing.Value {
				merged[k] = v
			}
		}
		secretField := func(field string, setKey string) {
			if v, ok := req.Value[field]; ok {
				if s, ok := v.(string); ok && s != "" {
					merged[field] = crypto.EncryptOrPlain(s)
					merged[setKey] = true
				}
			}
		}
		for k, v := range req.Value {
			switch k {
			case "google_client_secret", "github_client_secret", "gitlab_client_secret":
				continue
			default:
				merged[k] = v
			}
		}
		secretField("google_client_secret", "google_client_secret_set")
		secretField("github_client_secret", "github_client_secret_set")
		secretField("gitlab_client_secret", "gitlab_client_secret_set")
		value = merged
	}
	if key == "github_app" {
		// Merge with existing; encrypt secrets and set *_set flags. Empty
		// strings are ignored so the admin can edit one field without resetting
		// the others.
		existing, _ := h.Settings.Get(c.Request.Context(), "github_app")
		merged := model.JSONMap{}
		if existing != nil {
			for k, v := range existing.Value {
				merged[k] = v
			}
		} else {
			for k, v := range defaultSettingValue("github_app") {
				merged[k] = v
			}
		}
		// Plain (non-secret) fields.
		for _, field := range []string{"app_id", "app_name", "client_id"} {
			if v, ok := req.Value[field]; ok {
				merged[field] = v
			}
		}
		// Secret fields: encrypt, set the *_set flag, never expose back.
		setSecret := func(field, setKey string) {
			if v, ok := req.Value[field]; ok {
				if s, ok := v.(string); ok && s != "" {
					merged[field] = crypto.EncryptOrPlain(s)
					merged[setKey] = true
				}
			}
		}
		setSecret("client_secret", "client_secret_set")
		setSecret("private_key", "private_key_set")
		setSecret("webhook_secret", "webhook_secret_set")
		value = merged
	}
	if err := h.Settings.Upsert(c.Request.Context(), key, value); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save settings"})
		return
	}
	// Hot-reload integrations (e.g. github_app) so the new credentials take
	// effect without an API restart. Errors here are logged-and-ignored — the
	// settings are already saved; a stale client is preferable to a 500.
	if h.OnSectionUpdated != nil {
		h.OnSectionUpdated(c.Request.Context(), key)
	}
	// Never echo secrets back; the client already has the value it just set.
	responseValue := sanitizeSectionForResponse(key, value)
	c.JSON(http.StatusOK, gin.H{"key": key, "value": responseValue})
}

// unsplashPhoto is a single photo from Unsplash API response.
type unsplashPhoto struct {
	ID   string `json:"id"`
	URLs struct {
		Full    string `json:"full"`
		Regular string `json:"regular"`
		Thumb   string `json:"thumb"`
	} `json:"urls"`
}

// unsplashSearchResponse is the Unsplash search API response.
type unsplashSearchResponse struct {
	Results []unsplashPhoto `json:"results"`
}

// UnsplashSearchResult is a simplified photo returned by our proxy.
type UnsplashSearchResult struct {
	ID    string `json:"id"`
	URL   string `json:"url"`
	Thumb string `json:"thumb"`
}

// UnsplashSearch proxies search to Unsplash API using instance image settings key (auth required).
// GET /api/instance/unsplash/search?q=...
func (h *InstanceSettingsHandler) UnsplashSearch(c *gin.Context) {
	if h.Settings == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Settings not available"})
		return
	}
	row, err := h.Settings.Get(c.Request.Context(), "image")
	if err != nil || row == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsplash is not configured"})
		return
	}
	decrypted := decryptSectionSecretsInternal("image", row.Value)
	keyVal, _ := decrypted["unsplash_access_key"].(string)
	keyVal = strings.TrimSpace(keyVal)
	if keyVal == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsplash is not configured"})
		return
	}

	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		c.JSON(http.StatusOK, gin.H{"results": []UnsplashSearchResult{}})
		return
	}

	apiURL := "https://api.unsplash.com/search/photos?query=" + url.QueryEscape(q) + "&per_page=20"
	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, apiURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search"})
		return
	}
	req.Header.Set("Authorization", "Client-ID "+keyVal)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, resp.Body)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Unsplash search failed"})
		return
	}

	var payload unsplashSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid response"})
		return
	}

	results := make([]UnsplashSearchResult, 0, len(payload.Results))
	for _, p := range payload.Results {
		u := p.URLs.Regular
		if u == "" {
			u = p.URLs.Full
		}
		results = append(results, UnsplashSearchResult{ID: p.ID, URL: u, Thumb: p.URLs.Thumb})
	}
	c.JSON(http.StatusOK, gin.H{"results": results})
}
