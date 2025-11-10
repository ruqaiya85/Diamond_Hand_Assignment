package db

import (
    "fmt"
    "os"
    "time"

    "github.com/jmoiron/sqlx"
    _ "github.com/lib/pq"
)

func NewDBFromEnv() (*sqlx.DB, error) {
    host := getenv("DB_HOST", "localhost")
    port := getenv("DB_PORT", "5432")
    user := getenv("DB_USER", "postgres")
    pass := getenv("DB_PASS", "Warmc0nnect")
    name := getenv("DB_NAME", "Assignment")

    
    dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        host, port, user, pass, name)
    db, err := sqlx.Connect("postgres", dsn)
    if err != nil {
        return nil, err
    }
    db.SetMaxOpenConns(20)
    db.SetConnMaxIdleTime(5 * time.Minute)
    db.SetConnMaxLifetime(60 * time.Minute)
    return db, nil
}

func getenv(k, fallback string) string {
    v := os.Getenv(k)
    if v == "" {
        return fallback
    }
    return v
}
