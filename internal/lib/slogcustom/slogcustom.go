package slogcustom

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"

	"github.com/fatih/color"
)

type CustomHandler struct {
	l     *log.Logger
	level slog.Level
}

func NewCustomHandler(out io.Writer, level slog.Level) *CustomHandler {
	return &CustomHandler{
		l:     log.New(out, "", 0),
		level: level,
	}
}

func (c *CustomHandler) Handle(_ context.Context, r slog.Record) error {
	level := r.Level.String() + ":"

	switch r.Level {
	case slog.LevelDebug:
		level = color.MagentaString(level)
	case slog.LevelInfo:
		level = color.HiBlueString(level)
	case slog.LevelWarn:
		level = color.YellowString(level)
	case slog.LevelError:
		level = color.RedString(level)
	}

	attrsStr := ""
	r.Attrs(func(a slog.Attr) bool {
		attrsStr += color.GreenString(a.Key) + "=" + fmt.Sprint(a.Value.Any()) + " "
		return true
	})

	c.l.Println(
		r.Time.Format("15:05:05.000"),
		level,
		r.Message,
		attrsStr,
	)
	return nil
}

func (c *CustomHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return c

}

func (c *CustomHandler) WithGroup(_ string) slog.Handler {
	return c
}

func (c *CustomHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= c.level
}
