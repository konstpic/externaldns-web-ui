package auth

import (
	"os"
	"strings"
	"time"
)

type Config struct {
	Enabled          bool
	IssuerURL        string
	ClientID         string
	ClientSecret     string
	RedirectURL      string
	Scopes           []string
	RoleClaim        string
	GroupClaim       string
	AuthentikPublic  string
	FrontendURL      string
	JWTSecret        string
	AccessTTL        time.Duration
	RefreshTTL       time.Duration
	AdminRoles       []string
	AuthRequired     bool
}

func LoadConfig() Config {
	frontendURL := strings.TrimRight(envOr("FRONTEND_URL", ""), "/")
	redirectURL := envOr("OIDC_REDIRECT_URL", "")
	if redirectURL == "" && frontendURL != "" {
		redirectURL = frontendURL + "/api/auth/callback"
	}

	scopes := strings.Fields(envOr("OIDC_SCOPES", "openid profile email"))
	adminRoles := splitCSV(envOr("ADMIN_ROLES", "admin,super_admin,authentik Admins"))

	authRequired := true
	if v := os.Getenv("AUTH_REQUIRED"); v == "0" || strings.EqualFold(v, "false") {
		authRequired = false
	}

	enabled := !strings.EqualFold(envOr("OIDC_ENABLED", "true"), "false") && envOr("OIDC_ENABLED", "true") != "0"

	return Config{
		Enabled:         enabled,
		IssuerURL:       strings.TrimRight(envOr("OIDC_ISSUER_URL", ""), "/") + "/",
		ClientID:        envOr("OIDC_CLIENT_ID", ""),
		ClientSecret:    envOr("OIDC_CLIENT_SECRET", ""),
		RedirectURL:     redirectURL,
		Scopes:          scopes,
		RoleClaim:       envOr("OIDC_ROLE_CLAIM", "groups"),
		GroupClaim:      envOr("OIDC_GROUP_CLAIM", "groups"),
		AuthentikPublic: strings.TrimRight(envOr("AUTHENTIK_PUBLIC_URL", ""), "/"),
		FrontendURL:     frontendURL,
		JWTSecret:       envOr("JWT_SECRET", ""),
		AccessTTL:       parseDuration(envOr("JWT_ACCESS_TTL", "15m"), 15*time.Minute),
		RefreshTTL:      parseDuration(envOr("JWT_REFRESH_TTL", "168h"), 168*time.Hour),
		AdminRoles:      adminRoles,
		AuthRequired:    authRequired,
	}
}

func (c Config) OIDCConfigured() bool {
	return c.Enabled && c.IssuerURL != "/" && c.ClientID != "" && c.ClientSecret != "" && c.RedirectURL != ""
}

func (c Config) JWTConfigured() bool {
	return c.JWTSecret != ""
}

func (c Config) IsAdmin(roles []string) bool {
	for _, r := range roles {
		for _, admin := range c.AdminRoles {
			if strings.EqualFold(r, admin) {
				return true
			}
		}
	}
	return false
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitCSV(v string) []string {
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func parseDuration(v string, fallback time.Duration) time.Duration {
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}
