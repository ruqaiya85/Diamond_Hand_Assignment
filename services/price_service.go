package services

import (
    "context"
    //"database/sql"
    "math/rand"
    //"time"

    "github.com/jmoiron/sqlx"
    "github.com/sirupsen/logrus"
    "github.com/shopspring/decimal"
)

type PriceService interface {
    FetchPrice(ctx context.Context, symbol string) (decimal.Decimal, error)
    FetchAndStorePrice(ctx context.Context, symbol string) (decimal.Decimal, error)
}

type RandomPriceService struct {
    rnd *rand.Rand
    db  *sqlx.DB
}

func NewRandomPriceService(seed int64, db *sqlx.DB) PriceService {
    return &RandomPriceService{
        rnd: rand.New(rand.NewSource(seed)),
        db:  db,
    }
}

func (r *RandomPriceService) FetchPrice(ctx context.Context, symbol string) (decimal.Decimal, error) {
    // Generate a pseudo-random price for demo. Real integration: NSE/BSE API.
    base := 100.0 + float64(len(symbol))*10.0
    jitter := r.rnd.Float64() * 100.0
    price := base + jitter
    d := decimal.NewFromFloat(price).Round(4)
    return d, nil
}

func (r *RandomPriceService) FetchAndStorePrice(ctx context.Context, symbol string) (decimal.Decimal, error) {
    p, err := r.FetchPrice(ctx, symbol)
    if err != nil {
        return p, err
    }
    q := `INSERT INTO stock_prices (symbol, price, fetched_at) VALUES ($1, $2, now())`
    if _, err := r.db.ExecContext(ctx, q, symbol, p.String()); err != nil {
        logrus.WithError(err).Warn("failed to persist price")
    }
    return p, nil
}
