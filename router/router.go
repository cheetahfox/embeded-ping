package router

import (
	"github.com/cheetahfox/embeded-ping/health"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func SetupRoutes(app *fiber.App) {

	// Setup the routes
	app.Get("/healthz", health.GetHealthz)
	app.Get("/readyz", health.GetReadyz)
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler())) // Prometheus metrics endpoint

}
