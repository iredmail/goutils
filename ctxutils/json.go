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

func JSONErrorInvalidParamValue(ctx *fiber.Ctx, param string, err error) error {
	return ctx.JSON(fiber.Map{
		"_success":   false,
		"_parameter": param,
		"_msg":       err.Error(),
	})
}

func JSONError500(ctx *fiber.Ctx) error {
	return ctx.JSON(fiber.Map{
		"_success": false,
		"_msg":     "INTERNAL_SERVER_ERROR",
	})
}

func JSONErrorString(ctx *fiber.Ctx, ecode string) error {
	return ctx.JSON(fiber.Map{
		"_success": false,
		"_msg":     ecode,
	})
}

// JSONSuccess 返回表示 http 请求成功的 JSON 数据：
// {"_success": true, "_msg": ""}
func JSONSuccess(ctx *fiber.Ctx) error {
	m := fiber.Map{
		"_success": true,
		"_msg":     "",
	}

	return ctx.JSON(m)
}

// JSONSuccessMsg 返回表示 http 请求成功的 JSON 数据。
//
//	{
//		"_success": true,
//		"_msg": msg,
//	}
func JSONSuccessMsg(ctx *fiber.Ctx, msg string) error {
	m := fiber.Map{
		"_success": true,
		"_msg":     msg,
	}

	return ctx.JSON(m)
}

// JSONSuccessMap 返回表示 http 请求成功的 JSON 数据，m 的内容也将包含在 JSON 中。
//
//	{
//		"_success": true,
//		"_msg": "",
//		...
//	}
func JSONSuccessMap(ctx *fiber.Ctx, m fiber.Map) error {
	m["_success"] = true

	return ctx.JSON(m)
}
