package validator

import (
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

var _ echo.Validator = (*internal)(nil)

type internal struct {
	validate *validator.Validate
}

func (i *internal) Validate(value interface{}) error {
	return i.validate.Struct(value)
}

func New() (echo.Validator, error) {
	validate := validator.New()

	if err := validate.RegisterValidation("network", ValidateNetwork); err != nil {
		return nil, err
	}

	if err := validate.RegisterValidation("ethereum_address", ValidateEthereumAddress); err != nil {
		return nil, err
	}

	return &internal{
		validate: validate,
	}, nil
}
