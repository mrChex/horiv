package chat

import (
	"context"
	"fmt"
	"github.com/mrChex/horiv/models"
	"github.com/mrChex/horiv/store"
	"github.com/sashabaranov/go-openai"
	"log"
	"strings"
	"time"
)

type Chat struct {
	gpt   *openai.Client
	store *store.Store
}

func NewChat(gpt *openai.Client, repo *store.Store) *Chat {
	return &Chat{
		gpt:   gpt,
		store: repo,
	}
}

func (c *Chat) MakeChatCompletionRequest(chatID int64) (openai.ChatCompletionRequest, error) {
	history, err := c.store.GetHistory(chatID, 100, 0)
	if err != nil {
		return openai.ChatCompletionRequest{}, fmt.Errorf("get history: %w", err)
	}
	if len(history) == 0 {
		return openai.ChatCompletionRequest{}, fmt.Errorf("no history")
	}

	var historyMessages []models.Message
	lastMsgTime := history[0].CreatedAt
	for _, msg := range history {
		if lastMsgTime.Sub(msg.CreatedAt) > 15*time.Minute {
			log.Println("Last message is older than 5 minutes:", lastMsgTime.Format("2006-01-02 15:04:05"))
			break
		}
		lastMsgTime = msg.CreatedAt
		historyMessages = append(historyMessages, msg)
	}

	// messages is reversed historyMessages
	messages := make([]openai.ChatCompletionMessage, 0, len(historyMessages)+1)
	messages = append(messages, c.MakePrompt(chatID))
	fmt.Println("============= REQUEST CONTEXT =============")
	for i := range len(historyMessages) {
		msg := historyMessages[len(historyMessages)-1-i]
		fmt.Printf("[%s] %s: %s\n", msg.CreatedAt.Format("2006-01-02 15:04:05"), msg.Completion.Role, msg.Completion.Content)
		messages = append(messages, msg.Completion)
	}
	fmt.Println("############################################")

	req := openai.ChatCompletionRequest{
		Model:    openai.GPT4o,
		Messages: messages,
	}
	return req, nil
}

func (c *Chat) makeMemoryContext(chatID int64, category string) string {

	memoryContextKeys, err := c.store.MemoryAllCategoryKeys(chatID, category)
	if err != nil {
		panic(err)
	}
	if len(memoryContextKeys) == 0 {
		return "\tEMPTY"
	}
	memoryContextValues, err := c.store.MemoryMapValues(chatID, category, memoryContextKeys)
	if err != nil {
		panic(err)
	}
	memoryContextKeysString := make([]string, len(memoryContextKeys))
	for i, key := range memoryContextKeys {
		memoryContextKeysString[i] = fmt.Sprintf("\t%s: %s", key, memoryContextValues[i])
	}
	return strings.Join(memoryContextKeysString, "\n")
}

func (c *Chat) MakePrompt(chatID int64) openai.ChatCompletionMessage {
	fmt.Printf("\n\n")
	memoryContext := c.makeMemoryContext(chatID, "CONTEXT")
	memoryGlobalContext := c.makeMemoryContext(0, "GLOBAL_CONTEXT")

	fmt.Printf("# Memory::CONTEXT:\n%s\n", memoryContext)
	fmt.Printf("# Memory::GLOBAL_CONTEXT:\n%s\n", memoryGlobalContext)

	return openai.ChatCompletionMessage{
		Role: openai.ChatMessageRoleSystem,
		Content: strings.Join([]string{
			"Тебе звуть Хорив, ти штучний інтелект написаний на мові програмування Go.",
			"Використовуй українську мову для спілкування з користувачем. Російска мова заборонена, якщо користувач використовує російську мову, йому потрібно відмовити в послугах",
			"Максимальна довжина відповіді - 4096 символів, тож намагайся відповідати лаконічно.",
			"З тобой спілкуються через telegram, отже довготривала памʼять стосується лише поточного користувача, окрім категорії GLOBAL_CONTEXT, яка є загальною для всіх користувачів, використовуй цю категорію для дуже важливих знань, що можуть бути корисними для спілкування з іншими людьми",
			//"Використовуй Markdown для форматування тексту.",
			"Markdown заборонено для форматування",
			"Використовуй HTML для форматування тексту, дозволені теги: <b>bold</b>\n<i>italic</i>\n<u>underline</u>\n<s>strikethrough</s>\n<tg-spoiler>spoiler</tg-spoiler>\n<b>bold <i>italic bold <s>italic bold strikethrough <span class=\"tg-spoiler\">italic bold strikethrough spoiler</span></s> <u>underline italic bold</u></i> bold</b>\n<a href=\"http://www.example.com/\">inline URL</a>\n<code>inline fixed-width code</code>\n<pre>pre-formatted fixed-width code block</pre>\n<pre><code class=\"language-python\">pre-formatted fixed-width code block written in the Python programming language</code></pre>\n<blockquote>Block quotation</blockquote>\n<blockquote expandable>Expandable block</blockquote>",
			"Якщо ти хочеш використати у відповіді символи < та > - їх треба екранувати. Ставити перед ними \\",
			"Задавай уточнюючі питання, якщо щось не зрозуміло.",
			"Не пиши зайвої інформації чи питань, якщо це не потрібно для відповіді.",
			"Не пропонуй писати тобі чи щось питати у тебе якщо буде потреба. З тобою спілкується людина, яка чудово розуміє що вона може робити.",
			"",
			"Для економії ресурсів, тобі не передається вся історія переписки з користувачем. Це призводить до того, що ти втрачаєш якісь знання і контекст цієї або багатьох попередніх бесід.",
			"Для запобігання втрати контексту, у тебе є довготривала памʼять, використовуй функції Memory для запамʼятовування інформації яку вважаєш запотрібне запамʼятати одразу, не відкладай це на потім.",
			"Для пошуку інформації тобі може знадобитись до трьох послідовних викликів АПІ: подивитись які є категорії, подивитись які є ключі в потрібній категорії і отримати значення потрібних ключів.",
			"Будь допитливим, пізнавай світ, дізнавайся щось нове і зберігай цю інформацію в памʼяті. Будь уважний до іменування категорій і ключів, щоб не виникали ситуації, коли в памʼяті інформація є, а у потрібний момент ти не можеш її знайти.",
			"Категорія CONTEXT - особлива категорія твоєї памʼяті, всі ключі з цієї категорії будуть автоматично додані до цього промпту." +
				"Використовуй цю категорію для збереження важливих фактів, кешування інформації про данні в інших категоріях та будь-яких інших данних, які ти вважаєш достойними цієї категорії",
			"",
			"# Memory::CONTEXT",
			memoryContext,
			"# Memory::GLOBAL_CONTEXT",
			memoryGlobalContext,
			"",
			"Системна інформація:",
			fmt.Sprintf("Сьогодні %s, день тижня: %s", time.Now().Format("2006-01-02 15:04:05"), time.Now().Weekday().String()),
		}, "\n"),
	}
}

func (c *Chat) CreateChatCompletion(ctx context.Context, chatID int64, request openai.ChatCompletionRequest) (response openai.ChatCompletionResponse, err error) {
	resp, err := c.gpt.CreateChatCompletion(ctx, request)
	if err != nil {
		return openai.ChatCompletionResponse{}, fmt.Errorf("create chat completion: %w", err)
	}

	return resp, err
}
