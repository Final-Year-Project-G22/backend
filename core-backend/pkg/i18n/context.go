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

	return GetDefaultLocale()
}

// T is a convenience function that resolves a message key using the locale from context.
// It wraps Resolve(key, LocaleFromContext(ctx), params...).
func T(ctx context.Context, key string, params ...map[string]string) string {
	return Resolve(key, LocaleFromContext(ctx), params...)
}
