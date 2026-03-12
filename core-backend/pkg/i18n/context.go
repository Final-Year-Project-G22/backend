package i18n

import "context"

type localeKey struct{}

var contextLocaleKey = localeKey{}

func WithLocale(ctx context.Context, locale string) context.Context {
	return context.WithValue(ctx, contextLocaleKey, locale)
}

func LocaleFromContext(ctx context.Context) string {
	if locale, ok := ctx.Value(contextLocaleKey).(string); ok && locale != "" {
		return locale
	}

	return "en"
}
