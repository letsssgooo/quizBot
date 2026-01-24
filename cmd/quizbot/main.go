package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/letsssgooo/quizBot/internal/auth"
	"github.com/letsssgooo/quizBot/internal/bot"
	"github.com/letsssgooo/quizBot/internal/client"
	"github.com/letsssgooo/quizBot/internal/events/engine"
	"github.com/letsssgooo/quizBot/internal/events/fetcher"
	"github.com/letsssgooo/quizBot/internal/events/sender"
	"github.com/letsssgooo/quizBot/internal/lib/slogcustom"
	"github.com/letsssgooo/quizBot/internal/storage/postgres"
)

const dsn string = ""

//func parseCLIArgs() (string, string, error) {
//	flagToken := pflag.String("token", "", "token of telegram bot")
//	flagBotName := pflag.String("bot-username", "", "telegram bot username")
//	pflag.Parse()
//
//	if *flagToken == "" {
//		return "", "", fmt.Errorf("empty tg bot token")
//	} else if *flagBotName == "" {
//		return "", "", fmt.Errorf("empty tg bot name")
//	}
//
//	return *flagToken, *flagBotName, nil
//}

func main() {
	logger := setupLogger()
	slog.SetDefault(logger)

	slog.Debug("Logger started with Debug level")
	slog.Debug("Parsing tokens...")

	ctx := context.Background() // для graceful shutdown

	token, botUsername := "", ""
	//if err != nil {
	//	slog.Error("Empty token or botUsername", "error", err)
	//	os.Exit(1)
	//}

	slog.Debug("Tokens parsed")
	slog.Debug("Creating bot entities...")

	httpClient := client.NewHTTPClient(token)
	botAuth := auth.NewBotAuth()
	quizEngine := engine.NewEngine()
	telegramFetcher := fetcher.NewTelegramFetcher(httpClient)
	telegramSender := sender.NewTelegramSender(httpClient)

	botStorage, err := postgres.NewStorage(ctx, dsn)
	if err != nil {
		slog.Error("Cannot initialize database", "error", err)
		os.Exit(1)
	}

	telegramBot := bot.NewBot(httpClient, botAuth, telegramFetcher, telegramSender, quizEngine, botStorage, botUsername)

	slog.Debug("Bot entities created")
	slog.Debug("Starting bot...")

	err = telegramBot.Run()
	if err != nil {
		slog.Error("Error while running bot", "error", err)
		os.Exit(1)
	}
}

func setupLogger() *slog.Logger {
	logHandler := slogcustom.NewCustomHandler(os.Stdout, slog.LevelDebug)

	return slog.New(logHandler)
}
