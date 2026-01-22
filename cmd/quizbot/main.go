package main

import (
	"log/slog"
	"os"

	// "github.com/letsssgooo/quizBot/internal/client"
	// "github.com/letsssgooo/quizBot/internal/events/engine"
	// "github.com/letsssgooo/quizBot/internal/events/fetcher"
	// "github.com/letsssgooo/quizBot/internal/events/sender"
	// "github.com/letsssgooo/quizBot/internal/bot"
	"github.com/letsssgooo/quizBot/internal/lib/slogcustom"
)

// func getTokenAndBotName() (string, string, error) {
// 	ptrToken := flag.String("token", "", "")
// 	ptrBotName := flag.String("bot-username", "", "")
// 	flag.Parse()

// 	token := *ptrToken

// 	botUsername := *ptrBotName
// 	if token == "" {
// 		return "", "", fmt.Errorf("empty tg bot token")
// 	} else if botUsername == "" {
// 		return "", "", fmt.Errorf("empty tg bot name")
// 	}

// 	return token, botUsername, nil
// }

func main() {
	logger := setupLogger()
	slog.SetDefault(logger)

	slog.Debug("debug mode")

	slog.Info("starting events bot...")

	slog.Warn("warn mode", "error", "some error")

	slog.Error("error mode", "error", "some error")

	// token, botUsername, err := getTokenAndBotName()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// client := client.NewHTTPClient(token)
	// quizEngine := engine.NewEngine()
	// telegramFetcher := fetcher.NewTelegramFetcher(client)
	// telegramSender := sender.NewTelegramSender(client)
	// bot := bot.NewBot(client, telegramFetcher, telegramSender, quizEngine, botUsername)

	// err = bot.Run()
	// if err != nil {
	// 	log.Fatal(err)
	// }

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
