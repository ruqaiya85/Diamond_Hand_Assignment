# Stocky Assignment (Golang)

This project implements a minimal backend for Stocky — a hypothetical service that rewards users with Indian stock units.

## Stack
- Golang
- Postgres
- Libraries: gin, logrus, sqlx, shopspring/decimal, robfig/cron


## Endpoints
- POST /api/reward — record reward event
- GET /api/today-stocks/{userId}
- GET /api/historical-inr/{userId}
- GET /api/stats/{userId}
- GET /api/portfolio/{userId}

## Notes
- Price fetching is a mocked random generator, hourly job stores price points in stock_prices.
- Ledger entries follow a simple double-entry pattern, holdings_cache caches per-user holdings for quick reads.

## Author
- Rukaiya Kochikar
