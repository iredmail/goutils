package ctxutils

import (
	"fmt"
	"strconv"

	"spider/internal/cfg"
	"spider/internal/email"
	"spider/internal/respcode"

	"github.com/gofiber/fiber/v2"
)

func QueryInt(ctx *fiber.Ctx, key string, defaultValue int) (num int) {
	i := ctx.Query(key, fmt.Sprintf("%d", defaultValue))
	num, err := strconv.Atoi(i)
	if err != nil {
		return defaultValue
	}

	return num
}

// QueryPage 用于查询 URL query parameters（`/?page=x`）里 `page` 参数的值。
// 如果没有指定则默认为 1。
func QueryPage(ctx *fiber.Ctx) (page int) {
	return QueryInt(ctx, "page", 1)
}

// QueryLimit 用于查询 URL query parameters（`/?limit=x`）里 `limit` 参数的值。
// 如果没有指定则默认为 cfg.WebPageSize。
func QueryLimit(ctx *fiber.Ctx) (page int) {
	return QueryInt(ctx, "limit", cfg.WebPageSize)
}

func QueryDomain(ctx *fiber.Ctx) (domain string, err error) {
	domain = ctx.Query("domain")

	if !email.IsDomain(domain) {
		return "", respcode.ErrInvalidEmailDomain
	}

	return domain, nil
}

func QueryParticipant(ctx *fiber.Ctx) (addr string, err error) {
	addr = ctx.Query("participant")

	if !email.IsEmail(addr) {
		return "", respcode.ErrInvalidEmail
	}

	return addr, nil
}
