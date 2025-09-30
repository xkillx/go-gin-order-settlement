package controller

import (
	"net/http"

	"github.com/Caknoooo/go-gin-clean-starter/modules/order/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/order/service"
	"github.com/Caknoooo/go-gin-clean-starter/modules/order/validation"
	pkgdto "github.com/Caknoooo/go-gin-clean-starter/pkg/dto"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

type (
	OrderController interface {
		Create(ctx *gin.Context)
		GetByID(ctx *gin.Context)
		List(ctx *gin.Context)
		Delete(ctx *gin.Context)
	}

	orderController struct {
		service   service.OrderService
		validate  *validation.OrderValidation
	}
)

func NewOrderController(_ *do.Injector, s service.OrderService) OrderController {
	return &orderController{
		service:  s,
		validate: validation.NewOrderValidation(),
	}
}

func (c *orderController) Create(ctx *gin.Context) {
	var req dto.OrderCreateRequest
	if err := ctx.ShouldBind(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	if err := c.validate.ValidateOrderCreateRequest(req); err != nil {
		res := utils.BuildResponseFailed("Validation failed", err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.service.Create(ctx.Request.Context(), req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_CREATE_ORDER, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_CREATE_ORDER, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *orderController) GetByID(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := c.service.GetByID(ctx.Request.Context(), id)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_ORDER, err.Error(), nil)
		ctx.JSON(http.StatusNotFound, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_ORDER, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *orderController) List(ctx *gin.Context) {
	var p pkgdto.PaginationRequest
	if err := ctx.ShouldBindQuery(&p); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_PROSES_REQUEST, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}
	p.Default()

	items, meta, err := c.service.List(ctx.Request.Context(), p)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_LIST_ORDER, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	payload := gin.H{
		"items":      items,
		"pagination": meta,
	}
	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_LIST_ORDER, payload)
	ctx.JSON(http.StatusOK, res)
}

func (c *orderController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.service.Delete(ctx.Request.Context(), id); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_DELETE_ORDER, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_DELETE_ORDER, nil)
	ctx.JSON(http.StatusOK, res)
}
