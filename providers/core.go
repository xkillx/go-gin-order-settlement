package providers

import (
	"github.com/xkillx/go-gin-order-settlement/config"
	orderController "github.com/xkillx/go-gin-order-settlement/modules/order/controller"
	orderRepo "github.com/xkillx/go-gin-order-settlement/modules/order/repository"
	orderService "github.com/xkillx/go-gin-order-settlement/modules/order/service"
	productController "github.com/xkillx/go-gin-order-settlement/modules/product/controller"
	productRepo "github.com/xkillx/go-gin-order-settlement/modules/product/repository"
	productService "github.com/xkillx/go-gin-order-settlement/modules/product/service"
	jobRepo "github.com/xkillx/go-gin-order-settlement/modules/job/repository"
	settlementRepo "github.com/xkillx/go-gin-order-settlement/modules/settlement/repository"
	settlementService "github.com/xkillx/go-gin-order-settlement/modules/settlement/service"
	transactionRepo "github.com/xkillx/go-gin-order-settlement/modules/transaction/repository"
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
	// Settlement job related repos
	txRepository := transactionRepo.NewTransactionRepository(db)
	stRepository := settlementRepo.NewSettlementRepository(db)
	jobRepository := jobRepo.NewJobRepository(db)

	productService := productService.NewProductService(productRepository, db)
	orderService := orderService.NewOrderService(orderRepository, productRepository, db)
	// Provide JobManager as a singleton service so controllers can access the same instance for cancellation
	do.Provide(
		injector, func(i *do.Injector) (*settlementService.JobManager, error) {
			return settlementService.NewJobManager(txRepository, stRepository, jobRepository), nil
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
