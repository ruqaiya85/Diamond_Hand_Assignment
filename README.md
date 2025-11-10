# Stocky Assignment (Golang)

This project implements a minimal backend for **Stocky** — a mock investment and reward system that grants users stock units based on their reward activities.  

---

## 🧠 Concept Overview
The system simulates how a user earns stock rewards, how their stock prices change hourly, and how their holdings and stats are calculated.  
It’s built like a simplified version of a real financial reward and portfolio system.

---

## ⚙️ Tech Stack
- **Language:** Golang  
- **Database:** PostgreSQL  
- **Libraries Used:**  
  - `gin` – For building REST APIs  
  - `logrus` – For structured logging  
  - `sqlx` – For easy SQL + Go struct mapping  
  - `shopspring/decimal` – For precise decimal calculations (important for currency)  
  - `robfig/cron` – For scheduling tasks (like hourly stock price updates)  

---

## 🗄️ Database Schema Explanation

### 1. `users`
Stores user information.  
Each user can receive rewards and hold stocks.

| Column | Type | Description |
|--------|------|-------------|
| id | SERIAL PRIMARY KEY | Unique user ID |
| name | TEXT | User's name |

---

### 2. `stocks`
Contains all stock types available in the system.  
For example: TCS, INFY, HDFC, etc.

| Column | Type | Description |
|--------|------|-------------|
| id | SERIAL PRIMARY KEY | Stock ID |
| symbol | TEXT | Stock symbol |
| name | TEXT | Full stock name |

---

### 3. `stock_prices`
Stores hourly mock prices for each stock.  
A background job (cron) updates this table every hour.

| Column | Type | Description |
|--------|------|-------------|
| id | SERIAL PRIMARY KEY | Unique record ID |
| stock_id | INT (FK → stocks.id) | Which stock the price belongs to |
| price | DECIMAL | The stock price in INR |
| recorded_at | TIMESTAMP | When the price was recorded |

📌 **Relation:**  
Each stock can have multiple hourly price entries — *One-to-Many* (`stocks → stock_prices`).

---

### 4. `ledger`
Acts as a transaction log (double-entry style).  
Whenever a user earns a stock reward, an entry is added here to record what happened.

| Column | Type | Description |
|--------|------|-------------|
| id | SERIAL PRIMARY KEY | Transaction ID |
| user_id | INT (FK → users.id) | Who received the reward |
| stock_id | INT (FK → stocks.id) | Which stock they got |
| units | DECIMAL | How many stock units were rewarded |
| created_at | TIMESTAMP | When the reward happened |

📌 **Relation:**  
Each user can have multiple transactions — *One-to-Many* (`users → ledger`).

---

### 5. `holdings_cache`
Stores the current total holdings per user per stock for quick reads.  
This helps avoid recalculating from the ledger every time.

| Column | Type | Description |
|--------|------|-------------|
| id | SERIAL PRIMARY KEY | Cache ID |
| user_id | INT (FK → users.id) | Who owns the stock |
| stock_id | INT (FK → stocks.id) | Which stock |
| total_units | DECIMAL | Total number of stock units held |

📌 **Relation:**  
This is a *Many-to-One* cache table combining user and stock totals.


Test Data

Users Table

Assignment=# select * from users;
                  id                  |      email       |   name
--------------------------------------+------------------+-----------
 123e4567-e89b-12d3-a456-426614174000 | test@example.com | Test User
 123e4567-e89b-12d3-a456-426614174001 | rukaiya@test.com | rukaiya
(2 rows)

rewards_events table.

Assignment=# select * from reward_events;
-[ RECORD 1 ]---+-------------------------------------
id              | 221f177a-9d24-4b0f-aff1-0ff8845bd452
user_id         | 123e4567-e89b-12d3-a456-426614174000
symbol          | RELIANCE
quantity        | 2.500000
rewarded_at     | 2025-11-09 19:30:00+05:30
idempotency_key |
created_at      | 2025-11-09 02:48:05.148496+05:30
-[ RECORD 2 ]---+-------------------------------------
id              | 93969bf8-752a-4825-a50f-4a3597b159a2
user_id         | 123e4567-e89b-12d3-a456-426614174001
symbol          | JIO
quantity        | 2.300000
rewarded_at     | 2025-11-10 19:30:00+05:30
idempotency_key |
created_at      | 2025-11-11 00:02:20.365726+05:30




---

## 🚀 API Endpoints

| Method | Endpoint | Description |
|--------|-----------|-------------|
| **POST** | `/api/reward` | Record a new stock reward event for a user |
| **GET** | `/api/today-stocks/{userId}` | Get today’s stock prices for a specific user |
| **GET** | `/api/historical-inr/{userId}` | Get historical stock value in INR |
| **GET** | `/api/stats/{userId}` | Show user’s overall stats and growth |
| **GET** | `/api/portfolio/{userId}` | Show user’s portfolio summary (bonus feature) |

