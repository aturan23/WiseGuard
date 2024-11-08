package client

import (
	"context"
	"fmt"
	"wiseguard/pkg/protocol"
	"wiseguard/pkg/utils"
)

// readMessage reads a message from the server
func (c *Client) readMessage(ctx context.Context) (*protocol.Message, error) {
	headerBuf := make([]byte, protocol.HeaderSize)
	if err := utils.ReadFull(c.conn, headerBuf, c.cfg.ReadTimeout); err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	// Header unmarshal
	msg, err := protocol.Unmarshal(headerBuf)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal header: %w", err)
	}

	// Read payload
	if msg.Length > 0 {
		payload := make([]byte, msg.Length)
		if err := utils.ReadFull(c.conn, payload, c.cfg.ReadTimeout); err != nil {
			return nil, fmt.Errorf("failed to read payload: %w", err)
		}
		msg.Payload = payload
	}

	return msg, nil
}

// sendMessage sends a message to the server
func (c *Client) sendMessage(ctx context.Context, msg *protocol.Message) error {
	data, err := msg.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	if err := utils.WriteFull(c.conn, data, c.cfg.WriteTimeout); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}
