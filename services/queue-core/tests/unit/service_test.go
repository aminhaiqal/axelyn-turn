package unit

import (
	"context"
	"testing"
	"time"
	"queue-core/internal/models"
	"queue-core/internal/repositories"
	"queue-core/internal/services"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestTicketService_CreateTicket(t *testing.T) {
	// Setup mock database
	db, dbMock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := repositories.NewTicketRepo(db)

	// Create a real Redis client that will fail gracefully
	// Since the service treats Redis operations as best-effort, failures are acceptable
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:99999", // Invalid address - operations will fail
	})
	defer rdb.Close()

	service := services.NewTicketService(repo, rdb, "queue.stream", "queue.%d.broadcast")

	ticket := &models.Ticket{
		QueueID:       1,
		CustomerName:  "Alice",
		Status:        "waiting",
		Priority:      1,
		EstimatedTime: 5,
	}

	// Expect database INSERT
	dbMock.ExpectQuery("INSERT INTO tickets").
		WithArgs(ticket.QueueID, ticket.CustomerName, ticket.Status, ticket.Priority, ticket.EstimatedTime).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "version"}).
			AddRow(1, time.Now(), time.Now(), 1))

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// CreateTicket should succeed even if Redis operations fail (best-effort)
	id, err := service.CreateTicket(ctx, ticket)
	assert.NoError(t, err, "CreateTicket should succeed even if Redis fails")
	assert.Equal(t, int64(1), id)
	assert.Equal(t, int64(1), ticket.ID, "Ticket ID should be set")

	// Verify database expectations were met
	assert.NoError(t, dbMock.ExpectationsWereMet())
}
