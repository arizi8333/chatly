// @title Chatly API
// @version 1.0
// @description Scalable Real-time Chat Backend in Go
// @host localhost:8090
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and then your personal token.
package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"chatly/internal/config"
	"chatly/internal/delivery"
	"chatly/internal/domain"
	"chatly/internal/repository"
	"chatly/internal/service"
	"chatly/internal/websocket"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.LoadConfig()

	// 1. Setup PostgreSQL
	db, err := repository.ConnectDB(cfg.DBURL)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close()

	if err := repository.RunMigrations(db, "migrations"); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	// 2. Setup Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisURL,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}

	// 3. Initialize Hub & PubSub
	hub := websocket.NewHub()
	go hub.Run()

	pubsub := websocket.NewPubSubManager(rdb, hub)
	go pubsub.Subscribe(context.Background())

	// 4. Repositories
	userRepo := repository.NewPostgresUserRepository(db)
	tokenRepo := repository.NewPostgresTokenRepository(db)
	messageRepo := repository.NewPostgresMessageRepository(db)

	// 5. Services
	authService := service.NewAuthService(userRepo, tokenRepo, cfg)

	// 6. Handlers
	authHandler := delivery.NewAuthHandler(authService)
	msgHandler := delivery.NewMessageHandler(messageRepo)

	// WS Logic: Persistence + PubSub
	onMessage := func(msg domain.Message) {
		msg.ID = uuid.New()
		msg.CreatedAt = time.Now()

		// Save to DB
		if err := messageRepo.Save(context.Background(), &msg); err != nil {
			log.Printf("error saving message: %v", err)
		}

		// Publish to Redis Pub/Sub for other instances
		if err := pubsub.Publish(context.Background(), msg); err != nil {
			log.Printf("error publishing message: %v", err)
		}

		// Broadcast to local hub
		hub.BroadcastToRoom(msg.Room, msg)
	}
	wsHandler := websocket.NewHandler(hub, userRepo, onMessage)

	// 7. Router
	router := delivery.NewRouter(cfg, rdb, authHandler, msgHandler, wsHandler)

	// 8. Start Server
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	log.Printf("Server starting on port %s", cfg.Port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
