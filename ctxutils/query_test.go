package ctxutils

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryEmail(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    string
		wantErr bool
	}{
		{
			name: "规范化邮件地址",
			url:  "/?email=User@Example.COM",
			want: "user@example.com",
		},
		{
			name:    "缺少 email 参数",
			url:     "/",
			want:    "",
			wantErr: true,
		},
		{
			name:    "非法邮件地址",
			url:     "/?email=not-an-email",
			want:    "",
			wantErr: true,
		},
	}

	app := fiber.New()
	app.Get("/", func(ctx *fiber.Ctx) error {
		addr, err := QueryEmail(ctx)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).SendString(err.Error())
		}

		return ctx.SendString(addr)
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, tt.url, nil))
			require.NoError(t, err)

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			if tt.wantErr {
				assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
				assert.Equal(t, "INVALID_EMAIL", string(body))
			} else {
				assert.Equal(t, fiber.StatusOK, resp.StatusCode)
				assert.Equal(t, tt.want, string(body))
			}
		})
	}
}

// TestQueryEmailIgnoresRouteParam 确认 QueryEmail 读取的是 query 参数，
// 而不是同名 route parameter。
func TestQueryEmailIgnoresRouteParam(t *testing.T) {
	app := fiber.New()
	app.Get("/:email", func(ctx *fiber.Ctx) error {
		addr, err := QueryEmail(ctx)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).SendString(err.Error())
		}

		return ctx.SendString(addr)
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/user@example.com", nil))
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	assert.Equal(t, "INVALID_EMAIL", string(body))
}
