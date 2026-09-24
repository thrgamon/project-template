// Package auth provides the shared server-side OIDC and session boundary.
//
// Auth0 owns identity. Applications own membership, local user identifiers,
// roles, and authorization. This package deliberately does not implement
// application API keys or public share links.
package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

var (
	ErrUnauthenticated   = errors.New("authentication required")
	ErrMembershipDenied  = errors.New("application membership denied")
	ErrInvalidState      = errors.New("invalid authentication state")
	ErrInvalidReturnPath = errors.New("invalid return path")
	ErrCSRF              = errors.New("csrf validation failed")
)

type Claims struct {
	Issuer  string
	Subject string
	Email   string
	Name    string
}

type Principal struct {
	Issuer      string `json:"issuer"`
	Subject     string `json:"subject"`
	Email       string `json:"email,omitempty"`
	Name        string `json:"name,omitempty"`
	LocalUserID string `json:"localUserId"`
	Role        string `json:"role,omitempty"`
	Status      string `json:"status"`
}

type IdentityStore interface {
	ResolveIdentity(context.Context, Claims) (Principal, error)
}

type SessionRecord struct {
	TokenHash []byte
	Principal Principal
	ExpiresAt time.Time
}

type SessionStore interface {
	CreateSession(context.Context, SessionRecord) error
	ResolveSession(context.Context, []byte) (Principal, error)
	DeleteSession(context.Context, []byte) error
	DeleteExpiredSessions(context.Context) error
}

type Config struct {
	IssuerURL     string
	ClientID      string
	ClientSecret  string
	RedirectURL   string
	PostLogoutURL string
	CookieName    string
	CookieSecure  bool
	// AllowInsecureCookies must be explicitly enabled for HTTP-only local
	// development. Production callers must leave it false and set CookieSecure.
	AllowInsecureCookies bool
	CookieDomain         string
	StateSecret          []byte
	SessionMaxAge        time.Duration
	RoutePrefix          string
	DefaultReturnPath    string
	TransactionMaxAge    time.Duration
}

func (c Config) withDefaults() Config {
	if c.CookieName == "" {
		c.CookieName = "session"
	}
	if c.SessionMaxAge <= 0 {
		c.SessionMaxAge = 7 * 24 * time.Hour
	}
	if c.RoutePrefix == "" {
		c.RoutePrefix = "/api/auth"
	}
	if c.DefaultReturnPath == "" {
		c.DefaultReturnPath = "/"
	}
	if c.TransactionMaxAge <= 0 {
		c.TransactionMaxAge = 10 * time.Minute
	}
	return c
}

type App struct {
	cfg       Config
	oauth     oauth2.Config
	verifier  *oidc.IDTokenVerifier
	identity  IdentityStore
	sessions  SessionStore
	stateMu   sync.Mutex
	usedState map[string]time.Time
}

func New(ctx context.Context, cfg Config, identity IdentityStore, sessions SessionStore) (*App, error) {
	cfg = cfg.withDefaults()
	if identity == nil || sessions == nil {
		return nil, errors.New("identity and session stores are required")
	}
	// Keep each application's session cookie host-only. A parent-domain cookie
	// would allow a sibling application to receive this application's session.
	if cfg.CookieDomain != "" {
		return nil, errors.New("cookie domain is not supported; use host-only cookies")
	}
	if !cfg.CookieSecure && !cfg.AllowInsecureCookies {
		return nil, errors.New("secure cookies are required; explicitly allow insecure cookies only for local development")
	}
	if len(cfg.StateSecret) < 32 {
		return nil, errors.New("state secret must be at least 32 bytes")
	}
	if cfg.IssuerURL == "" || cfg.ClientID == "" || cfg.ClientSecret == "" || cfg.RedirectURL == "" {
		return nil, errors.New("issuer, client credentials, and redirect URL are required")
	}
	issuer, err := url.Parse(cfg.IssuerURL)
	if err != nil || issuer.Scheme == "" || issuer.Host == "" {
		return nil, errors.New("issuer URL must be absolute")
	}
	redirect, err := url.Parse(cfg.RedirectURL)
	if err != nil || redirect.Scheme == "" || redirect.Host == "" || redirect.Fragment != "" {
		return nil, errors.New("redirect URL must be an absolute URL without a fragment")
	}
	if !cfg.AllowInsecureCookies && (issuer.Scheme != "https" || redirect.Scheme != "https") {
		return nil, errors.New("issuer and redirect URLs must use HTTPS outside local development")
	}
	provider, err := oidc.NewProvider(ctx, cfg.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("discover oidc provider: %w", err)
	}
	return newApp(cfg, provider, identity, sessions), nil
}

