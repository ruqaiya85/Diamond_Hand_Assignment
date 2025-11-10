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


## Postman API testing link
- https://rukaiyak890-8732883.postman.co/workspace/DIamond-hand-assignment~5e35feab-403c-4e3f-9332-d9e2ae3df207/request/49913030-37f16dc9-3f5d-4871-a367-bc4172891ce3?action=share&creator=49913030&ctx=documentation
