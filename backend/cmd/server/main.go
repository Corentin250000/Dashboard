package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "dashboard-backend/docs"
	"dashboard-backend/internal/about"
	"dashboard-backend/internal/models"
	"dashboard-backend/internal/routes/auth"
	"dashboard-backend/internal/utils"
	"dashboard-backend/internal/widgets"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/limiter"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/gofiber/swagger/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type RootResponse struct {
	Message string `json:"message" example:"dashboard API - see GET /about.json"`
}

type HealthResponse struct {
	Status string `json:"status" example:"ok"`
}

// @title Dashboard API
// @version 1.0
// @description API for the Dashboard services, widgets, and user accounts.
// @host localhost:8080
// @BasePath /
func main() {
	database := connectDatabase()
	if err := models.AutoMigrate(database); err != nil {
		log.Fatalf("database migration failed: %v", err)
	}

	var stationCount int64
	if err := database.Model(&models.SncfStation{}).Count(&stationCount).Error; err != nil {
		log.Fatalf("failed to check SNCF stations: %v", err)
	}
	if stationCount == 0 {
		go func() {
			if err := widgets.GetAllStations(database); err != nil {
				log.Printf("SNCF stations synchronization failed: %v", err)
			}
		}()
	}
	
	auth.SetDB(database)

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
	// credentials: 'include' on the frontend requires explicit origins —
	// a wildcard "*" is rejected by browsers when credentials are sent.
	allowOrigins := defaultAllowOrigins()
	app.Use(cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	getHead := []string{fiber.MethodGet, fiber.MethodHead}
	app.Add(getHead, "/about.json", handleAbout)
	app.Add(getHead, "/health", handleHealth)
	app.Add(getHead, "/", handleRoot)
	app.Post("/auth/register", auth.Register)
	app.Post("/auth/login", auth.Login)
	app.Get("/confirm", auth.ConfirmHandler)
	app.Post("/confirm/resend", limiter.New(limiter.Config{
		Max:        3,
		Expiration: 15 * time.Minute,
	}), auth.ResendConfirmationHandler)
	app.Get("/docs/swagger/*", swagger.New(swagger.Config{
		URL:   "/docs/swagger/doc.json",
		Title: "Dashboard API Reference",
	}))
	app.Use(func(c fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not found"})
	})

	go func() {
		log.Printf("server listening on %s", addr)
		if err := app.Listen(addr); err != nil {
			log.Printf("listen: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
		log.Printf("shutdown error: %v", err)
	}
	log.Println("server stopped")
}

func connectDatabase() *gorm.DB {
	host := utils.GetEnv("DB_HOST", "db")
	port := utils.GetEnv("DB_PORT", "5432")
	dsn := "host=" + host +
		" port=" + port +
		" user=" + os.Getenv("POSTGRES_USER") +
		" password=" + os.Getenv("POSTGRES_PASSWORD") +
		" dbname=" + os.Getenv("POSTGRES_DB") +
		" sslmode=disable"

	var database *gorm.DB
	var err error
	for attempt := 1; attempt <= 10; attempt++ {
		database, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			sqlDB, pingErr := database.DB()
			if pingErr == nil {
				pingErr = sqlDB.Ping()
			}
			if pingErr == nil {
				return database
			}
			err = pingErr
		}

		log.Printf("database unavailable (attempt %d/10): %v", attempt, err)
		time.Sleep(2 * time.Second)
	}

	log.Fatalf("unable to connect to database: %v", err)
	return nil
}

// @Summary Service catalog
// @Description Return the available services and widgets for dashboard clients.
// @Tags system
// @Produce json
// @Success 200 {object} about.Response
// @Router /about.json [get]
func handleAbout(c fiber.Ctx) error {
	host := about.ClientIP(
		c.Get(fiber.HeaderXForwardedFor),
		c.Get("X-Real-IP"),
		c.IP(),
	)
	return c.JSON(about.Build(host))
}

// @Summary Health check
// @Description Check whether the API process is responding.
// @Tags system
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func handleHealth(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok"})
}

// @Summary API root
// @Description Return a pointer to the API catalog.
// @Tags system
// @Produce json
// @Success 200 {object} RootResponse
// @Router / [get]
func handleRoot(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"message": "dashboard API — see GET /about.json"})
}

// defaultAllowOrigins returns the CORS origins for the frontend.
// FRONTEND_URL may hold a single URL or a comma-separated list;
// local dev/docker defaults are always included.
func defaultAllowOrigins() []string {
	seen := map[string]struct{}{}
	origins := []string{}
	add := func(o string) {
		o = strings.TrimSpace(o)
		if o == "" {
			return
		}
		o = strings.TrimSuffix(o, "/")
		if _, ok := seen[o]; !ok {
			seen[o] = struct{}{}
			origins = append(origins, o)
		}
	}
	for _, o := range strings.Split(utils.GetEnv("FRONTEND_URL", "http://localhost:8081"), ",") {
		add(o)
	}
	add("http://localhost:5173")
	add("http://localhost:8081")
	add("http://localhost:8080")
	return origins
}
