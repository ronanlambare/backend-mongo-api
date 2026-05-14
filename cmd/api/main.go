package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lambare/go-mongo-api/internal/config"
	"github.com/lambare/go-mongo-api/internal/handlers"
	"github.com/lambare/go-mongo-api/internal/middleware"
	"github.com/lambare/go-mongo-api/internal/repository"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	cfg := config.Load()
	gin.SetMode(cfg.GinMode)

	// ── MongoDB ───────────────────────────────────────────────────────────────
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatalf("MongoDB connect: %v", err)
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Printf("MongoDB disconnect: %v", err)
		}
	}()

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("MongoDB ping: %v", err)
	}
	log.Println("✓ Connected to MongoDB")

	db := client.Database(cfg.DBName)

	// ── Repositories ─────────────────────────────────────────────────────────
	itemRepo := repository.NewItemRepository(db)
	userRepo := repository.NewUserRepository(db)

	// ── Handlers ─────────────────────────────────────────────────────────────
	authHandler := handlers.NewAuthHandler(userRepo, cfg)
	itemHandler := handlers.NewItemHandler(itemRepo)

	// ── Router ───────────────────────────────────────────────────────────────
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), middleware.CORS())

	// Docs
	r.GET("/swagger", handlers.SwaggerUI)
	r.GET("/openapi.yaml", handlers.OpenAPISpec)

	// Health
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().UTC(),
		})
	})

	// Auth (public)
	auth := r.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.Refresh)
	}

	// API (protected by JWT)
	api := r.Group("/api/v1", middleware.JWTAuth(cfg.JWTSecret))
	{
		items := api.Group("/items")
		{
			items.GET("", itemHandler.List)
			items.POST("", itemHandler.Create)
			items.GET("/:id", itemHandler.GetByID)
			items.PUT("/:id", itemHandler.Update)
			items.DELETE("/:id", itemHandler.Delete)
		}
	}

	// ── HTTP server with graceful shutdown ────────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("✓ API listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server…")
	shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutCancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		log.Fatalf("Server forced shutdown: %v", err)
	}
	log.Println("Server exited cleanly")
}
