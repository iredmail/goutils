package ctxutils

import (
	"github.com/gofiber/fiber/v3"
)

// FuncTranslate 把 msg 按当前请求的语言翻译为最终展示的文本。
type FuncTranslate func(ctx fiber.Ctx, msg string, args ...any) string

// Response 用于构造统一的 JSON 响应。
//
// Response 持有当前请求的 fiber.Ctx，因此只能在 handler 内同步使用：
// 不要保存到全局变量、缓存，也不要交给 handler 返回后仍在运行的 goroutine，
// 因为 fiber.Ctx 会在请求结束后被回收并复用。
type Response struct {
	ctx     fiber.Ctx
	Suc     bool   `json:"_success"`
	Message string `json:"_msg"`
	Data    any    `json:"data"`
}

// New 基于当前请求的 ctx 创建 Response。
func New(ctx fiber.Ctx) Response {
	return Response{ctx: ctx}
}

// Msg 设置响应消息，可传入翻译函数。
func (r Response) Msg(msg string, fn ...FuncTranslate) Response {
	if len(fn) > 0 {
		r.Message = fn[0](r.ctx, msg)
	} else {
		r.Message = msg
	}

	return r
}

// Error 把 err 作为响应消息。
func (r Response) Error(err error, fn ...FuncTranslate) Response {
	return r.Msg(err.Error(), fn...)
}

// Map 设置响应数据。
func (r Response) Map(data fiber.Map) Response {
	r.Data = data

	return r
}

// Any 设置任意类型的响应数据。
func (r Response) Any(data any) Response {
	r.Data = data

	return r
}

// Success 标记响应为成功。
func (r Response) Success() Response {
	r.Suc = true

	return r
}

// JSON 把 Response 作为 JSON 写入当前请求的响应。
func (r Response) JSON() error {
	return r.ctx.JSON(r)
}
