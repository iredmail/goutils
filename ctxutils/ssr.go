package ctxutils

import (
	"github.com/gofiber/fiber/v2"
)

func SSRError(ctx *fiber.Ctx, err error) error {
	m := fiber.Map{"err": err.Error()}

	return ctx.Render("ssr_error", m)
}

func SSRHXRedirect(ctx *fiber.Ctx, location, msg string) error {
	ctx.Set("HX-Redirect", location+"?msg="+msg)

	return ctx.SendString(msg)
}

func SSRMsg(ctx *fiber.Ctx, msg string) error {
	return ctx.SendString(msg)
}
