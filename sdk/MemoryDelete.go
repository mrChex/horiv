package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-telegram/bot"
	tgModels "github.com/go-telegram/bot/models"
	"github.com/sashabaranov/go-openai"
)

const MemoryDeleteFunctionName = "MemoryDelete"

func (s *SDK) DefinitionMemoryDelete() openai.FunctionDefinition {
	return openai.FunctionDefinition{
		Name:        MemoryDeleteFunctionName,
		Description: "Видалення факту з памʼяті",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"category": map[string]interface{}{
					"type": "string",
				},
				"key": map[string]interface{}{
					"type": "string",
				},
			},
			"response": map[string]interface{}{
				"type": "string",
				"enum": []string{
					"OK",
				},
			},
			"required": []string{
				"category",
				"key",
			},
		},
	}
}

func (s *SDK) ExecuteMemoryDelete(chatID int64, call openai.ToolCall) ([]openai.ChatCompletionMessage, error) {
	if call.Type != openai.ToolTypeFunction || call.Function.Name != MemoryDeleteFunctionName {
		return nil, fmt.Errorf("invalid call type")
	}
	ctx := context.Background()
	_, _ = s.bot.SendChatAction(ctx, &bot.SendChatActionParams{
		ChatID: chatID,
		Action: tgModels.ChatActionChooseSticker,
	})

	var args struct {
		Category string `json:"category"`
		Key      string `json:"key"`
	}
	err := json.Unmarshal([]byte(call.Function.Arguments), &args)
	// TODO: maybe we should told GPT there is invalid json?
	if err != nil {
		return nil, fmt.Errorf("unmarshal arguments: %w", err)
	}

	memoryChatID := chatID
	if args.Category == "GLOBAL_CONTEXT" {
		memoryChatID = 0
	}

	err = s.store.MemoryDelete(memoryChatID, args.Category, args.Key)
	if err != nil {
		return nil, fmt.Errorf("memory delete: %w", err)
	}

	rsp := []openai.ChatCompletionMessage{
		{
			Role:       openai.ChatMessageRoleTool,
			ToolCallID: call.ID,
			Content:    "OK",
		},
	}

	return rsp, nil
}
