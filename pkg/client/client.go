package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"
	"wiseguard/pkg/logger"
	"wiseguard/pkg/protocol"
	"wiseguard/pkg/utils"
)

// Config client configuration
type Config struct {
	ServerAddress  string
	ConnectTimeout time.Duration
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	RetryConfig    utils.RetryConfig
}

// Client server client
type Client struct {
	cfg  *Config
	log  logger.Logger
	conn net.Conn
}

func NewClient(cfg *Config, log logger.Logger) *Client {
	return &Client{
		cfg: cfg,
		log: log.WithComponent("client"),
	}
}

// Connect connects to the server
func (c *Client) Connect(ctx context.Context) error {
	dialer := &net.Dialer{
		Timeout: c.cfg.ConnectTimeout,
	}

	var err error
	c.conn, err = dialer.DialContext(ctx, "tcp", c.cfg.ServerAddress)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	c.log.Info("connected to server", map[string]interface{}{
		"address": c.cfg.ServerAddress,
	})

	return nil
}

// GetQuote gets a quote from the server
func (c *Client) GetQuote(ctx context.Context) (*protocol.QuotePayload, error) {
	// Receive PoW challenge
	challenge, err := c.receiveChallenge(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to receive challenge: %w", err)
	}

	// Solve PoW challenge
	solution, err := utils.SolvePoW(ctx, challenge.Prefix, challenge.Difficulty)
	if err != nil {
		return nil, fmt.Errorf("failed to solve PoW: %w", err)
	}

	// Send solution
	if err := c.sendSolution(ctx, challenge, solution); err != nil {
		return nil, fmt.Errorf("failed to send solution: %w", err)
	}

	return c.receiveQuote(ctx)
}

// receiveChallenge receives a challenge from the server
func (c *Client) receiveChallenge(ctx context.Context) (*protocol.ChallengePayload, error) {
	msg, err := c.readMessage(ctx)
	if err != nil {
		return nil, err
	}

	if msg.Type != protocol.TypeChallenge {
		return nil, fmt.Errorf("unexpected message type: %v", msg.Type)
	}

	var challenge protocol.ChallengePayload
	if err := json.Unmarshal(msg.Payload, &challenge); err != nil {
		return nil, fmt.Errorf("failed to unmarshal challenge: %w", err)
	}

	c.log.Debug("received challenge", map[string]interface{}{
		"difficulty": challenge.Difficulty,
		"expires_at": challenge.ExpiresAt,
	})

	return &challenge, nil
}

// sendSolution sends a solution to the server
func (c *Client) sendSolution(ctx context.Context, challenge *protocol.ChallengePayload, solution string) error {
	payload := &protocol.SolutionPayload{
		Prefix:   challenge.Prefix,
		Solution: solution,
		Nonce:    challenge.Nonce,
	}

	msg, err := protocol.NewMessage(protocol.TypeSolution, payload)
	if err != nil {
		return err
	}

	return c.sendMessage(ctx, msg)
}

// receiveQuote receives a quote from the server
func (c *Client) receiveQuote(ctx context.Context) (*protocol.QuotePayload, error) {
	msg, err := c.readMessage(ctx)
	if err != nil {
		return nil, err
	}

	switch msg.Type {
	case protocol.TypeQuote:
		var quote protocol.QuotePayload
		if err := json.Unmarshal(msg.Payload, &quote); err != nil {
			return nil, fmt.Errorf("failed to unmarshal quote: %w", err)
		}
		return &quote, nil

	case protocol.TypeError:
		var errPayload protocol.ErrorPayload
		if err := json.Unmarshal(msg.Payload, &errPayload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal error: %w", err)
		}
		return nil, fmt.Errorf("server error: %s - %s", errPayload.Code, errPayload.Message)

	default:
		return nil, fmt.Errorf("unexpected message type: %v", msg.Type)
	}
}

// Close disconnects from the server
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
