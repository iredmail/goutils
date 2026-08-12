package ctxutils

import (
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func SSRError(ctx *fiber.Ctx, err error) error {
	return ctx.Render("ssr_error", fiber.Map{"err": err.Error()})
}

func SSRHXRedirect(ctx *fiber.Ctx, location, msg string) error {
	sep := "?"
	if strings.Contains(location, sep) {
		sep = "&"
	}

	ctx.Set("HX-Redirect", location+sep+"msg="+url.QueryEscape(msg))

	return ctx.SendString(msg)
}

func SSRMsg(ctx *fiber.Ctx, msg string) error {
	return ctx.SendString(msg)
}
