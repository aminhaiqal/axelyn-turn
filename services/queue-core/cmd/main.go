package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"

	"queue-core/internal/api"
	"queue-core/internal/db"
	"queue-core/internal/repositories"
	"queue-core/internal/services"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on actual environment variables")
	}

	// --- DATABASE CONNECTION ---
	connStr := os.Getenv("DATABASE_URL")
	dbConn, err := db.Connect(connStr)
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}

	// --- REDIS CONNECTION ---
	redisConn := db.NewRedisClient() // reads UPSTASH_REDIS_URL and UPSTASH_REDIS_TOKEN from env
	rdb := redisConn.Client

	// --- REPOSITORIES ---
	ticketRepo := repositories.NewTicketRepo(dbConn)

	// --- SERVICES ---
	streamName := "queue.stream"
	pubSubBase := "queue.%d.broadcast"
	ticketService := services.NewTicketService(ticketRepo, rdb, streamName, pubSubBase)

	// --- API ---
	apiHandler := api.NewAPI(ticketService, rdb, streamName, pubSubBase)

	// --- Dispatcher ---
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go ticketService.StartDispatcher(ctx, 1, 2*time.Second)

	// --- HTTP SERVER ---
	fmt.Println("Queue-Core running on :8080")
	log.Fatal(http.ListenAndServe(":8080", apiHandler.Router()))
}
