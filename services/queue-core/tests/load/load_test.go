package load

import (
	"context"
	"sync"
	"testing"
	"strconv"
	"time"
	"queue-core/internal/models"
	"queue-core/internal/services"
)

func TestConcurrentTicketCreation(t *testing.T) {
	// Setup your service (repo + Redis)
	var service *services.TicketService // init with real or mock connections

	ctx := context.Background()
	const n = 100 // number of concurrent requests

	wg := sync.WaitGroup{}
	wg.Add(n)

	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			ticket := &models.Ticket{
				QueueID:      1,
				CustomerName: "LoadTest #" + strconv.Itoa(i),
				Priority:     1,
				Status:       "waiting",
			}
			_, _ = service.CreateTicket(ctx, ticket)
		}(i)
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("load test timed out")
	}
}
