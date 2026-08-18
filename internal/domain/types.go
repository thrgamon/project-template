// Package domain holds the request and response types that make up the HTTP
// API. It is the single source of truth for the API shape: TypeScript types
// are generated from this package by tygo (see tygo.yaml).
package domain

// RegisterRequest is the body of POST /api/auth/register.
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// LoginRequest is the body of POST /api/auth/login.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UserResponse describes an authenticated user.
type UserResponse struct {
	ID    int32  `json:"id"`
	Email string `json:"email"`
}

// AuthResponse is returned by every endpoint that establishes or reports the
// current session: register, login and me.
type AuthResponse struct {
	User UserResponse `json:"user"`
}

// MessageResponse is returned by endpoints that report an outcome with no
// other payload, such as logout.
type MessageResponse struct {
	Message string `json:"message"`
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
