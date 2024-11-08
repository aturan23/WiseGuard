package protocol

import (
	"errors"
	"time"
)

// Validate for ChallengePayload
func (p *ChallengePayload) Validate() error {
	if p.Prefix == "" {
		return errors.New("empty prefix")
	}
	if p.Difficulty == 0 {
		return errors.New("zero difficulty")
	}
	if p.Nonce == "" {
		return errors.New("empty nonce")
	}
	if p.ExpiresAt.IsZero() {
		return errors.New("zero expiration time")
	}
	if p.ExpiresAt.Before(time.Now()) {
		return errors.New("challenge already expired")
	}
	return nil
}

// Validate for SolutionPayload
func (p *SolutionPayload) Validate() error {
	if p.Prefix == "" {
		return errors.New("empty prefix")
	}
	if p.Solution == "" {
		return errors.New("empty solution")
	}
	if p.Nonce == "" {
		return errors.New("empty nonce")
	}
	return nil
}

// Validate for QuotePayload
func (p *QuotePayload) Validate() error {
	if p.Text == "" {
		return errors.New("empty quote text")
	}
	return nil
}

// Validate for ErrorPayload
func (p *ErrorPayload) Validate() error {
	if p.Code == "" {
		return errors.New("empty error code")
	}
	if p.Message == "" {
		return errors.New("empty error message")
	}
	return nil
}
