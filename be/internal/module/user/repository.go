package usermodule

import (
	"database/sql"
	"fmt"
)

// GetAll retrieves all users ordered by creation date.
func (r *userRepository) GetAll() ([]User, error) {
	var users []User
	query := "SELECT id, name, email, phone, role, is_active, created_at, updated_at FROM users ORDER BY created_at DESC"
	err := r.db.Select(&users, query)
	return users, err
}

// GetByID retrieves a user by ID.
func (r *userRepository) GetByID(id string) (*User, error) {
	var user User
	query := "SELECT id, name, email, phone, password, role, is_active, created_at, updated_at FROM users WHERE id = $1"
	err := r.db.Get(&user, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// GetByEmail retrieves a user by email.
func (r *userRepository) GetByEmail(email string) (*User, error) {
	var user User
	query := "SELECT id, name, email, phone, password, role, is_active, created_at, updated_at FROM users WHERE email = $1"
	err := r.db.Get(&user, query, email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// Create inserts a new user.
func (r *userRepository) Create(user *User) error {
	query := `
		INSERT INTO users (id, name, email, phone, password, role, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(query, user.ID, user.Name, user.Email, user.Phone, user.Password, user.Role, user.IsActive, user.CreatedAt, user.UpdatedAt)
	return err
}

// Update applies the given columns and returns the updated user.
func (r *userRepository) Update(id string, updates map[string]interface{}) (*User, error) {
	query := "UPDATE users SET updated_at = NOW()"
	args := []interface{}{id}

	if name, ok := updates["name"].(string); ok && name != "" {
		query += fmt.Sprintf(", name = $%d", len(args)+1)
		args = append(args, name)
	}
	if email, ok := updates["email"].(string); ok && email != "" {
		query += fmt.Sprintf(", email = $%d", len(args)+1)
		args = append(args, email)
	}
	if phone, ok := updates["phone"].(string); ok && phone != "" {
		query += fmt.Sprintf(", phone = $%d", len(args)+1)
		args = append(args, phone)
	}
	if role, ok := updates["role"].(string); ok && role != "" {
		query += fmt.Sprintf(", role = $%d", len(args)+1)
		args = append(args, role)
	}
	if isActive, ok := updates["is_active"].(bool); ok {
		query += fmt.Sprintf(", is_active = $%d", len(args)+1)
		args = append(args, isActive)
	}

	query += " WHERE id = $1 RETURNING id, name, email, phone, role, is_active, created_at, updated_at"

	var user User
	if err := r.db.QueryRowx(query, args...).StructScan(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdatePassword updates user password by ID.
func (r *userRepository) UpdatePassword(id string, hashedPassword string) error {
	result, err := r.db.Exec("UPDATE users SET password = $1, updated_at = NOW() WHERE id = $2", hashedPassword, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// Delete soft deletes a user.
func (r *userRepository) Delete(id string) error {
	result, err := r.db.Exec("UPDATE users SET is_active = false, updated_at = NOW() WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// HardDelete permanently deletes a user.
func (r *userRepository) HardDelete(id string) error {
	result, err := r.db.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ExistsByID checks whether a user exists.
func (r *userRepository) ExistsByID(id string) (bool, error) {
	var exists bool
	err := r.db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", id)
	return exists, err
}

// ExistsByEmail checks whether an email is taken.
func (r *userRepository) ExistsByEmail(email string) (bool, error) {
	var exists bool
	err := r.db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email)
	return exists, err
}
