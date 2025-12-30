package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/feiji/feiji-backend/internal/admin"
	"github.com/feiji/feiji-backend/internal/api"
	"github.com/feiji/feiji-backend/internal/auth"
	"github.com/feiji/feiji-backend/internal/broadcasts"
	"github.com/feiji/feiji-backend/internal/calls"
	"github.com/feiji/feiji-backend/internal/channels"
	"github.com/feiji/feiji-backend/internal/chats"
	"github.com/feiji/feiji-backend/internal/config"
	"github.com/feiji/feiji-backend/internal/contacts"
	"github.com/feiji/feiji-backend/internal/files"
	"github.com/feiji/feiji-backend/internal/messages"
	"github.com/feiji/feiji-backend/internal/push"
	"github.com/feiji/feiji-backend/internal/search"
	"github.com/feiji/feiji-backend/internal/secretchats"
	"github.com/feiji/feiji-backend/internal/store"
	"github.com/feiji/feiji-backend/internal/updates"
	"github.com/feiji/feiji-backend/internal/users"
	"github.com/feiji/feiji-backend/internal/ws"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Set Gin mode
	gin.SetMode(cfg.GinMode)

	// Initialize MySQL store
	mysqlStore, err := store.NewMySQLStore(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}
	defer mysqlStore.Close()
	log.Println("Connected to MySQL")

	// Initialize Redis store
	redisStore, err := store.NewRedisStore(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisStore.Close()
	log.Println("Connected to Redis")

	// Initialize repositories
	authRepo := auth.NewRepository(mysqlStore.DB())
	messagesRepo := messages.NewRepository(mysqlStore.DB())
	contactsRepo := contacts.NewRepository(mysqlStore.DB())
	usersRepo := users.NewRepository(mysqlStore.DB())
	filesRepo := files.NewRepository(mysqlStore.DB())
	updatesRepo := updates.NewRepository(mysqlStore.DB())
	chatsRepo := chats.NewRepository(mysqlStore.DB())
	channelsRepo := channels.NewRepository(mysqlStore.DB())
	searchRepo := search.NewRepository(mysqlStore.DB())
	pushRepo := push.NewRepository(mysqlStore.DB())
	adminRepo := admin.NewRepository(mysqlStore.DB())
	callsRepo := calls.NewRepository(mysqlStore.DB())
	secretChatsRepo := secretchats.NewRepository(mysqlStore.DB())
	broadcastsRepo := broadcasts.NewRepository(mysqlStore.DB())

	// Initialize services
	authService := auth.NewService(authRepo, redisStore)
	messagesService := messages.NewService(messagesRepo, redisStore)
	contactsService := contacts.NewService(contactsRepo)
	usersService := users.NewService(usersRepo)
	filesService := files.NewService(filesRepo, cfg.UploadPath, cfg.BaseURL)
	updatesService := updates.NewService(updatesRepo)
	chatsService := chats.NewService(chatsRepo)
	channelsService := channels.NewService(channelsRepo)
	searchService := search.NewService(searchRepo)
	pushService := push.NewService(pushRepo)
	adminService := admin.NewService(adminRepo)
	callsService := calls.NewService(callsRepo)
	secretChatsService := secretchats.NewService(secretChatsRepo)
	broadcastsService := broadcasts.NewService(broadcastsRepo)

	// Initialize WebSocket hub
	hub := ws.NewHub()
	go hub.Run()
	log.Println("WebSocket hub started")

	// Initialize handlers
	handlers := &api.Handlers{
		Main:        api.NewHandler(authService, messagesService),
		Contacts:    api.NewContactsHandler(contactsService),
		Users:       api.NewUsersHandler(usersService),
		Files:       api.NewFilesHandler(filesService),
		Updates:     api.NewUpdatesHandler(updatesService),
		Chats:       api.NewChatsHandler(chatsService),
		Channels:    api.NewChannelsHandler(channelsService),
		Search:      api.NewSearchHandler(searchService),
		Push:        api.NewPushHandler(pushService),
		Admin:       api.NewAdminHandler(adminService),
		Calls:       api.NewCallsHandler(callsService),
		Media:       api.NewMediaHandler(cfg.UploadPath, cfg.BaseURL),
		SecretChats: api.NewSecretChatsHandler(secretChatsService),
		Broadcasts:  api.NewBroadcastsHandler(broadcastsService),
		WS:          ws.NewHandler(hub, authService),
	}

	// Setup router
	router := api.SetupRouter(handlers, authService)

	// Start server
	log.Printf("Starting API server on port %s", cfg.HTTPPort)

	// Graceful shutdown
	go func() {
		if err := router.Run(":" + cfg.HTTPPort); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
}
