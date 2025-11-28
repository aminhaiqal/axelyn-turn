package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"queue-core/internal/models"
	"queue-core/internal/services"
)

type API struct {
	TicketService *services.TicketService
	Rdb           *redis.Client
	StreamName    string
	PubSubBase    string
	upgrader      websocket.Upgrader
}

func NewAPI(ts *services.TicketService, rdb *redis.Client, streamName, pubSubBase string) *API {
	return &API{
		TicketService: ts,
		Rdb:           rdb,
		StreamName:    streamName,
		PubSubBase:    pubSubBase,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

func (a *API) Router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/tickets", a.createTicketHandler)
	mux.HandleFunc("/tickets/waiting", a.listWaitingHandler)
	mux.HandleFunc("/ws", a.wsHandler) // ?queue_id=1
	return mux
}

func (a *API) createTicketHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var ticket models.Ticket
	if err := json.NewDecoder(r.Body).Decode(&ticket); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	id, err := a.TicketService.CreateTicket(ctx, &ticket)
	if err != nil {
		http.Error(w, "failed create", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]int64{"id": id})
}

func (a *API) listWaitingHandler(w http.ResponseWriter, r *http.Request) {
	// queue_id as query param
	qidStr := r.URL.Query().Get("queue_id")
	if qidStr == "" {
		http.Error(w, "queue_id required", http.StatusBadRequest)
		return
	}
	qid, err := strconv.Atoi(qidStr)
	if err != nil {
		http.Error(w, "invalid queue_id", http.StatusBadRequest)
		return
	}
	tickets, err := a.TicketService.Repo.GetByStatus(r.Context(), qid, "waiting")
	if err != nil {
		http.Error(w, "failed list", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(tickets)
}

// wsHandler upgrades and subscribes to a Redis PubSub channel for queue updates.
// Clients connect with /ws?queue_id=1
func (a *API) wsHandler(w http.ResponseWriter, r *http.Request) {
	qidStr := r.URL.Query().Get("queue_id")
	if qidStr == "" {
		http.Error(w, "queue_id required", http.StatusBadRequest)
		return
	}
	qid, err := strconv.Atoi(qidStr)
	if err != nil {
		http.Error(w, "invalid queue_id", http.StatusBadRequest)
		return
	}

	conn, err := a.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	pubChName := fmt.Sprintf(a.PubSubBase, qid)
	ps := a.Rdb.Subscribe(ctx, pubChName)
	defer ps.Close()

	// consume messages and send to WS
	ch := ps.Channel()
	// ping loop to keep connection alive
	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			// forward raw message to client
			if err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
				return
			}
		case <-pingTicker.C:
			_ = conn.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(5*time.Second))
		}
	}
}
