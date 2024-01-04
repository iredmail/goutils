package response

import (
	"github.com/gofiber/fiber/v2"
)

type FuncTranslate func(ctx *fiber.Ctx, msg string, args ...any) string

type Response struct {
	ctx     *fiber.Ctx
	Suc     bool   `json:"_success"`
	Message string `json:"_msg"`
	Data    any    `json:"data"`
}

func New(ctx *fiber.Ctx) Response {
	return Response{ctx: ctx}
}

func (r Response) Msg(msg string, fn ...FuncTranslate) Response {
	if len(fn) > 0 {
		r.Message = fn[0](r.ctx, msg)
	} else {
		r.Message = msg
	}

	return r
}

func (r Response) Error(err error, fn ...FuncTranslate) Response {
	return r.Msg(err.Error(), fn...)
}

func (r Response) Map(data fiber.Map) Response {
	r.Data = data

	return r
}

func (r Response) Any(data any) Response {
	r.Data = data

	return r
}

func (r Response) Success() Response {
	r.Suc = true

	return r
}

func (r Response) JSON() error {
	return r.ctx.JSON(r)
}
