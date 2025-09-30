package dto

import "errors"

const (
	// Failed
	MESSAGE_FAILED_GET_DATA_FROM_BODY = "failed get data from body"
	MESSAGE_FAILED_CREATE_PRODUCT     = "failed create product"
	MESSAGE_FAILED_GET_PRODUCT        = "failed get product"
	MESSAGE_FAILED_GET_LIST_PRODUCT   = "failed get list product"
	MESSAGE_FAILED_UPDATE_PRODUCT     = "failed update product"
	MESSAGE_FAILED_DELETE_PRODUCT     = "failed delete product"
	MESSAGE_FAILED_PROSES_REQUEST     = "failed proses request"

	// Success
	MESSAGE_SUCCESS_CREATE_PRODUCT   = "success create product"
	MESSAGE_SUCCESS_GET_PRODUCT      = "success get product"
	MESSAGE_SUCCESS_GET_LIST_PRODUCT = "success get list product"
	MESSAGE_SUCCESS_UPDATE_PRODUCT   = "success update product"
	MESSAGE_SUCCESS_DELETE_PRODUCT   = "success delete product"
)

var (
	ErrProductNotFound    = errors.New("product not found")
	ErrInsufficientStock  = errors.New("insufficient stock")
	ErrFailedCreate       = errors.New("failed to create product")
	ErrFailedUpdate       = errors.New("failed to update product")
	ErrFailedDelete       = errors.New("failed to delete product")
)

type (
	ProductCreateRequest struct {
		Name  string `json:"name" form:"name" binding:"required,min=2"`
		Stock int    `json:"stock" form:"stock" binding:"required,min=0"`
	}

	ProductUpdateRequest struct {
		Name  string `json:"name" form:"name" binding:"omitempty,min=2"`
		Stock *int   `json:"stock" form:"stock" binding:"omitempty,min=0"`
	}

	ProductResponse struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Stock int    `json:"stock"`
	}
)
