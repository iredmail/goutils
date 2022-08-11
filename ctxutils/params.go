package ctxutils

import (
	"errors"

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
