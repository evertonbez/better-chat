package middleware

import (
	"errors"
	"log/slog"

	"evertonbez/better-chat/internal/apperr"
	"evertonbez/better-chat/internal/config"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

type ErrorResponse struct {
	Code      apperr.Code    `json:"code"`
	Message   string         `json:"message"`
	Details   map[string]any `json:"details,omitempty"`
	RequestID string         `json:"requestId,omitempty"`
}

func ErrorHandler(c fiber.Ctx, err error) error {
	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {

		return c.Status(fiberErr.Code).JSON(ErrorResponse{
			Code:      apperr.CodeInvalidRequest,
			Message:   fiberErr.Message,
			RequestID: requestid.FromContext(c),
		})
	}

	appErr := apperr.From(err)

	if appErr.Status >= fiber.StatusInternalServerError {
		slog.Error("request failed",
			"error", appErr.Error(),
			"code", appErr.Code,
			"path", c.Path(),
			"method", c.Method(),
			"request_id", requestid.FromContext(c),
		)
	}

	res := ErrorResponse{
		Code:      appErr.Code,
		Message:   appErr.Message,
		Details:   appErr.Details,
		RequestID: requestid.FromContext(c),
	}

	if config.IsDevelopment() && appErr.Unwrap() != nil {
		if res.Details == nil {
			res.Details = map[string]any{}
		}
		res.Details["cause"] = appErr.Unwrap().Error()
	}

	return c.Status(appErr.Status).JSON(res)
}
