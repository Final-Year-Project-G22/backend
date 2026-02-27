package utils

import (
	"fmt"
	"regexp"
)

// Define regex patterns once for efficiency.

// SafeStringRegex allows letters, numbers, hyphens, underscores, and dots.
// It is ideal for environment names, basic flags, or simple config values.
// It explicitly disallows shell metacharacters like ;, |, &, >, <, $, (, ), etc.
var SafeStringRegex = regexp.MustCompile(`^[a-zA-Z0-9-._]+$`)

// SafeDSNComponentRegex allows characters common in database connection strings
// (including colons, slashes, and '@' symbols) but still disallows shell metacharacters.
var SafeDSNComponentRegex = regexp.MustCompile(`^[a-zA-Z0-9-._:/@]+$`)

// IsSafeString checks if the input string contains only safe characters.
func IsSafeString(input string) error {
	if !SafeStringRegex.MatchString(input) {
		return fmt.Errorf("input string '%s' contains forbidden shell characters", input)
	}
	return nil
}

// IsSafeDSNComponent checks if the input string is safe to use as a database
// connection component (host, dbname, etc.).
func IsSafeDSNComponent(input string) error {
	if !SafeDSNComponentRegex.MatchString(input) {
		return fmt.Errorf("DSN component '%s' contains forbidden shell characters", input)
	}
	return nil
}
