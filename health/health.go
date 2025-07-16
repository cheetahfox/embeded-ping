package health

import (
	"github.com/cheetahfox/longping/config"
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
	if !InfluxReady && config.Config.InfluxEnabled {
		return c.SendStatus(503)
	}
	return c.SendStatus(200)
}
