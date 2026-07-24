package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type Service struct {
	cfg    Config
	jwt    *JWTManager
	states *stateStore

	provider *oidc.Provider
	oauth2   *oauth2.Config
}

func NewService(cfg Config) (*Service, error) {
	jwtMgr, err := NewJWTManager(cfg)
	if err != nil {
		return nil, err
	}

	s := &Service{
		cfg:    cfg,
		jwt:    jwtMgr,
		states: newStateStore(),
	}

	if cfg.OIDCConfigured() {
		if err := s.initOIDC(context.Background()); err != nil {
			log.Printf("oidc init deferred: %v", err)
		}
	}

	return s, nil
}

func (s *Service) initOIDC(ctx context.Context) error {
	provider, err := oidc.NewProvider(ctx, strings.TrimRight(s.cfg.IssuerURL, "/"))
	if err != nil {
		return err
	}
	s.provider = provider
	s.oauth2 = &oauth2.Config{
		ClientID:     s.cfg.ClientID,
		ClientSecret: s.cfg.ClientSecret,
		RedirectURL:  s.cfg.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       s.cfg.Scopes,
	}
	log.Printf("oidc provider ready: %s", s.cfg.IssuerURL)
	return nil
}

func (s *Service) Config() Config {
	return s.cfg
}

func (s *Service) JWT() *JWTManager {
	return s.jwt
}

func (s *Service) LoginRedirect(w http.ResponseWriter, r *http.Request) {
	if s.oauth2 == nil || s.provider == nil {
		if err := s.initOIDC(r.Context()); err != nil {
			http.Error(w, `{"error":"oidc not configured"}`, http.StatusServiceUnavailable)
			return
		}
	}
	state, err := s.states.create()
	if err != nil {
		http.Error(w, `{"error":"state error"}`, http.StatusInternalServerError)
		return
	}
	authURL := s.oauth2.AuthCodeURL(state, oauth2.AccessTypeOffline)
	if s.cfg.AuthentikPublic != "" {
		authURL = rewriteBrowserURL(authURL, s.cfg.AuthentikPublic)
	}
	http.Redirect(w, r, authURL, http.StatusFound)
}

func (s *Service) Callback(w http.ResponseWriter, r *http.Request) {
	if s.oauth2 == nil || s.provider == nil {
		http.Error(w, `{"error":"oidc not configured"}`, http.StatusServiceUnavailable)
		return
	}
	state := r.URL.Query().Get("state")
	if !s.states.validate(state) {
		http.Error(w, `{"error":"invalid state"}`, http.StatusBadRequest)
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, `{"error":"missing code"}`, http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	token, err := s.oauth2.Exchange(ctx, code)
	if err != nil {
		http.Error(w, `{"error":"token exchange failed"}`, http.StatusBadRequest)
		return
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, `{"error":"missing id_token"}`, http.StatusBadRequest)
		return
	}
	verifier := s.provider.Verifier(&oidc.Config{ClientID: s.oauth2.ClientID})
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		http.Error(w, `{"error":"invalid id_token"}`, http.StatusBadRequest)
		return
	}

	var claims map[string]any
	if err := idToken.Claims(&claims); err != nil {
		http.Error(w, `{"error":"invalid claims"}`, http.StatusBadRequest)
		return
	}

	email := claimString(claims, "email")
	if email == "" {
		email = claimString(claims, "preferred_username")
	}
	displayName := claimString(claims, "name")
	if displayName == "" {
		displayName = email
	}

	roles := extractStringSlice(claims[s.cfg.RoleClaim])
	if len(roles) == 0 {
		roles = extractStringSlice(claims[s.cfg.GroupClaim])
	}
	if len(roles) == 0 {
		roles = []string{"viewer"}
	}

	pair, err := s.jwt.IssuePair(idToken.Subject, email, displayName, roles)
	if err != nil {
		http.Error(w, `{"error":"token issue failed"}`, http.StatusInternalServerError)
		return
	}

	redirect := fmt.Sprintf("%s/auth/callback?access_token=%s&refresh_token=%s",
		s.cfg.FrontendURL,
		url.QueryEscape(pair.AccessToken),
		url.QueryEscape(pair.RefreshToken),
	)
	http.Redirect(w, r, redirect, http.StatusFound)
}

func (s *Service) Refresh(r *http.Request) (*TokenPair, error) {
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := readJSON(r, &body); err != nil || body.RefreshToken == "" {
		return nil, ErrInvalidToken
	}
	claims, err := s.jwt.Parse(body.RefreshToken, "refresh")
	if err != nil {
		return nil, err
	}
	return s.jwt.IssuePair(claims.UserID, claims.Email, claims.DisplayName, claims.Roles)
}

func (s *Service) UserFromRequest(r *http.Request) (*Claims, error) {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		return nil, ErrInvalidToken
	}
	return s.jwt.Parse(strings.TrimPrefix(auth, "Bearer "), "access")
}

type stateStore struct {
	mu   sync.Mutex
	data map[string]time.Time
}

func newStateStore() *stateStore {
	s := &stateStore{data: make(map[string]time.Time)}
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			s.cleanup()
		}
	}()
	return s
}

func (s *stateStore) create() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	state := base64.URLEncoding.EncodeToString(b)
	s.mu.Lock()
	s.data[state] = time.Now().Add(10 * time.Minute)
	s.mu.Unlock()
	return state, nil
}

func (s *stateStore) validate(state string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	exp, ok := s.data[state]
	if !ok || time.Now().After(exp) {
		return false
	}
	delete(s.data, state)
	return true
}

func (s *stateStore) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for k, exp := range s.data {
		if now.After(exp) {
			delete(s.data, k)
		}
	}
}

func rewriteBrowserURL(raw, publicBase string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	pub, err := url.Parse(publicBase)
	if err != nil {
		return raw
	}
	u.Scheme = pub.Scheme
	u.Host = pub.Host
	return u.String()
}

func claimString(claims map[string]any, key string) string {
	v, ok := claims[key]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

func extractStringSlice(v any) []string {
	switch t := v.(type) {
	case []any:
		out := make([]string, 0, len(t))
		for _, item := range t {
			if s, ok := item.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	case []string:
		return t
	case string:
		if t != "" {
			return []string{t}
		}
	}
	return nil
}
