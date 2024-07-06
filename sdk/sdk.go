package sdk

import (
	"context"
	"fmt"
	"github.com/go-telegram/bot"
	tgModels "github.com/go-telegram/bot/models"
	"github.com/mrChex/horiv/chat"
	"github.com/mrChex/horiv/models"
	"github.com/mrChex/horiv/store"
	"github.com/sashabaranov/go-openai"
	"log"
	"time"
)

type SDK struct {
	chat  *chat.Chat
	gpt   *openai.Client
	bot   *bot.Bot
	store *store.Store
}

func NewSDK(gpt *openai.Client, chat *chat.Chat, b *bot.Bot, repo *store.Store) *SDK {
	return &SDK{
		chat:  chat,
		gpt:   gpt,
		bot:   b,
		store: repo,
	}
}

func (s *SDK) GetTools() []openai.Tool {
	enabledTools := []openai.FunctionDefinition{
		s.DefinitionGenerateImage(),
		s.DefinitionMemoryAllCategories(),
		s.DefinitionMemoryAllCategoryKeys(),
		s.DefinitionMemoryMapValues(),
		s.DefinitionMemoryUpsert(),
		s.DefinitionMemoryDelete(),
	}

	tools := make([]openai.Tool, len(enabledTools))
	for i, tool := range enabledTools {
		tools[i] = openai.Tool{
			Type:     openai.ToolTypeFunction,
			Function: &tool,
		}
	}
	return tools
}

func (s *SDK) HandleToolCall(_ context.Context, chatID int64, aiMsg openai.ChatCompletionMessage) ([]models.Message, error) {
	var response []openai.ChatCompletionMessage
	var err error

	for _, call := range aiMsg.ToolCalls {
		log.Printf("... SDK::%s::%s(%s)", call.ID, call.Function.Name, call.Function.Arguments)
		var res []openai.ChatCompletionMessage
		switch call.Function.Name {
		case "GenerateImage":
			res, err = s.ExecuteGenerateImage(chatID, call)
		case "MemoryUpsert":
			res, err = s.ExecuteMemoryUpsert(chatID, call)
		case "MemoryAllCategories":
			res, err = s.ExecuteMemoryAllCategories(chatID, call)
		case "MemoryAllCategoryKeys":
			res, err = s.ExecuteMemoryAllCategoryKeys(chatID, call)
		case "MemoryMapValues":
			res, err = s.ExecuteMemoryMapValues(chatID, call)
		case "MemoryDelete":
			res, err = s.ExecuteMemoryDelete(chatID, call)
		default:
			err = fmt.Errorf("unknown function %s", call.Function.Name)
		}
		if err != nil {
			log.Printf("         |> ERROR: %s", err.Error())
			return nil, err
		}

		response = append(response, res...)
	}

	messages := make([]models.Message, len(response))
	for i, msg := range response {
		messages[i] = models.Message{
			Model:      openai.GPT4o,
			Completion: msg,
			CreatedAt:  time.Now(),
		}
	}

	return messages, nil
}

func (s *SDK) sendError(botMsg *tgModels.Message, act string, err error) error {
	_, _ = s.bot.SendMessage(context.Background(), &bot.SendMessageParams{
		ChatID: botMsg.Chat.ID,
		Text:   fmt.Sprintf("Сталась помилка виконання %s: %s", act, err.Error()),
	})
	return err
}
