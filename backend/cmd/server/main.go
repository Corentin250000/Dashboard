package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"dashboard-backend/internal/about"
	"dashboard-backend/internal/models"
	"dashboard-backend/internal/routes/auth"
	"dashboard-backend/internal/utils"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	database := connectDatabase()
	if err := models.AutoMigrate(database); err != nil {
		log.Fatalf("database migration failed: %v", err)
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
