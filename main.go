package main

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/go-telegram/bot"
	tgModels "github.com/go-telegram/bot/models"
	"github.com/sashabaranov/go-openai"

	chatS "github.com/mrChex/gustav/chat"
	"github.com/mrChex/gustav/models"
	sdkS "github.com/mrChex/gustav/sdk"
	"github.com/mrChex/gustav/store"
)

type Config struct {
	TelegramToken string
	OpenAIToken   string
	StorePath     string
}

func NewConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error while loading .env file: ", err)
	}

	cfg := Config{
		TelegramToken: os.Getenv("TELEGRAM_TOKEN"),
		OpenAIToken:   os.Getenv("OPENAI_TOKEN"),
		StorePath:     os.Getenv("STORE_PATH"),
	}

	if cfg.TelegramToken == "" {
		log.Fatal("env TELEGRAM_TOKEN is not set")
	}
	if cfg.OpenAIToken == "" {
		log.Fatal("env OPENAI_TOKEN is not set")
	}
	if cfg.StorePath == "" {
		cfg.StorePath = "store.db"
	}

	return cfg
}

var repo *store.Store // global
var chat *chatS.Chat  // global
var sdk *sdkS.SDK     // global

func main() {
	cfg := NewConfig()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// store
	repo = store.NewStore(cfg.StorePath)

	// telegram
	b, err := bot.New(
		cfg.TelegramToken,
		bot.WithDefaultHandler(handler),
	)
	if err != nil {
		panic(err)
	}

	// gpt
	gpt := openai.NewClient(cfg.OpenAIToken)
	chat = chatS.NewChat(gpt, repo)

	// sdk
	sdk = sdkS.NewSDK(gpt, chat, b, repo)

	log.Println("Bot started and listening telegram updates...")
	b.Start(ctx)
}

func handler(ctx context.Context, b *bot.Bot, update *tgModels.Update) {
	_, _ = b.SendChatAction(ctx, &bot.SendChatActionParams{
		ChatID: update.Message.Chat.ID,
		Action: tgModels.ChatActionTyping,
	})

	userMessage := models.Message{
		Model: openai.GPT4o,
		Completion: openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: update.Message.Text,
		},
		CreatedAt: time.Now(),
	}

	handleUserMessage(ctx, b, update, userMessage)
}

func handleUserMessage(ctx context.Context, b *bot.Bot, update *tgModels.Update, userMessage models.Message) {
	sendError := func(act string, err error) {
		_, _ = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("Сталась помилка виконання %s: %s", act, err.Error()),
		})
	}

	err := repo.HistoryPush(update.Message.Chat.ID, userMessage)
	if err != nil {
		sendError("HistoryPush", err)
		return
	}

	respMsg, err := sendGPTRequest(ctx, update.Message.Chat.ID)
	if err != nil {
		sendError("GPTRequest", err)
		return
	}

	text := respMsg.Completion.Content
	//text = escapeTelegramMarkdown(text)

	log.Println("\t\t\t|> ", strings.Join(strings.Split(text, "\n"), "\n\t\t\t   "))

	_, err = b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:    update.Message.Chat.ID,
		Text:      text,
		ParseMode: tgModels.ParseModeHTML,
	})
	if err != nil {
		log.Println("Error sending Telegram message:", err)
		_, err = b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "TG SENDING ERROR. Plain text:\n\n" + text,
		})
	}
}

func escapeTelegramMarkdown(str string) string {
	for _, specialChar := range []string{"\\", "_", "*", "[", "]", "(", ")", "~", "`", ">", "<", "&", "#", "+", "-", "=", "|", "{", "}", ".", "!"} {
		str = strings.ReplaceAll(str, specialChar, "\\"+specialChar)
	}
	return str
}

func sendGPTRequest(ctx context.Context, chatID int64) (models.Message, error) {

	req, err := chat.MakeChatCompletionRequest(chatID)
	if err != nil {
		return models.Message{}, fmt.Errorf("MakeChatCompletionRequest: %w", err)
	}

	req.Tools = sdk.GetTools()

	log.Println("Waiting for completion response...")
	resp, err := chat.CreateChatCompletion(ctx, chatID, req)
	if err != nil {
		return models.Message{}, fmt.Errorf("CreateChatCompletion: %w", err)
	}

	respMsg := models.Message{
		Model:      req.Model,
		Completion: resp.Choices[0].Message,
		CreatedAt:  time.Now(),
	}

	// when error happen in function call - we should not save those request to prevent
	// deadlock. Saving request alongside with response or nothing
	if resp.Choices[0].Message.ToolCalls != nil {
		callResponseMessages, err := sdk.HandleToolCall(ctx, chatID, respMsg.Completion)
		if err != nil {
			return models.Message{}, fmt.Errorf("HandleToolCall: %w", err)
		}

		if err := repo.HistoryPush(chatID, respMsg); err != nil {
			return models.Message{}, fmt.Errorf("HistoryPush function call: %w", err)
		}
		for _, callResponseMessage := range callResponseMessages {
			err := repo.HistoryPush(chatID, callResponseMessage)
			if err != nil {
				return models.Message{}, fmt.Errorf("HistoryPush functon response: %w", err)
			}
		}

		return sendGPTRequest(ctx, chatID)
	} else {
		if err := repo.HistoryPush(chatID, respMsg); err != nil {
			return models.Message{}, fmt.Errorf("HistoryPush function call: %w", err)
		}
	}

	return respMsg, nil
}
