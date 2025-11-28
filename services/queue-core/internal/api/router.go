package api

import (
    "encoding/json"
    "net/http"
    "queue-core/internal/models"
    "queue-core/internal/services"
)

type API struct {
    TicketService *services.TicketService
}

func NewAPI(ts *services.TicketService) *API {
    return &API{TicketService: ts}
}

func (a *API) Router() http.Handler {
    mux := http.NewServeMux()
    mux.HandleFunc("/tickets", a.createTicketHandler)
    mux.HandleFunc("/tickets/waiting", a.listWaitingHandler)
    return mux
}

func (a *API) createTicketHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
    var ticket models.Ticket
    json.NewDecoder(r.Body).Decode(&ticket)
    id, err := a.TicketService.CreateTicket(&ticket)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        return
    }
    json.NewEncoder(w).Encode(map[string]int{"id": id})
}

func (a *API) listWaitingHandler(w http.ResponseWriter, r *http.Request) {
    queueID := 1 // hardcoded for Day-1 simplicity
    tickets, _ := a.TicketService.ListWaitingTickets(queueID)
    json.NewEncoder(w).Encode(tickets)
}
