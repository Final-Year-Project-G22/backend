package event

type UserEmailOTPRequestedEvent struct {
	AccountID      string `json:"account_id"`
	Email          string `json:"email"`
	FirstName      string `json:"first_name"`
	OTPCode        string `json:"otp_code"`
	ExpiresMinutes int    `json:"expires_minutes"`
	Locale         string `json:"locale"`
}
