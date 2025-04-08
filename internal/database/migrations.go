package database

import "context"

// CreateTables creates all necessary database tables
func CreateTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS auth_users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			role VARCHAR(50) NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			balance DECIMAL(15,2) NOT NULL DEFAULT 0.00,
			auth_user_id INTEGER REFERENCES auth_users(id),
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS transactions (
			id SERIAL PRIMARY KEY,
			from_user_id INTEGER REFERENCES users(id),
			to_user_id INTEGER REFERENCES users(id),
			amount DECIMAL(15,2) NOT NULL,
			transaction_type VARCHAR(50) NOT NULL,
			description TEXT,
			created_at TIMESTAMP NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(from_user_id, to_user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions(created_at)`,
	}

	for _, query := range queries {
		_, err := GetPool().Exec(context.Background(), query)
		if err != nil {
			return err
		}
	}

	return nil
} 