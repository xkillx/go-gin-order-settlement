package dto

import "errors"

const (
	// Failed
	MESSAGE_FAILED_GET_DATA_FROM_BODY = "failed get data from body"
	MESSAGE_FAILED_CREATE_ORDER       = "failed create order"
	MESSAGE_FAILED_GET_ORDER          = "failed get order"
	MESSAGE_FAILED_GET_LIST_ORDER     = "failed get list order"
	MESSAGE_FAILED_DELETE_ORDER       = "failed delete order"
	MESSAGE_FAILED_PROSES_REQUEST     = "failed proses request"

	// Success
	MESSAGE_SUCCESS_CREATE_ORDER   = "success create order"
	MESSAGE_SUCCESS_GET_ORDER      = "success get order"
	MESSAGE_SUCCESS_GET_LIST_ORDER = "success get list order"
	MESSAGE_SUCCESS_DELETE_ORDER   = "success delete order"
)

var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrInsufficientStock  = errors.New("insufficient stock")
)

type (
	OrderCreateRequest struct {
		ProductID string `json:"product_id" form:"product_id" binding:"required,uuid4"`
		BuyerID   string `json:"buyer_id" form:"buyer_id" binding:"required,min=1"`
		Quantity  int    `json:"quantity" form:"quantity" binding:"required,min=1"`
	}

	OrderResponse struct {
		ID        string `json:"id"`
		ProductID string `json:"product_id"`
		BuyerID   string `json:"buyer_id"`
		Quantity  int    `json:"quantity"`
	}
)
