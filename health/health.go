package health

import (
	"github.com/gofiber/fiber/v2"
)

var InfluxReady bool

func init() {
	InfluxReady = false
}

func GetHealthz(c *fiber.Ctx) error {
	// return &fiber.Error{}
	return c.SendStatus(200)
}

func GetReadyz(c *fiber.Ctx) error {
	if !InfluxReady {
		return c.SendStatus(503)
	}
	return c.SendStatus(200)
}
