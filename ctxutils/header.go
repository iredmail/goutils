package ctxutils

import (
	"github.com/gofiber/fiber/v2"
)

func SetSSEHeader(ctx *fiber.Ctx) {
	ctx.Set("Content-Type", "text/event-stream")
	ctx.Set("Cache-Control", "no-cache")
}
