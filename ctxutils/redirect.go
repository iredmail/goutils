package ctxutils

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

func addQueryMark(uri string) (newURI string) {
	if strings.Contains(uri, "?") {
		return uri + "&"
	} else {
		return uri + "?"
	}
}

func RedirectWithString(ctx *fiber.Ctx, uri, msg string) error {
	// TODO URI 是否需要添加 HomePath 前缀？
	return ctx.Redirect(addQueryMark(uri) + "msg=" + msg)
}

func RedirectWithError(ctx *fiber.Ctx, uri string, err error) error {
	if err != nil {
		return ctx.Redirect(addQueryMark(uri) + "msg=" + err.Error())
	}

	// TODO URI 是否需要添加 HomePath 前缀？
	return ctx.Redirect(uri)
}
