package pow

import (
	"fmt"
	"strings"
	"testing"
	"time"
	"wiseguard/pkg/logger"
	"wiseguard/pkg/protocol"
)

type mockLogger struct{}

func (m *mockLogger) Debug(msg string, fields map[string]interface{})            {}
func (m *mockLogger) Info(msg string, fields map[string]interface{})             {}
func (m *mockLogger) Error(msg string, err error, fields map[string]interface{}) {}
func (m *mockLogger) Fatal(msg string, err error, fields map[string]interface{}) {}
func (m *mockLogger) WithComponent(component string) logger.Logger               { return m }
func (m *mockLogger) WithFields(fields map[string]interface{}) logger.Logger     { return m }

func TestCreateChallenge(t *testing.T) {
	tests := []struct {
		name       string
		difficulty uint8
		ttl        time.Duration
		wantErr    bool
	}{
		{
			name:       "valid challenge",
			difficulty: 4,
			ttl:        time.Minute,
			wantErr:    false,
		},
		{
			name:       "zero difficulty",
			difficulty: 0,
			ttl:        time.Minute,
			wantErr:    true,
		},
		{
			name:       "valid challenge with max difficulty",
			difficulty: 8,
			ttl:        time.Minute,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewService(&mockLogger{}, tt.ttl)
			challenge, err := s.CreateChallenge(tt.difficulty)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateChallenge() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if challenge.Difficulty != tt.difficulty {
					t.Errorf("Difficulty = %v, want %v", challenge.Difficulty, tt.difficulty)
				}
				if challenge.Prefix == "" {
					t.Error("Prefix is empty")
				}
				if challenge.Nonce == "" {
					t.Error("Nonce is empty")
				}
				if challenge.ExpiresAt.Before(time.Now()) {
					t.Error("Challenge already expired")
				}
			}
		})
	}
}

func TestVerifySolution(t *testing.T) {
	s := NewService(&mockLogger{}, time.Minute)
	challenge, err := s.CreateChallenge(1) // Используем маленькую сложность для теста
	if err != nil {
		t.Fatalf("Failed to create challenge: %v", err)
	}

	solution := findSolution(challenge.Prefix, challenge.Difficulty)

	tests := []struct {
		name     string
		solution *protocol.SolutionPayload
		want     bool
	}{
		{
			name: "valid solution",
			solution: &protocol.SolutionPayload{
				Prefix:   challenge.Prefix,
				Solution: solution,
				Nonce:    challenge.Nonce,
			},
			want: true,
		},
		{
			name: "invalid prefix",
			solution: &protocol.SolutionPayload{
				Prefix:   "wrong",
				Solution: solution,
				Nonce:    challenge.Nonce,
			},
			want: false,
		},
		{
			name: "invalid nonce",
			solution: &protocol.SolutionPayload{
				Prefix:   challenge.Prefix,
				Solution: solution,
				Nonce:    "wrong",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := s.VerifySolution(challenge, tt.solution); got != tt.want {
				t.Errorf("VerifySolution() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChallengeTTL(t *testing.T) {
	ttl := 100 * time.Millisecond
	s := NewService(&mockLogger{}, ttl)

	challenge, err := s.CreateChallenge(1) // Используем минимальную сложность
	if err != nil {
		t.Fatalf("Failed to create challenge: %v", err)
	}

	solution := findSolution(challenge.Prefix, challenge.Difficulty)

	solutionPayload := &protocol.SolutionPayload{
		Prefix:   challenge.Prefix,
		Solution: solution,
		Nonce:    challenge.Nonce,
	}

	if !s.VerifySolution(challenge, solutionPayload) {
		t.Error("Solution should be valid immediately after creation")
		t.Logf("Challenge: %+v", challenge)
		t.Logf("Solution: %+v", solutionPayload)

		hash := calculateHash(challenge.Prefix + solution)
		t.Logf("Hash: %s", hash)
		t.Logf("Expected prefix: %s", strings.Repeat("0", int(challenge.Difficulty)))
	}

	// Wait for TTL
	time.Sleep(2 * ttl)

	if s.VerifySolution(challenge, solutionPayload) {
		t.Error("Solution should be invalid after TTL")
	}
}

func findSolution(prefix string, difficulty uint8) string {
	target := strings.Repeat("0", int(difficulty))
	nonce := 0
	for {
		solution := fmt.Sprintf("%d", nonce)
		hash := calculateHash(prefix + solution)
		if strings.HasPrefix(hash, target) {
			return solution
		}
		nonce++
		if nonce > 1000000 {
			panic("Could not find solution in reasonable time")
		}
	}
}
