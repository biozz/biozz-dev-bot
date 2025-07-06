# biozz.dev bot

A personal Telegram bot on top of PocketBase with a growing number of features, starting with simple GPT chats interoperable with LibreChat.

The bot is intended to live in a Telegram Super Group with multiple topics. One for GPT chats, one for various alerts from apps, and more topics if necessary.

## Features

- **AI Chat**: OpenAI/OpenRouter integration with persistent LibreChat storage
- **Home Assistant**: Control smart home devices via interactive keyboards
- **Access Control**: Whitelist-based user access
- **PocketBase Backend**: Built-in database and API management

## Setup

1. **Environment Variables**:

```bash
# Telegram Bot Configuration
TELEGRAM_BOT_TOKEN=your-bot-token-here
SUPERGROUP_ID=-1001234567890
SUPERUSER_ID=123456789

# GPT Configuration
GPT_THREAD_ID=your-thread-id
OPENAI_API_KEY=sk-proj-your-openai-api-key-here
OPENROUTER_API_KEY=your-openrouter-api-key
CONVO_MODEL=gpt-4o
SUMMARY_MODEL=gpt-4o-mini

# LibreChat Configuration
LIBRECHAT_MONGO_URI=mongodb://localhost:27017/LibreChat
LIBRECHAT_USER_ID=your-librechat-user-id
LIBRECHAT_TAG=your-bot-tag

# Home Assistant Configuration
HOME_ASSISTANT_URL=http://your-home-assistant-url:8123
HOME_ASSISTANT_TOKEN=your_home_assistant_long_lived_token
```

2. **Run**:

```bash
go run main.go serve
```

## Commands

- `/gpt` - Start a new GPT conversation
- `/ha` - Show Home Assistant devices with interactive control panel
