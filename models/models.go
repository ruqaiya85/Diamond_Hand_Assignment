package models

import "time"

type RewardRequest struct {
    UserID         string  `json:"user_id" binding:"required,uuid"`
    Symbol         string  `json:"symbol" binding:"required"`
    Quantity       string  `json:"quantity" binding:"required"` // send as string to avoid float errors
    RewardedAt     string  `json:"rewarded_at" binding:"required"`
    IdempotencyKey *string `json:"idempotency_key,omitempty"`
    }

type RewardEvent struct {
    ID             string    `db:"id" json:"id"`
    UserID         string    `db:"user_id" json:"user_id"`
    Symbol         string    `db:"symbol" json:"symbol"`
    Quantity       string    `db:"quantity" json:"quantity"`
    RewardedAt     time.Time `db:"rewarded_at" json:"rewarded_at"`
    IdempotencyKey *string   `db:"idempotency_key" json:"idempotency_key,omitempty"`
    CreatedAt      time.Time `db:"created_at" json:"created_at"`
}

type LedgerEntry struct {
    ID            string  `db:"id" json:"id"`
    RewardEventID *string `db:"reward_event_id" json:"reward_event_id,omitempty"`
    Account       string  `db:"account" json:"account"`
    Debit         string  `db:"debit" json:"debit"`
    Credit        string  `db:"credit" json:"credit"`
    StockSymbol   *string `db:"stock_symbol" json:"stock_symbol,omitempty"`
    StockQuantity string  `db:"stock_quantity" json:"stock_quantity"`
    Notes         *string `db:"notes" json:"notes,omitempty"`
    CreatedAt     string  `db:"created_at" json:"created_at"`
}
