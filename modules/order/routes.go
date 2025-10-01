package order

import (
	"github.com/xkillx/go-gin-order-settlement/modules/order/controller"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	ctrl := do.MustInvoke[controller.OrderController](injector)

	r := server.Group("/api/orders")
	{
		r.GET("", ctrl.List)
		r.GET("/:id", ctrl.GetByID)
		r.POST("", ctrl.Create)
		r.DELETE("/:id", ctrl.Delete)
	}
}
