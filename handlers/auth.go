package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"weather-app/models"
)

// AuthHandler handles authentication related routes
type AuthHandler struct {
	userStore models.UserStore
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userStore models.UserStore) *AuthHandler {
	return &AuthHandler{
		userStore: userStore,
	}
}

// GetSignup displays signup form
func (h *AuthHandler) GetSignup(c *gin.Context) {
	c.HTML(http.StatusOK, "signup.html", gin.H{
		"title": "Sign Up - Weather App",
	})
}

// PostSignup processes signup form
func (h *AuthHandler) PostSignup(c *gin.Context) {
	// Get form data
	username := c.PostForm("username")
	email := c.PostForm("email")
	password := c.PostForm("password")
	homeCity := c.PostForm("home_city")
	acceptTerms := c.PostForm("terms")

	log.Printf("Signup attempt: username=%s, email=%s, homeCity=%s, termsAccepted=%s",
		username, email, homeCity, acceptTerms)

	// Validate input
	if username == "" || email == "" || password == "" {
		c.HTML(http.StatusBadRequest, "signup.html", gin.H{
			"title": "Sign Up - Weather App",
			"error": "All fields are required",
		})
		return
	}

	// Check if terms are accepted
	if acceptTerms != "on" {
		c.HTML(http.StatusBadRequest, "signup.html", gin.H{
			"title": "Sign Up - Weather App",
			"error": "You must accept the Terms & Conditions",
		})
		return
	}

	// Create new user
	user := &models.User{
		Username:             username,
		Email:                email,
		Password:             password, // Will be hashed in CreateUser
		HomeCity:             homeCity,
		NotificationsEnabled: false,
		AlertThreshold:       "severe",      // Default to severe alerts only
		ProfilePhoto:         "default.jpg", // Default profile photo
	}

	// Try to save the user
	err := h.userStore.CreateUser(user)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		c.HTML(http.StatusBadRequest, "signup.html", gin.H{
			"title": "Sign Up - Weather App",
			"error": err.Error(),
		})
		return
	}

	log.Printf("User created successfully: ID=%d, username=%s", user.ID, user.Username)

	// Log the user in
	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	err = session.Save()
	if err != nil {
		log.Printf("Error saving session: %v", err)
		c.HTML(http.StatusInternalServerError, "signup.html", gin.H{
			"title": "Sign Up - Weather App",
			"error": "Failed to create session. Please try logging in.",
		})
		return
	}

	// Redirect to home
	c.Redirect(http.StatusFound, "/")
}

// GetLogin displays login form
func (h *AuthHandler) GetLogin(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", gin.H{
		"title": "Login - Weather App",
	})
}

// PostLogin processes login form
func (h *AuthHandler) PostLogin(c *gin.Context) {
	// Get form data
	username := c.PostForm("username")
	password := c.PostForm("password")

	log.Printf("Login attempt: username=%s", username)

	// Validate input
	if username == "" || password == "" {
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"title": "Login - Weather App",
			"error": "Username and password are required",
		})
		return
	}

	// Get user by username
	user, err := h.userStore.GetUserByUsername(username)
	if err != nil {
		log.Printf("Login error: User not found: %v", err)
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{
			"title": "Login - Weather App",
			"error": "Invalid username or password",
		})
		return
	}

	// Validate password
	if !user.ValidatePassword(password) {
		log.Printf("Login error: Invalid password for user %s", username)
		c.HTML(http.StatusUnauthorized, "login.html", gin.H{
			"title": "Login - Weather App",
			"error": "Invalid username or password",
		})
		return
	}

	log.Printf("Login successful: username=%s, ID=%d", username, user.ID)

	// Set session
	session := sessions.Default(c)
	session.Set("user_id", user.ID)
	err = session.Save()
	if err != nil {
		log.Printf("Error saving session: %v", err)
		c.HTML(http.StatusInternalServerError, "login.html", gin.H{
			"title": "Login - Weather App",
			"error": "Failed to create session. Please try again.",
		})
		return
	}

	// Redirect to home
	c.Redirect(http.StatusFound, "/")
}

