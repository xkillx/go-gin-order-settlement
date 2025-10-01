package validation

import (
	"github.com/xkillx/go-gin-order-settlement/modules/order/dto"
	"github.com/go-playground/validator/v10"
)

type OrderValidation struct {
	validate *validator.Validate
}

func NewOrderValidation() *OrderValidation {
	return &OrderValidation{validate: validator.New()}
}

func (v *OrderValidation) ValidateOrderCreateRequest(req dto.OrderCreateRequest) error {
	return v.validate.Struct(req)
}
