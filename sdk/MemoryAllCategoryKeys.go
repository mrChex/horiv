package sdk

import (
	"encoding/json"
	"fmt"
	"github.com/sashabaranov/go-openai"
)

const MemoryAllCategoryKeysFunctionName = "MemoryAllCategoryKeys"

func (s *SDK) DefinitionMemoryAllCategoryKeys() openai.FunctionDefinition {
	return openai.FunctionDefinition{
		Name:        MemoryAllCategoryKeysFunctionName,
		Description: "Отримати всі ключі в конкретній категорії",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"category": map[string]interface{}{
					"type": "string",
				},
			},
			"response": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "string",
				},
			},
		},
	}
}

func (s *SDK) ExecuteMemoryAllCategoryKeys(chatID int64, call openai.ToolCall) ([]openai.ChatCompletionMessage, error) {
	if call.Type != openai.ToolTypeFunction || call.Function.Name != MemoryAllCategoryKeysFunctionName {
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

func (s *SDK) MemoryAllCategoryKeys(chatID int64, category string) ([]string, error) {
	return s.store.MemoryAllCategoryKeys(chatID, category)
}
