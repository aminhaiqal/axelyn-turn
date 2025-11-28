package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"queue-core/internal/models"
	"queue-core/internal/repositories"
)

type TicketService struct {
	Repo       *repositories.TicketRepository
	Rdb        *redis.Client
	StreamName string // e.g. "queue.jobs" OR per-queue "queue.<id>.events"
	PubSubBase string // base channel for websocket broadcasts, e.g. "queue.%d.broadcast"
}

// NewTicketService requires repo and a configured redis client.
func NewTicketService(repo *repositories.TicketRepository, rdb *redis.Client, streamName, pubSubBase string) *TicketService {
	return &TicketService{Repo: repo, Rdb: rdb, StreamName: streamName, PubSubBase: pubSubBase}
}

// CreateTicket writes DB then publishes to Redis Stream and PubSub.
// Returns created ticket id.
func (s *TicketService) CreateTicket(ctx context.Context, ticket *models.Ticket) (int64, error) {
	// ensure defaults
	if ticket.Status == "" {
		ticket.Status = "waiting"
	}
	if ticket.Priority == 0 {
		ticket.Priority = 1
	}

	if err := s.Repo.Create(ctx, ticket); err != nil {
		return 0, err
	}

	// Push to Redis stream (minimal payload)
	streamKey := fmt.Sprintf("%s", s.StreamName) // you may choose fmt.Sprintf("%s.%d", s.StreamName, ticket.QueueID)
	values := map[string]interface{}{
		"ticket_id":  ticket.ID,
		"queue_id":   ticket.QueueID,
		"event":      "ticket.created",
		"created_at": time.Now().UTC().Format(time.RFC3339),
	}
	_, err := s.Rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: streamKey,
		Values: values,
	}).Result()
	if err != nil {
		// log and continue — creation already persisted.
		// In production you may implement retry/poison queue.
	}

	// Publish to Pub/Sub for websocket clients (fast fanout)
	pubChannel := fmt.Sprintf(s.PubSubBase, ticket.QueueID)
	b, _ := json.Marshal(map[string]interface{}{
		"event":     "ticket.created",
		"ticket_id": ticket.ID,
		"queue_id":  ticket.QueueID,
	})
	_ = s.Rdb.Publish(ctx, pubChannel, b).Err() // best-effort

	return ticket.ID, nil
}

// StartDispatcher starts a goroutine that continuously attempts to reserve tickets
// for a given queueID and publishes reservation events to Redis Stream + PubSub.
// This keeps workers decoupled: workers consume the stream and be sure a ticket was reserved.
func (s *TicketService) StartDispatcher(ctx context.Context, queueID int, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			// Attempt to reserve as many as available in tight loop until no ticket or until next tick
			for {
				t, err := s.Repo.ReserveNext(ctx, queueID)
				if err != nil {
					// Log and break to avoid tight error loop
					// Use your logger; here we use fmt
					fmt.Printf("ReserveNext error: %v\n", err)
					break
				}
				if t == nil {
					// nothing waiting right now
					break
				}

				// build event payload
				event := map[string]interface{}{
					"event":      "ticket.reserved",
					"ticket_id":  t.ID,
					"queue_id":   t.QueueID,
					"status":     "processing",
					"version":    t.Version + 1, // ReserveNext bumped the version
					"reserved_at": time.Now().UTC().Format(time.RFC3339),
				}

				// push to stream (for workers)
				streamKey := fmt.Sprintf("%s.%d", s.StreamName, queueID)
				_, err = s.Rdb.XAdd(ctx, &redis.XAddArgs{
					Stream: streamKey,
					Values: event,
				}).Result()
				if err != nil {
					fmt.Printf("XAdd error: %v\n", err)
					// you might want to requeue the ticket in DB or mark for reconciliation
				}

				// also publish to pubsub for websocket frontends
				pubChannel := fmt.Sprintf(s.PubSubBase, queueID)
				b, _ := json.Marshal(event)
				_ = s.Rdb.Publish(ctx, pubChannel, b).Err()
			}

			// wait for next tick
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

// PublishWorkerUpdate allows other components (e.g., a worker) to send back computed updates.
// Useful when worker wants Core to persist estimated_time or other improvements.
func (s *TicketService) PublishWorkerUpdate(ctx context.Context, queueID int, payload map[string]interface{}) error {
	streamKey := fmt.Sprintf("%s.worker.updates.%d", s.StreamName, queueID)
	_, err := s.Rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: streamKey,
		Values: payload,
	}).Result()
	if err != nil {
		return err
	}
	// also publish to pubsub so WebSocket clients see it in real-time
	pubChannel := fmt.Sprintf(s.PubSubBase, queueID)
	b, _ := json.Marshal(payload)
	return s.Rdb.Publish(ctx, pubChannel, b).Err()
}
