package protocol

import (
	"errors"
	"time"
)

// Protocol constants
const (
	Version        = 1
	MaxMessageSize = 64 * 1024 // 64KB max message size
	HeaderSize     = 8         // 1 + 1 + 2 + 4 (version + type + flags + payload length)
)

// MessageType type of message
type MessageType uint8

const (
	TypeChallenge MessageType = iota + 1
	TypeSolution
	TypeQuote
	TypeError
)

// PayloadProvider interface for message payloads
type PayloadProvider interface {
	Validate() error
}

// Possible errors
var (
	ErrInvalidVersion  = errors.New("invalid protocol version")
	ErrMessageTooLarge = errors.New("message exceeds max size")
	ErrInvalidPayload  = errors.New("invalid payload")
)

// ChallengePayload contains PoW challenge
type ChallengePayload struct {
	Prefix     string    `json:"prefix"`
	Difficulty uint8     `json:"difficulty"`
	Nonce      string    `json:"nonce"`
	ExpiresAt  time.Time `json:"expires_at"`
}

// SolutionPayload contains PoW solution
type SolutionPayload struct {
	Prefix   string `json:"prefix"`
	Solution string `json:"solution"`
	Nonce    string `json:"nonce"`
}

// QuotePayload contains a quote
type QuotePayload struct {
	Text   string `json:"text"`
	Author string `json:"author"`
}

// ErrorPayload contains an error
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewMessage creates a new message
func NewMessage(msgType MessageType, payload interface{}) (*Message, error) {
	// Check if payload implements PayloadProvider
	p, ok := payload.(PayloadProvider)
	if !ok {
		return nil, errors.New("payload must implement PayloadProvider interface")
	}

	msg := &Message{
		Version: Version,
		Type:    msgType,
		Flags:   0,
	}

	if err := msg.SetPayload(p); err != nil {
		return nil, err
	}

	return msg, nil
}
