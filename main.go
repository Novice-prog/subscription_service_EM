package main

import (
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"log"
	"os"
	"rest-service/handlers"
	"rest-service/storage"
)

// @title Subscription API
// @version 1.0
// @description REST-сервис для управления подписками пользователей.
// @host localhost:8080
// @BasePath /
func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found")
	}

	db, err := storage.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	h := handlers.NewHandler(db)

	r := gin.Default()

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	subs := r.Group("/subscriptions")
	{
		subs.GET("", h.ListSubscriptions)
		subs.GET("/:id", h.GetSubscription)
		subs.POST("", h.CreateSubscription)
		subs.PUT("/:id", h.UpdateSubscription)
		subs.DELETE("/:id", h.DeleteSubscription)
	}

	// Ручка для подсчёта суммы
	r.GET("/summary", h.GetSummary)

	// Запуск сервера
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server is running on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
