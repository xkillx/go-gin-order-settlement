package product

import (
	"github.com/Caknoooo/go-gin-clean-starter/modules/product/controller"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	ctrl := do.MustInvoke[controller.ProductController](injector)

	r := server.Group("/api/products")
	{
		r.GET("", ctrl.List)
		r.GET("/:id", ctrl.GetByID)
		r.POST("", ctrl.Create)
		r.PUT("/:id", ctrl.Update)
		r.DELETE("/:id", ctrl.Delete)
	}
}
