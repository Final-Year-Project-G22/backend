package errors

import (
	"context"
	"log"

	"github.com/Final-Year-Project-G22/backend/core/pkg/i18n"
	"github.com/danielgtaylor/huma/v2"
)

func ToHumaError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	locale := i18n.LocaleFromContext(ctx)

	if appErr, ok := err.(*AppError); ok {
		message := appErr.GetMessage(locale)
		if appErr.Err != nil {
			log.Printf("huma adapter app error: status=%d code=%s message=%q wrapped_err=%v", appErr.GetStatus(), appErr.Code, appErr.Message, appErr.Err)
			return huma.NewError(appErr.GetStatus(), message, appErr.Err)
		}
		return huma.NewError(appErr.GetStatus(), message)
	}

	if statusErr, ok := err.(huma.StatusError); ok {
		return statusErr
	}

	message := i18n.Resolve("errors.internalError", locale)
	return huma.NewError(500, message)
}
