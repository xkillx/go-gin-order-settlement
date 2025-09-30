package controller

import (
	"net/http"

	"github.com/Caknoooo/go-gin-clean-starter/modules/product/dto"
	"github.com/Caknoooo/go-gin-clean-starter/modules/product/service"
	"github.com/Caknoooo/go-gin-clean-starter/modules/product/validation"
	pkgdto "github.com/Caknoooo/go-gin-clean-starter/pkg/dto"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

type (
	ProductController interface {
		Create(ctx *gin.Context)
		GetByID(ctx *gin.Context)
		List(ctx *gin.Context)
		Update(ctx *gin.Context)
		Delete(ctx *gin.Context)
	}

	productController struct {
		service    service.ProductService
		validator  *validation.ProductValidation
	}
)

func NewProductController(_ *do.Injector, s service.ProductService) ProductController {
	return &productController{
		service:   s,
		validator: validation.NewProductValidation(),
	}
}

func (c *productController) Create(ctx *gin.Context) {
	var req dto.ProductCreateRequest
	if err := ctx.ShouldBind(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	if err := c.validator.ValidateProductCreateRequest(req); err != nil {
		res := utils.BuildResponseFailed("Validation failed", err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.service.Create(ctx.Request.Context(), req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_CREATE_PRODUCT, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_CREATE_PRODUCT, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *productController) GetByID(ctx *gin.Context) {
	id := ctx.Param("id")
	result, err := c.service.GetByID(ctx.Request.Context(), id)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_PRODUCT, err.Error(), nil)
		ctx.JSON(http.StatusNotFound, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_PRODUCT, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *productController) List(ctx *gin.Context) {
	var p pkgdto.PaginationRequest
	if err := ctx.ShouldBindQuery(&p); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_PROSES_REQUEST, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}
	p.Default()

	items, meta, err := c.service.List(ctx.Request.Context(), p)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_LIST_PRODUCT, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	payload := gin.H{
		"items":      items,
		"pagination": meta,
	}
	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_GET_LIST_PRODUCT, payload)
	ctx.JSON(http.StatusOK, res)
}

func (c *productController) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	var req dto.ProductUpdateRequest
	if err := ctx.ShouldBind(&req); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_GET_DATA_FROM_BODY, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	if err := c.validator.ValidateProductUpdateRequest(req); err != nil {
		res := utils.BuildResponseFailed("Validation failed", err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	result, err := c.service.Update(ctx.Request.Context(), id, req)
	if err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_UPDATE_PRODUCT, err.Error(), nil)
		ctx.JSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_UPDATE_PRODUCT, result)
	ctx.JSON(http.StatusOK, res)
}

func (c *productController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.service.Delete(ctx.Request.Context(), id); err != nil {
		res := utils.BuildResponseFailed(dto.MESSAGE_FAILED_DELETE_PRODUCT, err.Error(), nil)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, res)
		return
	}

	res := utils.BuildResponseSuccess(dto.MESSAGE_SUCCESS_DELETE_PRODUCT, nil)
	ctx.JSON(http.StatusOK, res)
}
