package database

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// we need these variables in different parts of the code
var (
	// this is where we store database connections
	pool *pgxpool.Pool
	// we use this for database operations
	ctx = context.Background()
)

// Config has all the database settings we need
type Config struct {
	Host     string // where the database is running
	Port     int    // which port to connect to
	User     string // database username
	Password string // database password
	DBName   string // what we named our database
	SSLMode  string // if we want secure connection
	MaxConns int32  // how many connections we can have
	MinConns int32  // minimum connections to keep ready
}

// LoadConfig gets database settings from environment variables
// we keep passwords and stuff in environment variables to be safe
func LoadConfig() (*Config, error) {
	// convert port from string to number
	port, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %v", err)
	}

	// set max connections, default is 100
	maxConns := int32(100)
	if maxStr := os.Getenv("DB_MAX_CONNECTIONS"); maxStr != "" {
		if max, err := strconv.Atoi(maxStr); err == nil {
			maxConns = int32(max)
		}
	}

	// set min connections, default is 10
	minConns := int32(10)
	if minStr := os.Getenv("DB_MIN_CONNECTIONS"); minStr != "" {
		if min, err := strconv.Atoi(minStr); err == nil {
			minConns = int32(min)
		}
	}

	// pack everything into our config
	return &Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     port,
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSL_MODE"),
		MaxConns: maxConns,
		MinConns: minConns,
	}, nil
}

// InitDB connects to the database when the program starts
func InitDB() error {
	// load our settings
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	// make the connection string that tells database how to connect
	connString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s pool_max_conns=%d pool_min_conns=%d",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.DBName,
		config.SSLMode,
		config.MaxConns,
		config.MinConns,
	)

	// set up connection settings
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return fmt.Errorf("unable to parse pool config: %v", err)
	}

	// connect to database
	pool, err = pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return fmt.Errorf("unable to create connection pool: %v", err)
	}

	// check if connection works
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("unable to ping database: %v", err)
	}

	return nil
}

// GetPool gives us the database connection when we need it
func GetPool() *pgxpool.Pool {
	return pool
}

// CloseDB disconnects from database when program ends
func CloseDB() {
	if pool != nil {
		pool.Close()
	}
}

// RunInTransaction makes sure database operations happen together
// if something fails, everything gets undone
func RunInTransaction(fn func(pgx.Tx) error) error {
	// get a connection to use
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %v", err)
	}
	defer conn.Release()

	// start the transaction
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	// if something crashes, undo everything
	defer func() {
		if p := recover(); p != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				fmt.Printf("Error rolling back transaction after panic: %v\n", rbErr)
			}
			panic(p)
		}
	}()

	// run the database operations
	if err := fn(tx); err != nil {
		// if there's an error, undo everything
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("error rolling back transaction: %v (original error: %v)", rbErr, err)
		}
		return err
	}

	// if everything worked, save all changes
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	return nil
} 