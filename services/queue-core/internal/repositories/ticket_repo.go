package repositories

import (
    "context"
    "database/sql"
    "queue-core/internal/models"
)

type TicketRepository struct {
    db *sql.DB
}

func NewTicketRepo(db *sql.DB) *TicketRepository {
    return &TicketRepository{db: db}
}

// Create ticket
func (r *TicketRepository) Create(ctx context.Context, t *models.Ticket) error {
    query := `
        INSERT INTO tickets (queue_id, customer_name, status, priority, estimated_time, created_at, updated_at)
        VALUES ($1,$2,$3,$4,$5,NOW(),NOW())
        RETURNING id, created_at, updated_at
    `
    return r.db.QueryRowContext(ctx, query, t.QueueID, t.CustomerName, t.Status, t.Priority, t.EstimatedTime).
        Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
}

// Get ticket by ID
func (r *TicketRepository) GetByID(ctx context.Context, id int) (*models.Ticket, error) {
    t := &models.Ticket{}
    query := `SELECT id, queue_id, customer_name, status, priority, created_at, updated_at, estimated_time FROM tickets WHERE id=$1`
    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &t.ID, &t.QueueID, &t.CustomerName, &t.Status, &t.Priority, &t.CreatedAt, &t.UpdatedAt, &t.EstimatedTime,
    )
    if err != nil {
        return nil, err
    }
    return t, nil
}

// Update status with optimistic locking
func (r *TicketRepository) UpdateStatus(ctx context.Context, id int, oldStatus, newStatus string, version int64) (bool, error) {
    query := `
        UPDATE tickets
        SET status=$1, updated_at=NOW(), version=version+1
        WHERE id=$2 AND status=$3 AND version=$4
    `
    res, err := r.db.ExecContext(ctx, query, newStatus, id, oldStatus, version)
    if err != nil {
        return false, err
    }
    rowsAffected, _ := res.RowsAffected()
    return rowsAffected > 0, nil
}

func (r *TicketRepository) GetByStatus(ctx context.Context, queueID int, status string) ([]*models.Ticket, error) {
    query := `
        SELECT id, queue_id, customer_name, status, priority, created_at, updated_at, estimated_time
        FROM tickets
        WHERE queue_id=$1 AND status=$2
        ORDER BY created_at ASC
    `
    rows, err := r.db.QueryContext(ctx, query, queueID, status)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var tickets []*models.Ticket
    for rows.Next() {
        t := &models.Ticket{}
        if err := rows.Scan(
            &t.ID,
            &t.QueueID,
            &t.CustomerName,
            &t.Status,
            &t.Priority,
            &t.CreatedAt,
            &t.UpdatedAt,
            &t.EstimatedTime,
        ); err != nil {
            return nil, err
        }
        tickets = append(tickets, t)
    }

    if err = rows.Err(); err != nil {
        return nil, err
    }

    return tickets, nil
}