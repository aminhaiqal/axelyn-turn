package services

import (
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
    id, err := s.Repo.Create(ticket)
    if err != nil {
        return 0, err
    }
    // Push event to Redis (later)
    return id, nil
}

func (s *TicketService) ListWaitingTickets(queueID int) ([]models.Ticket, error) {
    return s.Repo.GetByStatus(queueID, "waiting")
}