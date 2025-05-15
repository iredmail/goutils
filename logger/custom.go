package logger

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"
)

type CustomHandler struct {
	w          io.Writer
	level      slog.Level
	timeFormat string
}

func (c *CustomHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= c.level
}

func (c *CustomHandler) Handle(_ context.Context, r slog.Record) error {
	r.Attrs(func(attr slog.Attr) bool {
		fmt.Println(attr)

		return true
	})

	var b bytes.Buffer
	if c.timeFormat != "" {
		b.WriteString(time.Now().Format(c.timeFormat) + " ")
	} else {
		b.WriteString(time.Now().String() + " ")
	}

	b.WriteString(r.Level.String() + " ")
	b.WriteString(r.Message + "\n")

	_, err := c.w.Write(b.Bytes())

	return err
}

func (c *CustomHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	return c
}

func (c *CustomHandler) WithGroup(_ string) slog.Handler {
	return c
}
