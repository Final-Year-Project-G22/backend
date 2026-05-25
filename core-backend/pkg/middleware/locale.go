package middleware

import (
	"strings"

	"github.com/Final-Year-Project-G22/backend/core/pkg/i18n"
	"github.com/danielgtaylor/huma/v2"
)

func LocaleResolver() func(ctx huma.Context, next func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		raw := ctx.Header("Accept-Language")
		locale := normalizeLocale(raw)
		enriched := i18n.WithLocale(ctx.Context(), locale)
		ctx = huma.WithContext(ctx, enriched)
		next(ctx)
	}
}

func normalizeLocale(raw string) string {
	if raw == "" {
		return "en"
	}
	base := strings.Split(strings.ToLower(strings.TrimSpace(raw)), "-")[0]
	switch base {
	case "en":
		return "en"
	case "am":
		return "am"
	default:
		return "en"
	}
}
