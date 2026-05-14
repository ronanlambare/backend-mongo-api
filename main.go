// @title           Backend Mongo API
// @version         1.0
// @description     REST API backed by MongoDB with JWT authentication.
//
// @host            localhost:8080
// @BasePath        /api
//
// @securityDefinitions.apikey BearerAuth
// @in              header
// @name            Authorization
// @description     JWT token. Format: "Bearer {token}"
package main

import (
	"context"
	"log"
	"os"
	"time"

	_ "github.com/ronanlambare/backend-mongo-api/docs"
	"github.com/ronanlambare/backend-mongo-api/internal/handler"
	"github.com/ronanlambare/backend-mongo-api/internal/middleware"
	"github.com/ronanlambare/backend-mongo-api/internal/repository"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	mongoURI  := getEnv("MONGODB_URI", "mongodb://localhost:27017")
	dbName    := getEnv("MONGODB_DB", "mydb")
	jwtSecret := getEnv("JWT_SECRET", "change-me-in-production")
	port      := getEnv("PORT", "8080")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(context.Background())

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("failed to ping MongoDB: %v", err)
	}
	log.Println("connected to MongoDB")

	db := client.Database(dbName)
	itemRepo := repository.NewItemRepository(db)
	userRepo := repository.NewUserRepository(db)

	// OIDC providers: add env vars for each provider you support.
	// The key is the issuer URL, the value is your OAuth2 client ID for that provider.
	oidcProviders := map[string]string{}
	if id := getEnv("OIDC_GOOGLE_CLIENT_ID", ""); id != "" {
		oidcProviders["https://accounts.google.com"] = id
	}
	if id := getEnv("OIDC_APPLE_CLIENT_ID", ""); id != "" {
		oidcProviders["https://appleid.apple.com"] = id
	}

	itemHandler := handler.NewItemHandler(itemRepo)
	authHandler, err := handler.NewAuthHandler(userRepo, jwtSecret, oidcProviders)
	if err != nil {
		log.Fatalf("failed to initialize auth handler: %v", err)
	}

	router := gin.Default()

	// Swagger UI — available at /swagger/index.html
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.GET("/health", func(c *gin.Context) { c.Status(200) })

	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/oidc", authHandler.OIDCLogin)
		}

		items := api.Group("/items")
		items.Use(middleware.JWTAuth(jwtSecret))
		{
			items.GET("", itemHandler.List)
			items.POST("", itemHandler.Create)
			items.GET("/:id", itemHandler.GetByID)
			items.PUT("/:id", itemHandler.Update)
			items.DELETE("/:id", itemHandler.Delete)
		}
	}

	log.Printf("server listening on :%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

