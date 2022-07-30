package validator

import (
	"github.com/go-playground/validator/v10"
	"github.com/kallydev/web3api/common/ethereum"
)

func ValidateNetwork(fieldLevel validator.FieldLevel) bool {
	switch fieldLevel.Field().String() {
	case ethereum.NetworkEthereum, ethereum.NetworkPolygon, ethereum.NetworkOptimism, ethereum.NetworkArbitrum:
		return true
	default:
		return false
	}
}
