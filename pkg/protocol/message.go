package protocol

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
)

// Message protocol
type Message struct {
	Version uint8
	Type    MessageType
	Flags   uint16
	Length  uint32
	Payload []byte
}

// Marshal serializes message to bytes
func (m *Message) Marshal() ([]byte, error) {
	if m.Version != Version {
		return nil, ErrInvalidVersion
	}

	if len(m.Payload) > MaxMessageSize-HeaderSize {
		return nil, ErrMessageTooLarge
	}

	totalSize := HeaderSize + len(m.Payload)
	buf := make([]byte, totalSize)

	buf[0] = m.Version
	buf[1] = byte(m.Type)
	binary.BigEndian.PutUint16(buf[2:4], m.Flags)
	binary.BigEndian.PutUint32(buf[4:8], uint32(len(m.Payload)))

	if len(m.Payload) > 0 {
		copy(buf[HeaderSize:], m.Payload)
	}

	return buf, nil
}

// Unmarshal deserializes bytes to message
func Unmarshal(data []byte) (*Message, error) {
	if len(data) < HeaderSize {
		return nil, ErrInvalidPayload
	}

	msg := &Message{
		Version: data[0],
		Type:    MessageType(data[1]),
		Flags:   binary.BigEndian.Uint16(data[2:4]),
		Length:  binary.BigEndian.Uint32(data[4:8]),
	}

	if msg.Version != Version {
		return nil, ErrInvalidVersion
	}

	if msg.Length > MaxMessageSize-HeaderSize {
		return nil, ErrMessageTooLarge
	}

	if msg.Length > 0 && len(data) >= HeaderSize+int(msg.Length) {
		msg.Payload = make([]byte, msg.Length)
		copy(msg.Payload, data[HeaderSize:HeaderSize+msg.Length])
	}

	return msg, nil
}

// GetPayload deserializes payload
func (m *Message) GetPayload(v PayloadProvider) error {
	if len(m.Payload) == 0 {
		return nil
	}

	if err := json.Unmarshal(m.Payload, v); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	if err := v.Validate(); err != nil {
		return fmt.Errorf("payload validation failed: %w", err)
	}

	return nil
}

// SetPayload serializes payload
func (m *Message) SetPayload(v PayloadProvider) error {
	if err := v.Validate(); err != nil {
		return fmt.Errorf("payload validation failed: %w", err)
	}

	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	if len(data) > MaxMessageSize-HeaderSize {
		return ErrMessageTooLarge
	}

	m.Payload = data
	m.Length = uint32(len(data))
	return nil
}
