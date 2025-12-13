#!/bin/bash

go build -ldflags="\
    -X main.tgBotToken=$TG_BOT_TOKEN \
    -X main.tgWhitelist=$TG_WHITELIST \
    -X main.geminiApiKey=$GEMINI_API_KEY \
    -X main.geminiModel=$GEMINI_MODEL" \
    -o build/bot
