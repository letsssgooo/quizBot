//go:build !solution

package main

import (
	"log/slog"
	"os"

	"github.com/spf13/pflag"
	"gitlab.com/slon/shad-go/Exam-1-QuizBot/quizbot/internal/quiz"
	"gitlab.com/slon/shad-go/Exam-1-QuizBot/quizbot/internal/telegram"
)

func main() {
	log := setupLogger()
	slog.SetDefault(log)
	slog.Info("starting quiz bot...")

	flagToken := pflag.String("token", "", "token of telegram bot")
	flagBotUsername := pflag.String("bot-username", "", "username of the telegram bot")
	pflag.Parse()

	client := telegram.NewHTTPClient(*flagToken)
	engine := quiz.NewEngine()

	bot := telegram.NewBot(client, engine, *flagBotUsername)

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
