package ctxutils

import (
	"spider/internal/email"
	"spider/internal/respcode"
	"spider/internal/types"

	"github.com/gofiber/fiber/v2"
)

func ParamHUID(ctx *fiber.Ctx) (types.HUID, bool) {
	s := ctx.Params("huid")

	return types.StringToHUID(s)
}

func ParamAUID(ctx *fiber.Ctx) (types.AUID, bool) {
	s := ctx.Params("auid")

	return types.StringToAUID(s)
}

func ParamDomain(ctx *fiber.Ctx) (domain string, err error) {
	domain = ctx.Params("domain")

	if !email.IsDomain(domain) {
		return "", respcode.ErrInvalidEmailDomain
	}

	return domain, nil
}

func ParamEmail(ctx *fiber.Ctx) (addr string, err error) {
	addr = ctx.Params("email")

	if !email.IsEmail(addr) {
		return "", respcode.ErrInvalidEmail
	}

	return addr, nil
}
