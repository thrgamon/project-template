package domain

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UserResponse struct {
	ID    int32  `json:"id"`
	Email string `json:"email"`
}

type AuthResponse struct {
	User UserResponse `json:"user"`
}
