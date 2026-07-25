package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/otelforge/otelforge/api/internal/handlers"
	"github.com/otelforge/otelforge/internal/auth"
	"github.com/otelforge/otelforge/internal/config"
	"github.com/otelforge/otelforge/internal/db"
	"github.com/otelforge/otelforge/internal/events"
	"github.com/otelforge/otelforge/internal/queue"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	if err := db.Migrate(ctx, pool); err != nil {
		log.Fatal(err)
	}

	q, err := queue.Connect(cfg.RabbitMQURL)
	if err != nil {
		log.Fatal(err)
	}
	defer q.Close()
	if err := q.Declare(ctx); err != nil {
		log.Fatal(err)
	}

	store := db.NewStore(pool)
	tokens := auth.NewTokenService(cfg.JWTSecret)
	authMW := auth.NewMiddleware(tokens, store)
	eventSvc := events.NewService(store, q)

	authH := handlers.NewAuthHandler(store, tokens)
	instH := handlers.NewInstancesHandler(store, cfg.EncryptionKey)
	evH := handlers.NewEventsHandler(store, eventSvc)
	adminH := handlers.NewAdminHandler(store)

	app := fiber.New(fiber.Config{
		AppName: "OtelForge API",
	})
	app.Use(recover.New(), logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSOrigin,
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Authorization,Content-Type,Accept",
		AllowCredentials: true,
	}))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	api := app.Group("/api/v1")
	api.Post("/auth/login", authH.Login)

	protected := api.Group("", authMW.RequireAuth)
	protected.Get("/instances", instH.List)
	protected.Post("/instances", instH.Create)
	protected.Post("/instances/bulk", instH.BulkCreate)
	protected.Put("/instances/:id", instH.Update)
	protected.Delete("/instances/:id", instH.Delete)

	protected.Get("/events", evH.List)
	protected.Post("/events", evH.Create)
	protected.Get("/events/:id", evH.Get)
	protected.Post("/events/:id/rerun-failed", evH.RerunFailed)
	protected.Post("/events/:id/clone", evH.Clone)
	protected.Post("/events/:id/jobs/:jobId/rollback", evH.RollbackJob)

	admin := api.Group("/admin", authMW.RequireAuth, authMW.RequireAdmin)
	admin.Get("/events", adminH.ListEvents)
	admin.Get("/instances", adminH.ListInstances)

	go func() {
		log.Printf("api listening on %s", cfg.APIAddr)
		if err := app.Listen(cfg.APIAddr); err != nil {
			log.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	if err := app.Shutdown(); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
