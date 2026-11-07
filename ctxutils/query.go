package ctxutils

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/iredmail/goutils/emailutils"
)

const (
	defaultPageSizeLimit = 50
)

// QueryString 获取指定的 query 的值，并移除首尾的空白字符。
// 如果未指定参数名，则返回默认值。如果没有默认值则值为空字符串。
func QueryString(ctx fiber.Ctx, name string, defaultValue ...string) (isEmpty bool, value string) {
	if len(defaultValue) > 0 {
		value = defaultValue[0]
	}

	if len(name) == 0 {
		isEmpty = len(value) == 0

		return
	}

	value = ctx.Query(name, value)
	value = strings.TrimSpace(value)

	isEmpty = len(value) == 0

	return
}

// QueryPage 用于查询 URL query parameters（`/?page=x`）里 `page` 参数的值。
// 如果没有指定或小于 1 则设置为 1。
func QueryPage(ctx fiber.Ctx) (page int) {
	page = max(fiber.Query[int](ctx, "page", 1), 1)

	return
}

// QueryLimit 用于查询 URL query parameters（`/?limit=x`）里 `limit` 参数的值。
// 如果没有指定则默认为 cfg.WebPageSize。
func QueryLimit(ctx fiber.Ctx, defaultValue ...int) (limit int) {
	limit = fiber.Query[int](ctx, "limit", 0)
	if limit <= 0 {
		if len(defaultValue) == 1 {
			limit = defaultValue[0]
		} else {
			limit = defaultPageSizeLimit
		}
	} else if limit > 500 {
		limit = defaultPageSizeLimit
	}

	return
}

func QueryDomain(ctx fiber.Ctx) (domain string, err error) {
	domain = strings.ToLower(strings.TrimSpace(ctx.Query("domain")))

	if !emailutils.IsDomain(domain) {
		err = errors.New("INVALID_EMAIL_DOMAIN")
	}

	return
}

func QueryEmail(ctx fiber.Ctx) (addr string, err error) {
	addr = ctx.Params("email")

	if !emailutils.IsEmail(addr) {
		return "", errors.New("INVALID_EMAIL")
	}

	return strings.ToLower(addr), nil
}

func QueryParticipant(ctx fiber.Ctx) (addr string, err error) {
	addr = ctx.Query("participant")

	if !emailutils.IsEmail(addr) {
		return "", errors.New("INVALID_EMAIL")
	}

	return strings.TrimSpace(addr), nil
}
