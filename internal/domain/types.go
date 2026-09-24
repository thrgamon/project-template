// Package domain holds the request and response types that make up the HTTP
// API. Browser-facing DTOs are deliberately hand-written beside the SvelteKit
// API client; validate this boundary with endpoint tests rather than codegen.
package domain

// UserResponse describes an authenticated user.
type UserResponse struct {
	ID    int32  `json:"id"`
	Email string `json:"email"`
}

// AuthResponse is returned by GET /api/auth/me after an Auth0 session has
// been mapped to an active local membership.
type AuthResponse struct {
	User      UserResponse `json:"user"`
	CSRFToken string       `json:"csrfToken"`
}

// HealthResponse is returned by GET /api/health.
type HealthResponse struct {
	Status string `json:"status"`
}

// DashboardResponse is the example protected payload.
type DashboardResponse struct {
	Message string `json:"message"`
	Email   string `json:"email"`
}

// ErrorResponse is the body of every non-2xx API response.
type ErrorResponse struct {
	Error string `json:"error"`
}
