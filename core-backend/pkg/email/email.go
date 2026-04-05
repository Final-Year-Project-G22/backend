package email

import (
	"context"
)

type Emailer interface {
	Send(ctx context.Context, input SendInput) error
	SendTemplate(ctx context.Context, input SendTemplateInput) error
}

type SendInput struct {
	To       []string
	CC       []string
	BCC      []string
	Subject  string
	Body     string
	IsHTML   bool
	From     string
	FromName string
}

type SendTemplateInput struct {
	To       []string
	CC       []string
	BCC      []string
	Subject  string
	Template string
	Data     interface{}
	Locale   string
	From     string
	FromName string
}

type I18nBody struct {
	Key    string
	Locale string
	Params map[string]string
}

type BodyType string

const (
	BodyTypeText BodyType = "text"
	BodyTypeHTML BodyType = "html"
	BodyTypeI18n BodyType = "i18n"
)
