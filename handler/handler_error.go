package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func (i *internal) Error(err error, c echo.Context) {
	_ = c.JSONPretty(http.StatusInternalServerError, &ErrorResponse{
		Error: err.Error(),
	}, "\t")
}
