package entity

type ThreadStatus string

const (
	ThreadStatusActive   ThreadStatus = "active"
	ThreadStatusClosed   ThreadStatus = "closed"
	ThreadStatusArchived ThreadStatus = "archived"
	ThreadStatusHidden   ThreadStatus = "hidden"
)

type TargetType string

const (
	TargetTypeThread TargetType = "thread"
	TargetTypePost   TargetType = "post"
	TargetTypeUser   TargetType = "user"
	TargetTypeHidden TargetType = "hidden"
)

type ReportStatus string

const (
	ReportStatusPending     ReportStatus = "pending"
	ReportStatusUnderReview ReportStatus = "under_review"
	ReportStatusResolved    ReportStatus = "resolved"
	ReportStatusDismissed   ReportStatus = "dismissed"
)
