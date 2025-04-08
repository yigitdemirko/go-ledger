package api

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yigit-demirko/go-ledger/internal/auth"
	"github.com/yigit-demirko/go-ledger/internal/models"
)

// some default values we use
const (
	defaultLimit  = 10 // how many items to show per page
	defaultOffset = 0  // start from the beginning
)

// what we need to make a new account
type RegisterRequest struct {
	Username string         `json:"username" binding:"required"`
	Password string         `json:"password" binding:"required,min=6"`
	Name     string         `json:"name" binding:"required"`
	Role     models.UserRole `json:"role,omitempty"`
}

// what we need to log in
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// what we need to send money
type TransferRequest struct {
	FromUserID int64   `json:"from_user_id" binding:"required"`
	ToUserID   int64   `json:"to_user_id" binding:"required"`
	Amount     float64 `json:"amount" binding:"required,gt=0"`
}

// what we need to see transaction history
type TransactionHistoryRequest struct {
	StartTime string `form:"start_time"`
	EndTime   string `form:"end_time"`
	Limit     int    `form:"limit"`
	Offset    int    `form:"offset"`
}

// what we need to check old balance
type HistoricalBalanceRequest struct {
	Timestamp string `form:"timestamp" binding:"required"`
}

// Register makes a new user account
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// if no role picked, make them a regular user
	if req.Role == "" {
		req.Role = models.RoleUser
	}

	// make sure role is valid
	if req.Role != models.RoleUser && req.Role != models.RoleAdmin {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		return
	}

	// create login account
	authUser, err := models.CreateAuthUser(req.Username, req.Password, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// create user profile
	user, err := models.CreateUser(req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ledger user"})
		return
	}

	// make a login token
	token, err := auth.GenerateToken(authUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"name":     user.Name,
			"username": authUser.Username,
			"role":     authUser.Role,
		},
	})
}

// Login checks password and gives a token
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// find the user
	authUser, err := models.GetAuthUserByUsername(req.Username)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	// check if password is right
	if !authUser.ValidatePassword(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// make a login token
	token, err := auth.GenerateToken(authUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":       authUser.ID,
			"username": authUser.Username,
			"role":     authUser.Role,
		},
	})
}

// GetUser shows user info
func GetUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := models.GetUserByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetAllUsers lists all users (admin only)
func GetAllUsers(c *gin.Context) {
	users, err := models.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get users"})
		return
	}

	c.JSON(http.StatusOK, users)
}

// TransferCredits moves money between users
func TransferCredits(c *gin.Context) {
	var req TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// check if sender has enough money
	fromUser, err := models.GetUserWithBalance(req.FromUserID, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get source user"})
		return
	}
	if fromUser == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Source user not found or insufficient balance"})
		return
	}

	// check if receiver exists
	toUser, err := models.GetUserByID(req.ToUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get destination user"})
		return
	}
	if toUser == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Destination user not found"})
		return
	}

	// take money from sender
	if err := fromUser.UpdateBalance(-req.Amount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update source user balance"})
		return
	}

	// give money to receiver
	if err := toUser.UpdateBalance(req.Amount); err != nil {
		// if this fails, give money back to sender
		_ = fromUser.UpdateBalance(req.Amount)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update destination user balance"})
		return
	}

	// save the transfer in history
	fromUserID := req.FromUserID
	toUserID := req.ToUserID
	transaction, err := models.CreateTransaction(&fromUserID, &toUserID, req.Amount, models.TransactionTypeTransfer)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "Transfer successful but failed to log transaction",
			"from_user": gin.H{
				"id":      fromUser.ID,
				"balance": fromUser.Balance,
			},
			"to_user": gin.H{
				"id":      toUser.ID,
				"balance": toUser.Balance,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Transfer successful",
		"from_user": gin.H{
			"id":      fromUser.ID,
			"balance": fromUser.Balance,
		},
		"to_user": gin.H{
			"id":      toUser.ID,
			"balance": toUser.Balance,
		},
		"transaction": transaction,
	})
}

// GetUserTransactions shows money movement history
func GetUserTransactions(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req TransactionHistoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// use default values if not specified
	if req.Limit <= 0 {
		req.Limit = defaultLimit
	}
	if req.Offset < 0 {
		req.Offset = defaultOffset
	}

	var transactions []models.Transaction
	if req.StartTime != "" && req.EndTime != "" {
		// if dates given, find transactions between those dates
		startTime, err := time.Parse(time.RFC3339, req.StartTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_time format"})
			return
		}

		endTime, err := time.Parse(time.RFC3339, req.EndTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_time format"})
			return
		}

		transactions, err = models.GetUserTransactionsInTimeRange(userID, startTime, endTime, req.Limit, req.Offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get transactions"})
			return
		}
	} else {
		// if no dates, just get recent transactions
		transactions, err = models.GetTransactionsByUserID(userID, req.Limit, req.Offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get transactions"})
			return
		}
	}

	c.JSON(http.StatusOK, transactions)
}

func GetHistoricalBalance(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req HistoricalBalanceRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	timestamp, err := time.Parse(time.RFC3339, req.Timestamp)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid timestamp format"})
		return
	}

	balance, err := models.GetBalanceAtTime(userID, timestamp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get historical balance"})
		return
	}

	c.JSON(http.StatusOK, models.BalanceWithTimestamp{
		Balance:   balance,
		Timestamp: timestamp,
	})
}

func ChangePassword(c *gin.Context) {
	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword    string `json:"new_password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claims, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userClaims := claims.(*auth.Claims)

	authUser, err := models.GetAuthUserByUsername(userClaims.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}

	if !authUser.ValidatePassword(req.CurrentPassword) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid current password"})
		return
	}

	if err := authUser.UpdatePassword(req.NewPassword); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

// HealthCheck tells us if the server is running
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

// InitializeBalance lets admins set a user's initial balance
func InitializeBalance(c *gin.Context) {
	// get user ID from URL
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// get amount from request body
	var req struct {
		Amount float64 `json:"amount" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// initialize balance
	if err := models.InitializeUserBalance(userID, req.Amount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Balance initialized successfully"})
}
