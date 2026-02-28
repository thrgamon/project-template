package domain

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email" validate:"required"`
	Password string `json:"password" binding:"required,min=8" validate:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" validate:"required"`
	Password string `json:"password" binding:"required" validate:"required"`
}

type UserResponse struct {
	ID    int32  `json:"id" validate:"required"`
	Email string `json:"email" validate:"required"`
}

type AuthResponse struct {
	User UserResponse `json:"user" validate:"required"`
}
