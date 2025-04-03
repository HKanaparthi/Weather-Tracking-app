package middleware

import (
	"log"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// AuthRequired ensures that a user is logged in
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")
		path := c.Request.URL.Path

		log.Printf("AuthRequired: Path=%s, UserID=%v", path, userID)

		if userID == nil {
			// If we're trying to access the login page, signup page, or historical page, continue
			if c.Request.URL.Path == "/login" ||
				c.Request.URL.Path == "/signup" ||
				c.Request.URL.Path == "/historical-comparison" ||
				c.Request.URL.Path == "/historical-comparison.html" {
				log.Printf("Allowing access to public page: %s", path)
				c.Next()
				return
			}
			// Otherwise redirect to login
			log.Printf("No user_id found, redirecting to login")
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		// Set user ID for handlers to use
		log.Printf("User authenticated (ID=%v), proceeding to: %s", userID, path)
		c.Set("user_id", userID)
		c.Next()
	}
}

// RedirectIfLoggedIn redirects to home if user is already logged in
func RedirectIfLoggedIn() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userID := session.Get("user_id")
		path := c.Request.URL.Path

		log.Printf("RedirectIfLoggedIn: Path=%s, UserID=%v", path, userID)

		if userID != nil {
			log.Printf("User already logged in (ID=%v), redirecting to home", userID)
			c.Redirect(http.StatusFound, "/")
			c.Abort()
			return
		}
		log.Printf("No user_id found, allowing access to: %s", path)
		c.Next()
	}
}
