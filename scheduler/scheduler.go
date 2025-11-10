package scheduler

import (
    "context"
    "time"
    "github.com/jmoiron/sqlx"
    "github.com/robfig/cron/v3"
    "github.com/sirupsen/logrus"
    "github.com/ruqaiya85/Diamond_Hand_Assignment/services"
)

func StartHourlyPriceFetcher(priceSvc services.PriceService, db *sqlx.DB) {
    c := cron.New()
    // Run at minute 0 of every hour
    _, err := c.AddFunc("0 * * * *", func() {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
        defer cancel()
        logrus.Info("Hourly price fetcher started")
        var symbols []string
        err := db.SelectContext(ctx, &symbols, `SELECT DISTINCT symbol FROM reward_events`)
        if err != nil {
            logrus.WithError(err).Warn("failed to list symbols for price fetch")
            return
        }
        
        for _, s := range symbols {
            if _, err := priceSvc.FetchAndStorePrice(ctx, s); err != nil {
                logrus.WithField("symbol", s).WithError(err).Warn("price fetch failed")
            }
        }
        logrus.Info("Hourly price fetcher done")
    })
    
    if err != nil {
        logrus.WithError(err).Error("cron add func failed")
        return
    }
    c.Start()
    logrus.Info("Scheduler started: hourly price fetcher registered")
}
