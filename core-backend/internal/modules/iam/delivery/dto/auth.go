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

type UpdateUserProfileRequest struct {
	FirstName string  `json:"firstName" doc:"First name" minLength:"1" maxLength:"100"`
	LastName  string  `json:"lastName" doc:"Last name" minLength:"1" maxLength:"100"`
	Bio       *string `json:"bio,omitempty" doc:"Bio"`
}

type UpdateUserProfileInput struct {
	Body UpdateUserProfileRequest
}

type UpdateUserProfileResponseBody struct {
	FirstName string `json:"firstName" doc:"First name"`
	LastName  string `json:"lastName" doc:"Last name"`
	Bio       string `json:"bio" doc:"Bio"`
}

type UpdateUserProfileOutput struct {
	Body UpdateUserProfileResponseBody
}
type UpdateAccountPasswordRequest struct {
	ExistingPassword string `json:"existingPassword" doc:"Password" minLength:"1" maxLength:"128"`
	NewPassword      string `json:"newPassword" doc:"Password (min 8 chars, 1 uppercase, 1 lowercase, 1 digit)" minLength:"8" maxLength:"128"`
	ConfirmPassword  string `json:"confirmPassword" doc:"Password (min 8 chars, 1 uppercase, 1 lowercase, 1 digit)" minLength:"8" maxLength:"128"`
}

type UpdateAccountPasswordInput struct {
	Body UpdateAccountPasswordRequest
}
type UpdateAccountPasswordResponseBody struct {
	Message string `json:"message" doc:"Message"`
}
type UpdateAccountPasswordOutput struct {
	Body UpdateAccountPasswordResponseBody
}
type GetCurrentUserResponseBody struct {
	User    UserDTO    `json:"user" doc:"Current user"`
	Account AccountDTO `json:"account" doc:"Current user account"`
}

type GetCurrentUserInput struct{}

type GetCurrentUserOutput struct {
	Body GetCurrentUserResponseBody
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

type VerifyEmailOTPRequest struct {
	OTP string `json:"otp" doc:"6-digit verification code" minLength:"6" maxLength:"6"`
}

type VerifyEmailOTPInput struct {
	Body VerifyEmailOTPRequest
}

type VerifyEmailOTPResponseBody struct {
	Message string `json:"message" doc:"Verification status message"`
}

type VerifyEmailOTPOutput struct {
	Body VerifyEmailOTPResponseBody
}

type ResendEmailOTPInput struct{}

type ResendEmailOTPResponseBody struct {
	Message string `json:"message" doc:"OTP resend status message"`
}

type ResendEmailOTPOutput struct {
	Body ResendEmailOTPResponseBody
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

// OAuthProviderDTO represents an OAuth provider.
type OAuthProviderDTO struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Icon        string `json:"icon"`
}

// OAuthProvidersResponse is the response for listing OAuth providers.
type OAuthProvidersResponse struct {
	Providers []OAuthProviderDTO `json:"providers"`
}

// OAuthCallbackResponse is the response body for successful OAuth callback.
type OAuthCallbackResponse struct {
	AccessToken   string                      `json:"accessToken,omitempty"`
	RefreshToken  string                      `json:"refreshToken,omitempty"`
	ExpiresAt     time.Time                   `json:"expiresAt,omitempty"`
	User          *UserDTO                    `json:"user,omitempty"`
	Account       *AccountDTO                 `json:"account,omitempty"`
	IsNewUser     bool                        `json:"isNewUser,omitempty"`
	EmailRequired *OAuthEmailRequiredResponse `json:"emailRequired,omitempty"`
}

// OAuthCallbackOutput is the full response for OAuth callback.
type OAuthCallbackOutput struct {
	SetCookie http.Cookie `header:"Set-Cookie"`
	Body      OAuthCallbackResponse
}

// OAuthEmailRequiredResponse indicates email is required to complete OAuth.
type OAuthEmailRequiredResponse struct {
	Provider   string `json:"provider"`
	Subject    string `json:"subject"`
	Name       string `json:"name"`
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	PictureURL string `json:"pictureUrl,omitempty"`
	State      string `json:"state"`
}

// OAuthEmailRequiredOutput is the response when email is required.
type OAuthEmailRequiredOutput struct {
	Body OAuthEmailRequiredResponse
}

// OAuthLinkCallbackOutput is the response for OAuth link callback.
type OAuthLinkCallbackOutput struct {
	Body struct {
		Provider string `json:"provider"`
	} `json:"body"`
	EmailRequired *OAuthEmailRequiredResponse `json:"emailRequired,omitempty"`
}

// OAuthLinkEmailRequiredOutput is the response when email is required for linking.
type OAuthLinkEmailRequiredOutput struct {
	Body OAuthEmailRequiredResponse
}

// OAuthCompleteEmailRequest is the input for completing OAuth with email.
type OAuthCompleteEmailRequest struct {
	Email string `json:"email" doc:"Email address" format:"email" minLength:"1" maxLength:"255"`
	State string `json:"state" doc:"State token" minLength:"1"`
}

// OAuthCompleteEmailInput wraps the request body.
type OAuthCompleteEmailInput struct {
	Body OAuthCompleteEmailRequest
}

// OAuthLinkResponse is the response body for linking OAuth provider.
type OAuthLinkResponse struct {
	Provider string    `json:"provider"`
	LinkedAt time.Time `json:"linkedAt"`
}

// OAuthLinkOutput is the response for linking OAuth provider.
type OAuthLinkOutput struct {
	Body OAuthLinkResponse
}

// OAuthIdentityDTO represents a linked OAuth identity.
type OAuthIdentityDTO struct {
	Provider      string     `json:"provider"`
	ProviderEmail string     `json:"providerEmail,omitempty"`
	LinkedAt      *time.Time `json:"linkedAt,omitempty"`
	LastUsedAt    *time.Time `json:"lastUsedAt,omitempty"`
}

// OAuthIdentitiesResponse is the response for listing OAuth identities.
type OAuthIdentitiesResponse struct {
	Identities []OAuthIdentityDTO `json:"identities"`
}

// OAuthIdentitiesOutput is the response for listing OAuth identities.
type OAuthIdentitiesOutput struct {
	Body OAuthIdentitiesResponse
}

// OAuthUnlinkOutput is the response for unlinking OAuth provider.
type OAuthUnlinkOutput struct {
	Body struct {
		Unlinked string `json:"unlinked"`
	}
}
