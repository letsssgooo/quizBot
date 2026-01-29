package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/letsssgooo/quizBot/internal/auth"
	"github.com/letsssgooo/quizBot/internal/bot"
	"github.com/letsssgooo/quizBot/internal/client"
	"github.com/letsssgooo/quizBot/internal/events/engine"
	"github.com/letsssgooo/quizBot/internal/events/fetcher"
	"github.com/letsssgooo/quizBot/internal/events/sender"
	"github.com/letsssgooo/quizBot/internal/lib/slogcustom"
	"github.com/letsssgooo/quizBot/internal/storage/postgres"
	"golang.org/x/sync/errgroup"
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

const timeoutGraceful = 5 * time.Second

func main() {
	logger := setupLogger()
	slog.SetDefault(logger)

	slog.Debug("Logger started with Debug level")
	slog.Debug("Parsing tokens...")

	rootCtx, cancelFunc := context.WithCancel(context.Background()) // graceful shutdown
	g, gCtx := errgroup.WithContext(rootCtx)

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
	telegramFetcher := fetcher.NewTelegramFetcher(gCtx, httpClient)
	telegramSender := sender.NewTelegramSender(httpClient)

	botStorage, err := postgres.NewStorage(gCtx, dsn)
	if err != nil {
		slog.Error("Cannot initialize database", "error", err)
		os.Exit(1)
	}

	telegramBot := bot.NewBot(
		httpClient,
		botAuth,
		telegramFetcher,
		telegramSender,
		quizEngine,
		botStorage,
		botUsername,
	)

	slog.Debug("Bot entities created")
	slog.Debug("Starting bot...")

	g.Go(func() error {
		return telegramBot.Run(gCtx)
	})

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)

	done := make(chan error, 1)
	go func() {
		done <- g.Wait()
	}()

	select {
	case sig := <-stopCh:
		slog.Info("Received signal, stopping bot...", "signal", sig.String())
		cancelFunc()

		select {
		case err = <-done:
			switch {
			case errors.Is(err, context.Canceled):
				slog.Info("Context cancelled, bot stopped", "error", err)
			default:
				slog.Error("Error while running bot", "error", err)
			}
		case <-time.After(timeoutGraceful):
			slog.Warn("Graceful timeout is over, bot stopped")
		}
	case err = <-done:
		if err != nil {
			switch {
			case errors.Is(err, context.Canceled):
				slog.Info("Context cancelled, bot stopped", "error", err)
			default:
				slog.Error("Error while running bot", "error", err)
			}
		} else {
			slog.Info("Bot stopped")
		}
	}
}

func setupLogger() *slog.Logger {
	logHandler := slogcustom.NewCustomHandler(os.Stdout, slog.LevelDebug)

	return slog.New(logHandler)
}
