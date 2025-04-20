package handlers

import (
	"net/http"
	"weather-app/middleware"
	"weather-app/models"
	"weather-app/services"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// ChatHandler handles all chat-related HTTP requests
type ChatHandler struct {
	chatService *services.ChatService
}

// NewChatHandler creates a new instance of ChatHandler
func NewChatHandler(chatService *services.ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

// RegisterRoutes registers all chat-related routes
func (h *ChatHandler) RegisterRoutes(router *gin.Engine) {
	// API routes with authentication middleware
	chatGroup := router.Group("/api/chat")
	chatGroup.Use(middleware.AuthRequired())
	{
		chatGroup.GET("/room", h.getChatRoom)
		chatGroup.POST("/message", h.postMessage)
		chatGroup.POST("/activity", h.updateActivity)
	}

	// Page routes with authentication middleware
	pageGroup := router.Group("")
	pageGroup.Use(middleware.AuthRequired())
	{
		pageGroup.GET("/chat", h.handleChatPage)
		pageGroup.GET("/chats", h.handleChatsListPage)
	}
}

// getChatRoom handles GET requests to retrieve a chat room
func (h *ChatHandler) getChatRoom(c *gin.Context) {
	cityName := c.Query("city")
	if cityName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "City name is required"})
		return
	}

	// Get the chat room
	room := h.chatService.GetOrCreateChatRoom(cityName)

	// Update user activity in this room
	session := sessions.Default(c)
	userID := session.Get("user_id").(int)
	h.chatService.UpdateUserActivity(userID, cityName)

	// Return the chat room
	c.JSON(http.StatusOK, room)
}

// postMessage handles POST requests to add a message to a chat room
func (h *ChatHandler) postMessage(c *gin.Context) {
	// Parse request body
	var req struct {
		CityName string `json:"city_name" binding:"required"`
		Message  string `json:"message" binding:"required"`
		ImageURL string `json:"image_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Get user info from session
	session := sessions.Default(c)
	userID := session.Get("user_id").(int)

	// Get user from store
	userStore := c.MustGet("user_store").(models.UserStore)
	user, err := userStore.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user data"})
		return
	}

	// Get avatar URL
	avatarURL := "/static/profile_photos/default.jpg"
	if user.ProfilePhoto != "" {
		avatarURL = "/static/profile_photos/" + user.ProfilePhoto
	}

	// Create message
	msg := models.ChatMessage{
		UserID:    userID,
		Username:  user.Username,
		CityName:  req.CityName,
		Message:   req.Message,
		ImageURL:  req.ImageURL,
		AvatarURL: avatarURL,
	}

	// Add message to chat room
	newMsg, err := h.chatService.AddMessage(msg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add message: " + err.Error()})
		return
	}

	// Update user activity
	h.chatService.UpdateUserActivity(userID, req.CityName)

	// Return the new message
	c.JSON(http.StatusOK, newMsg)
}

// updateActivity handles POST requests to update user activity in a chat room
func (h *ChatHandler) updateActivity(c *gin.Context) {
	// Parse request body
	var req struct {
		CityName string `json:"city_name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Get user ID from session
	session := sessions.Default(c)
	userID := session.Get("user_id").(int)

	// Update user activity
	if err := h.chatService.UpdateUserActivity(userID, req.CityName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update activity: " + err.Error()})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// handleChatPage serves the chat page for a specific city
func (h *ChatHandler) handleChatPage(c *gin.Context) {
	cityName := c.Query("city")
	if cityName == "" {
		c.Redirect(http.StatusSeeOther, "/chats")
		return
	}

	// Get user ID and username from session
	session := sessions.Default(c)
	userID := session.Get("user_id").(int)

	// Get user from store
	userStore := c.MustGet("user_store").(models.UserStore)
	user, err := userStore.GetUserByID(userID)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/logout")
		return
	}

	// Render template
	c.HTML(http.StatusOK, "chat.html", gin.H{
		"title":    "Weather Chat - " + cityName,
		"User":     user,
		"CityName": cityName,
		"UserID":   userID,
		"Username": user.Username,
	})
}

// handleChatsListPage serves the chats list page
func (h *ChatHandler) handleChatsListPage(c *gin.Context) {
	// Get user's saved cities from database
	session := sessions.Default(c)
	userID := session.Get("user_id").(int)

	// Get user from store
	userStore := c.MustGet("user_store").(models.UserStore)
	user, err := userStore.GetUserByID(userID)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/logout")
		return
	}

	// Query database for saved cities
	rows, err := h.chatService.GetDB().Query(
		"SELECT city_name FROM saved_cities WHERE user_id = ?",
		userID,
	)

	var savedCities []string
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var city string
			rows.Scan(&city)
			savedCities = append(savedCities, city)
		}
	}

	// Get popular chat rooms (rooms with most messages or active users)
	rows, err = h.chatService.GetDB().Query(
		`SELECT city_name, COUNT(*) as message_count 
		FROM chat_messages 
		WHERE created_at > NOW() - INTERVAL 24 HOUR 
		GROUP BY city_name 
		ORDER BY message_count DESC 
		LIMIT 5`,
	)

	type PopularCity struct {
		Name         string
		MessageCount int
		ActiveUsers  int
	}

	var popularCities []PopularCity
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var city PopularCity
			rows.Scan(&city.Name, &city.MessageCount)

			// Get active users count
			h.chatService.GetDB().QueryRow(
				"SELECT COUNT(*) FROM chat_active_users WHERE city_name = ? AND last_active > NOW() - INTERVAL 10 MINUTE",
				city.Name,
			).Scan(&city.ActiveUsers)

			popularCities = append(popularCities, city)
		}
	}

	// Render template
	c.HTML(http.StatusOK, "chats_list.html", gin.H{
		"title":         "Weather Chats",
		"User":          user,
		"SavedCities":   savedCities,
		"PopularCities": popularCities,
	})
}
