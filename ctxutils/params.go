package ctxutils

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/iredmail/goutils/emailutils"
)

func ParamDomain(ctx *fiber.Ctx) (domain string, err error) {
	domain = ctx.Params("domain")

	if !emailutils.IsDomain(domain) {
		return "", errors.New("INVALID_EMAIL_DOMAIN")
	}

	return domain, nil
}

func ParamEmail(ctx *fiber.Ctx) (addr string, err error) {
	addr = ctx.Params("email")

	if !emailutils.IsEmail(addr) {
		return "", errors.New("INVALID_EMAIL")
	}

	return addr, nil
}

// ParamPage 用于查询 URL parameters（`/xxx/:page`）的 `page` 参数的值。
// 如果没有指定则默认为 1。
func ParamPage(ctx *fiber.Ctx) (page int) {
	page, _ = ctx.ParamsInt("page", 1)
	if page < 1 {
		page = 1
	}

	return
}

func ParamInt64(ctx *fiber.Ctx, key string) (i int64) {
	s := ctx.Params(key, "0")

	i, _ = strconv.ParseInt(s, 10, 64)

	return
}
