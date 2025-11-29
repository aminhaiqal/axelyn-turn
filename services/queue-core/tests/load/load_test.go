package load

import (
	"context"
	"os"
	"sync"
	"testing"
	"strconv"
	"time"
	"queue-core/internal/db"
	"queue-core/internal/models"
	"queue-core/internal/repositories"
	"queue-core/internal/services"
)

func TestConcurrentTicketCreation(t *testing.T) {
	// Setup service with real database and Redis connections
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		t.Skip("DATABASE_URL not set, skipping load test")
	}

	dbConn, err := db.Connect(connStr)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()

	rdb := db.NewRedisClient().Client
	repo := repositories.NewTicketRepo(dbConn)
	service := services.NewTicketService(repo, rdb, "queue.stream", "queue.%d.broadcast")

	ctx := context.Background()
	const n = 100 // number of concurrent requests

	wg := sync.WaitGroup{}
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			ticket := &models.Ticket{
				QueueID:      1,
				CustomerName: "LoadTest #" + strconv.Itoa(i),
				Priority:     1,
				Status:       "waiting",
			}
			_, _ = service.CreateTicket(ctx, ticket)
		}(i)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("load test timed out")
	}
}
