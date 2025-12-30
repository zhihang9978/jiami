package api

import (
	"github.com/feiji/feiji-backend/internal/auth"
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	Main     *Handler
	Contacts *ContactsHandler
	Users    *UsersHandler
}

func SetupRouter(handlers *Handlers, authService *auth.Service) *gin.Engine {
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Auth-Key-ID")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Health check
	r.GET("/health", handlers.Main.Health)
	r.GET("/", handlers.Main.Health)

	// Auth routes (no authentication required)
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/sendCode", handlers.Main.SendCode)
		authGroup.POST("/signIn", handlers.Main.SignIn)
		authGroup.POST("/signUp", handlers.Main.SignUp)
		authGroup.POST("/logOut", handlers.Main.LogOut)
	}

	// Protected routes (authentication required)
	protected := r.Group("/")
	protected.Use(AuthMiddleware(authService))
	{
		// Messages
		messagesGroup := protected.Group("/messages")
		{
			messagesGroup.POST("/sendMessage", handlers.Main.SendMessage)
			messagesGroup.POST("/getHistory", handlers.Main.GetHistory)
			messagesGroup.POST("/getDialogs", handlers.Main.GetDialogs)
			messagesGroup.GET("/getDialogs", handlers.Main.GetDialogs)
			messagesGroup.POST("/readHistory", handlers.Main.ReadHistory)
			messagesGroup.POST("/deleteMessages", handlers.Main.DeleteMessages)
			messagesGroup.POST("/editMessage", handlers.Main.EditMessage)
		}

		// Contacts
		contactsGroup := protected.Group("/contacts")
		{
			contactsGroup.POST("/getContacts", handlers.Contacts.GetContacts)
			contactsGroup.GET("/getContacts", handlers.Contacts.GetContacts)
			contactsGroup.POST("/importContacts", handlers.Contacts.ImportContacts)
			contactsGroup.POST("/addContact", handlers.Contacts.AddContact)
			contactsGroup.POST("/deleteContacts", handlers.Contacts.DeleteContacts)
			contactsGroup.POST("/block", handlers.Contacts.Block)
			contactsGroup.POST("/unblock", handlers.Contacts.Unblock)
			contactsGroup.POST("/getBlocked", handlers.Contacts.GetBlocked)
			contactsGroup.GET("/getBlocked", handlers.Contacts.GetBlocked)
			contactsGroup.POST("/resolveUsername", handlers.Users.ResolveUsername)
			contactsGroup.POST("/search", handlers.Users.Search)
		}

		// Users
		usersGroup := protected.Group("/users")
		{
			usersGroup.POST("/getUsers", handlers.Users.GetUsers)
			usersGroup.POST("/getFullUser", handlers.Users.GetFullUser)
		}

		// Account
		accountGroup := protected.Group("/account")
		{
			accountGroup.POST("/updateProfile", handlers.Users.UpdateProfile)
			accountGroup.POST("/updateUsername", handlers.Users.UpdateUsername)
			accountGroup.POST("/checkUsername", handlers.Users.CheckUsername)
			accountGroup.POST("/updateStatus", handlers.Users.UpdateStatus)
		}
	}

	return r
}

// AuthMiddleware validates the user session
func AuthMiddleware(authService *auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		authKeyID := c.GetHeader("X-Auth-Key-ID")
		if authKeyID == "" {
			// Try to get from Authorization header
			authKeyID = c.GetHeader("Authorization")
		}

		if authKeyID == "" {
			c.JSON(401, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		user, err := authService.GetUserByAuthKey(c.Request.Context(), authKeyID)
		if err != nil || user == nil {
			c.JSON(401, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		// Update user online status
		authService.UpdateUserOnline(c.Request.Context(), user.ID)

		c.Set("user", user)
		c.Next()
	}
}
