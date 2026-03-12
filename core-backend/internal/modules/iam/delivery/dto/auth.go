// Package dto contains Data Transfer Objects for HTTP request/response bodies.
package dto

import (
	"net/http"
	"time"

	"github.com/google/uuid"
)

// UserDTO represents a user in API responses.
type UserDTO struct {
	ID        uuid.UUID `json:"id" doc:"User's unique identifier"`
	FirstName string    `json:"firstName" doc:"User's first name"`
	LastName  string    `json:"lastName" doc:"User's last name"`
	ImageURL  *string   `json:"imageUrl,omitempty" doc:"URL to user's profile image"`
	Bio       *string   `json:"bio,omitempty" doc:"User's biography"`
}

// AccountDTO represents an account in API responses.
type AccountDTO struct {
	ID     uuid.UUID `json:"id" doc:"Account's unique identifier"`
	Email  string    `json:"email" doc:"Account email address"`
	Status string    `json:"status" doc:"Account status (pending_verification, active, locked, suspended, disabled)"`
}

// RegisterRequest is the input for user registration.
type RegisterRequest struct {
	Email     string `json:"email" doc:"Email address" format:"email" minLength:"1" maxLength:"255"`
	Password  string `json:"password" doc:"Password (min 8 chars, 1 uppercase, 1 lowercase, 1 digit)" minLength:"8" maxLength:"128"`
	FirstName string `json:"firstName" doc:"First name" minLength:"1" maxLength:"100"`
	LastName  string `json:"lastName" doc:"Last name" minLength:"1" maxLength:"100"`
}

// RegisterInput wraps the register request body for Huma.
type RegisterInput struct {
	Body RegisterRequest
}

// RegisterResponseBody is the response body for successful registration.
type RegisterResponseBody struct {
	AccessToken string     `json:"accessToken" doc:"JWT access token"`
	ExpiresAt   time.Time  `json:"expiresAt" doc:"When the access token expires"`
	User        UserDTO    `json:"user" doc:"Created user"`
	Account     AccountDTO `json:"account" doc:"Created account"`
}

// RegisterOutput is the full response for registration, including the Set-Cookie header.
type RegisterOutput struct {
	SetCookie http.Cookie `header:"Set-Cookie"`
	Body      RegisterResponseBody
}

// LoginRequest is the input for user login.
type LoginRequest struct {
	Email    string `json:"email" doc:"Email address" format:"email" minLength:"1" maxLength:"255"`
	Password string `json:"password" doc:"Password" minLength:"1" maxLength:"128"`
}

// LoginInput wraps the login request body for Huma.
type LoginInput struct {
	Body LoginRequest
}

// LoginResponseBody is the response body for successful login.
type LoginResponseBody struct {
	AccessToken string     `json:"accessToken" doc:"JWT access token"`
	ExpiresAt   time.Time  `json:"expiresAt" doc:"When the access token expires"`
	User        UserDTO    `json:"user" doc:"Authenticated user"`
	Account     AccountDTO `json:"account" doc:"Authenticated account"`
}

// LoginOutput is the full response for login, including the Set-Cookie header.
type LoginOutput struct {
	SetCookie http.Cookie `header:"Set-Cookie"`
	Body      LoginResponseBody
}

// RefreshInput reads the refresh token from cookies.
type RefreshInput struct {
	RefreshToken string `cookie:"refresh_token"`
}

// RefreshResponseBody is the response body for token refresh.
type RefreshResponseBody struct {
	AccessToken string    `json:"accessToken" doc:"New JWT access token"`
	ExpiresAt   time.Time `json:"expiresAt" doc:"When the access token expires"`
}

// RefreshOutput is the full response for refresh, including the Set-Cookie header.
type RefreshOutput struct {
	SetCookie http.Cookie `header:"Set-Cookie"`
	Body      RefreshResponseBody
}

// LogoutInput has no body - authentication comes from middleware context.
type LogoutInput struct{}

// LogoutOutput clears the refresh token cookie.
type LogoutOutput struct {
	SetCookie http.Cookie `header:"Set-Cookie"`
}

// LogoutAllInput has no body - authentication comes from middleware context.
type LogoutAllInput struct{}

// LogoutAllOutput clears the refresh token cookie.
type LogoutAllOutput struct {
	SetCookie http.Cookie `header:"Set-Cookie"`
}
