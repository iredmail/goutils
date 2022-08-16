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

func ApiRespError(ctx *fiber.Ctx, err error) error {
	return ctx.JSON(fiber.Map{
		"_success": false,
		"_msg":     err.Error(),
	})
}

func ApiRespData(ctx *fiber.Ctx, key string, data map[string]any) error {
	m := fiber.Map{
		"_success": true,
		"_msg":     "",
		key:        data,
	}

	return ctx.JSON(m)
}

func RouteURI(ctx *fiber.Ctx, name string, params fiber.Map) (uri string) {
	uri, err := ctx.GetRouteURL(name, params)
	if err != nil {
		uri = "/"
	}

	return uri
}
