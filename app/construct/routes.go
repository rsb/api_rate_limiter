package construct

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rsb/api_rate_limiter/app"
	"github.com/rsb/api_rate_limiter/app/api/handlers/health"
	"github.com/rsb/api_rate_limiter/app/api/handlers/ping"
)

func AddAllRoutes(a *fiber.App, d *app.Dependencies) *fiber.App {
	a = AddHealthCheckRoutes(a, d)
	a = AddPingRoutes(a, d)

	return a
}

func AddHealthCheckRoutes(a *fiber.App, d *app.Dependencies) *fiber.App {
	checker := health.NewCheckHandler(d)
	a.Get("/readiness", checker.Readiness)

	return a
}

func AddPingRoutes(a *fiber.App, _ *app.Dependencies) *fiber.App {
	h := &ping.PongHandler{}
	a.Get("/ping", h.Ping)
	return a
}
