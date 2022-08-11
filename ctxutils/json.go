package ctxutils

import (
	"github.com/gofiber/fiber/v2"
)

// JSONError 返回表示 http 请求错误的 JSON 数据：
// {"_success": false, "_msg": "<具体错误原因>"}
func JSONError(ctx *fiber.Ctx, err error) error {
	return ctx.JSON(fiber.Map{
		"_success": false,
		"_msg":     err.Error(),
	})
}

func JSONErrorString(ctx *fiber.Ctx, ecode string) error {
	return ctx.JSON(fiber.Map{
		"_success": false,
		"_msg":     ecode,
	})
}

// JSONSuccess 返回表示 http 请求成功的 JSON 数据：
// {"_success": true, fiber.Map{...}}
func JSONSuccess(ctx *fiber.Ctx) error {
	m := fiber.Map{"_success": true, "_msg": ""}

	return ctx.JSON(m)
}

func JSONSuccessMap(ctx *fiber.Ctx, dataMap fiber.Map) error {
	dataMap["_success"] = true

	return ctx.JSON(dataMap)
}
