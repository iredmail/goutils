package ctxutils

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/iredmail/goutils/emailutils"
)

const (
	defaultPageSizeLimit = 100
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
	return QueryInt(ctx, "limit", defaultPageSizeLimit)
}

func QueryDomain(ctx *fiber.Ctx) (domain string, err error) {
	domain = ctx.Query("domain")

	if !emailutils.IsDomain(domain) {
		return "", errors.New("INVALID_EMAIL_DOMAIN")
	}

	return domain, nil
}

func QueryParticipant(ctx *fiber.Ctx) (addr string, err error) {
	addr = ctx.Query("participant")

	if !emailutils.IsEmail(addr) {
		return "", errors.New("INVALID_EMAIL")
	}

	return addr, nil
}
