package quotes

import (
	"math/rand"
	"wiseguard/pkg/logger"
)

type Quote struct {
	Text   string
	Author string
}

// Service qoutes service
type Service interface {
	GetRandomQuote() *Quote
}

// service for quotes
type service struct {
	quotes []Quote
	log    logger.Logger
}

// NewService new quotes service
func NewService(log logger.Logger) Service {
	return &service{
		quotes: []Quote{
			{Text: "The only true wisdom is in knowing you know nothing.", Author: "Socrates"},
			{Text: "Life is really simple, but we insist on making it complicated.", Author: "Confucius"},
			{Text: "The unexamined life is not worth living.", Author: "Socrates"},
			{Text: "The journey of a thousand miles begins with one step.", Author: "Lao Tzu"},
		},
		log: log.WithComponent("quotes"),
	}
}

// GetRandomQuote returns random quote
func (s *service) GetRandomQuote() *Quote {
	if len(s.quotes) == 0 {
		s.log.Error("quotes collection is empty", nil, nil)
		return &Quote{
			Text:   "No quotes available",
			Author: "System",
		}
	}

	quote := s.quotes[rand.Intn(len(s.quotes))]
	s.log.Debug("returning quote", map[string]interface{}{
		"quote": quote,
	})
	return &quote
}

// GetQuotesCount returns count of quotes
func (s *service) GetQuotesCount() int {
	return len(s.quotes)
}
