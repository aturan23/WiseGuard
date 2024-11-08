package quotes

import (
	"testing"
	"wiseguard/pkg/logger"
)

type mockLogger struct{}

func (m *mockLogger) Debug(msg string, fields map[string]interface{})            {}
func (m *mockLogger) Info(msg string, fields map[string]interface{})             {}
func (m *mockLogger) Error(msg string, err error, fields map[string]interface{}) {}
func (m *mockLogger) Fatal(msg string, err error, fields map[string]interface{}) {}
func (m *mockLogger) WithComponent(component string) logger.Logger               { return m }
func (m *mockLogger) WithFields(fields map[string]interface{}) logger.Logger     { return m }

func TestGetRandomQuote(t *testing.T) {
	s := NewService(&mockLogger{})

	// Check that we can get a quote
	for i := 0; i < 10; i++ {
		quote := s.GetRandomQuote()
		if quote == nil {
			t.Error("GetRandomQuote() returned nil")
			continue
		}

		if quote.Text == "" {
			t.Error("Quote text is empty")
		}
		if quote.Author == "" {
			t.Error("Quote author is empty")
		}
	}
}

func TestQuoteUniqueness(t *testing.T) {
	s := NewService(&mockLogger{})
	quotes := make(map[string]bool)
	iterations := 100

	// Check that we get different quotes
	for i := 0; i < iterations; i++ {
		quote := s.GetRandomQuote()
		quotes[quote.Text] = true
	}

	// Should have more than one quote
	if len(quotes) == 1 {
		t.Error("All quotes are the same, randomization might not work")
	}
}
