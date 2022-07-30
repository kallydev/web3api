package validator

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/go-playground/validator/v10"
)

func ValidateEthereumAddress(fieldLevel validator.FieldLevel) bool {
	var address common.Address

	if err := address.UnmarshalText([]byte(fieldLevel.Field().String())); err != nil {
		return false
	}

	return true
}
