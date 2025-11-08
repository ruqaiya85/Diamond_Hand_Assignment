package handlers

import (
    "context"
    "database/sql"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/jmoiron/sqlx"
    "github.com/shopspring/decimal"
    "github.com/sirupsen/logrus"

    "github.com/ruqaiya85/Diamond_Hand_Assignment/models"
    "github.com/ruqaiya85/Diamond_Hand_Assignment/services"
)

type Handler struct {
    db       *sqlx.DB
    priceSvc services.PriceService
}

func NewHandler(db *sqlx.DB, ps services.PriceService) *Handler {
    return &Handler{db: db, priceSvc: ps}
}

// POST /api/reward
// body: { user_id, symbol, quantity (string), rewarded_at (RFC3339), idempotency_key? }
func (h *Handler) PostReward(c *gin.Context) {
    var req models.RewardRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    ctx := c.Request.Context()
    // parse quantity as decimal
    qty, err := decimal.NewFromString(req.Quantity)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid quantity"})
        return
    }
    // parse time
    rewardedAt, err := time.Parse(time.RFC3339, req.RewardedAt)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rewarded_at format (RFC3339 expected)"})
        return
    }

    // Idempotency prevention
    if req.IdempotencyKey != nil && *req.IdempotencyKey != "" {
        var existing string
        err = h.db.GetContext(ctx, &existing, `SELECT id FROM reward_events WHERE idempotency_key=$1 AND user_id=$2`, *req.IdempotencyKey, req.UserID)
        if err == nil {
            // already processed
            c.JSON(http.StatusOK, gin.H{"status": "already_processed", "id": existing})
            return
        }
        if err != sql.ErrNoRows && err != nil {
            logrus.WithError(err).Warn("idempotency check failed")
        }
    }

    tx, err := h.db.BeginTxx(ctx, nil)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "db begin failed"})
        return
    }
    defer func() {
        if p := recover(); p != nil {
            tx.Rollback()
            panic(p)
        }
    }()

    // insert reward_event
    var eventID string
    q := `INSERT INTO reward_events (user_id, symbol, quantity, rewarded_at, idempotency_key) 
          VALUES ($1,$2,$3,$4,$5) RETURNING id`
    if err := tx.GetContext(ctx, &eventID, q, req.UserID, req.Symbol, qty.String(), rewardedAt, req.IdempotencyKey); err != nil {
        tx.Rollback()
        c.JSON(http.StatusInternalServerError, gin.H{"error": "insert reward failed", "detail": err.Error()})
        return
    }

    // compute cost and fees using mocked price service: use latest price or fetch now
    price, _ := h.priceSvc.FetchPrice(ctx, req.Symbol)
    cost := price.Mul(qty) // decimal
    // compute example fees: brokerage 0.2% + stt 0.1% + gst 18% on brokerage
    brokerage := cost.Mul(decimal.NewFromFloat(0.002)).Round(4)
    stt := cost.Mul(decimal.NewFromFloat(0.001)).Round(4)
    gst := brokerage.Mul(decimal.NewFromFloat(0.18)).Round(4)
    totalFees := brokerage.Add(stt).Add(gst).Round(4)
    totalOutflow := cost.Add(totalFees).Round(4)

    // Create ledger entries (double-entry)
    // Debit company_stock_SYMBOL (asset) with quantity (stock units)
    _, err = tx.ExecContext(ctx, `INSERT INTO ledger_entries (reward_event_id, account, stock_symbol, stock_quantity, created_at) VALUES ($1,$2,$3,$4,now())`,
        eventID, "company_stock_"+req.Symbol, req.Symbol, qty.String())
    if err != nil {
        tx.Rollback()
        c.JSON(http.StatusInternalServerError, gin.H{"error": "ledger insert failed", "detail": err.Error()})
        return
    }
    // Credit company_cash for INR outflow (company pays cash)
    _, err = tx.ExecContext(ctx, `INSERT INTO ledger_entries (reward_event_id, account, debit, credit, created_at, notes) VALUES ($1,$2,$3,$4,now(),$5)`,
        eventID, "company_cash", totalOutflow.String(), "0", "cash outflow for stock purchase")
    if err != nil {
        tx.Rollback()
        c.JSON(http.StatusInternalServerError, gin.H{"error": "ledger insert failed", "detail": err.Error()})
        return
    }
    // Record fees as expense (fee entries)
    _, err = tx.ExecContext(ctx, `INSERT INTO ledger_entries (reward_event_id, account, debit, credit, created_at, notes) VALUES ($1,$2,$3,$4,now(),$5)`,
        eventID, "fee_expense", totalFees.String(), "0", "brokerage+stt+gst")
    if err != nil {
        tx.Rollback()
        c.JSON(http.StatusInternalServerError, gin.H{"error": "ledger insert failed", "detail": err.Error()})
        return
    }

    // update holdings_cache (upsert)
    upq := `
    INSERT INTO holdings_cache (user_id, symbol, quantity, last_updated)
    VALUES ($1,$2,$3, now())
    ON CONFLICT (user_id, symbol)
    DO UPDATE SET quantity = (holdings_cache.quantity::numeric + $3::numeric), last_updated = now()
    `
    if _, err := tx.ExecContext(ctx, upq, req.UserID, req.Symbol, qty.String()); err != nil {
        tx.Rollback()
        c.JSON(http.StatusInternalServerError, gin.H{"error": "holdings update failed", "detail": err.Error()})
        return
    }

    if err := tx.Commit(); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "commit failed", "detail": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "status":         "reward_recorded",
        "reward_event_id": eventID,
        "cost_in_inr":    cost.String(),
        "fees_in_inr":    totalFees.String(),
        "total_outflow":  totalOutflow.String(),
    })
}

