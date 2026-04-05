package event

type AdminCreatedEvent struct {
	AccountID string `json:"account_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Password  string `json:"password"`
	Locale    string `json:"locale"`
}
