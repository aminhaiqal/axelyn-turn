package models

import "time"

type TicketStatus string

const (
    StatusWaiting    TicketStatus = "waiting"
    StatusInProgress TicketStatus = "in_progress"
    StatusDone       TicketStatus = "done"
)

type Ticket struct {
    ID              int64        `json:"id"`
    QueueID         int64        `json:"queue_id"`
    CustomerName    string       `json:"customer_name"`
    Status          TicketStatus `json:"status"` 
    Priority        int          `json:"priority"`
    AssignedWorker  int64        `json:"assigned_worker,omitempty"`
    EstimatedTime   int          `json:"estimated_time"` // seconds
    CreatedAt       time.Time    `json:"created_at"`
    UpdatedAt       time.Time    `json:"updated_at"`
    CompletedAt     *time.Time   `json:"completed_at,omitempty"`
    Version         int64        `json:"version"` // optimistic locking
}
