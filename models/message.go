package models

import (
	"encoding/json"
	"github.com/sashabaranov/go-openai"
	"time"
)

type Message struct {
	Model      string                       `json:"model"`
	Completion openai.ChatCompletionMessage `json:"completion"`
	CreatedAt  time.Time                    `json:"created_at"`
}

func NewMessageFromJSON(data []byte) (Message, error) {
	var m Message
	err := json.Unmarshal(data, &m)
	if err != nil {
		return Message{}, err
	}
	return m, nil
}
