package request

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestParsePagination(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c *fiber.Ctx) error {
		page, pageSize := ParsePagination(c)
		return c.JSON(fiber.Map{
			"page":     page,
			"pageSize": pageSize,
		})
	})

	t.Run("defaults when no query parameters provided", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("valid page and pageSize query parameters", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test?page=3&pageSize=25", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("clamps negative or zero page and pageSize to 1 and 10", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test?page=-2&pageSize=0", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("falls back to limit query parameter if pageSize is missing or invalid", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test?page=2&limit=50", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("non-numeric parameters fall back to defaults", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test?page=invalid&pageSize=bad", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})
}
