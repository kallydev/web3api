package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func ZapLogger(logger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			fields := make([]zap.Field, 0)

			if err := next(c); err != nil {
				c.Error(err)

				fields = append(fields, zap.Error(err))
			}

			status := c.Response().Status

			fields = append(
				fields,
				zap.Int("status", c.Response().Status),
				zap.String("method", c.Request().Method),
				zap.String("uri", c.Request().RequestURI),
				zap.String("client_ip", c.RealIP()),
			)

			switch {
			case status >= http.StatusInternalServerError:
				logger.Error("an error occurred while processing the request", fields...)
			case status >= http.StatusBadRequest:
				logger.Warn("invalid request", fields...)
			case status >= http.StatusMultipleChoices:
				logger.Info("the request was redirected", fields...)
			default:
				logger.Info("the request was processed", fields...)
			}

			return nil
		}
	}
}
