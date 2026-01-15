package main

import (
	"log/slog"
	"os"

	"github.com/letsssgooo/quizBot/internal/bot"
	"github.com/letsssgooo/quizBot/internal/client"
	"github.com/letsssgooo/quizBot/internal/events/engine"
	"github.com/spf13/pflag"
)

func main() {
	log := setupLogger()
	slog.SetDefault(log)
	slog.Info("starting events bot...")

	flagToken := pflag.String("token", "", "token of client bot")
	flagBotUsername := pflag.String("bot-username", "", "username of the client bot")
	pflag.Parse()

	client := client.NewHTTPClient(*flagToken)
	engine := engine.NewEngine()

	bot := bot.NewBot(client, engine, *flagBotUsername)

	err := bot.Run()
	if err != nil {
		return
	}

	// Storage
}

func setupLogger() *slog.Logger {
	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	return log
}
