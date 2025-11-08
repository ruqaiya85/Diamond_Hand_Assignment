package main

import (
    //"fmt"
    "os"

    "github.com/gin-gonic/gin"
    "github.com/sirupsen/logrus"

    "github.com/ruqaiya85/Diamond_Hand_Assignment/logger"
    "github.com/ruqaiya85/Diamond_Hand_Assignment/db"
    "github.com/ruqaiya85/Diamond_Hand_Assignment/handlers"
    "github.com/ruqaiya85/Diamond_Hand_Assignment/scheduler"
    "github.com/ruqaiya85/Diamond_Hand_Assignment/services"
)

func main() {
    logger.Init()
    log := logrus.New()

    cfg := loadEnv()

    d, err := db.NewDBFromEnv()
    if err != nil {
        log.Fatalf("db init failed: %v", err)
    }
    defer d.Close()

    priceSvc := services.NewRandomPriceService(cfg.PriceSeed, d)
    scheduler.StartHourlyPriceFetcher(priceSvc, d)

    r := gin.Default()
    api := r.Group("/api")
    {
        h := handlers.NewHandler(d, priceSvc)
        api.POST("/reward", h.PostReward)
        api.GET("/today-stocks/:userId", h.GetTodayStocks)
        api.GET("/historical-inr/:userId", h.GetHistoricalINR)
        api.GET("/stats/:userId", h.GetStats)
        api.GET("/portfolio/:userId", h.GetPortfolio)
    }

    port := cfg.Port
    if port == "" {
        port = "8080"
    }
    log.Infof("Starting server on :%s", port)
    if err := r.Run(":" + port); err != nil {
        log.Fatalf("server failed: %v", err)
    }
}

type Config struct {
    Port      string
    PriceSeed int64
}

func loadEnv() Config {
    // minimal: read env vars; in production use a proper env loader
    seed := int64(42)
    if s := os.Getenv("PRICE_SEED"); s != "" {
        // ignore parse error for brevity
    }
    return Config{
        Port:      os.Getenv("PORT"),
        PriceSeed: seed,
    }
}
