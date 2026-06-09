package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthRequired reads X-User-ID and X-User-Email headers forwarded by nginx ingress
// after it validates the token via gocars-auth /auth/validate (auth-url).
//
// K8s ingress annotations required on gocars-api routes:
//   nginx.ingress.kubernetes.io/auth-url: http://gocars-auth/auth/validate
//   nginx.ingress.kubernetes.io/auth-response-headers: X-User-ID,X-User-Email
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetHeader("X-User-ID")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Set("email", c.GetHeader("X-User-Email"))

		c.Next()
	}
}
