package auth

import (
	"strings"

	"golang/pkg/auth"
	customErrors "golang/pkg/errors"

	"github.com/gofiber/fiber/v2"
)

// bearerScheme is the Authorization header scheme used everywhere in the app.
const bearerScheme = "Bearer"

// extractAccessToken gets token from cookie "access_token", query "token", or Authorization header
func extractAccessToken(c *fiber.Ctx) string {
	if cookie := c.Cookies("access_token"); cookie != "" {
		return cookie
	}
	if queryToken := c.Query("token"); queryToken != "" {
		return queryToken
	}
	authHeader := c.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == bearerScheme {
			return parts[1]
		}
	}
	return ""
}

// extractRefreshToken gets token from cookie "refresh_token" or Authorization header
func extractRefreshToken(c *fiber.Ctx) string {
	if cookie := c.Cookies("refresh_token"); cookie != "" {
		return cookie
	}
	authHeader := c.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == bearerScheme {
			return parts[1]
		}
	}
	return ""
}

// AuthMiddleware creates a JWT authentication middleware
func AuthMiddleware(jwtService *auth.JWTService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString := extractAccessToken(c)
		if tokenString == "" {
			return customErrors.Unauthorized("Authorization token is required")
		}

		// Validate access token
		claims, err := jwtService.ValidateToken(tokenString, auth.AccessToken)
		if err != nil {
			return err
		}

		// Check if user tokens are revoked
		if jwtService.IsUserRevoked(claims.UserID) {
			return customErrors.Unauthorized("Token has been revoked")
		}

		// Store claims in context
		c.Locals("user_id", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("role", claims.Role)
		c.Locals("token_id", claims.TokenID)
		c.Locals("claims", claims)

		return c.Next()
	}
}

// OptionalAuthMiddleware is similar to AuthMiddleware but doesn't fail if token is missing
func OptionalAuthMiddleware(jwtService *auth.JWTService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString := extractAccessToken(c)
		if tokenString == "" {
			return c.Next()
		}

		claims, err := jwtService.ValidateToken(tokenString, auth.AccessToken)
		if err != nil {
			return c.Next()
		}

		if !jwtService.IsUserRevoked(claims.UserID) {
			c.Locals("user_id", claims.UserID)
			c.Locals("email", claims.Email)
			c.Locals("role", claims.Role)
			c.Locals("token_id", claims.TokenID)
			c.Locals("claims", claims)
		}

		return c.Next()
	}
}

// RefreshTokenMiddleware validates refresh token
func RefreshTokenMiddleware(jwtService *auth.JWTService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString := extractRefreshToken(c)
		if tokenString == "" {
			return customErrors.Unauthorized("Refresh token is required")
		}

		// Validate refresh token
		claims, err := jwtService.ValidateToken(tokenString, auth.RefreshToken)
		if err != nil {
			return err
		}

		if jwtService.IsUserRevoked(claims.UserID) {
			return customErrors.Unauthorized("Token has been revoked")
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("email", claims.Email)
		c.Locals("token_id", claims.TokenID)
		c.Locals("claims", claims)
		c.Locals("refresh_token", tokenString)

		return c.Next()
	}
}

// GetUserID extracts user ID from context
func GetUserID(c *fiber.Ctx) (string, error) {
	userID, ok := c.Locals("user_id").(string)
	if !ok || userID == "" {
		return "", customErrors.Unauthorized("User not authenticated")
	}
	return userID, nil
}

// GetEmail extracts email from context
func GetEmail(c *fiber.Ctx) (string, error) {
	email, ok := c.Locals("email").(string)
	if !ok || email == "" {
		return "", customErrors.Unauthorized("User not authenticated")
	}
	return email, nil
}

// GetClaims extracts JWT claims from context
func GetClaims(c *fiber.Ctx) (*auth.JWTClaims, error) {
	claims, ok := c.Locals("claims").(*auth.JWTClaims)
	if !ok {
		return nil, customErrors.Unauthorized("User not authenticated")
	}
	return claims, nil
}

// MustGetUserID extracts user ID or panics (use with recover middleware)
func MustGetUserID(c *fiber.Ctx) string {
	userID, err := GetUserID(c)
	if err != nil {
		panic(err)
	}
	return userID
}

// GetRole extracts user role from context
func GetRole(c *fiber.Ctx) (string, error) {
	role, ok := c.Locals("role").(string)
	if !ok || role == "" {
		return "", customErrors.Unauthorized("Role not found in token")
	}
	return role, nil
}

// RequireRole creates an RBAC middleware ensuring the authenticated user has one of allowed roles.
func RequireRole(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		role, ok := c.Locals("role").(string)
		if !ok || role == "" {
			return customErrors.Forbidden("Access denied: no role assigned")
		}

		for _, allowed := range allowedRoles {
			if strings.EqualFold(role, allowed) {
				return c.Next()
			}
		}

		return customErrors.Forbidden("Access denied: insufficient permissions")
	}
}
