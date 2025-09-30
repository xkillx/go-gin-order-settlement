package order

import (
	"github.com/Caknoooo/go-gin-clean-starter/middlewares"
	authService "github.com/Caknoooo/go-gin-clean-starter/modules/auth/service"
	"github.com/Caknoooo/go-gin-clean-starter/modules/order/controller"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	ctrl := do.MustInvoke[controller.OrderController](injector)
	jwtService := do.MustInvokeNamed[authService.JWTService](injector, constants.JWTService)

	r := server.Group("/api/orders")
	{
		r.GET("", ctrl.List)
		r.GET("/:id", ctrl.GetByID)
		r.POST("", middlewares.Authenticate(jwtService), ctrl.Create)
		r.DELETE("/:id", middlewares.Authenticate(jwtService), ctrl.Delete)
	}
}
