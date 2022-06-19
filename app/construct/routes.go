package construct

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rsb/api_rate_limiter/app"
	"github.com/rsb/api_rate_limiter/app/api/handlers/health"
)

func AddAllRoutes(r *fiber.App, d *app.Dependencies) *fiber.App {
	r = AddHealthCheckRoutes(r, d)
	return r
}

func AddHealthCheckRoutes(r *fiber.App, d *app.Dependencies) *fiber.App {
	checker := health.NewCheckHandler(d)
	r.Get("/readiness", checker.Readiness)

	return r
}
