package ctxutils

import (
	"github.com/gofiber/fiber/v2"
)

func EmptyContent(ctx *fiber.Ctx) error {
	_, err := ctx.WriteString("")

	return err
}
