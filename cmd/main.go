package main

import (
    "log"
    "os"

    "github.com/xkillx/go-gin-order-settlement/middlewares"
    "github.com/xkillx/go-gin-order-settlement/modules/order"
    "github.com/xkillx/go-gin-order-settlement/modules/product"
    "github.com/xkillx/go-gin-order-settlement/modules/settlement"
    "github.com/xkillx/go-gin-order-settlement/providers"
    "github.com/xkillx/go-gin-order-settlement/script"
    "github.com/samber/do"

	"github.com/common-nighthawk/go-figure"
	"github.com/gin-gonic/gin"
)

func args(injector *do.Injector) bool {
    if len(os.Args) > 1 {
        flag := script.Commands(injector)
        return flag
    }

    return true
}

func run(server *gin.Engine) {
    server.Static("/assets", "./assets")

    port := os.Getenv("GOLANG_PORT")
    if port == "" {
        port = "8888"
    }

    var serve string
    if os.Getenv("APP_ENV") == "localhost" {
        serve = "0.0.0.0:" + port
    } else {
        serve = ":" + port
    }

    myFigure := figure.NewColorFigure("Caknoo", "", "green", true)
    myFigure.Print()

    if err := server.Run(serve); err != nil {
        log.Fatalf("error running server: %v", err)
    }
}

func main() {
    var (
        injector = do.New()
    )

    providers.RegisterDependencies(injector)

    if !args(injector) {
        return
    }

    server := gin.Default()
    server.Use(middlewares.CORSMiddleware())

    // Register module routes
    product.RegisterRoutes(server, injector)
    order.RegisterRoutes(server, injector)
    settlement.RegisterRoutes(server, injector)

    run(server)
}
