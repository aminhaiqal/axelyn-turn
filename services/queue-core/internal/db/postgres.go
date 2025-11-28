package db

import (
    "database/sql"
    _ "github.com/lib/pq"
    "fmt"
)

func Connect(connStr string) (*sql.DB, error) {
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return nil, err
    }
    if err = db.Ping(); err != nil {
        return nil, err
    }
    fmt.Println("Connected to Postgres")
    return db, nil
}
