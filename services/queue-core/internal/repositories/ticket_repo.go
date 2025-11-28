package repositories

import (
	"context"
	"database/sql"
	"errors"

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
        RETURNING id, created_at, updated_at, version
    `
	return r.db.QueryRowContext(ctx, query, t.QueueID, t.CustomerName, t.Status, t.Priority, t.EstimatedTime).
		Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt, &t.Version)
}

// Get ticket by ID
func (r *TicketRepository) GetByID(ctx context.Context, id int64) (*models.Ticket, error) {
	t := &models.Ticket{}
	query := `SELECT id, queue_id, customer_name, status, priority, created_at, updated_at, estimated_time, version FROM tickets WHERE id=$1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&t.ID, &t.QueueID, &t.CustomerName, &t.Status, &t.Priority, &t.CreatedAt, &t.UpdatedAt, &t.EstimatedTime, &t.Version,
	)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// GetByStatus — also return version (for API listing if needed)
func (r *TicketRepository) GetByStatus(ctx context.Context, queueID int, status string) ([]*models.Ticket, error) {
	query := `
        SELECT id, queue_id, customer_name, status, priority, created_at, updated_at, estimated_time, version
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
			&t.ID, &t.QueueID, &t.CustomerName, &t.Status, &t.Priority,
			&t.CreatedAt, &t.UpdatedAt, &t.EstimatedTime, &t.Version,
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

// ReserveNext - atomically pick one waiting ticket and set it to processing.
// Returns ticket with its original version value (pre-update).
func (r *TicketRepository) ReserveNext(ctx context.Context, queueID int) (*models.Ticket, error) {
	// Begin a short-lived transaction
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, err
	}
	// ensure rollback if not committed
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	q := `
        SELECT id, queue_id, customer_name, status, priority, created_at, updated_at, estimated_time, version
        FROM tickets
        WHERE queue_id=$1 AND status='waiting'
        ORDER BY created_at ASC
        FOR UPDATE SKIP LOCKED
        LIMIT 1
    `
	t := &models.Ticket{}
	err = tx.QueryRowContext(ctx, q, queueID).Scan(
		&t.ID, &t.QueueID, &t.CustomerName, &t.Status, &t.Priority,
		&t.CreatedAt, &t.UpdatedAt, &t.EstimatedTime, &t.Version,
	)
	if err == sql.ErrNoRows {
		// nothing to reserve
		_ = tx.Rollback()
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Update to processing and bump version
	_, err = tx.ExecContext(ctx, `
        UPDATE tickets
        SET status='processing', updated_at=NOW(), version=version+1
        WHERE id = $1
    `, t.ID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	committed = true

	// t.Version is the old version before increment
	return t, nil
}

// UpdateStatus uses optimistic locking. It expects the caller to pass the CURRENT
// version value as seen by caller. If update succeeds, returns true and newVersion.
func (r *TicketRepository) UpdateStatus(ctx context.Context, id int64, oldStatus, newStatus string, expectedVersion int64) (bool, int64, error) {
	q := `
        UPDATE tickets
        SET status=$1, updated_at=NOW(), version=version+1
        WHERE id=$2 AND status=$3 AND version=$4
        RETURNING version
    `
	var newVersion int64
	err := r.db.QueryRowContext(ctx, q, newStatus, id, oldStatus, expectedVersion).Scan(&newVersion)
	if err == sql.ErrNoRows {
		return false, 0, nil
	}
	if err != nil {
		return false, 0, err
	}
	return true, newVersion, nil
}

// Optional: Move ticket to history (archival). Not strictly required but recommended.
func (r *TicketRepository) Archive(ctx context.Context, id int64) error {
	// simplistic example; adapt to your schema
	_, err := r.db.ExecContext(ctx, `
        INSERT INTO ticket_history (id, queue_id, customer_name, status, priority, estimated_time, version, created_at, updated_at)
        SELECT id, queue_id, customer_name, status, priority, estimated_time, version, created_at, updated_at
        FROM tickets WHERE id=$1
    `, id)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `DELETE FROM tickets WHERE id=$1`, id)
	return err
}

// Optional helper to requeue a ticket to waiting (e.g., on failure)
func (r *TicketRepository) RequeueToWaiting(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `
        UPDATE tickets
        SET status='waiting', updated_at=NOW(), version=version+1
        WHERE id=$1
    `, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("requeue failed: no rows updated")
	}
	return nil
}
