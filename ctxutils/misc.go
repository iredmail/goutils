package ctxutils

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func EmptyContent(ctx *fiber.Ctx) error {
	_, err := ctx.WriteString("")

	return err
}

func ReplyEmptyOK(ctx *fiber.Ctx) error {
	return ctx.Status(fiber.StatusOK).SendString("")
}

func RedirectWithString(ctx *fiber.Ctx, uri, msg string) error {
	// TODO URI 是否需要添加 HomePath 前缀？
	return ctx.Redirect(uri + "?msg=" + msg)
}

func RedirectWithError(ctx *fiber.Ctx, uri string, err error) error {
	if err != nil {
		return ctx.Redirect(uri + "?msg=" + err.Error())
	}

	// TODO URI 是否需要添加 HomePath 前缀？
	return ctx.Redirect(uri)
}

func CtxSSRError(ctx *fiber.Ctx, err error) error {
	m := fiber.Map{"err": err.Error()}

	return ctx.Render("ssr_error", m)
}

func SSRHXRedirect(ctx *fiber.Ctx, location, msg string) error {
	ctx.Set("HX-Redirect", location+"?msg="+msg)
	return ctx.SendString(msg)
}

func CtxSSRMsg(ctx *fiber.Ctx, msg string) error {
	return ctx.SendString(msg)
}

func CtxSetSSEHeader(ctx *fiber.Ctx) {
	ctx.Set("Content-Type", "text/event-stream")
	ctx.Set("Cache-Control", "no-cache")
	ctx.Set("Connection", "keep-alive")
	ctx.Set("Transfer-Encoding", "chunked")
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

func QueryInt64(ctx *fiber.Ctx, key string, defaultValue ...int64) (num int64) {
	dv := int64(0)

	if len(defaultValue) > 0 {
		dv = defaultValue[0]
	}

	i := ctx.Query(key, fmt.Sprintf("%d", dv))
	v, err := strconv.ParseInt(i, 10, 64)
	if err != nil {
		return 0
	}

	return v
}

func QueryBool(ctx *fiber.Ctx, key string) bool {
	query := ctx.Query(key, "false")
	parseBool, _ := strconv.ParseBool(query)

	return parseBool
}

func RouteURI(ctx *fiber.Ctx, name string, params fiber.Map) (uri string) {
	uri, err := ctx.GetRouteURL(name, params)
	if err != nil {
		uri = "/"
	}

	return uri
}