---

## 🧩 How It Works 

1. **Reward Entry:**  
   When you hit the `/api/reward` endpoint, a new reward is created in the **ledger** and **holdings_cache** is updated for that user.

2. **Stock Prices:**  
   Every hour, a scheduled job (using `cron`) generates random prices and stores them in the **stock_prices** table.

3. **User Portfolio & Stats:**  
   The `/api/portfolio` and `/api/stats` endpoints calculate a user’s total worth using:
   - Latest prices from `stock_prices`
   - Total units from `holdings_cache`

4. **Speed Optimization:**  
   Instead of recalculating from `ledger` every time, data is quickly fetched from `holdings_cache`, improving performance.

---

## 🧪 Postman API Testing Links

- **POST /reward** – [Open in Postman](https://rukaiyak890-8732883.postman.co/workspace/DIamond-hand-assignment~5e35feab-403c-4e3f-9332-d9e2ae3df207/request/49913030-e5f502db-bd35-418e-ae50-5998cfb3d695?action=share&creator=49913030)
- **GET /today-stocks/{userId}** – [Open in Postman](https://rukaiyak890-8732883.postman.co/workspace/DIamond-hand-assignment~5e35feab-403c-4e3f-9332-d9e2ae3df207/request/49913030-bc28403c-4877-4f11-befd-0f7636216600?action=share&creator=49913030)
- **GET /historical-inr/{userId}** – [Open in Postman](https://rukaiyak890-8732883.postman.co/workspace/DIamond-hand-assignment~5e35feab-403c-4e3f-9332-d9e2ae3df207/request/49913030-af282792-c109-4dac-9936-c61ecf97b5a1?action=share&creator=49913030)
- **GET /stats/{userId}** – [Open in Postman](https://rukaiyak890-8732883.postman.co/workspace/DIamond-hand-assignment~5e35feab-403c-4e3f-9332-d9e2ae3df207/request/49913030-82e7ebb4-8056-4adc-afbd-de4983625b23?action=share&creator=49913030)
- **GET /portfolio/{userId}** – [Open in Postman](https://rukaiyak890-8732883.postman.co/workspace/DIamond-hand-assignment~5e35feab-403c-4e3f-9332-d9e2ae3df207/request/49913030-37f16dc9-3f5d-4871-a367-bc4172891ce3?action=share&creator=49913030)

---

## 🧾 Code Explanation 

- The project mimics how reward points convert to stock units.  
- The main logic lies in:
  - `reward.go`: handles reward creation and ledger updates.  
  - `cron.go`: runs hourly to generate stock prices.  
  - `portfolio.go` and `stats.go`: calculate holdings and portfolio values.  
- Each module is designed cleanly with real-world-like data flow between users, rewards, and stock value updates.  

In Details :

- **`main.go`**  
  Entry point of the app.  
  It connects to the database, sets up routes, and starts the HTTP server on port **8080**.  
  It also initializes the **cron job** for fetching prices every hour.

- **`db/db.go`**  
  Handles database connection using PostgreSQL.  
  It connects using `sqlx` and logs any errors during initialization.

- **`handlers/` folder**  
  Contains all the route functions (APIs).  
  Each file handles a specific feature like rewards, portfolio, or stats.  
  Examples:  
  - `PostReward` → adds a new reward record for a user.  
  - `GetPortfolio` → shows user’s current stock holdings and total worth.  
  - `GetStats` → gives user summary info like total rewards and performance.

- **`models/` folder**  
  Contains Go structs that define the data structure used in both the database and the API requests.

- **`services/` folder**  
  Contains the main logic for calculations, price fetching, and business rules.

- **`scheduler/` folder**  
  Uses the `cron` library to automatically update stock prices every hour.

There are **four main tables** in `sql/schema.sql`:

### 1. `users`
Stores basic information about each user.  
- `id` (UUID) is the unique identifier.  
Used as a reference in other tables.

### 2. `reward_events`
Stores the history of rewards given to users (like stock bonuses).  
- Has a **foreign key (`user_id`)** linked to `users.id`.  
Each row means a user got a stock reward (symbol + quantity + date).

### 3. `stock_prices`
Keeps the **current and historical stock prices**.  
- Updated automatically by the scheduler every hour.  
- Helps calculate portfolio values and statistics.

### 4. `user_portfolio`
Tracks how many shares each user owns for each stock.  
- Linked to `users` via `user_id`.  
- Gets updated whenever a new reward event is added.

**In short:**  
> This project demonstrates a clean backend architecture using Go and PostgreSQL to simulate real-time stock reward and tracking system.

---

## ✨ Author
Developed by **Ruqaiya** for the **Diamond Hand Assignment**.