// Logout logs out a user
func (h *AuthHandler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	if userID != nil {
		log.Printf("Logout: user_id=%v", userID)
	}
	session.Clear()
	session.Save()
	c.Redirect(http.StatusFound, "/")
}

// GetProfile displays user profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	// Get user ID from context
	userID := c.MustGet("user_id").(int)
	log.Printf("Getting profile for user ID: %d", userID)

	// Get user from store
	user, err := h.userStore.GetUserByID(userID)
	if err != nil {
		log.Printf("Error getting user profile: %v", err)
		c.Redirect(http.StatusFound, "/logout")
		return
	}

	// Get success parameter if any
	success := c.Query("success")
	errorMsg := c.Query("error")

	// Render profile page
	c.HTML(http.StatusOK, "profile.html", gin.H{
		"title":   "Your Profile - Weather App",
		"user":    user,
		"success": success,
		"error":   errorMsg,
	})
}

// PostProfile updates user profile
func (h *AuthHandler) PostProfile(c *gin.Context) {
	// Get user ID from context
	userID := c.MustGet("user_id").(int)

	// Get user from store
	user, err := h.userStore.GetUserByID(userID)
	if err != nil {
		log.Printf("Error getting user for profile update: %v", err)
		c.Redirect(http.StatusFound, "/logout")
		return
	}

	// Get form data
	email := c.PostForm("email")
	homeCity := c.PostForm("home_city")
	notificationsEnabled := c.PostForm("notifications_enabled") == "on"
	alertThreshold := c.PostForm("alert_threshold")
	profilePhoto := c.PostForm("profile_photo")

	// Debug logging - add this
	log.Printf("Profile update - User ID: %d, Username: %s", user.ID, user.Username)
	log.Printf("Profile photo from form: '%s'", profilePhoto)

	// Update user fields
	user.Email = email
	user.HomeCity = homeCity
	user.NotificationsEnabled = notificationsEnabled
	user.AlertThreshold = alertThreshold

	// Update profile photo if provided
	if profilePhoto != "" {
		// Validate profile photo
		validPhotos := map[string]bool{
			"photo1.jpg":  true,
			"photo2.jpg":  true,
			"photo3.jpg":  true,
			"photo4.jpg":  true,
			"photo5.jpg":  true,
			"default.jpg": true,
		}

		if validPhotos[profilePhoto] {
			log.Printf("Setting profile photo to: %s", profilePhoto)
			user.ProfilePhoto = profilePhoto
		} else {
			log.Printf("Invalid profile photo selection: %s", profilePhoto)
		}
	}

	// If alert threshold is not set, default to "severe"
	if user.AlertThreshold == "" {
		user.AlertThreshold = "severe"
	}

	// Only update password if provided
	newPassword := c.PostForm("new_password")
	if newPassword != "" {
		// Verify current password
		currentPassword := c.PostForm("current_password")
		if !user.ValidatePassword(currentPassword) {
			c.HTML(http.StatusBadRequest, "profile.html", gin.H{
				"title": "Your Profile - Weather App",
				"user":  user,
				"error": "Current password is incorrect",
			})
			return
		}
		user.Password = newPassword
	}

	// Update timestamp
	user.UpdatedAt = time.Now()

	// Save changes - use UpdateUserProfile for more validation
	err = h.userStore.UpdateUserProfile(user)
	if err != nil {
		log.Printf("Error updating user profile: %v", err)
		c.HTML(http.StatusInternalServerError, "profile.html", gin.H{
			"title": "Your Profile - Weather App",
			"user":  user,
			"error": "Failed to update profile: " + err.Error(),
		})
		return
	}

	log.Printf("Profile successfully updated for user ID: %d", user.ID)
	log.Printf("Profile photo set to: %s", user.ProfilePhoto)

	// Redirect back to profile with success message
	c.Redirect(http.StatusSeeOther, "/profile?success=true")
}
