package db

import (
    "database/sql"
    _ "github.com/jackc/pgx/v5/stdlib"
    "time"
)

func Connect(connStr string) (*sql.DB, error) {
    db, err := sql.Open("pgx", connStr)
    if err != nil {
        return nil, err
    }

    db.SetMaxOpenConns(10)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(15 * time.Minute)

    return db, nil
}
