package models

import (
	"context"
	"time"

	"github.com/yigit-demirko/go-ledger/internal/database"
)

// what kind of money movements we track
type TransactionType string

const (
	TransactionTypeTransfer TransactionType = "TRANSFER" // when users send money to each other
	TransactionTypeDeposit TransactionType = "DEPOSIT"  // when money comes in
	TransactionTypeWithdraw TransactionType = "WITHDRAW" // when money goes out
)

// Transaction keeps track of money movements
type Transaction struct {
	ID              int64           `json:"id"`
	FromUserID      *int64          `json:"from_user_id"`      // who sent the money (can be null for deposits)
	ToUserID        *int64          `json:"to_user_id"`        // who got the money (can be null for withdrawals)
	Amount          float64         `json:"amount"`            // how much money moved
	TransactionType TransactionType `json:"transaction_type"`  // what kind of movement it was
	CreatedAt       time.Time       `json:"created_at"`        // when it happened
}

// CreateTransaction saves a new money movement in the database
func CreateTransaction(fromUserID, toUserID *int64, amount float64, transactionType TransactionType) (*Transaction, error) {
	var transaction Transaction
	err := database.GetPool().QueryRow(
		context.Background(),
		`INSERT INTO transactions (from_user_id, to_user_id, amount, transaction_type, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, from_user_id, to_user_id, amount, transaction_type, created_at`,
		fromUserID, toUserID, amount, transactionType, time.Now(),
	).Scan(
		&transaction.ID,
		&transaction.FromUserID,
		&transaction.ToUserID,
		&transaction.Amount,
		&transaction.TransactionType,
		&transaction.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &transaction, nil
}

// GetTransactionsByUserID finds all money movements for a user
func GetTransactionsByUserID(userID int64, limit, offset int) ([]Transaction, error) {
	rows, err := database.GetPool().Query(
		context.Background(),
		`SELECT id, from_user_id, to_user_id, amount, transaction_type, created_at
		FROM transactions 
		WHERE from_user_id = $1 OR to_user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var transaction Transaction
		err := rows.Scan(
			&transaction.ID,
			&transaction.FromUserID,
			&transaction.ToUserID,
			&transaction.Amount,
			&transaction.TransactionType,
			&transaction.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

// GetUserTransactionsInTimeRange finds money movements between two dates
func GetUserTransactionsInTimeRange(userID int64, startTime, endTime time.Time, limit, offset int) ([]Transaction, error) {
	rows, err := database.GetPool().Query(
		context.Background(),
		`SELECT id, from_user_id, to_user_id, amount, transaction_type, created_at
		FROM transactions 
		WHERE (from_user_id = $1 OR to_user_id = $1)
		AND created_at BETWEEN $2 AND $3
		ORDER BY created_at DESC
		LIMIT $4 OFFSET $5`,
		userID, startTime, endTime, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []Transaction
	for rows.Next() {
		var transaction Transaction
		err := rows.Scan(
			&transaction.ID,
			&transaction.FromUserID,
			&transaction.ToUserID,
			&transaction.Amount,
			&transaction.TransactionType,
			&transaction.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return transactions, nil
}

// GetBalanceAtTime calculates a user's balance at a specific point in time
func GetBalanceAtTime(userID int64, targetTime time.Time) (float64, error) {
	var balance float64
	err := database.GetPool().QueryRow(
		context.Background(),
		`WITH balance_changes AS (
			SELECT 
				CASE 
					WHEN from_user_id = $1 THEN -amount
					WHEN to_user_id = $1 THEN amount
				END as change
			FROM transactions
			WHERE (from_user_id = $1 OR to_user_id = $1)
			AND created_at <= $2
		)
		SELECT COALESCE(SUM(change), 0)
		FROM balance_changes`,
		userID, targetTime,
	).Scan(&balance)

	if err != nil {
		return 0, err
	}

	return balance, nil
}

// BalanceWithTimestamp represents a balance at a specific time
type BalanceWithTimestamp struct {
	Balance   float64   `json:"balance"`
	Timestamp time.Time `json:"timestamp"`
} 