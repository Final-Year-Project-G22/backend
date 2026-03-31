package error

import "errors"

var (
	ErrUserNotFound                 = errors.New("iam: user not found")
	ErrAccountNotFound              = errors.New("iam: account not found")
	ErrAccountAlreadyExists         = errors.New("iam: account already exists")
	ErrBusinessProfileNotFound      = errors.New("iam: business profile not found")
	ErrBusinessProfileAlreadyExists = errors.New("iam: business profile already exists")
	ErrOAuthIdentityNotFound        = errors.New("iam: oauth identity not found")
	ErrOAuthIdentityAlreadyLinked   = errors.New("iam: oauth identity already linked")
	ErrRoleNotFound                 = errors.New("iam: role not found")
	ErrRoleAlreadyAssigned          = errors.New("iam: role already assigned")
	ErrRoleNotAssigned              = errors.New("iam: role not assigned")
	ErrPermissionNotFound           = errors.New("iam: permission not found")
	ErrSessionNotFound              = errors.New("iam: session not found")
	ErrSessionRevoked               = errors.New("iam: session revoked")
	ErrEmailOTPNotFound             = errors.New("iam: email otp not found")
	ErrEmailOTPExpired              = errors.New("iam: email otp expired")
	ErrEmailOTPAlreadyConsumed      = errors.New("iam: email otp already consumed")
)
