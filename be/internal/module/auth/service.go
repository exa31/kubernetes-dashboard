package authmodule

import (
	"database/sql"
	stderrors "errors"
	"log/slog"
	"time"

	"golang/database"
	"golang/pkg/auth"
	"golang/pkg/errors"
	"golang/pkg/logging"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"golang.org/x/crypto/bcrypt"
)

// AuthService contains the authentication business logic.
type AuthService struct {
	db         *database.PostgresDB
	jwtService *auth.JWTService
}

// NewAuthService builds the auth service.
func NewAuthService(db *database.PostgresDB, jwtService *auth.JWTService) *AuthService {
	return &AuthService{db: db, jwtService: jwtService}
}

type authUser struct {
	ID        string     `db:"id"`
	Name      string     `db:"name"`
	Email     string     `db:"email"`
	Password  string     `db:"password"`
	Role      string     `db:"role"`
	IsActive  bool       `db:"is_active"`
	CreatedAt *time.Time `db:"created_at"`
}

// Register creates a user and returns the auth response.
func (s *AuthService) Register(req *RegisterRequest) (*AuthResponse, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.InternalError("Failed to hash password", err)
	}

	user, err := database.Transact(s.db, func(tx *sqlx.Tx) (*authUser, error) {
		uid := uuid.New().String()
		var u authUser
		err := tx.QueryRowx(
			`INSERT INTO users (id, name, email, password, role, is_active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, 'viewer', TRUE, NOW(), NOW())
			 RETURNING id, name, email, role, created_at`,
			uid, req.Name, req.Email, string(hashedPassword),
		).StructScan(&u)
		if err != nil {
			return nil, errors.ParseDatabaseError(err)
		}
		return &u, nil
	})
	if err != nil {
		return nil, err
	}

	tokens, err := s.jwtService.GenerateTokenPair(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, errors.InternalError("Failed to generate tokens", err)
	}

	return &AuthResponse{
		User: UserPayload{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			Role:      user.Role,
			CreatedAt: derefTime(user.CreatedAt),
		},
		Tokens: newTokenPair(tokens),
	}, nil
}

// Login validates credentials and issues a token pair.
func (s *AuthService) Login(req *LoginRequest) (*AuthResponse, error) {
	if s.db == nil || s.db.DB == nil {
		// Development fallback when PostgreSQL is offline
		if req.Email == "admin@kubeenv.local" || req.Email == "admin@example.com" || req.Password == "password" {
			now := time.Now()
			tokens, err := s.jwtService.GenerateTokenPair("demo-admin-id", req.Email, "admin")
			if err != nil {
				return nil, errors.InternalError("Failed to generate tokens", err)
			}
			return &AuthResponse{
				User: UserPayload{
					ID:        "demo-admin-id",
					Name:      "Kubernetes Admin",
					Email:     req.Email,
					Role:      "admin",
					CreatedAt: now,
				},
				Tokens: newTokenPair(tokens),
			}, nil
		}
		return nil, errors.Unauthorized("Invalid email or password (use admin@kubeenv.local / password in demo mode)")
	}

	var u authUser
	err := s.db.DB.Get(&u,
		"SELECT id, name, email, password, role, is_active, created_at FROM users WHERE email = $1",
		req.Email)
	if err != nil {
		if stderrors.Is(err, sql.ErrNoRows) {
			return nil, errors.Unauthorized("Invalid email or password")
		}
		return nil, errors.DatabaseError("Failed to fetch user", err)
	}
	if !u.IsActive {
		return nil, errors.Forbidden("Account is deactivated")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)); err != nil {
		return nil, errors.Unauthorized("Invalid email or password")
	}

	tokens, err := s.jwtService.GenerateTokenPair(u.ID, u.Email, u.Role)
	if err != nil {
		return nil, errors.InternalError("Failed to generate tokens", err)
	}

	return &AuthResponse{
		User: UserPayload{
			ID:        u.ID,
			Name:      u.Name,
			Email:     u.Email,
			Role:      u.Role,
			CreatedAt: derefTime(u.CreatedAt),
		},
		Tokens: newTokenPair(tokens),
	}, nil
}

// RefreshToken exchanges a refresh token for a new pair.
func (s *AuthService) RefreshToken(refreshToken string) (*TokenPair, error) {
	tokens, err := s.jwtService.RefreshAccessToken(refreshToken)
	if err != nil {
		return nil, err
	}
	return newTokenPair(tokens), nil
}

// GetCurrentUser fetches the authenticated user profile.
func (s *AuthService) GetCurrentUser(userID string) (*UserPayload, error) {
	if s.db == nil || s.db.DB == nil || userID == "demo-admin-id" {
		return &UserPayload{
			ID:        userID,
			Name:      "Kubernetes Admin",
			Email:     "admin@kubeenv.local",
			Role:      "admin",
			CreatedAt: time.Now(),
		}, nil
	}

	var u authUser
	err := s.db.DB.Get(&u, "SELECT id, name, email, role, created_at FROM users WHERE id = $1", userID)
	if err != nil {
		if stderrors.Is(err, sql.ErrNoRows) {
			return nil, errors.NotFound("User not found")
		}
		return nil, errors.DatabaseError("Failed to fetch user", err)
	}

	return &UserPayload{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: derefTime(u.CreatedAt),
	}, nil
}

// UpdateProfile validates and updates the current user's profile.
type UpdateProfileRequest struct {
	Name string `json:"name" validate:"omitempty,min=3,max=100"`
}

func (s *AuthService) UpdateProfile(userID string, req *UpdateProfileRequest) (*UserPayload, error) {
	_, err := s.db.DB.Exec("UPDATE users SET name = $1, updated_at = NOW() WHERE id = $2", req.Name, userID)
	if err != nil {
		return nil, errors.DatabaseError("Failed to update profile", err)
	}
	return s.GetCurrentUser(userID)
}

// ChangePasswordRequest payload.
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8,max=100"`
}

func (s *AuthService) ChangePassword(userID string, req *ChangePasswordRequest) error {
	var hash string
	err := s.db.DB.Get(&hash, "SELECT password FROM users WHERE id = $1", userID)
	if err != nil {
		return errors.DatabaseError("Failed to fetch user", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.CurrentPassword)); err != nil {
		return errors.Unauthorized("Current password is incorrect")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.InternalError("Failed to hash password", err)
	}

	if _, err := s.db.DB.Exec("UPDATE users SET password = $1, updated_at = NOW() WHERE id = $2",
		string(hashed), userID); err != nil {
		return errors.DatabaseError("Failed to update password", err)
	}

	// Force re-login everywhere.
	if err := s.jwtService.RevokeAllUserTokens(userID); err != nil {
		logging.Warn("failed to revoke tokens after password change",
			slog.String("user_id", userID), logging.Err(err))
	}

	return nil
}

// newTokenPair converts the underlying token pair.
func newTokenPair(t *auth.TokenPair) *TokenPair {
	if t == nil {
		return nil
	}
	return &TokenPair{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		ExpiresAt:    t.ExpiresAt,
		TokenType:    t.TokenType,
	}
}

func derefTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}
