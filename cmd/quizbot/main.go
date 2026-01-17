package main

import (
	"log/slog"
	"os"

	"github.com/letsssgooo/quizBot/internal/lib/slogcustom"
)

func main() {
	logger := setupLogger()
	slog.SetDefault(logger)

	slog.Debug("debug mode")

	slog.Info("starting events bot...")

	slog.Warn("warn mode", "error", "some error")

	slog.Error("error mode", "error", "some error")

	//flagToken := pflag.String("token", "", "token of client bot")
	//flagBotUsername := pflag.String("bot-username", "", "username of the client bot")
	//pflag.Parse()
	//
	//client := client.NewHTTPClient(*flagToken)
	//engine := engine.NewEngine()
	//
	//bot := bot.NewBot(client, engine, *flagBotUsername)
	//
	//err := bot.Run()
	//if err != nil {
	//	return
	//}

	// Storage
}

func setupLogger() *slog.Logger {
	logHandler := slogcustom.NewCustomHandler(os.Stdout, slog.LevelDebug)

	return slog.New(logHandler)
}
