package authmodule

import (
	"strings"

	"golang/pkg/errors"
	"golang/pkg/logging"
	"golang/pkg/response"
	"golang/pkg/validation"

	"github.com/gofiber/fiber/v2"
)

// AuthHandler exposes the auth endpoints.
type AuthHandler struct {
	service *AuthService
}

// NewAuthHandler creates the auth HTTP handler.
func NewAuthHandler(service *AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// Register handles POST /api/v1/auth/register.
func (h *AuthHandler) Register() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req RegisterRequest
		if err := validation.Default.BindAndValidate(c, &req); err != nil {
			return err
		}

		res, err := h.service.Register(&req)
		if err != nil {
			return err
		}

		logging.LoggerFromFiber(c).Info("user registered", "email", req.Email)
		return response.CreatedResponse(c, res, "User registered successfully")
	}
}

// Login handles POST /api/v1/auth/login.
func (h *AuthHandler) Login() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req LoginRequest
		if err := validation.Default.BindAndValidate(c, &req); err != nil {
			return err
		}

		res, err := h.service.Login(&req)
		if err != nil {
			return err
		}

		setAuthCookies(c, res.Tokens.AccessToken, res.Tokens.RefreshToken)
		logging.LoggerFromFiber(c).Info("user logged in", "user_id", res.User.ID)
		return response.SuccessResponse(c, res, "Login successful")
	}
}

// RefreshToken handles POST /api/v1/auth/refresh.
func (h *AuthHandler) RefreshToken() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req RefreshTokenRequest
		_ = c.BodyParser(&req)

		refreshToken := extractRefreshToken(c, req.RefreshToken)
		if refreshToken == "" {
			return errors.Unauthorized("Refresh token is required")
		}

		tokens, err := h.service.RefreshToken(refreshToken)
		if err != nil {
			return err
		}

		setAuthCookies(c, tokens.AccessToken, tokens.RefreshToken)
		return response.SuccessResponse(c, tokens, "Token refreshed successfully")
	}
}

// Logout handles POST /api/v1/auth/logout.
func (h *AuthHandler) Logout() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return errors.Unauthorized("User not authenticated")
		}

		token := extractBearer(c)
		if token != "" {
			if err := h.service.jwtService.RevokeToken(token, "access"); err != nil {
				logging.LoggerFromFiber(c).Warn("failed to revoke token on logout", logging.Err(err))
			}
		}

		clearAuthCookies(c)
		return response.SuccessResponse(c, fiber.Map{"user_id": userID}, "Logout successful")
	}
}

// LogoutAll handles POST /api/v1/auth/logout-all.
func (h *AuthHandler) LogoutAll() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return errors.Unauthorized("User not authenticated")
		}

		if err := h.service.jwtService.RevokeAllUserTokens(userID); err != nil {
			return err
		}

		clearAuthCookies(c)
		return response.SuccessResponse(c, fiber.Map{"user_id": userID}, "Logout successful")
	}
}

// GetProfile handles GET /api/v1/auth/profile.
func (h *AuthHandler) GetProfile() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return errors.Unauthorized("User not authenticated")
		}

		user, err := h.service.GetCurrentUser(userID)
		if err != nil {
			return err
		}

		return response.SuccessResponse(c, user, "Profile retrieved successfully")
	}
}

// UpdateProfile handles PUT /api/v1/auth/profile.
func (h *AuthHandler) UpdateProfile() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return errors.Unauthorized("User not authenticated")
		}

		var req UpdateProfileRequest
		if err := validation.Default.BindAndValidate(c, &req); err != nil {
			return err
		}

		user, err := h.service.UpdateProfile(userID, &req)
		if err != nil {
			return err
		}

		return response.SuccessResponse(c, user, "Profile updated successfully")
	}
}

// ChangePassword handles POST /api/v1/auth/change-password.
func (h *AuthHandler) ChangePassword() fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(string)
		if !ok || userID == "" {
			return errors.Unauthorized("User not authenticated")
		}

		var req ChangePasswordRequest
		if err := validation.Default.BindAndValidate(c, &req); err != nil {
			return err
		}

		if err := h.service.ChangePassword(userID, &req); err != nil {
			return err
		}

		clearAuthCookies(c)
		return response.SuccessResponse(c, fiber.Map{}, "Password changed successfully. Please login again.")
	}
}

// extractBearer strips the "Bearer " prefix from Authorization header or checks access_token cookie.
func extractBearer(c *fiber.Ctx) string {
	if cookie := c.Cookies("access_token"); cookie != "" {
		return cookie
	}
	header := c.Get("Authorization")
	if len(header) > 7 && header[:7] == "Bearer " {
		return header[7:]
	}
	return ""
}

// extractRefreshToken extracts refresh token from body, cookies, Authorization header, X-Refresh-Token, or locals.
func extractRefreshToken(c *fiber.Ctx, bodyToken string) string {
	if strings.TrimSpace(bodyToken) != "" {
		return strings.TrimSpace(bodyToken)
	}
	if cookie := c.Cookies("refresh_token"); strings.TrimSpace(cookie) != "" {
		return strings.TrimSpace(cookie)
	}
	if token, ok := c.Locals("refresh_token").(string); ok && strings.TrimSpace(token) != "" {
		return strings.TrimSpace(token)
	}
	if xToken := c.Get("X-Refresh-Token"); strings.TrimSpace(xToken) != "" {
		return strings.TrimSpace(xToken)
	}
	authHeader := c.Get("Authorization")
	if len(authHeader) > 7 && strings.EqualFold(authHeader[:7], "Bearer ") {
		return strings.TrimSpace(authHeader[7:])
	} else if authHeader != "" && !strings.Contains(authHeader, " ") {
		return strings.TrimSpace(authHeader)
	}
	if qToken := c.Query("refresh_token"); strings.TrimSpace(qToken) != "" {
		return strings.TrimSpace(qToken)
	}
	return ""
}

// setAuthCookies sets httpOnly cookies for access and refresh tokens.
func setAuthCookies(c *fiber.Ctx, accessToken, refreshToken string) {
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		HTTPOnly: true,
		SameSite: "Lax",
		MaxAge:   15 * 60, // 15 minutes
	})
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		HTTPOnly: true,
		SameSite: "Lax",
		MaxAge:   7 * 24 * 3600, // 7 days
	})
}

// clearAuthCookies clears the access and refresh token cookies.
func clearAuthCookies(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		HTTPOnly: true,
		SameSite: "Lax",
		MaxAge:   -1,
	})
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HTTPOnly: true,
		SameSite: "Lax",
		MaxAge:   -1,
	})
}