// GET /api/today-stocks/:userId
func (h *Handler) GetTodayStocks(c *gin.Context) {
    userId := c.Param("userId")
    ctx := c.Request.Context()
    today := time.Now().UTC()
    start := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
    end := start.Add(24 * time.Hour)
    var events []models.RewardEvent
    q := `SELECT id, user_id, symbol, quantity, rewarded_at, idempotency_key, created_at FROM reward_events
          WHERE user_id=$1 AND rewarded_at >= $2 AND rewarded_at < $3 ORDER BY rewarded_at DESC`
    if err := h.db.SelectContext(ctx, &events, q, userId, start, end); err != nil {
        logrus.WithError(err).Warn("today stocks query failed")
        c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
        return
    }
    c.JSON(http.StatusOK, events)
}

// GET /api/historical-inr/:userId --> returns date->INR value up to yesterday
func (h *Handler) GetHistoricalINR(c *gin.Context) {
    userId := c.Param("userId")
    ctx := c.Request.Context()

    // For each distinct date in reward_events for user up to yesterday, compute total INR value using closing price (we use latest price on that day)
    rows, err := h.db.QueryxContext(ctx, `
    SELECT date(rewarded_at) as day, symbol, SUM(quantity::numeric) as total_qty
    FROM reward_events
    WHERE user_id=$1 AND rewarded_at::date < now()::date
    GROUP BY day, symbol
    ORDER BY day DESC
    `, userId)
    if err != nil {
        logrus.WithError(err).Warn("historical query failed")
        c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
        return
    }
    defer rows.Close()

    // map[day] -> list of (symbol, qty)
    type item struct {
        Day    string          `json:"day"`
        Symbol string          `json:"symbol"`
        Qty    decimal.Decimal `json:"quantity"`
        INR    decimal.Decimal `json:"inr_value"`
    }
    results := make(map[string][]item)

    for rows.Next() {
        var day string
        var sym string
        var qtyStr string
        if err := rows.Scan(&day, &sym, &qtyStr); err != nil {
            continue
        }
        qty, _ := decimal.NewFromString(qtyStr)
        // find price for that day: take latest price fetched_at on that day
        var priceStr string
        err := h.db.GetContext(ctx, &priceStr, `SELECT price FROM stock_prices WHERE symbol=$1 AND date(fetched_at)=date($2) ORDER BY fetched_at DESC LIMIT 1`, sym, day)
        var price decimal.Decimal
        if err != nil {
            // fallback: latest available price
            var latest string
            if err := h.db.GetContext(ctx, &latest, `SELECT price FROM stock_prices WHERE symbol=$1 ORDER BY fetched_at DESC LIMIT 1`, sym); err == nil {
                price, _ = decimal.NewFromString(latest)
            } else {
                price = decimal.Zero
            }
        } else {
            price, _ = decimal.NewFromString(priceStr)
        }
        inr := price.Mul(qty).Round(4)
        results[day] = append(results[day], item{
            Day:    day,
            Symbol: sym,
            Qty:    qty,
            INR:    inr,
        })
    }

    c.JSON(http.StatusOK, results)
}

