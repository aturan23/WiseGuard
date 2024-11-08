package server

import (
	"encoding/json"
	"fmt"
	"net"
	"wiseguard/pkg/pow"
	"wiseguard/pkg/protocol"
)

// readMessage reads a message from the client
func (s *Server) readMessage(conn net.Conn) (*protocol.Message, error) {
	headerBuf := make([]byte, protocol.HeaderSize)
	_, err := conn.Read(headerBuf)
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	msg, err := protocol.Unmarshal(headerBuf)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal header: %w", err)
	}

	if msg.Length > 0 {
		payload := make([]byte, msg.Length)
		_, err := conn.Read(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to read payload: %w", err)
		}
		msg.Payload = payload
	}

	return msg, nil
}

// sendMessage sends a message to the client
func (s *Server) sendMessage(conn net.Conn, msg *protocol.Message) error {
	data, err := msg.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	_, err = conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}

// sendChallenge sends a challenge to the client
func (s *Server) sendChallenge(conn net.Conn, challenge *pow.Challenge) error {
	payload := &protocol.ChallengePayload{
		Prefix:     challenge.Prefix,
		Difficulty: challenge.Difficulty,
		Nonce:      challenge.Nonce,
		ExpiresAt:  challenge.ExpiresAt,
	}

	s.log.Debug("sending challenge", map[string]interface{}{
		"prefix":     payload.Prefix,
		"difficulty": payload.Difficulty,
		"nonce":      payload.Nonce,
		"expires_at": payload.ExpiresAt,
	})

	msg, err := protocol.NewMessage(protocol.TypeChallenge, payload)
	if err != nil {
		return fmt.Errorf("failed to create challenge message: %w", err)
	}

	return s.sendMessage(conn, msg)
}

// readSolution reads a solution from the client
func (s *Server) readSolution(conn net.Conn) (*protocol.SolutionPayload, error) {
	msg, err := s.readMessage(conn)
	if err != nil {
		return nil, err
	}

	if msg.Type != protocol.TypeSolution {
		return nil, fmt.Errorf("unexpected message type: %v", msg.Type)
	}

	var solution protocol.SolutionPayload
	if err := json.Unmarshal(msg.Payload, &solution); err != nil {
		return nil, fmt.Errorf("failed to unmarshal solution: %w", err)
	}

	s.log.Debug("received solution", map[string]interface{}{
		"prefix":   solution.Prefix,
		"solution": solution.Solution,
		"nonce":    solution.Nonce,
	})

	return &solution, nil
}

// sendQuote sends a quote to the client
func (s *Server) sendQuote(conn net.Conn) error {
	quote := s.quoteService.GetRandomQuote()
	payload := &protocol.QuotePayload{
		Text:   quote.Text,
		Author: quote.Author,
	}

	msg, err := protocol.NewMessage(protocol.TypeQuote, payload)
	if err != nil {
		return fmt.Errorf("failed to create quote message: %w", err)
	}

	return s.sendMessage(conn, msg)
}

// sendError sends an error to the client
func (s *Server) sendError(conn net.Conn, code, message string) error {
	payload := &protocol.ErrorPayload{
		Code:    code,
		Message: message,
	}

	msg, err := protocol.NewMessage(protocol.TypeError, payload)
	if err != nil {
		return fmt.Errorf("failed to create error message: %w", err)
	}

	return s.sendMessage(conn, msg)
}
