package entity

type NotificationType string

const (
	NotificationTypeSystemAnnouncement   NotificationType = "system_announcement"
	NotificationTypePolicyUpdate         NotificationType = "policy_update"
	NotificationTypeWelcomeMessage       NotificationType = "welcome_message"
	NotificationTypeCommunityReply       NotificationType = "community_reply"
	NotificationTypeCommunitySolution    NotificationType = "community_solution"
	NotificationTypeCommunityMention     NotificationType = "community_mention"
	NotificationTypeGuideStepCompleted   NotificationType = "guide_step_completed"
	NotificationTypeGuideDeadline        NotificationType = "guide_deadline"
	NotificationTypeGuideUpdate          NotificationType = "guide_update"
	NotificationTypeAIQuotaLimit         NotificationType = "ai_quota_limit"
	NotificationTypeAIResponseReady      NotificationType = "ai_response_ready"
	NotificationTypeAccountAlert         NotificationType = "account_alert"
	NotificationTypeAccountAlertCritical NotificationType = "account_alert_critical"
	NotificationTypeAccountAlertInfo     NotificationType = "account_alert_info"
	NotificationTypeAccountVerification  NotificationType = "account_verification"
	NotificationTypePasswordReset        NotificationType = "password_reset"
	NotificationTypePaymentConfirmation  NotificationType = "payment_confirmation"
	NotificationTypeCampaign             NotificationType = "campaign"
	NotificationTypeUserScheduled        NotificationType = "user_scheduled"
	NotificationTypeComplianceInfo       NotificationType = "compliance_info"
)

type NotificationPriority int8

const (
	NotificationPriorityLow    NotificationPriority = 0
	NotificationPriorityMedium NotificationPriority = 1
	NotificationPriorityHigh   NotificationPriority = 2
	NotificationPriorityUrgent NotificationPriority = 3
)

type Channel string

const (
	ChannelInApp Channel = "in_app"
	ChannelEmail Channel = "email"
	ChannelPush  Channel = "push"
	ChannelSMS   Channel = "sms"
)

type NotificationStatus string

const (
	NotificationStatusPending    NotificationStatus = "pending"
	NotificationStatusProcessing NotificationStatus = "processing"
	NotificationStatusDelivered  NotificationStatus = "delivered"
	NotificationStatusFailed     NotificationStatus = "failed"
	NotificationStatusCancelled  NotificationStatus = "cancelled"
)

type DeliveryStatus string

const (
	DeliveryStatusSent      DeliveryStatus = "sent"
	DeliveryStatusDelivered DeliveryStatus = "delivered"
	DeliveryStatusFailed    DeliveryStatus = "failed"
	DeliveryStatusBounced   DeliveryStatus = "bounced"
)

type CampaignType string

const (
	CampaignTypeBroadcast CampaignType = "broadcast"
	CampaignTypeSegmented CampaignType = "segmented"
)

type CampaignStatus string

const (
	CampaignStatusDraft     CampaignStatus = "draft"
	CampaignStatusScheduled CampaignStatus = "scheduled"
	CampaignStatusSending   CampaignStatus = "sending"
	CampaignStatusCompleted CampaignStatus = "completed"
	CampaignStatusCancelled CampaignStatus = "cancelled"
)

type ScheduleStatus string

const (
	ScheduleStatusPending   ScheduleStatus = "pending"
	ScheduleStatusSent      ScheduleStatus = "sent"
	ScheduleStatusCancelled ScheduleStatus = "cancelled"
)

type ComplianceType string

const (
	ComplianceTypeTaxRegistration      ComplianceType = "tax_registration"
	ComplianceTypeTradeLicense         ComplianceType = "trade_license"
	ComplianceTypeBusinessRegistration ComplianceType = "business_registration"
)

type ComplianceEntryStatus string

const (
	ComplianceEntryStatusActive  ComplianceEntryStatus = "active"
	ComplianceEntryStatusExpired ComplianceEntryStatus = "expired"
	ComplianceEntryStatusRenewed ComplianceEntryStatus = "renewed"
)

type ComplianceSource string

const (
	ComplianceSourceAuto   ComplianceSource = "auto"
	ComplianceSourceManual ComplianceSource = "manual"
)

type DeviceType string

const (
	DeviceTypeAndroid DeviceType = "android"
	DeviceTypeIOS     DeviceType = "ios"
	DeviceTypeWeb     DeviceType = "web"
)