// GET /api/stats/:userId -> total shares rewarded today (by symbol) + current INR value of portfolio
func (h *Handler) GetStats(c *gin.Context) {
    userId := c.Param("userId")
    ctx := c.Request.Context()
    // total shares rewarded today grouped by symbol
    var rows []struct {
        Symbol string `db:"symbol" json:"symbol"`
        Qty    string `db:"quantity" json:"quantity"`
    }

    if err := h.db.SelectContext(ctx, &rows, `
        SELECT symbol, SUM(quantity::numeric) as quantity FROM reward_events
        WHERE user_id=$1 AND rewarded_at >= date_trunc('day', now()) AND rewarded_at < date_trunc('day', now()) + interval '1 day'
        GROUP BY symbol
    `, userId); err != nil {
        logrus.WithError(err).Warn("today grouped query failed")
        c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
        return
    }

    // current INR value of portfolio: use holdings_cache join latest price
    var holdings []struct {
        Symbol string `db:"symbol"`
        Qty    string `db:"quantity"`
    }
    if err := h.db.SelectContext(ctx, &holdings, `SELECT symbol, quantity FROM holdings_cache WHERE user_id=$1`, userId); err != nil {
        logrus.WithError(err).Warn("holdings fetch failed")
    }

    totalINR := decimal.Zero
    portfolio := []gin.H{}
    for _, hld := range holdings {
        qty, _ := decimal.NewFromString(hld.Qty)
        var priceStr string
        if err := h.db.GetContext(ctx, &priceStr, `SELECT price FROM stock_prices WHERE symbol=$1 ORDER BY fetched_at DESC LIMIT 1`, hld.Symbol); err != nil {
            priceStr = "0"
        }
        price, _ := decimal.NewFromString(priceStr)
        value := price.Mul(qty).Round(4)
        totalINR = totalINR.Add(value)
        portfolio = append(portfolio, gin.H{"symbol": hld.Symbol, "quantity": qty.String(), "price": price.String(), "value_inr": value.String()})
    }

    c.JSON(http.StatusOK, gin.H{
        "total_shares_today": rows,
        "current_portfolio_inr": gin.H{
            "total":     totalINR.String(),
            "breakdown": portfolio,
        },
    })
}

// GET /api/portfolio/:userId -> holdings per symbol with current INR value
func (h *Handler) GetPortfolio(c *gin.Context) {
    userId := c.Param("userId")
    ctx := c.Request.Context()
    var holdings []struct {
        Symbol string `db:"symbol"`
        Qty    string `db:"quantity"`
    }
    if err := h.db.SelectContext(ctx, &holdings, `SELECT symbol, quantity FROM holdings_cache WHERE user_id=$1`, userId); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
        return
    }
    totalINR := decimal.Zero
    data := []gin.H{}
    for _, hld := range holdings {
        qty, _ := decimal.NewFromString(hld.Qty)
        var priceStr string
        if err := h.db.GetContext(ctx, &priceStr, `SELECT price FROM stock_prices WHERE symbol=$1 ORDER BY fetched_at DESC LIMIT 1`, hld.Symbol); err != nil {
            priceStr = "0"
        }
        price, _ := decimal.NewFromString(priceStr)
        value := price.Mul(qty).Round(4)
        totalINR = totalINR.Add(value)
        data = append(data, gin.H{
            "symbol":      hld.Symbol,
            "quantity":    qty.String(),
            "price":       price.String(),
            "value_inr":   value.String(),
        })
    }
    c.JSON(http.StatusOK, gin.H{
        "total_inr": totalINR.String(),
        "holdings":  data,
    })
}
