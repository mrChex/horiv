package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-telegram/bot"
	tgModels "github.com/go-telegram/bot/models"
	"github.com/sashabaranov/go-openai"
	"log"
)

const GenerateImageFunctionName = "GenerateImage"
const GenerateImageSize = openai.CreateImageSize256x256

func (s *SDK) DefinitionGenerateImage() openai.FunctionDefinition {
	return openai.FunctionDefinition{
		Name:        GenerateImageFunctionName,
		Description: "Генерація зображення на основі тексту яке буде одразу відправлено користувачу",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"prompt": map[string]interface{}{
					"type":        "string",
					"description": "Текст, на основі якого буде згенеровано зображення",
				},
			},
			"response": map[string]interface{}{
				"type": "string",
				"enum": []string{
					"OK",
				},
			},
			"required": []string{
				"prompt",
			},
		},
	}
}

func (s *SDK) GenerateImage(ctx context.Context, prompt string) (string, error) {
	log.Println("Generating image for prompt:", prompt)
	reqBase64 := openai.ImageRequest{
		Prompt:         prompt,
		Size:           GenerateImageSize,
		ResponseFormat: openai.CreateImageResponseFormatURL,
		N:              1,
	}

	imageResponse, err := s.gpt.CreateImage(ctx, reqBase64)
	if err != nil {
		fmt.Printf("Image creation error: %v\n", err)
		return "", err
	}

	return imageResponse.Data[0].URL, nil
}

func (s *SDK) ExecuteGenerateImage(chatID int64, call openai.ToolCall) ([]openai.ChatCompletionMessage, error) {
	if call.Type != openai.ToolTypeFunction || call.Function.Name != GenerateImageFunctionName {
		return nil, fmt.Errorf("invalid call type")
	}
	ctx := context.Background()
	_, _ = s.bot.SendChatAction(ctx, &bot.SendChatActionParams{
		ChatID: chatID,
		Action: tgModels.ChatActionUploadPhoto,
	})

	var args struct {
		Prompt string `json:"prompt"`
	}
	err := json.Unmarshal([]byte(call.Function.Arguments), &args)
	if err != nil {
		return nil, fmt.Errorf("unmarshal arguments: %w", err)
	}

	imgURL, err := s.GenerateImage(ctx, args.Prompt)
	if err != nil {
		return nil, fmt.Errorf("generate image: %w", err)
	}

	log.Println("imgURL: ", imgURL)

	_, _ = s.bot.SendPhoto(ctx, &bot.SendPhotoParams{
		ChatID: chatID,
		Photo:  &tgModels.InputFileString{Data: imgURL},
	})

	rsp := []openai.ChatCompletionMessage{
		{
			Role:       openai.ChatMessageRoleTool,
			ToolCallID: call.ID,
			Content:    "OK",
		},
	}

	return rsp, nil
}
