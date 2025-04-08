package models

import (
	"context"
	"log"
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

// InitializeUserBalance sets a user's initial balance
func InitializeUserBalance(userID int, amount float64) error {
	// get database connection
	db := database.GetPool()
	ctx := context.Background()

	// start transaction
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
			log.Printf("Error rolling back transaction: %v", err)
		}
	}()

	// update user balance
	_, err = tx.Exec(ctx, `
		UPDATE users 
		SET balance = $1, 
			updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2`,
		amount, userID)
	if err != nil {
		return err
	}

	// create initial transaction
	_, err = tx.Exec(ctx, `
		INSERT INTO transactions (
			from_user_id, 
			to_user_id, 
			amount,
			transaction_type,
			description,
			created_at
		) VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)`,
		nil, userID, amount, TransactionTypeDeposit, "Initial balance")
	if err != nil {
		return err
	}

	// commit transaction
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
} 