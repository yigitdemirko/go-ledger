package models

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/yigit-demirko/go-ledger/internal/database"
)

// User holds info about each user and their money
type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateUser adds a new user to database
func CreateUser(name string) (*User, error) {
	var user User
	err := database.GetPool().QueryRow(
		context.Background(),
		`INSERT INTO users (name, balance, created_at, updated_at)
		VALUES ($1, $2, $3, $3)
		RETURNING id, name, balance, created_at, updated_at`,
		name, 0.0, time.Now(),
	).Scan(&user.ID, &user.Name, &user.Balance, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID finds a user using their ID
func GetUserByID(id int64) (*User, error) {
	var user User
	err := database.GetPool().QueryRow(
		context.Background(),
		`SELECT id, name, balance, created_at, updated_at
		FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Name, &user.Balance, &user.CreatedAt, &user.UpdatedAt)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetAllUsers gets a list of all users
func GetAllUsers() ([]User, error) {
	rows, err := database.GetPool().Query(
		context.Background(),
		`SELECT id, name, balance, created_at, updated_at
		FROM users 
		ORDER BY id`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Name, &user.Balance, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// UpdateBalance changes how much money a user has
func (u *User) UpdateBalance(amount float64) error {
	return database.RunInTransaction(func(tx pgx.Tx) error {
		var updatedBalance float64
		err := tx.QueryRow(
			context.Background(),
			`UPDATE users
			SET balance = balance + $1, updated_at = $2
			WHERE id = $3 AND balance + $1 >= 0
			RETURNING balance`,
			amount, time.Now(), u.ID,
		).Scan(&updatedBalance)

		if err != nil {
			return err
		}

		u.Balance = updatedBalance
		return nil
	})
}

// GetUserWithBalance finds a user and checks if they have enough money
func GetUserWithBalance(id int64, requiredBalance float64) (*User, error) {
	var user User
	err := database.GetPool().QueryRow(
		context.Background(),
		`SELECT id, name, balance, created_at, updated_at
		FROM users 
		WHERE id = $1 AND balance >= $2
		FOR UPDATE`,
		id, requiredBalance,
	).Scan(&user.ID, &user.Name, &user.Balance, &user.CreatedAt, &user.UpdatedAt)

	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
} 