package providers

import (
	"github.com/xkillx/go-gin-order-settlement/config"
	orderController "github.com/xkillx/go-gin-order-settlement/modules/order/controller"
	orderRepo "github.com/xkillx/go-gin-order-settlement/modules/order/repository"
	orderService "github.com/xkillx/go-gin-order-settlement/modules/order/service"
	productController "github.com/xkillx/go-gin-order-settlement/modules/product/controller"
	productRepo "github.com/xkillx/go-gin-order-settlement/modules/product/repository"
	productService "github.com/xkillx/go-gin-order-settlement/modules/product/service"
	"github.com/xkillx/go-gin-order-settlement/pkg/constants"
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
