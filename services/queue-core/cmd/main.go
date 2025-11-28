package main

import (
    "fmt"
    "log"
    "net/http"
    "os"
    "queue-core/internal/api"
    "queue-core/internal/db"
    "queue-core/internal/repositories"
    "queue-core/internal/services"
)

func main() {
    connStr := os.Getenv("POSTGRES_URL") // e.g., postgres://user:pass@localhost:5432/queue
    dbConn, err := db.Connect(connStr)
    if err != nil {
        log.Fatal(err)
    }

    ticketRepo := repositories.NewTicketRepo(dbConn)
    ticketService := services.NewTicketService(ticketRepo)
    apiHandler := api.NewAPI(ticketService)

    fmt.Println("Queue-Core running on :8080")
    log.Fatal(http.ListenAndServe(":8080", apiHandler.Router()))
}
