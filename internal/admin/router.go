package admin

import (
	"database/sql"

	"github.com/feiji/feiji-backend/internal/admin/handlers"
	"github.com/feiji/feiji-backend/internal/admin/middleware"
	"github.com/feiji/feiji-backend/internal/admin/websocket"
	"github.com/gin-gonic/gin"
)

// SetupAdminRouter sets up all admin panel routes
func SetupAdminRouter(r *gin.Engine, db *sql.DB) {
	// Initialize WebSocket hub
	websocket.InitAdminHub()

	// Create handlers
	adminHandlers := handlers.NewAdminHandlersWithDB(db)
	monitorHandlers := handlers.NewMonitorHandlers()

	// Admin API group with CORS
	adminAPI := r.Group("/api/admin")
	adminAPI.Use(middleware.CORS())

	// Public routes (no authentication required)
	{
		adminAPI.POST("/login", adminHandlers.Login)
	}

	// Protected routes (authentication required)
	protected := adminAPI.Group("/")
	protected.Use(middleware.AdminAuth())
	{
		// Auth
		protected.POST("/logout", adminHandlers.Logout)
		protected.GET("/profile", adminHandlers.GetProfile)
		protected.PUT("/profile/password", adminHandlers.ChangePassword)
		protected.GET("/profile/login-logs", adminHandlers.GetLoginLogs)

		// Notifications
		protected.GET("/notifications", adminHandlers.GetNotifications)
		protected.PUT("/notifications/:id/read", adminHandlers.MarkNotificationRead)
		protected.PUT("/notifications/read-all", adminHandlers.MarkAllNotificationsRead)
		protected.DELETE("/notifications/:id", adminHandlers.DeleteNotification)

		// Users
		protected.GET("/users", adminHandlers.GetUsers)
		protected.GET("/users/:id", adminHandlers.GetUser)
		protected.POST("/users", adminHandlers.CreateUser)
		protected.PUT("/users/:id", adminHandlers.UpdateUser)
		protected.DELETE("/users/:id", adminHandlers.DeleteUser)
		protected.POST("/users/batch", adminHandlers.BatchUpdateUsers)

		// Calls
		protected.GET("/calls", adminHandlers.GetCalls)
		protected.GET("/calls/:id", adminHandlers.GetCall)

		// Broadcasts
		protected.GET("/broadcasts", adminHandlers.GetBroadcasts)
		protected.GET("/broadcasts/:id", adminHandlers.GetBroadcast)
		protected.POST("/broadcasts", adminHandlers.SendBroadcast)
		protected.POST("/broadcasts/:id/retry", adminHandlers.RetryBroadcast)

		// Message Templates
		protected.GET("/templates", adminHandlers.GetMessageTemplates)
		protected.POST("/templates", adminHandlers.CreateMessageTemplate)
		protected.PUT("/templates/:id", adminHandlers.UpdateMessageTemplate)
		protected.DELETE("/templates/:id", adminHandlers.DeleteMessageTemplate)

		// Auto Messages
		protected.GET("/auto-messages", adminHandlers.GetAutoMessageConfigs)
		protected.PUT("/auto-messages/:type", adminHandlers.UpdateAutoMessageConfig)

		// System Monitoring
		protected.GET("/monitor/server", monitorHandlers.GetServerStatus)
		protected.GET("/monitor/application", monitorHandlers.GetApplicationStatus)
		protected.GET("/monitor/services", monitorHandlers.GetAllServicesStatus)
		protected.GET("/monitor/services/:name", monitorHandlers.GetServiceStatus)
		protected.POST("/monitor/services/:name/restart", monitorHandlers.RestartService)
		protected.POST("/monitor/services/:name/stop", monitorHandlers.StopService)
		protected.POST("/monitor/services/:name/start", monitorHandlers.StartService)
		protected.GET("/monitor/services/:name/logs", monitorHandlers.GetServiceLogs)
		protected.GET("/monitor/realtime", monitorHandlers.GetMonitorData)

		// Statistics
		protected.GET("/statistics/users", adminHandlers.GetUserStatistics)
		protected.GET("/statistics/calls", adminHandlers.GetCallStatistics)
		protected.GET("/statistics/bot", adminHandlers.GetBotStatistics)
		protected.GET("/dashboard", adminHandlers.GetDashboard)

		// System Settings
		protected.GET("/settings", adminHandlers.GetSystemConfigs)
		protected.PUT("/settings/:key", adminHandlers.UpdateSystemConfig)

		// Audit Logs
		protected.GET("/audit-logs", adminHandlers.GetAuditLogs)
	}

	// WebSocket endpoint for real-time notifications
	r.GET("/api/admin/ws", websocket.HandleAdminWebSocket)
}
