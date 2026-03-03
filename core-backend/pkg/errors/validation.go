// Package errors provides validation error handling integration with go-playground/validator.
package errors

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/Final-Year-Project-G22/backend/core/pkg/i18n"
	"github.com/go-playground/validator/v10"
)

// FieldValidationError holds validation error details for a single field.
type FieldValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Tag     string `json:"tag"`
	Value   any    `json:"value,omitempty"`
}

// FieldValidationErrors holds a collection of field validation errors.
type FieldValidationErrors struct {
	Errors []FieldValidationError `json:"errors"`
}

// HandleValidationErrors converts go-playground/validator errors to AppError.
// It maps validation tags to i18n keys for localized messages.
func HandleValidationErrors(err error) *AppError {
	if err == nil {
		return nil
	}

	validationErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return InternalError("errors.databaseError", err)
	}

	var validationErrors FieldValidationErrors
	for _, fe := range validationErrs {
		fieldName := fe.Field()
		tag := fe.Tag()
		value := fe.Value()

		msgKey := getValidationMessageKey(fieldName, tag)
		message := i18n.Resolve(msgKey, i18n.GetDefaultLocale(), map[string]string{
			"field": fieldName,
			"value": fmt.Sprintf("%v", value),
		})

		validationErrors.Errors = append(validationErrors.Errors, FieldValidationError{
			Field:   fieldName,
			Message: message,
			Tag:     tag,
			Value:   value,
		})
	}

	return &AppError{
		Code:    ErrCodeValidation,
		Message: "errors.validationError",
		Status:  GetStatus(ErrCodeValidation),
		Details: validationErrors,
	}
}

// HandleBindingErrors handles gin binding errors.
func HandleBindingErrors(err error) *AppError {
	if err == nil {
		return nil
	}

	errMsg := err.Error()
	if strings.Contains(errMsg, "EOF") {
		return BadRequestError("errors.badRequest")
	}

	return BadRequestError("errors.badRequest").WithDetails(map[string]string{
		"details": errMsg,
	})
}

// getValidationMessageKey returns the i18n key for a validation error.
func getValidationMessageKey(field, tag string) string {
	fieldLower := strings.ToLower(field)

	switch tag {
	case "required":
		return "errors.requiredField"
	case "email":
		return "errors.invalidEmail"
	case "min":
		return "errors.invalidPassword"
	case "max":
		return "errors.invalidInput"
	case "eqfield":
		return "errors.passwordMismatch"
	default:
		return fmt.Sprintf("errors.validation.%s.%s", fieldLower, tag)
	}
}

// GetValidationErrors extracts validation errors from a struct using validator.
func GetValidationErrors(v any) FieldValidationErrors {
	validate := validator.New()
	err := validate.Struct(v)
	if err == nil {
		return FieldValidationErrors{}
	}

	return handleValidatorErrors(err)
}

func handleValidatorErrors(err error) FieldValidationErrors {
	validationErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return FieldValidationErrors{}
	}

	var errors FieldValidationErrors
	for _, fe := range validationErrs {
		errors.Errors = append(errors.Errors, FieldValidationError{
			Field:   fe.Field(),
			Message: fe.Error(),
			Tag:     fe.Tag(),
			Value:   fe.Value(),
		})
	}

	return errors
}

// AddValidationError adds a custom validation error.
func AddValidationError(field, message string) *FieldValidationErrors {
	return &FieldValidationErrors{
		Errors: []FieldValidationError{
			{
				Field:   field,
				Message: message,
			},
		},
	}
}

// FormatValidationErrors formats validation errors for display.
func FormatValidationErrors(errs FieldValidationErrors, locale string) []map[string]string {
	var result []map[string]string
	for _, e := range errs.Errors {
		msg := e.Message
		if !strings.HasPrefix(msg, "errors.") {
			msg = i18n.Resolve(e.Message, locale, map[string]string{
				"field": e.Field,
			})
		}
		result = append(result, map[string]string{
			"field":   e.Field,
			"message": msg,
			"tag":     e.Tag,
		})
	}
	return result
}

// GetFieldName returns the JSON field name from a struct field.
// It uses reflection to get the json tag.
func GetFieldName(v any, field string) string {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Name == field {
			tag := f.Tag.Get("json")
			if tag != "" && tag != "-" {
				return strings.Split(tag, ",")[0]
			}
			return strings.ToLower(field[:1]) + field[1:]
		}
	}
	return field
}
