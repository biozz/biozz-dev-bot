# biozz.dev bot

A personal Telegram bot on top of PocketBase with a growing number of features, starting with simple GPT chats interoperable with LibreChat.

The bot is intended to live in a Telegram Super Group with multiple topics. One for GPT chats, one for various alerts from apps, and more topics if necessary.

## Features

- **GPT Integration**: Chat with OpenAI GPT models via `/gpt` command
- **LibreChat Integration**: Persistent conversation storage through MongoDB
- **Access Control**: Whitelist-based user access (superuser only)
- **PocketBase Backend**: Built-in database and API management
- **Multiple AI Providers**: Support for OpenAI and OpenRouter APIs
- **Conversation Management**: Create, store, and manage chat sessions
- **State Management**: Context-aware message handling

## Setup

1. **Environment Variables**:
   ```bash
   TELEGRAM_BOT_TOKEN=your_telegram_bot_token
   SUPERUSER_ID=your_telegram_user_id
   SUPERGROUP_ID=your_telegram_group_id
   GPT_THREAD_ID=your_gpt_thread_id
   LIBRECHAT_MONGO_URI=mongodb://localhost:27017/LibreChat
   LIBRECHAT_USER_ID=your_librechat_user_id
   LIBRECHAT_TAG=your_tag
   OPENAI_API_KEY=your_openai_api_key
   OPENROUTER_API_KEY=your_openrouter_api_key
   CONVO_MODEL=gpt-4o
   SUMMARY_MODEL=gpt-4
   ```

2. **Run**:
   ```bash
   go run main.go serve
   ```

## Commands

- `/gpt` - Start a new GPT conversation