func newApp(cfg Config, provider *oidc.Provider, identity IdentityStore, sessions SessionStore) *App {
	return &App{
		cfg:      cfg,
		oauth:    oauth2.Config{ClientID: cfg.ClientID, ClientSecret: cfg.ClientSecret, Endpoint: provider.Endpoint(), RedirectURL: cfg.RedirectURL, Scopes: []string{oidc.ScopeOpenID, "profile", "email"}},
		verifier: provider.Verifier(&oidc.Config{ClientID: cfg.ClientID}),
		identity: identity,
		sessions: sessions,
	}
}

func (a *App) SetStores(identity IdentityStore, sessions SessionStore) {
	a.identity, a.sessions = identity, sessions
}

type transaction struct {
	State, Nonce, Verifier, ReturnPath string
	IssuedAt                           int64
}

func (a *App) Routes(mux *http.ServeMux) {
	p := strings.TrimRight(a.cfg.RoutePrefix, "/")
	mux.HandleFunc("GET "+p+"/login", a.Login)
	mux.HandleFunc("GET "+p+"/callback", a.Callback)
	mux.HandleFunc("POST "+p+"/logout", a.Logout)
	mux.HandleFunc("GET "+p+"/me", a.Me)
}

func (a *App) Login(w http.ResponseWriter, r *http.Request) {
	ret, err := ValidateReturnPath(r.URL.Query().Get("return"), a.cfg.DefaultReturnPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	state, err := randomString(32)
	if err != nil {
		http.Error(w, "internal error", 500)
		return
	}
	nonce, err := randomString(32)
	if err != nil {
		http.Error(w, "internal error", 500)
		return
	}
	verifier, err := randomString(32)
	if err != nil {
		http.Error(w, "internal error", 500)
		return
	}
	tx := transaction{State: state, Nonce: nonce, Verifier: verifier, ReturnPath: ret, IssuedAt: time.Now().Unix()}
	value, err := a.signTransaction(tx)
	if err != nil {
		http.Error(w, "internal error", 500)
		return
	}
	http.SetCookie(w, a.transactionCookie(value))
	url := a.oauth.AuthCodeURL(state, oauth2.SetAuthURLParam("nonce", nonce), oauth2.SetAuthURLParam("code_challenge", pkceChallenge(verifier)), oauth2.SetAuthURLParam("code_challenge_method", "S256"))
	http.Redirect(w, r, url, http.StatusFound)
}

func (a *App) Callback(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie(a.transactionCookieName())
	if err != nil {
		http.Error(w, "invalid authentication state", 400)
		return
	}
	tx, err := a.verifyTransaction(c.Value)
	if err != nil {
		http.Error(w, "invalid authentication state", 400)
		return
	}
	age := time.Since(time.Unix(tx.IssuedAt, 0))
	if tx.IssuedAt <= 0 || age < 0 || age > a.cfg.TransactionMaxAge || !hmac.Equal([]byte(tx.State), []byte(r.URL.Query().Get("state"))) {
		http.Error(w, "invalid authentication state", 400)
		return
	}
	if !a.consumeState(tx.State, time.Unix(tx.IssuedAt, 0)) {
		http.Error(w, "invalid authentication state", http.StatusBadRequest)
		return
	}
	// State is one-time from the browser's perspective even when the provider
	// sends an error: clear it before doing any exchange or redirect.
	http.SetCookie(w, a.clearCookie(a.transactionCookieName()))
	if providerErr := r.URL.Query().Get("error"); providerErr != "" {
		http.Error(w, "authentication failed", 401)
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "authentication code missing", 400)
		return
	}
	tok, err := a.oauth.Exchange(r.Context(), code, oauth2.SetAuthURLParam("code_verifier", tx.Verifier))
	if err != nil {
		http.Error(w, "authentication failed", 401)
		return
	}
	raw, ok := tok.Extra("id_token").(string)
	if !ok || raw == "" {
		http.Error(w, "identity token missing", 401)
		return
	}
	idToken, err := a.verifier.Verify(r.Context(), raw)
	if err != nil {
		http.Error(w, "identity token invalid", 401)
		return
	}
	var cclaims struct {
		Issuer  string `json:"iss"`
		Subject string `json:"sub"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Nonce   string `json:"nonce"`
	}
	if err := idToken.Claims(&cclaims); err != nil || cclaims.Nonce != tx.Nonce || cclaims.Issuer == "" || cclaims.Subject == "" {
		http.Error(w, "identity token invalid", 401)
		return
	}
	principal, err := a.identity.ResolveIdentity(r.Context(), Claims{Issuer: cclaims.Issuer, Subject: cclaims.Subject, Email: cclaims.Email, Name: cclaims.Name})
	if err != nil {
		http.Error(w, "application access denied", 403)
		return
	}
	if principal.Status != "active" || principal.LocalUserID == "" {
		http.Error(w, "application access denied", 403)
		return
	}
	if principal.Issuer != cclaims.Issuer || principal.Subject != cclaims.Subject {
		http.Error(w, "application access denied", http.StatusForbidden)
		return
	}
	rawSession, err := a.StartSession(r.Context(), principal)
	if err != nil {
		http.Error(w, "internal error", 500)
		return
	}
	http.SetCookie(w, a.sessionCookie(rawSession))
	csrf, err := a.csrfCookie()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, csrf)
	http.Redirect(w, r, tx.ReturnPath, http.StatusFound)
}

func (a *App) Logout(w http.ResponseWriter, r *http.Request) {
	if err := a.CheckCSRF(r); err != nil {
		http.Error(w, "csrf validation failed", 403)
		return
	}
	if c, err := r.Cookie(a.cfg.CookieName); err == nil && c.Value != "" {
		_ = a.sessions.DeleteSession(r.Context(), HashToken(c.Value))
	}
	http.SetCookie(w, a.clearCookie(a.cfg.CookieName))
	http.SetCookie(w, a.clearCookie(a.csrfCookieName()))
	w.WriteHeader(http.StatusNoContent)
}

// RequireRole protects a handler with an exact active membership role.
func (a *App) RequireRole(role string, next http.Handler) http.Handler {
	return a.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, _ := PrincipalFromContext(r.Context())
		if role == "" || p.Role != role {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	}))
}

func (a *App) Me(w http.ResponseWriter, r *http.Request) {
	p, err := a.Authenticate(r)
	if err != nil {
		http.Error(w, "authentication required", 401)
		return
	}
	csrf, err := a.EnsureCSRF(w, r)
	if err != nil {
		http.Error(w, "internal error", 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		User      Principal `json:"user"`
		CSRFToken string    `json:"csrfToken"`
	}{p, csrf})
}

func (a *App) StartSession(ctx context.Context, p Principal) (string, error) {
	raw, err := randomString(32)
	if err != nil {
		return "", err
	}
	return raw, a.sessions.CreateSession(ctx, SessionRecord{TokenHash: HashToken(raw), Principal: p, ExpiresAt: time.Now().Add(a.cfg.SessionMaxAge)})
}

func (a *App) Authenticate(r *http.Request) (Principal, error) {
	c, err := r.Cookie(a.cfg.CookieName)
	if err != nil || c.Value == "" {
		return Principal{}, ErrUnauthenticated
	}
	p, err := a.sessions.ResolveSession(r.Context(), HashToken(c.Value))
	if err != nil || p.Status != "active" {
		return Principal{}, ErrUnauthenticated
	}
	return p, nil
}

func (a *App) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p, err := a.Authenticate(r)
		if err != nil {
			http.Error(w, "authentication required", 401)
			return
		}
		next.ServeHTTP(w, r.WithContext(WithPrincipal(r.Context(), p)))
	})
}
func (a *App) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p, err := a.Authenticate(r); err == nil {
			r = r.WithContext(WithPrincipal(r.Context(), p))
		}
		next.ServeHTTP(w, r)
	})
}
func (a *App) RequireCSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := a.CheckCSRF(r); err != nil {
			http.Error(w, "csrf validation failed", 403)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (a *App) CheckCSRF(r *http.Request) error {
	if !a.sameOrigin(r) {
		return ErrCSRF
	}
	c, err := r.Cookie(a.csrfCookieName())
	if err != nil || c.Value == "" {
		return ErrCSRF
	}
	supplied := r.Header.Get("X-CSRF-Token")
	if supplied == "" || subtle.ConstantTimeCompare([]byte(c.Value), []byte(supplied)) != 1 {
		return ErrCSRF
	}
	return nil
}

// sameOrigin rejects an explicitly cross-origin browser request. The token
// comparison below remains necessary because Origin may be omitted by some
// same-origin requests. RedirectURL names the canonical server origin, which
// keeps the check independent of untrusted forwarded request headers.
func (a *App) sameOrigin(r *http.Request) bool {
	if r.Header.Get("Sec-Fetch-Site") == "cross-site" {
		return false
	}
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	u, err := url.Parse(origin)
	if err != nil || u.Scheme == "" || u.Host == "" || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return false
	}
	redirect, err := url.Parse(a.cfg.RedirectURL)
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(u.Scheme+"://"+u.Host), []byte(redirect.Scheme+"://"+redirect.Host)) == 1
}
func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(principalKey{}).(Principal)
	return p, ok
}
func WithPrincipal(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, principalKey{}, p)
}

type principalKey struct{}

func ValidateReturnPath(raw, fallback string) (string, error) {
	if raw == "" {
		raw = fallback
	}
	if raw == "" || strings.ContainsAny(raw, "\\\r\n\x00") {
		return "", ErrInvalidReturnPath
	}
	u, err := url.Parse(raw)
	if err != nil || !u.IsAbs() && (u.Host != "" || !strings.HasPrefix(u.Path, "/") || strings.HasPrefix(u.Path, "//")) {
		return "", ErrInvalidReturnPath
	}
	if u.IsAbs() || u.Host != "" || strings.HasPrefix(raw, "//") {
		return "", ErrInvalidReturnPath
	}
	return raw, nil
}
func HashToken(raw string) []byte { sum := sha256.Sum256([]byte(raw)); return sum[:] }

func (a *App) transactionCookieName() string { return a.cfg.CookieName + "_oidc_tx" }
func (a *App) csrfCookieName() string        { return a.cfg.CookieName + "_csrf" }
func (a *App) transactionCookie(value string) *http.Cookie {
	return &http.Cookie{Name: a.transactionCookieName(), Value: value, Path: "/", MaxAge: int(a.cfg.TransactionMaxAge.Seconds()), HttpOnly: true, Secure: a.cfg.CookieSecure, SameSite: http.SameSiteLaxMode}
}
func (a *App) csrfCookie() (*http.Cookie, error) {
	value, err := randomString(32)
	if err != nil {
		return nil, err
	}
	return &http.Cookie{Name: a.csrfCookieName(), Value: value, Path: "/", MaxAge: int(a.cfg.SessionMaxAge.Seconds()), HttpOnly: false, Secure: a.cfg.CookieSecure, SameSite: http.SameSiteLaxMode}, nil
}

// EnsureCSRF returns the existing browser CSRF token or issues one when the
// caller has already authenticated the request. Framework adapters use this
// to preserve an application's established /api/auth/me response shape.
func (a *App) EnsureCSRF(w http.ResponseWriter, r *http.Request) (string, error) {
	if c, err := r.Cookie(a.csrfCookieName()); err == nil && c.Value != "" {
		return c.Value, nil
	}
	c, err := a.csrfCookie()
	if err != nil {
		return "", err
	}
	http.SetCookie(w, c)
	return c.Value, nil
}
func (a *App) sessionCookie(value string) *http.Cookie {
	return &http.Cookie{Name: a.cfg.CookieName, Value: value, Path: "/", MaxAge: int(a.cfg.SessionMaxAge.Seconds()), HttpOnly: true, Secure: a.cfg.CookieSecure, SameSite: http.SameSiteLaxMode}
}
func (a *App) clearCookie(name string) *http.Cookie {
	return &http.Cookie{Name: name, Value: "", Path: "/", MaxAge: -1, HttpOnly: name != a.csrfCookieName(), Secure: a.cfg.CookieSecure, SameSite: http.SameSiteLaxMode}
}
func (a *App) signTransaction(tx transaction) (string, error) {
	raw, err := json.Marshal(tx)
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(raw)
	mac := hmac.New(sha256.New, a.cfg.StateSecret)
	_, _ = mac.Write([]byte(encoded))
	return encoded + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}
func (a *App) verifyTransaction(value string) (transaction, error) {
	parts := strings.Split(value, ".")
	if len(parts) != 2 {
		return transaction{}, ErrInvalidState
	}
	mac := hmac.New(sha256.New, a.cfg.StateSecret)
	_, _ = mac.Write([]byte(parts[0]))
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !hmac.Equal(sig, mac.Sum(nil)) {
		return transaction{}, ErrInvalidState
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return transaction{}, ErrInvalidState
	}
	var tx transaction
	if json.Unmarshal(raw, &tx) != nil || tx.State == "" || tx.Nonce == "" || tx.Verifier == "" {
		return transaction{}, ErrInvalidState
	}
	return tx, nil
}

// consumeState gives each signed browser transaction one callback attempt.
// The map is process-local; deployments with multiple replicas should route
// the callback to the initiating replica or add a shared replay store around
// this package before enabling active-active login callbacks.
func (a *App) consumeState(state string, issuedAt time.Time) bool {
	a.stateMu.Lock()
	defer a.stateMu.Unlock()
	if a.usedState == nil {
		a.usedState = make(map[string]time.Time)
	}
	cutoff := time.Now().Add(-a.cfg.TransactionMaxAge)
	for key, at := range a.usedState {
		if at.Before(cutoff) {
			delete(a.usedState, key)
		}
	}
	if _, ok := a.usedState[state]; ok {
		return false
	}
	a.usedState[state] = issuedAt
	return true
}
func randomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
func pkceChallenge(v string) string {
	sum := sha256.Sum256([]byte(v))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
