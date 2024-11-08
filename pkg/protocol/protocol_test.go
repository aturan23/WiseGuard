package protocol

import (
	"testing"
	"time"
)

func TestMessageMarshaling(t *testing.T) {
	tests := []struct {
		name    string
		message *Message
		wantErr bool
	}{
		{
			name: "valid message with payload",
			message: &Message{
				Version: Version,
				Type:    TypeQuote,
				Flags:   0,
				Payload: []byte(`{"text":"test quote","author":"test author"}`),
			},
			wantErr: false,
		},
		{
			name: "invalid version",
			message: &Message{
				Version: Version + 1,
				Type:    TypeQuote,
				Flags:   0,
				Payload: []byte(`{"text":"test"}`),
			},
			wantErr: true,
		},
		{
			name: "payload too large",
			message: &Message{
				Version: Version,
				Type:    TypeQuote,
				Flags:   0,
				Payload: make([]byte, MaxMessageSize),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Тест маршалинга
			data, err := tt.message.Marshal()
			if (err != nil) != tt.wantErr {
				t.Errorf("Message.Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Тест анмаршалинга
			got, err := Unmarshal(data)
			if err != nil {
				t.Errorf("Unmarshal() error = %v", err)
				return
			}

			if got.Version != tt.message.Version {
				t.Errorf("Version = %v, want %v", got.Version, tt.message.Version)
			}
			if got.Type != tt.message.Type {
				t.Errorf("Type = %v, want %v", got.Type, tt.message.Type)
			}
			if got.Flags != tt.message.Flags {
				t.Errorf("Flags = %v, want %v", got.Flags, tt.message.Flags)
			}
			if len(got.Payload) != len(tt.message.Payload) {
				t.Errorf("Payload length = %v, want %v", len(got.Payload), len(tt.message.Payload))
			}
		})
	}
}

func TestPayloadValidation(t *testing.T) {
	tests := []struct {
		name    string
		payload PayloadProvider
		wantErr bool
	}{
		{
			name: "valid quote payload",
			payload: &QuotePayload{
				Text:   "test quote",
				Author: "test author",
			},
			wantErr: false,
		},
		{
			name: "invalid quote payload - empty text",
			payload: &QuotePayload{
				Text:   "",
				Author: "test author",
			},
			wantErr: true,
		},
		{
			name: "valid challenge payload",
			payload: &ChallengePayload{
				Prefix:     "test",
				Difficulty: 4,
				Nonce:      "testnonce",
				ExpiresAt:  time.Now().Add(time.Hour),
			},
			wantErr: false,
		},
		{
			name: "invalid challenge payload - zero difficulty",
			payload: &ChallengePayload{
				Prefix:     "test",
				Difficulty: 0,
				Nonce:      "testnonce",
				ExpiresAt:  time.Now().Add(time.Hour),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.payload.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
