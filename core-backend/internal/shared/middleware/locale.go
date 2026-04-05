package middleware

import (
	"strings"

	"github.com/Final-Year-Project-G22/backend/core/internal/shared/constants"
	"github.com/gin-gonic/gin"
)

const LocaleKey = "locale"

func LocaleResolver() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.Query("locale")
		if raw == "" {
			raw = c.GetHeader("Accept-Language")
		}

		locale := NormalizeLocale(raw)
		c.Set(LocaleKey, locale)
		c.Next()
	}
}

func NormalizeLocale(raw string) constants.Locale {
	if raw == "" {
		return constants.LocaleEnglish
	}
	base := strings.Split(strings.ToLower(strings.TrimSpace(raw)), "-")[0]
	switch base {
	case "en":
		return constants.LocaleEnglish
	case "am":
		return constants.LocaleAmharic
	default:
		return constants.LocaleEnglish
	}
}
