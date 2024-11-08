package pow

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"
	"wiseguard/pkg/logger"
	"wiseguard/pkg/protocol"
)

// Challenge presents a PoW challenge
type Challenge struct {
	Prefix     string
	Difficulty uint8
	Nonce      string
	ExpiresAt  time.Time
}

// Service interface provides PoW functionality
type Service interface {
	CreateChallenge(difficulty uint8) (*Challenge, error)
	VerifySolution(challenge *Challenge, solution *protocol.SolutionPayload) bool
}

// service provides PoW functionality
type service struct {
	log        logger.Logger
	ttl        time.Duration
	challenges sync.Map
}

func NewService(log logger.Logger, ttl time.Duration) Service {
	if ttl == 0 {
		ttl = 5 * time.Minute
	}

	return &service{
		log: log.WithComponent("pow"),
		ttl: ttl,
	}
}

// CreateChallenge creates a new PoW challenge
func (s *service) CreateChallenge(difficulty uint8) (*Challenge, error) {
	s.log.Debug("creating challenge", map[string]interface{}{
		"requested_difficulty": difficulty,
	})

	if difficulty == 0 {
		return nil, fmt.Errorf("difficulty must be greater than 0")
	}

	prefix, err := generateRandomString(16)
	if err != nil {
		return nil, fmt.Errorf("failed to generate prefix: %w", err)
	}

	nonce, err := generateRandomString(16)
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	challenge := &Challenge{
		Prefix:     prefix,
		Difficulty: difficulty,
		Nonce:      nonce,
		ExpiresAt:  time.Now().Add(s.ttl),
	}

	// save challenge
	s.challenges.Store(challenge.Nonce, challenge)

	// remove challenge after ttl
	go func() {
		time.Sleep(s.ttl)
		s.challenges.Delete(challenge.Nonce)
		s.log.Debug("challenge expired", map[string]interface{}{
			"nonce": challenge.Nonce,
		})
	}()

	return challenge, nil
}

// VerifySolution verifies the PoW solution
func (s *service) VerifySolution(challenge *Challenge, solution *protocol.SolutionPayload) bool {
	savedChallenge, exists := s.challenges.Load(solution.Nonce)
	if !exists {
		s.log.Debug("challenge not found", map[string]interface{}{
			"nonce": solution.Nonce,
		})
		return false
	}

	ch, ok := savedChallenge.(*Challenge)
	if !ok {
		s.log.Error("invalid challenge type in storage", nil, nil)
		return false
	}

	if ch.Prefix != solution.Prefix {
		return false
	}

	if time.Now().After(ch.ExpiresAt) {
		return false
	}

	hash := calculateHash(solution.Prefix + solution.Solution)
	return validateDifficulty(hash, ch.Difficulty)
}

func generateRandomString(length int) (string, error) {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func calculateHash(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

func validateDifficulty(hash string, difficulty uint8) bool {
	prefix := strings.Repeat("0", int(difficulty))
	return strings.HasPrefix(hash, prefix)
}
