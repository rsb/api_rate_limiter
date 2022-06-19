// Package ping is responsible for the wep entry point logic of our
// example rate limiting app. It will have behavior for when the user
// queries the ping endpoint it will return with a json pong response
package ping

import "github.com/gofiber/fiber/v2"

type PongHandler struct{}

func (p *PongHandler) Ping(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"value": "pong"})
}
