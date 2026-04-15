package error

import "errors"

var (
	ErrCategoryNotFound          = errors.New("community: category not found")
	ErrThreadNotFound            = errors.New("community: thread not found")
	ErrPostNotFound              = errors.New("community: post not found")
	ErrThreadSettingNotFound     = errors.New("community: thread setting not found")
	ErrCategorySettingNotFound   = errors.New("community: category setting not found")
	ErrReportNotFound            = errors.New("community: report not found")
	ErrThreadBlockedUserNotFound = errors.New("community: blocked user not found")
	ErrThreadInactive            = errors.New("community: thread not active")
	ErrThreadBlocked             = errors.New("community: user blocked from thread")
	ErrPostEditWindowExpired     = errors.New("community: post edit window expired")
	ErrPostDeleteWindowExpired   = errors.New("community: post delete window expired")
	ErrPostSolutionLocked        = errors.New("community: post locked by solution")
)
