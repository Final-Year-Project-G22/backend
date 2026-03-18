package core

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"strings"
)

var envPlaceholderPattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)(?:(:-|:\?)([^}]*))?\}`)

func resolvePlaceholdersInString(input string, debug bool) (string, error) {
	if input == "" {
		return input, nil
	}

	var unresolved []string

	result := envPlaceholderPattern.ReplaceAllStringFunc(input, func(match string) string {
		parts := envPlaceholderPattern.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}

		envKey := parts[1]
		sep := parts[2]
		value := parts[3]

		if envVal, ok := os.LookupEnv(envKey); ok {
			return envVal
		}

		if sep == ":?" {
			unresolved = append(unresolved, fmt.Sprintf("${%s:?%s}", envKey, value))
			return match
		}

		if sep == ":-" {
			return value
		}

		if debug {
			unresolved = append(unresolved, fmt.Sprintf("${%s}", envKey))
		}
		return ""
	})

	if len(unresolved) > 0 {
		return result, fmt.Errorf("unresolved required placeholders: %s", strings.Join(unresolved, ", "))
	}

	return result, nil
}

func resolvePlaceholdersInAny(v reflect.Value, debug bool) error {
	if !v.IsValid() {
		return nil
	}

	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			if !field.CanSet() && field.Kind() != reflect.Struct && field.Kind() != reflect.Ptr && field.Kind() != reflect.Slice && field.Kind() != reflect.Map {
				continue
			}
			if err := resolvePlaceholdersInAny(field, debug); err != nil {
				return err
			}
		}
		return nil

	case reflect.String:
		if v.CanSet() {
			resolved, err := resolvePlaceholdersInString(v.String(), debug)
			if err != nil {
				return err
			}
			v.SetString(resolved)
		}
		return nil

	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if err := resolvePlaceholdersInAny(v.Index(i), debug); err != nil {
				return err
			}
		}
		return nil

	case reflect.Map:
		if v.IsNil() {
			return nil
		}

		iter := v.MapRange()
		for iter.Next() {
			key := iter.Key()
			val := iter.Value()

			if val.Kind() == reflect.String {
				resolved, err := resolvePlaceholdersInString(val.String(), debug)
				if err != nil {
					return err
				}
				v.SetMapIndex(key, reflect.ValueOf(resolved).Convert(val.Type()))
				continue
			}

			clone := reflect.New(val.Type()).Elem()
			clone.Set(val)
			if err := resolvePlaceholdersInAny(clone, debug); err != nil {
				return err
			}
			v.SetMapIndex(key, clone)
		}
		return nil

	default:
		return nil
	}
}

func resolveConfigPlaceholders(cfg any, debug bool) error {
	if cfg == nil {
		return fmt.Errorf("nil config passed to resolveConfigPlaceholders")
	}

	val := reflect.ValueOf(cfg)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return fmt.Errorf("resolveConfigPlaceholders expects a non-nil pointer")
	}

	err := resolvePlaceholdersInAny(val, debug)
	if err != nil {
		if debug {
			log.Printf("[config] placeholder resolution error: %v", err)
		}
		return err
	}

	return nil
}
