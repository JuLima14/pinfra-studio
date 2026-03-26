package middleware

import (
	"context"

	"github.com/JuLima14/pinfra-studio/internal/auth"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type contextKey string

const (
	UserContextKey contextKey = "user"
)

// AuthMiddleware validates JWT from Auth0 (same token as infra-platform).
// Reads from cookie "id_token" or Authorization Bearer header.
func AuthMiddleware(validator *auth.JWTValidator, db *gorm.DB, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var token string

		// Try Authorization header first
		authHeader := c.Get("Authorization")
		if authHeader != "" {
			t, err := auth.ExtractTokenFromHeader(authHeader)
			if err == nil {
				token = t
			}
		}

		// Fallback to cookie (shared with infra-platform on same domain)
		if token == "" {
			token = c.Cookies("id_token")
		}

		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: missing token",
			})
		}

		claims, err := validator.ValidateToken(token)
		if err != nil {
			logger.Warn("Invalid JWT token", zap.String("path", c.Path()), zap.Error(err))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: invalid token",
			})
		}

		// Look up user in shared DB by Auth0 ID
		var user auth.User
		if err := db.Where("auth0_id = ?", claims.Subject).First(&user).Error; err != nil {
			logger.Error("User not found in database", zap.String("auth0_id", claims.Subject), zap.Error(err))
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "User not found",
			})
		}

		// Store user in context
		ctx := context.WithValue(c.UserContext(), UserContextKey, &user)
		c.SetUserContext(ctx)

		return c.Next()
	}
}

// GetUser extracts the authenticated user from the request context.
func GetUser(c *fiber.Ctx) *auth.User {
	user, _ := c.UserContext().Value(UserContextKey).(*auth.User)
	return user
}
