package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dashboard-backend/internal/about"
	"dashboard-backend/internal/utils"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

func main() {
	port := utils.GetEnv("PORT", "8080")
	addr := ":" + port

	app := fiber.New(fiber.Config{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	})

	app.Use(requestid.New())
	app.Use(logger.New(logger.Config{
		Format: "${time} ${respHeader:X-Request-Id} ${status} ${method} ${path} ${latency}\n",
	}))
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Content-Type", "Authorization"},
	}))

	// GET routes also answer HEAD (proper HTTP semantics; the Docker
	// healthcheck probes /health with `wget --spider`, i.e. HEAD).
	getHead := []string{fiber.MethodGet, fiber.MethodHead}
	app.Add(getHead, "/about.json", handleAbout)
	app.Add(getHead, "/health", handleHealth)
	app.Add(getHead, "/", handleRoot)
	app.Use(func(c fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not found"})
	})

	go func() {
		log.Printf("server listening on %s", addr)
		if err := app.Listen(addr); err != nil {
			log.Printf("listen: %v", err)
		}
	}()

	// Graceful shutdown on SIGINT/SIGTERM (docker stop).
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
		log.Printf("shutdown error: %v", err)
	}
	log.Println("server stopped")
}

func handleAbout(c fiber.Ctx) error {
	host := about.ClientIP(
		c.Get(fiber.HeaderXForwardedFor),
		c.Get("X-Real-IP"),
		c.IP(),
	)
	return c.JSON(about.Build(host))
}

func handleHealth(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}

func handleRoot(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "dashboard API — see GET /about.json"})
}
