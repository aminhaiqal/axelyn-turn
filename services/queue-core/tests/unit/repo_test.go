package unit

import (
	"context"
	"testing"
	"time"
	"queue-core/internal/models"
	"queue-core/internal/repositories"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestTicketRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := repositories.NewTicketRepo(db)

	now := time.Now()
	mock.ExpectQuery("INSERT INTO tickets").
		WithArgs(1, "John Doe", "waiting", 1, 10).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at", "version"}).
			AddRow(1, now, now, 1))

	ticket := &models.Ticket{
		QueueID:       1,
		CustomerName:  "John Doe",
		Status:        "waiting",
		Priority:      1,
		EstimatedTime: 10,
	}

	err = repo.Create(context.Background(), ticket)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), ticket.ID)
}

func TestTicketRepository_UpdateStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := repositories.NewTicketRepo(db)

	mock.ExpectQuery("UPDATE tickets").
		WithArgs("processing", int64(1), "waiting", int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"version"}).AddRow(2))

	ok, newVersion, err := repo.UpdateStatus(context.Background(), 1, "waiting", "processing", 1)
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, int64(2), newVersion)
}
