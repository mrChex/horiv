package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-telegram/bot"
	tgModels "github.com/go-telegram/bot/models"
	"github.com/sashabaranov/go-openai"
)

const MemoryUpsertFunctionName = "MemoryUpsert"

func (s *SDK) DefinitionMemoryUpsert() openai.FunctionDefinition {
	return openai.FunctionDefinition{
		Name:        MemoryUpsertFunctionName,
		Description: "Додати або оновити данні в довготривалій памʼяті",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"category": map[string]interface{}{
					"type":        "string",
					"description": "Категорія факту для групування знань по категоріям. Якщо раніше такої категорії не було - вона створиться автоматично",
				},
				"key": map[string]interface{}{
					"type":        "string",
					"description": "Унікальний в рамках категорії ключ для збереження факту, має однозначно характеризувати значення яке міститься під цим ключем",
				},
				"value": map[string]interface{}{
					"type":        "string",
					"description": "Факт який необхідно запамʼятати",
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
				"value",
			},
		},
	}
}

func (s *SDK) ExecuteMemoryUpsert(chatID int64, call openai.ToolCall) ([]openai.ChatCompletionMessage, error) {
	if call.Type != openai.ToolTypeFunction || call.Function.Name != MemoryUpsertFunctionName {
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
		Value    string `json:"value"`
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

	err = s.MemoryUpsert(memoryChatID, args.Category, args.Key, args.Value)
	if err != nil {
		return nil, fmt.Errorf("memory upsert: %w", err)
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

func (s *SDK) MemoryUpsert(chatID int64, category, key, value string) error {
	return s.store.MemoryUpsert(chatID, category, key, value)
}
