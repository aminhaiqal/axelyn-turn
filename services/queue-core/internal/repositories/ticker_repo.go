package repositories

import (
    "database/sql"
    "queue-core/internal/models"
    "time"
)

type TicketRepository struct {
    DB *sql.DB
}

func NewTicketRepo(db *sql.DB) *TicketRepository {
    return &TicketRepository{DB: db}
}

func (r *TicketRepository) Create(ticket *models.Ticket) (int, error) {
    var id int
    err := r.DB.QueryRow(`
        INSERT INTO tickets(queue_id, customer_name, status, priority, created_at, updated_at, estimated_time)
        VALUES($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
        ticket.QueueID, ticket.CustomerName, ticket.Status, ticket.Priority, time.Now(), time.Now(), ticket.EstimatedTime,
    ).Scan(&id)
    return id, err
}

// Fetch by status
func (r *TicketRepository) GetByStatus(queueID int, status string) ([]models.Ticket, error) {
    rows, err := r.DB.Query("SELECT id, queue_id, customer_name, status, priority, created_at, updated_at, estimated_time FROM tickets WHERE queue_id=$1 AND status=$2", queueID, status)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    tickets := []models.Ticket{}
    for rows.Next() {
        var t models.Ticket
        if err := rows.Scan(&t.ID, &t.QueueID, &t.CustomerName, &t.Status, &t.Priority, &t.CreatedAt, &t.UpdatedAt, &t.EstimatedTime); err != nil {
            return nil, err
        }
        tickets = append(tickets, t)
    }
    return tickets, nil
}
