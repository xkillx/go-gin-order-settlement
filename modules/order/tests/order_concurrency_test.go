package order_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/xkillx/go-gin-order-settlement/config"
	"github.com/xkillx/go-gin-order-settlement/database/entities"
	orderModule "github.com/xkillx/go-gin-order-settlement/modules/order"
	orderController "github.com/xkillx/go-gin-order-settlement/modules/order/controller"
	orderRepo "github.com/xkillx/go-gin-order-settlement/modules/order/repository"
	orderService "github.com/xkillx/go-gin-order-settlement/modules/order/service"
	productRepo "github.com/xkillx/go-gin-order-settlement/modules/product/repository"
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
	"gorm.io/gorm"
)

type orderCreateRequest struct {
	ProductID string `json:"product_id"`
	BuyerID   string `json:"buyer_id"`
	Quantity  int    `json:"quantity"`
}

func setupTestServer(t *testing.T) (*gin.Engine, *do.Injector, *gorm.DB, *entities.Product, func()) {
	t.Helper()

	gin.SetMode(gin.TestMode)
	// Use test PostgreSQL connection that does not require a .env file
	db := config.SetUpTestDatabaseConnection()

	// Ensure schema
	if err := db.AutoMigrate(&entities.Product{}, &entities.Order{}); err != nil {
		t.Fatalf("automigrate failed: %v", err)
	}
	// Truncate tables (orders first due to FK)
	if err := db.Exec("DELETE FROM orders").Error; err != nil {
		t.Fatalf("failed to truncate orders: %v", err)
	}
	if err := db.Exec("DELETE FROM products").Error; err != nil {
		t.Fatalf("failed to truncate products: %v", err)
	}

	// Create repositories and service bound to this DB
	prdRepo := productRepo.NewProductRepository(db)
	ordRepo := orderRepo.NewOrderRepository(db)
	svc := orderService.NewOrderService(ordRepo, prdRepo, db)

	// Create a dummy product with stock = 100
	ctx := context.Background()
	product, err := prdRepo.Create(ctx, db, entities.Product{
		Name:  "Test Product",
		Stock: 100,
	})
	if err != nil {
		t.Fatalf("failed to create product: %v", err)
	}

	// Wire only the Orders routes with a DI container that provides OrderController
	inj := do.New()
	do.Provide(inj, func(i *do.Injector) (orderController.OrderController, error) {
		return orderController.NewOrderController(i, svc), nil
	})

	engine := gin.New()
	orderModule.RegisterRoutes(engine, inj)

	cleanup := func() {
		config.CloseDatabaseConnection(db)
	}

	return engine, inj, db, &product, cleanup
}

func TestConcurrentBuyers500(t *testing.T) {
	server, _, db, product, cleanup := setupTestServer(t)
	defer cleanup()

	// Use UnstartedServer to configure connection limits before starting
	ts := httptest.NewUnstartedServer(server)
	// Increase MaxHeaderBytes to handle more concurrent requests
	ts.Config.MaxHeaderBytes = 1 << 20
	ts.Start()
	defer ts.Close()

	url := ts.URL + "/api/orders"

	var (
		wg            sync.WaitGroup
		successCount  int32
		conflictCount int32
		otherCount    int32
	)

	// Barrier to start all goroutines at once
	start := make(chan struct{})
	
	// Semaphore to limit concurrent HTTP connections (to avoid overwhelming test server)
	// Limit to 100 concurrent connections at a time
	concurrencySem := make(chan struct{}, 100)

	// Configure HTTP client to handle high concurrency
	transport := &http.Transport{
		MaxIdleConns:        200,
		MaxIdleConnsPerHost: 200,
		MaxConnsPerHost:     200,
		IdleConnTimeout:     90 * time.Second,
	}
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}

	for i := 0; i < 500; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			
			// Acquire semaphore slot before making HTTP call
			concurrencySem <- struct{}{}
			defer func() { <-concurrencySem }()

			payload := orderCreateRequest{
				ProductID: product.ID.String(),
				BuyerID:   fmt.Sprintf("buyer-%d", i),
				Quantity:  1,
			}
			b, _ := json.Marshal(payload)
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
			if err != nil {
				atomic.AddInt32(&otherCount, 1)
				return
			}
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				atomic.AddInt32(&otherCount, 1)
				return
			}
			defer resp.Body.Close()

			switch resp.StatusCode {
			case http.StatusCreated: // 201 success expected
				atomic.AddInt32(&successCount, 1)
			case http.StatusConflict: // 409 out-of-stock expected for losers
				atomic.AddInt32(&conflictCount, 1)
			default:
				atomic.AddInt32(&otherCount, 1)
			}
		}(i)
	}

	// Release all goroutines simultaneously
	close(start)
	wg.Wait()
	if successCount != 100 || conflictCount != 400 {
		t.Fatalf("unexpected counts: success=%d conflict=%d other=%d (expected success=100, conflict=400, other=0)", successCount, conflictCount, otherCount)
	}

	// Verify final stock and orders count using the same DB connection
	ctx := context.Background()
	var reloaded entities.Product
	if err := db.WithContext(ctx).Where("id = ?", product.ID).Take(&reloaded).Error; err != nil {
		t.Fatalf("failed to reload product: %v", err)
	}
	if reloaded.Stock != 0 {
		t.Fatalf("expected product stock 0, got %d", reloaded.Stock)
	}

	var orderCount int64
	if err := db.WithContext(ctx).Model(&entities.Order{}).Count(&orderCount).Error; err != nil {
		t.Fatalf("failed counting orders: %v", err)
	}
	if orderCount != 100 {
		t.Fatalf("expected orders count 100, got %d", orderCount)
	}
}
