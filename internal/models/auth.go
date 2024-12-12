package models

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/yigit-demirko/go-ledger/internal/database"
	"golang.org/x/crypto/bcrypt"
)

// what kind of user roles we have
type UserRole string

const (
	RoleUser  UserRole = "USER"  // regular users
	RoleAdmin UserRole = "ADMIN" // admins can do everything
)

// AuthUser is for login and permissions
type AuthUser struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"` // we never show the password hash
	Role         UserRole  `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateAuthUser makes a new user account
func CreateAuthUser(username, password string, role UserRole) (*AuthUser, error) {
	// make the password safe to store
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	var user AuthUser
	now := time.Now()

	// save the user in database
	err = database.GetPool().QueryRow(
		context.Background(),
		`INSERT INTO auth_users (username, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $4)
		RETURNING id, username, password_hash, role, created_at, updated_at`,
		username, string(hashedPassword), role, now,
	).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetAuthUserByUsername finds a user by their username
func GetAuthUserByUsername(username string) (*AuthUser, error) {
	var user AuthUser
	err := database.GetPool().QueryRow(
		context.Background(),
		`SELECT id, username, password_hash, role, created_at, updated_at
		FROM auth_users WHERE username = $1`,
		username,
	).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// ValidatePassword checks if the password is correct
func (u *AuthUser) ValidatePassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// UpdatePassword lets users change their password
func (u *AuthUser) UpdatePassword(newPassword string) error {
	// hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// save it in database
	err = database.GetPool().QueryRow(
		context.Background(),
		`UPDATE auth_users
		SET password_hash = $1, updated_at = $2
		WHERE id = $3
		RETURNING updated_at`,
		string(hashedPassword), time.Now(), u.ID,
	).Scan(&u.UpdatedAt)

	return err
}

// error messages we might need
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound      = errors.New("user not found")
) 