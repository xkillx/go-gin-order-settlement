package validation

import (
	"github.com/Caknoooo/go-gin-clean-starter/modules/product/dto"
	"github.com/go-playground/validator/v10"
)

type ProductValidation struct {
	validate *validator.Validate
}

func NewProductValidation() *ProductValidation {
	return &ProductValidation{validate: validator.New()}
}

func (v *ProductValidation) ValidateProductCreateRequest(req dto.ProductCreateRequest) error {
	return v.validate.Struct(req)
}

func (v *ProductValidation) ValidateProductUpdateRequest(req dto.ProductUpdateRequest) error {
	return v.validate.Struct(req)
}
