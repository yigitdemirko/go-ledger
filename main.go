package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/yigit-demirko/go-ledger/internal/api"
	"github.com/yigit-demirko/go-ledger/internal/database"
)

func main() {
	// try to load settings from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	// connect to database (if this fails, we can't start)
	if err := database.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	
	// make sure we close database when done
	defer database.CloseDB()

	// create database tables if they don't exist yet
	if err := database.CreateTables(); err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	// create a new web server
	r := gin.Default()

	// set up all our API endpoints
	api.SetupRouter(r)

	// get the port to run on, default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// start the server
	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
} 