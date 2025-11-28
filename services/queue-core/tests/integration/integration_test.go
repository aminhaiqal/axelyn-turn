package integration

import (
	"context"
	"os"
	"testing"
	"queue-core/internal/db"
	"queue-core/internal/models"
	"queue-core/internal/repositories"
	"queue-core/internal/services"

	"github.com/stretchr/testify/assert"
)

func TestCreateAndReserveTicket(t *testing.T) {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		t.Skip("DATABASE_URL not set")
	}
	dbConn, err := db.Connect(connStr)
	assert.NoError(t, err)

	rdb := db.NewRedisClient().Client
	repo := repositories.NewTicketRepo(dbConn)
	service := services.NewTicketService(repo, rdb, "queue.stream", "queue.%d.broadcast")

	ctx := context.Background()
	ticket := &models.Ticket{
		QueueID:      1,
		CustomerName: "Integration Test",
		Status:       "waiting",
		Priority:     1,
	}

	// Create ticket
	tid, err := service.CreateTicket(ctx, ticket)
	assert.NoError(t, err)
	assert.Greater(t, tid, int64(0))

	// Reserve ticket
	reserved, err := repo.ReserveNext(ctx, 1)
	assert.NoError(t, err)
	assert.NotNil(t, reserved)
	assert.Equal(t, "processing", reserved.Status)
}
