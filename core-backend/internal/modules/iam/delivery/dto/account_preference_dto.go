package dto

// --- Account Preferences ---

type AccountPreferenceResponse struct {
	Language string `json:"language" doc:"User's language preference (en, am)"`
	Timezone string `json:"timezone" doc:"User's timezone (e.g. UTC, Africa/Addis_Ababa)"`
}

type UpdateAccountPreferenceRequest struct {
	Language *string `json:"language,omitempty" doc:"Language code (en, am)" maxLength:"10"`
}

type UpdateAccountPreferenceInput struct {
	Body UpdateAccountPreferenceRequest
}

type UpdateAccountPreferenceOutput struct {
	Body AccountPreferenceResponse
}

type GetAccountPreferenceOutput struct {
	Body AccountPreferenceResponse
}
