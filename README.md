# Telegram Bot to OpenAI API Bridge

This is a Horiv natural language assistant Telegram bot that has long-term memory ability and its own SDK for custom functionalities.

You can provide custom functions to it for your needs.

## Configuration

1. Create a new bot on Telegram and get the token.
2. Create a new OpenAI account and get the API key.
3. Ensure you have Go installed on your system.
4. Clone this repository and navigate into the project directory.
5. Create a new file named `.env` in the root directory of the project and add the following content:
```
TELEGRAM_TOKEN=<YOUR_TELEGRAM_BOT_TOKEN>
OPENAI_TOKEN=<YOUR_OPENAI_API_KEY>
```

## Running the bot

```go
go run main.go
```

The bot starts listening for Telegram updates and interacts with the OpenAI API to process and respond to messages.
It logs its activities and errors for monitoring and debugging purposes.
