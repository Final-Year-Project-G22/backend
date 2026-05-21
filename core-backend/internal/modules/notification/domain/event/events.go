package event

const (
	AccountRegistered   = "account.registered"
	AccountVerification = "account.verification"
	PasswordReset       = "password.reset"
	AccountAlert        = "account.alert"
	WelcomeMessage      = "welcome.message"

	ThreadReply    = "thread.reply"
	ThreadSolution = "thread.solution"
	ThreadMention  = "thread.mention"

	GuideStepCompleted = "guide.step_completed"
	GuideDeadline      = "guide.deadline"
	GuideUpdate        = "guide.update"

	AIQuotaLimit    = "ai.quota_limit"
	AIResponseReady = "ai.response_ready"

	PaymentConfirmation = "payment.confirmation"

	UserEmailOTPRequested = "user.email_otp_requested"

	BusinessProfileUpdated        = "business.profile.updated"
	GuideComplianceStepCompleted  = "guide.compliance_step_completed"
	GuideComplianceStepRolledBack = "guide.compliance_step_rolled_back"
	ComplianceAlert               = "compliance.alert"

	NotificationFailed = "notification.failed"
)
