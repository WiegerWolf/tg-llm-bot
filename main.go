package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	tele "gopkg.in/telebot.v4"
	"gopkg.in/telebot.v4/middleware"

	"google.golang.org/genai"
)

var (
	tgBotToken   string
	tgWhitelist  string
	geminiApiKey string
	geminiModel  string
)

type App struct {
	bot       *tele.Bot
	llm       *genai.Client
	model     string
	llmConfig *genai.GenerateContentConfig
}

func (app *App) onText(c tele.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()
	resp, err := app.llm.Models.GenerateContent(ctx, app.model, genai.Text(c.Text()), app.llmConfig)
	if err != nil {
		return err
	}
	return c.Send(resp.Text())
}

func main() {
	if len(tgBotToken) == 0 {
		log.Println("WARN: No TG_BOT_TOKEN was baked into the binary, using environment variable instead")
		tgBotToken = os.Getenv("TG_BOT_TOKEN")
		if len(tgBotToken) == 0 {
			log.Fatalln("ERROR: TG_BOT_TOKEN environment variable is not set. Please set it and try again.")
		}
	}

	bot, err := tele.NewBot(tele.Settings{
		Token: tgBotToken,
	})
	if err != nil {
		log.Fatalln("ERROR: Failed to create bot:", err)
	}

	if len(geminiApiKey) == 0 {
		log.Println("WARN: No GEMINI_API_KEY was baked into the binary, using environment variable instead")
		geminiApiKey = os.Getenv("GEMINI_API_KEY")
		if len(geminiApiKey) == 0 {
			log.Fatalln("ERROR: GEMINI_API_KEY environment variable is not set. Please set it and try again.")
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	llm, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: geminiApiKey,
	})
	if err != nil {
		log.Fatalln("ERROR: Failed to create Gemini client:", err)
	}

	if len(geminiModel) == 0 {
		log.Println("WARN: No GEMINI_MODEL was baked into the binary, using environment variable instead")
		geminiModel = os.Getenv("GEMINI_MODEL")
		if len(geminiModel) == 0 {
			log.Println("WARN: GEMINI_MODEL environment variable is not set.")
			geminiModel = "gemini-flash-lite-latest"
			log.Println("INFO: Using default model:", geminiModel)
		}
	}

	app := &App{
		bot:   bot,
		llm:   llm,
		model: geminiModel,
		llmConfig: &genai.GenerateContentConfig{
			ThinkingConfig: &genai.ThinkingConfig{
				ThinkingBudget: genai.Ptr[int32](24576),
			},
			Tools: []*genai.Tool{
				{URLContext: &genai.URLContext{}},
				{CodeExecution: &genai.ToolCodeExecution{}},
				{GoogleSearch: &genai.GoogleSearch{}},
			},
		},
	}

	if len(tgWhitelist) == 0 {
		log.Println("WARN: No TG_WHITELIST was baked into the binary, using environment variable instead")
		tgWhitelist = os.Getenv("TG_WHITELIST")
	}
	var whitelistIDs []int64
	if len(tgWhitelist) > 0 {
		parts := strings.Split(tgWhitelist, ",")
		whitelistIDs = make([]int64, 0, len(parts))
		for _, part := range parts {
			id, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64)
			if err != nil {
				log.Printf("WARN: Invalid whitelist ID '%s', skipping", part)
				continue
			}
			whitelistIDs = append(whitelistIDs, id)
		}
		log.Printf("INFO: Whitelist enabled with %d IDs", len(whitelistIDs))
		app.bot.Use(middleware.Whitelist(whitelistIDs...))
	} else {
		log.Println("WARN: No whitelist IDs found. Skipping whitelist middleware, meaning that ALL PUBLIC CHATS are allowed.")
	}

	app.bot.Handle(tele.OnText, app.onText)
	app.bot.Start()
}
