package ctxutils

import (
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResponseJSON(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(ctx fiber.Ctx) error {
		return New(ctx).
			Success().
			Msg("OK").
			Map(fiber.Map{"foo": "bar"}).
			JSON()
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.JSONEq(t, `{"_success": true, "_msg": "OK", "data": {"foo": "bar"}}`, string(body))
}

func TestResponseError(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(ctx fiber.Ctx) error {
		return New(ctx).Error(errors.New("boom")).JSON()
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.JSONEq(t, `{"_success": false, "_msg": "boom", "data": null}`, string(body))
}

func TestResponseMsgWithTranslator(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(ctx fiber.Ctx) error {
		translate := func(ctx fiber.Ctx, msg string, args ...any) string {
			return msg + "!"
		}

		return New(ctx).Success().Msg("hello", translate).JSON()
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.JSONEq(t, `{"_success": true, "_msg": "hello!", "data": null}`, string(body))
}

func TestResponseAny(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(ctx fiber.Ctx) error {
		return New(ctx).Success().Msg("OK").Any([]string{"a", "b"}).JSON()
	})

	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, "/", nil))
	require.NoError(t, err)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.JSONEq(t, `{"_success": true, "_msg": "OK", "data": ["a", "b"]}`, string(body))
}
