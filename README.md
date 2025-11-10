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
POST /reward -

https://rukaiyak890-8732883.postman.co/workspace/DIamond-hand-assignment~5e35feab-403c-4e3f-9332-d9e2ae3df207/request/49913030-e5f502db-bd35-418e-ae50-5998cfb3d695?action=share&creator=49913030

GET /today-stocks/{userId} – 

https://rukaiyak890-8732883.postman.co/workspace/DIamond-hand-assignment~5e35feab-403c-4e3f-9332-d9e2ae3df207/request/49913030-bc28403c-4877-4f11-befd-0f7636216600?action=share&creator=49913030

GET /historical-inr/{userId} – 

https://rukaiyak890-8732883.postman.co/workspace/DIamond-hand-assignment~5e35feab-403c-4e3f-9332-d9e2ae3df207/request/49913030-af282792-c109-4dac-9936-c61ecf97b5a1?action=share&creator=49913030

GET /stats/{userId} – 

https://rukaiyak890-8732883.postman.co/workspace/DIamond-hand-assignment~5e35feab-403c-4e3f-9332-d9e2ae3df207/request/49913030-82e7ebb4-8056-4adc-afbd-de4983625b23?action=share&creator=49913030

(Bonus: Add /portfolio/{userId} -

https://rukaiyak890-8732883.postman.co/workspace/DIamond-hand-assignment~5e35feab-403c-4e3f-9332-d9e2ae3df207/request/49913030-37f16dc9-3f5d-4871-a367-bc4172891ce3?action=share&creator=49913030
