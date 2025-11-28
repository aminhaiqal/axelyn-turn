package unit

import (
	"context"
	"testing"
	"time"
	"queue-core/internal/models"
	"queue-core/internal/repositories"
	"queue-core/internal/services"

	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestTicketService_CreateTicket(t *testing.T) {
	db := &repositories.TicketRepository{} // Mock repo can be expanded with testify/mock
	rdb, mock := redismock.NewClientMock()
	service := services.NewTicketService(db, rdb, "queue.stream", "queue.%d.broadcast")

	ticket := &models.Ticket{
		ID:            1,
		QueueID:       1,
		CustomerName:  "Alice",
		Status:        "waiting",
		Priority:      1,
		EstimatedTime: 5,
	}

	// Expect XAdd to be called for Redis Stream
	mock.ExpectXAdd(&redis.XAddArgs{
		Stream: "queue.stream",
		Values: map[string]interface{}{
			"ticket_id":  ticket.ID,
			"queue_id":   ticket.QueueID,
			"event":      "ticket.created",
			"created_at": time.Now().UTC().Format(time.RFC3339),
		},
	}).SetVal("1")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// You may need to adjust repo mock for a full test
	// Here just testing Redis push
	_, _ = service.CreateTicket(ctx, ticket)
	assert.NoError(t, mock.ExpectationsWereMet())
}
