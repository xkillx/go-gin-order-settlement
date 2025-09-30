package providers

import (
	"github.com/Caknoooo/go-gin-clean-starter/config"
	orderController "github.com/Caknoooo/go-gin-clean-starter/modules/order/controller"
	orderRepo "github.com/Caknoooo/go-gin-clean-starter/modules/order/repository"
	orderService "github.com/Caknoooo/go-gin-clean-starter/modules/order/service"
	productController "github.com/Caknoooo/go-gin-clean-starter/modules/product/controller"
	productRepo "github.com/Caknoooo/go-gin-clean-starter/modules/product/repository"
	productService "github.com/Caknoooo/go-gin-clean-starter/modules/product/service"
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

	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)

	productRepository := productRepo.NewProductRepository(db)
	orderRepository := orderRepo.NewOrderRepository(db)

	productService := productService.NewProductService(productRepository, db)
	orderService := orderService.NewOrderService(orderRepository, productRepository, db)

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
