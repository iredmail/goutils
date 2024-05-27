package ctxutils

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/iredmail/goutils/emailutils"
)

const (
	defaultPageSizeLimit = 50
)

// QueryString 获取指定的 query 的值，并移除首尾的空白字符。
// 如果未指定参数名，则返回默认值。如果没有默认值则值为空字符串。
func QueryString(ctx *fiber.Ctx, name string, defaultValue ...string) (isEmpty bool, value string) {
	if len(defaultValue) > 0 {
		value = defaultValue[0]
	}

	if len(name) == 0 {
		isEmpty = len(value) == 0

		return
	}

	value = ctx.Query(name, value)
	value, _ = url.QueryUnescape(value)
	value = strings.TrimSpace(value)

	isEmpty = len(value) == 0

	return
}

func QueryInt(ctx *fiber.Ctx, key string, defaultValue int) (num int) {
	i := ctx.Query(key, fmt.Sprintf("%d", defaultValue))
	num, err := strconv.Atoi(i)
	if err != nil {
		return defaultValue
	}

	return num
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

// QueryPage 用于查询 URL query parameters（`/?page=x`）里 `page` 参数的值。
// 如果没有指定或小于 1 则设置为 1。
func QueryPage(ctx *fiber.Ctx) (page int) {
	page = QueryInt(ctx, "page", 1)
	if page < 1 {
		page = 1
	}

	return
}

// QueryLimit 用于查询 URL query parameters（`/?limit=x`）里 `limit` 参数的值。
// 如果没有指定则默认为 cfg.WebPageSize。
func QueryLimit(ctx *fiber.Ctx, defaultValue ...int) (limit int) {
	limit = QueryInt(ctx, "limit", 0)
	if limit <= 0 {
		if len(defaultValue) == 1 {
			limit = defaultValue[0]
		} else {
			limit = defaultPageSizeLimit
		}
	}

	return
}

func QueryDomain(ctx *fiber.Ctx) (domain string, err error) {
	domain = ctx.Query("domain")

	if !emailutils.IsDomain(domain) {
		return "", errors.New("INVALID_EMAIL_DOMAIN")
	}

	return strings.ToLower(domain), nil
}

func QueryEmail(ctx *fiber.Ctx) (addr string, err error) {
	addr = ctx.Params("email")

	if !emailutils.IsEmail(addr) {
		return "", errors.New("INVALID_EMAIL")
	}

	return strings.ToLower(addr), nil
}

func QueryParticipant(ctx *fiber.Ctx) (addr string, err error) {
	addr = ctx.Query("participant")

	if !emailutils.IsEmail(addr) {
		return "", errors.New("INVALID_EMAIL")
	}

	return addr, nil
}
