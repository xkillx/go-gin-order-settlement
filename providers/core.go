package providers

import (
	"github.com/Caknoooo/go-gin-clean-starter/config"
	authController "github.com/Caknoooo/go-gin-clean-starter/modules/auth/controller"
	authRepo "github.com/Caknoooo/go-gin-clean-starter/modules/auth/repository"
	authService "github.com/Caknoooo/go-gin-clean-starter/modules/auth/service"
	orderController "github.com/Caknoooo/go-gin-clean-starter/modules/order/controller"
	orderRepo "github.com/Caknoooo/go-gin-clean-starter/modules/order/repository"
	orderService "github.com/Caknoooo/go-gin-clean-starter/modules/order/service"
	productController "github.com/Caknoooo/go-gin-clean-starter/modules/product/controller"
	productRepo "github.com/Caknoooo/go-gin-clean-starter/modules/product/repository"
	productService "github.com/Caknoooo/go-gin-clean-starter/modules/product/service"
	userController "github.com/Caknoooo/go-gin-clean-starter/modules/user/controller"
	userRepo "github.com/Caknoooo/go-gin-clean-starter/modules/user/repository"
	userService "github.com/Caknoooo/go-gin-clean-starter/modules/user/service"
	"github.com/Caknoooo/go-gin-clean-starter/pkg/constants"
	"github.com/samber/do"
	"gorm.io/gorm"
)

func InitDatabase(injector *do.Injector) {
	do.ProvideNamed(injector, constants.DB, func(i *do.Injector) (*gorm.DB, error) {
		return config.SetUpDatabaseConnection(), nil
	})
}

func RegisterDependencies(injector *do.Injector) {
	InitDatabase(injector)

	do.ProvideNamed(injector, constants.JWTService, func(i *do.Injector) (authService.JWTService, error) {
		return authService.NewJWTService(), nil
	})

	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	jwtService := do.MustInvokeNamed[authService.JWTService](injector, constants.JWTService)

	userRepository := userRepo.NewUserRepository(db)
	refreshTokenRepository := authRepo.NewRefreshTokenRepository(db)
	productRepository := productRepo.NewProductRepository(db)
	orderRepository := orderRepo.NewOrderRepository(db)

	userService := userService.NewUserService(userRepository, db)
	authService := authService.NewAuthService(userRepository, refreshTokenRepository, jwtService, db)
	productService := productService.NewProductService(productRepository, db)
	orderService := orderService.NewOrderService(orderRepository, productRepository, db)

	do.Provide(
		injector, func(i *do.Injector) (userController.UserController, error) {
			return userController.NewUserController(i, userService), nil
		},
	)

	do.Provide(
		injector, func(i *do.Injector) (authController.AuthController, error) {
			return authController.NewAuthController(i, authService), nil
		},
	)

	do.Provide(
		injector, func(i *do.Injector) (productController.ProductController, error) {
			return productController.NewProductController(i, productService), nil
		},
	)

	do.Provide(
		injector, func(i *do.Injector) (orderController.OrderController, error) {
			return orderController.NewOrderController(i, orderService), nil
		},
	)
}
