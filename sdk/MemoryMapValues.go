package sdk

import (
	"encoding/json"
	"fmt"
	"github.com/sashabaranov/go-openai"
)

const MemoryMapValuesFunctionName = "MemoryMapValues"

func (s *SDK) DefinitionMemoryMapValues() openai.FunctionDefinition {
	return openai.FunctionDefinition{
		Name:        MemoryMapValuesFunctionName,
		Description: "Отримати значення всіх наданих ключів",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"category": map[string]interface{}{
					"type": "string",
				},
				"keys": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
			},
			"response": map[string]interface{}{
				"type":        "array",
				"description": "Масив значень в тому ж порядку як і ключі",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
			"required": []string{
				"category",
				"keys",
			},
		},
	}
}

func (s *SDK) ExecuteMemoryMapValues(chatID int64, call openai.ToolCall) ([]openai.ChatCompletionMessage, error) {
	if call.Type != openai.ToolTypeFunction || call.Function.Name != MemoryMapValuesFunctionName {
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

func (s *SDK) MemoryMemoryMapValues(chatID int64, category string) ([]string, error) {
	return s.store.MemoryAllCategoryKeys(chatID, category)
}
