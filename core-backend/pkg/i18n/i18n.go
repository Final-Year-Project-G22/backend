// Package i18n provides internationalization support for the application.
// It loads message files and resolves messages based on locale.
package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	messages      = make(map[string]map[string]string)
	mu            sync.RWMutex
	defaultLocale = "en"
	loaded        = false
)

// Init loads all message files from the messages directory.
func Init(messagesPath string) error {
	mu.Lock()
	defer mu.Unlock()

	if loaded {
		return nil
	}

	if messagesPath == "" {
		messagesPath = "pkg/i18n/messages"
	}

	absMessagesPath, err := filepath.Abs(messagesPath)
	if err != nil {
		return fmt.Errorf("failed to resolve messages path: %w", err)
	}

	files, err := os.ReadDir(absMessagesPath)
	if err != nil {
		return fmt.Errorf("failed to read messages directory: %w", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		locale := strings.TrimSuffix(file.Name(), ".json")

		if strings.Contains(file.Name(), "..") || filepath.IsAbs(file.Name()) {
			continue
		}

		fullPath := filepath.Join(absMessagesPath, file.Name())
		absFullPath, err := filepath.Abs(fullPath)
		if err != nil {
			continue
		}
		absMessagesPathAbs, err := filepath.Abs(absMessagesPath)
		if err != nil {
			continue
		}
		if !strings.HasPrefix(absFullPath, absMessagesPathAbs+string(filepath.Separator)) {
			continue
		}

		data, err := os.ReadFile(fullPath) // nolint:gosec
		if err != nil {
			return fmt.Errorf("failed to read message file %s: %w", file.Name(), err)
		}

		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			return fmt.Errorf("failed to parse message file %s: %w", file.Name(), err)
		}

		flattened := make(map[string]string)
		flattenMessages("", raw, flattened)

		messages[locale] = flattened
	}

	loaded = true

	if _, ok := messages[defaultLocale]; !ok {
		return fmt.Errorf("default locale %s not found", defaultLocale)
	}

	return nil
}

func flattenMessages(prefix string, input map[string]any, out map[string]string) {
	for key, value := range input {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case string:
			out[fullKey] = v
		case map[string]any:
			flattenMessages(fullKey, v, out)
		}
	}
}

// Resolve returns the message for the given key and locale.
// Falls back to default locale if the locale is not found.
// Supports template placeholders like {{field}} and {{resource}}.
func Resolve(key string, locale string, params ...map[string]string) string {
	mu.RLock()
	defer mu.RUnlock()

	locale = normalizeLocale(locale)

	msg := getMessage(locale, key)
	if msg == "" {
		msg = getMessage(defaultLocale, key)
	}

	if msg == "" {
		return key
	}

	if len(params) > 0 {
		for k, v := range params[0] {
			msg = strings.ReplaceAll(msg, fmt.Sprintf("{{%s}}", k), v)
		}
	}

	return msg
}

// getMessage retrieves a message from the specified locale.
func getMessage(locale, key string) string {
	if localeMsgs, ok := messages[locale]; ok {
		if msg, ok := localeMsgs[key]; ok {
			return msg
		}
	}
	return ""
}

// normalizeLocale normalizes the locale string (e.g., "en-US" -> "en").
func normalizeLocale(locale string) string {
	if locale == "" {
		return defaultLocale
	}
	parts := strings.Split(locale, "-")
	return parts[0]
}

// GetSupportedLocales returns a list of supported locale codes.
func GetSupportedLocales() []string {
	mu.RLock()
	defer mu.RUnlock()

	locales := make([]string, 0, len(messages))
	for locale := range messages {
		locales = append(locales, locale)
	}
	return locales
}

// HasLocale checks if a locale is supported.
func HasLocale(locale string) bool {
	mu.RLock()
	defer mu.RUnlock()

	_, ok := messages[normalizeLocale(locale)]
	return ok
}

// GetDefaultLocale returns the default locale code.
func GetDefaultLocale() string {
	return defaultLocale
}

// SetDefaultLocale sets the default locale.
func SetDefaultLocale(locale string) {
	mu.Lock()
	defer mu.Unlock()
	defaultLocale = locale
}
