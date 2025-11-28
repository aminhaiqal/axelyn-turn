package services

import (
	"context"
	"queue-core/internal/models"
	"queue-core/internal/repositories"
)

type TicketService struct {
    Repo *repositories.TicketRepository
}

func NewTicketService(repo *repositories.TicketRepository) *TicketService {
    return &TicketService{Repo: repo}
}

func (s *TicketService) CreateTicket(ticket *models.Ticket) (int, error) {
    err := s.Repo.Create(context.Background(), ticket)
    if err != nil {
        return 0, err
    }
    // Push event to Redis (later)
    return int(ticket.ID), nil
}

func (s *TicketService) ListWaitingTickets(queueID int) ([]*models.Ticket, error) {
    return s.Repo.GetByStatus(context.Background(), queueID, "waiting")
}