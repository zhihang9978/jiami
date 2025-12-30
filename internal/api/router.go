package api

import (
	"github.com/feiji/feiji-backend/internal/auth"
	"github.com/feiji/feiji-backend/internal/ws"
	"github.com/gin-gonic/gin"
)

type Handlers struct {
	Main        *Handler
	Contacts    *ContactsHandler
	Users       *UsersHandler
	Files       *FilesHandler
	Updates     *UpdatesHandler
	Chats       *ChatsHandler
	Channels    *ChannelsHandler
	Search      *SearchHandler
	Push        *PushHandler
	Admin       *AdminHandler
	Calls       *CallsHandler
	Media       *MediaHandler
	SecretChats *SecretChatsHandler
	Broadcasts  *BroadcastsHandler
	WS          *ws.Handler
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

		// Upload (Files)
		uploadGroup := protected.Group("/upload")
		{
			uploadGroup.POST("/saveFilePart", handlers.Files.SaveFilePart)
			uploadGroup.POST("/saveBigFilePart", handlers.Files.SaveBigFilePart)
			uploadGroup.POST("/getFile", handlers.Files.GetFile)
			uploadGroup.POST("/completeUpload", handlers.Files.CompleteUpload)
			uploadGroup.GET("/getNextFileID", handlers.Files.GetNextFileID)
		}

		// Updates
		updatesGroup := protected.Group("/updates")
		{
			updatesGroup.POST("/getState", handlers.Updates.GetState)
			updatesGroup.GET("/getState", handlers.Updates.GetState)
			updatesGroup.POST("/getDifference", handlers.Updates.GetDifference)
			updatesGroup.POST("/getChannelDifference", handlers.Updates.GetChannelDifference)
		}

		// Chats (Groups)
		if handlers.Chats != nil {
			chatsGroup := protected.Group("/messages")
			{
				chatsGroup.POST("/createChat", handlers.Chats.CreateChat)
				chatsGroup.POST("/getFullChat", handlers.Chats.GetFullChat)
				chatsGroup.POST("/editChatTitle", handlers.Chats.EditChatTitle)
				chatsGroup.POST("/editChatPhoto", handlers.Chats.EditChatPhoto)
				chatsGroup.POST("/addChatUser", handlers.Chats.AddChatUser)
				chatsGroup.POST("/deleteChatUser", handlers.Chats.DeleteChatUser)
				chatsGroup.POST("/leaveChat", handlers.Chats.LeaveChat)
				chatsGroup.POST("/editChatAdmin", handlers.Chats.EditChatAdmin)
				chatsGroup.POST("/getCommonChats", handlers.Chats.GetCommonChats)
			}
		}

		// Channels
		if handlers.Channels != nil {
			channelsGroup := protected.Group("/channels")
			{
				channelsGroup.POST("/createChannel", handlers.Channels.CreateChannel)
				channelsGroup.POST("/getFullChannel", handlers.Channels.GetFullChannel)
				channelsGroup.POST("/editTitle", handlers.Channels.EditTitle)
				channelsGroup.POST("/editAbout", handlers.Channels.EditAbout)
				channelsGroup.POST("/updateUsername", handlers.Channels.UpdateUsername)
				channelsGroup.POST("/joinChannel", handlers.Channels.JoinChannel)
				channelsGroup.POST("/leaveChannel", handlers.Channels.LeaveChannel)
				channelsGroup.POST("/inviteToChannel", handlers.Channels.InviteToChannel)
				channelsGroup.POST("/kickFromChannel", handlers.Channels.KickFromChannel)
				channelsGroup.POST("/getParticipants", handlers.Channels.GetParticipants)
				channelsGroup.POST("/getChannels", handlers.Channels.GetChannels)
				channelsGroup.GET("/getChannels", handlers.Channels.GetChannels)
				channelsGroup.POST("/checkUsername", handlers.Channels.CheckUsername)
			}
		}

		// Search
		if handlers.Search != nil {
			searchGroup := protected.Group("/messages")
			{
				searchGroup.POST("/searchGlobal", handlers.Search.SearchGlobal)
				searchGroup.POST("/search", handlers.Search.SearchMessages)
				searchGroup.POST("/searchHashtag", handlers.Search.SearchHashtag)
			}
			contactsSearchGroup := protected.Group("/contacts")
			{
				contactsSearchGroup.POST("/getRecentSearch", handlers.Search.GetRecentSearch)
				contactsSearchGroup.POST("/clearRecentSearch", handlers.Search.ClearRecentSearch)
				contactsSearchGroup.POST("/searchUsers", handlers.Search.SearchUsers)
				contactsSearchGroup.POST("/searchChannels", handlers.Search.SearchChannels)
			}
		}

		// Push Notifications
		if handlers.Push != nil {
			accountPushGroup := protected.Group("/account")
			{
				accountPushGroup.POST("/registerDevice", handlers.Push.RegisterDevice)
				accountPushGroup.POST("/unregisterDevice", handlers.Push.UnregisterDevice)
				accountPushGroup.POST("/getNotifySettings", handlers.Push.GetNotifySettings)
				accountPushGroup.POST("/updateNotifySettings", handlers.Push.UpdateNotifySettings)
				accountPushGroup.POST("/resetNotifySettings", handlers.Push.ResetNotifySettings)
				accountPushGroup.POST("/getAllNotifySettings", handlers.Push.GetAllNotifySettings)
			}
		}

		// Phone Calls (VoIP)
		if handlers.Calls != nil {
			phoneGroup := protected.Group("/phone")
			{
				phoneGroup.POST("/requestCall", handlers.Calls.RequestCall)
				phoneGroup.POST("/acceptCall", handlers.Calls.AcceptCall)
				phoneGroup.POST("/discardCall", handlers.Calls.DiscardCall)
				phoneGroup.POST("/confirmCall", handlers.Calls.ConfirmCall)
				phoneGroup.POST("/receivedCall", handlers.Calls.ReceivedCall)
				phoneGroup.POST("/setCallRating", handlers.Calls.SetCallRating)
				phoneGroup.POST("/saveCallDebug", handlers.Calls.SaveCallDebug)
				phoneGroup.POST("/getCallConfig", handlers.Calls.GetCallConfig)
				phoneGroup.GET("/getCallConfig", handlers.Calls.GetCallConfig)
			}
		}
	}

	// REST-style Media Upload (Devin.md endpoints)
	if handlers.Media != nil {
		// Simple upload endpoints
		uploadMediaGroup := protected.Group("/api/v1/upload")
		{
			uploadMediaGroup.POST("/voice", handlers.Media.UploadVoice)
			uploadMediaGroup.POST("/video", handlers.Media.UploadVideo)
			uploadMediaGroup.POST("/file", handlers.Media.UploadFile)
			uploadMediaGroup.POST("/image", handlers.Media.UploadImage)
		}

		// Multipart upload endpoints
		mediaGroup := protected.Group("/api/v1/media")
		{
			mediaGroup.POST("/init", handlers.Media.MediaInit)
			mediaGroup.POST("/upload", handlers.Media.MediaUpload)
			mediaGroup.POST("/complete", handlers.Media.MediaComplete)
			mediaGroup.GET("/:file_id", handlers.Media.GetMedia)
		}
	}

	// Secret Chats (E2E encrypted)
	if handlers.SecretChats != nil {
		secretChatsGroup := protected.Group("/api/v1/secret_chats")
		{
			secretChatsGroup.POST("/create", handlers.SecretChats.CreateSecretChat)
			secretChatsGroup.POST("/status", handlers.SecretChats.UpdateSecretChatStatus)
			secretChatsGroup.GET("/list", handlers.SecretChats.GetSecretChatList)
			secretChatsGroup.POST("/close", handlers.SecretChats.CloseSecretChat)
		}

		secretMessagesGroup := protected.Group("/api/v1/secret_messages")
		{
			secretMessagesGroup.POST("/send", handlers.SecretChats.SendSecretMessage)
			secretMessagesGroup.GET("/history", handlers.SecretChats.GetSecretMessageHistory)
		}
	}

	// Broadcasts
	if handlers.Broadcasts != nil {
		broadcastsGroup := protected.Group("/api/v1/broadcasts")
		{
			broadcastsGroup.POST("/create", handlers.Broadcasts.CreateBroadcast)
			broadcastsGroup.GET("/list", handlers.Broadcasts.GetBroadcastList)
			broadcastsGroup.GET("/:id", handlers.Broadcasts.GetBroadcast)
			broadcastsGroup.PUT("/:id", handlers.Broadcasts.UpdateBroadcast)
			broadcastsGroup.DELETE("/:id", handlers.Broadcasts.DeleteBroadcast)
			broadcastsGroup.POST("/:id/send", handlers.Broadcasts.SendBroadcast)
		}
	}

	// Admin routes (with authentication middleware)
	if handlers.Admin != nil {
		adminGroup := r.Group("/admin")
		adminGroup.Use(AdminAuthMiddleware())
		{
			adminGroup.GET("/stats", handlers.Admin.GetStats)
			adminGroup.GET("/users", handlers.Admin.GetUsers)
			adminGroup.GET("/users/:id", handlers.Admin.GetUser)
			adminGroup.POST("/users", handlers.Admin.CreateUser)
			adminGroup.PUT("/users/:id", handlers.Admin.UpdateUser)
			adminGroup.DELETE("/users/:id", handlers.Admin.DeleteUser)
			adminGroup.POST("/users/:id/ban", handlers.Admin.BanUser)
			adminGroup.POST("/users/:id/unban", handlers.Admin.UnbanUser)
			adminGroup.GET("/messages", handlers.Admin.GetMessages)
			adminGroup.DELETE("/messages/:id", handlers.Admin.DeleteMessage)
			adminGroup.GET("/chats", handlers.Admin.GetChats)
			adminGroup.GET("/channels", handlers.Admin.GetChannels)
			adminGroup.GET("/users/:id/sessions", handlers.Admin.GetSessions)
			adminGroup.DELETE("/sessions/:id", handlers.Admin.TerminateSession)
			adminGroup.DELETE("/users/:id/sessions", handlers.Admin.TerminateAllSessions)
		}
	}

	// WebSocket endpoint
	if handlers.WS != nil {
		r.GET("/ws", handlers.WS.HandleWebSocket)
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
		c.Set("user_id", user.ID)
		c.Next()
	}
}

// AdminAuthMiddleware validates admin JWT token (section 3.11.13 of BACKEND_COMPLETE_SPECIFICATION.md)
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(401, gin.H{"error": "unauthorized", "message": "Missing Authorization header"})
			c.Abort()
			return
		}

		// Remove "Bearer " prefix if present
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		// For now, accept any non-empty token (in production, validate JWT)
		// TODO: Implement proper JWT validation with secret key
		if token == "" {
			c.JSON(401, gin.H{"error": "unauthorized", "message": "Invalid token"})
			c.Abort()
			return
		}

		// Set admin info in context
		c.Set("admin_token", token)
		c.Set("admin_role", "super_admin") // Default role
		c.Next()
	}
}
