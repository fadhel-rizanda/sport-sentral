package request

import "github.com/gofiber/fiber/v2"

// ParsePagination safely extracts and clamps page and pageSize (or limit) query parameters.
// Ensures page >= 1 and pageSize >= 1 to prevent downstream gRPC validation errors.
func ParsePagination(c *fiber.Ctx) (page int, pageSize int) {
	page = c.QueryInt("page", 1)
	if page < 1 {
		page = 1
	}

	pageSize = c.QueryInt("pageSize", 0)
	if pageSize < 1 {
		pageSize = c.QueryInt("limit", 10)
		if pageSize < 1 {
			pageSize = 10
		}
	}

	return page, pageSize
}
