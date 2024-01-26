package ctxutils

import (
	"errors"
	"net/url"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/iredmail/goutils"
	"github.com/iredmail/goutils/emailutils"
	"github.com/iredmail/goutils/respcode"
)

func ParamDomain(ctx *fiber.Ctx) (domain string, err error) {
	domain = ctx.Params("domain")
	domain, err = url.QueryUnescape(domain)
	if err != nil {
		return "", respcode.ErrInvalidDomain
	}

	if !emailutils.IsDomain(domain) {
		return "", respcode.ErrInvalidDomain
	}

	return strings.ToLower(domain), nil
}

func ParamEmail(ctx *fiber.Ctx, name ...string) (addr string, err error) {
	param := "email"

	if len(name) > 0 {
		param = name[0]
	}

	addr = ctx.Params(param)
	addr, err = url.QueryUnescape(addr)
	if err != nil {
		return "", respcode.ErrInvalidEmailAddress
	}

	if !emailutils.IsEmail(addr) {
		return "", respcode.ErrInvalidEmailAddress
	}

	return emailutils.ToLowerWithExt(addr), nil
}

func ParamEmailWithoutExt(ctx *fiber.Ctx, name ...string) (addr string, err error) {
	param := "email"

	if len(name) > 0 {
		param = name[0]
	}

	addr = ctx.Params(param)
	addr, err = url.QueryUnescape(addr)
	if err != nil {
		return "", errors.New("INVALID_EMAIL")
	}

	if !emailutils.IsEmail(addr) {
		return "", errors.New("INVALID_EMAIL")
	}

	return emailutils.ToLowerWithoutExt(addr), nil
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

// ParamUUIDLicenseKey 用于查询 URL parameter（`/xxx/:license_key`）里 `license_key` 参数的值（必须是 UUID 格式）。
func ParamUUIDLicenseKey(ctx *fiber.Ctx) (key string, err error) {
	key = strings.ToUpper(ctx.Params("license_key"))

	if !goutils.IsUUIDLicenseKey(key) {
		err = respcode.ErrInvalidLicenseKey
	
		return
	}

	if err = uuid.Validate(key); err != nil {
		return "", respcode.ErrInvalidLicenseKey
	}

	return
}
