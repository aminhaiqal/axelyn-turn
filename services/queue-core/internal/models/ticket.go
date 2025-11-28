package models

import "time"

type Ticket struct {
    ID            int       `json:"id"`
    QueueID       int       `json:"queue_id"`
    CustomerName  string    `json:"customer_name"`
    Status        string    `json:"status"`       // waiting, in_progress, done
    Priority      int       `json:"priority"`
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
    EstimatedTime int       `json:"estimated_time"` // seconds
}
