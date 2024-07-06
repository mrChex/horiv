package sdk

import (
	"encoding/json"
	"fmt"
	"github.com/sashabaranov/go-openai"
)

const MemoryAllCategoriesFunctionName = "MemoryAllCategories"

func (s *SDK) DefinitionMemoryAllCategories() openai.FunctionDefinition {
	return openai.FunctionDefinition{
		Name:        MemoryAllCategoriesFunctionName,
		Description: "Отримати всі категорії знань",
		Parameters: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
			"response": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
		},
	}
}

func (s *SDK) ExecuteMemoryAllCategories(chatID int64, call openai.ToolCall) ([]openai.ChatCompletionMessage, error) {
	if call.Type != openai.ToolTypeFunction || call.Function.Name != MemoryAllCategoriesFunctionName {
		return nil, fmt.Errorf("invalid call type")
	}

	categories, err := s.store.MemoryAllCategories(chatID)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}

	rsp, _ := json.Marshal(categories)

	return []openai.ChatCompletionMessage{
		{
			Role:       openai.ChatMessageRoleTool,
			ToolCallID: call.ID,
			Content:    string(rsp),
		},
	}, nil

}

func (s *SDK) MemoryAllCategories(chatID int64) ([]string, error) {
	return s.store.MemoryAllCategories(chatID)
}
