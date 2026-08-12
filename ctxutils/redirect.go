package ctxutils

import (
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func addQueryMark(uri string) (newURI string) {
	if i := strings.IndexByte(uri, '#'); i >= 0 {
		uri = uri[:i]
	}

	if strings.Contains(uri, "?") {
		return uri + "&"
	}

	return uri + "?"
}

func RedirectWithString(ctx *fiber.Ctx, uri, msg string) error {
	return ctx.Redirect(addQueryMark(uri) + "msg=" + url.QueryEscape(msg))
}

func RedirectWithError(ctx *fiber.Ctx, uri string, err error) error {
	if err != nil {
		return ctx.Redirect(addQueryMark(uri) + "msg=" + url.QueryEscape(err.Error()))
	}

	return ctx.Redirect(uri)
}
