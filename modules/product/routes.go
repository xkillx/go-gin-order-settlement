package product

import (
	"github.com/Caknoooo/go-gin-clean-starter/middlewares"
	authService "github.com/Caknoooo/go-gin-clean-starter/modules/auth/service"
	"github.com/Caknoooo/go-gin-clean-starter/modules/product/controller"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	ctrl := do.MustInvoke[controller.ProductController](injector)
	jwtService := do.MustInvokeNamed[authService.JWTService](injector, constants.JWTService)

	r := server.Group("/api/products")
	{
		r.GET("", ctrl.List)
		r.GET("/:id", ctrl.GetByID)
		r.POST("", middlewares.Authenticate(jwtService), ctrl.Create)
		r.PUT("/:id", middlewares.Authenticate(jwtService), ctrl.Update)
		r.DELETE("/:id", middlewares.Authenticate(jwtService), ctrl.Delete)
	}
}
