package ctxutils

import (
	"github.com/gofiber/fiber/v2"
)

func EmptyContent(ctx *fiber.Ctx) error {
	_, err := ctx.WriteString("")

	return err
}

func ReplyEmptyOK(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).SendString("")
}

func RouteURI(ctx *fiber.Ctx, name string, params fiber.Map) (uri string) {
	uri, err := ctx.GetRouteURL(name, params)
	if err != nil {
		uri = "/"
	}

	return uri
}
