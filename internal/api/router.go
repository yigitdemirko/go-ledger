package api

import (
	"github.com/gin-gonic/gin"
	"github.com/yigit-demirko/go-ledger/internal/middleware"
	"github.com/yigit-demirko/go-ledger/internal/models"
)

// SetupRouter tells the server what to do when users visit different URLs
func SetupRouter(r *gin.Engine) {
	r.GET("/health", HealthCheck)

	// group all our URLs under /api/v1
	v1 := r.Group("/api/v1")
	{
		// these URLs don't need login
		auth := v1.Group("/auth")
		{
			auth.POST("/register", Register)
			auth.POST("/login", Login)
		}

		// these URLs need login to use
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			// stuff about users
			users := protected.Group("/users")
			{
				// only admins can list all users
				users.GET("", middleware.RequireRole(models.RoleAdmin), GetAllUsers)

				// users can only see their own info (or admins can see anyone)
				users.GET("/:id", middleware.RequireOwnershipOrAdmin(), GetUser)
				users.GET("/:id/transactions", middleware.RequireOwnershipOrAdmin(), GetUserTransactions)
				users.GET("/:id/balance/historical", middleware.RequireOwnershipOrAdmin(), GetHistoricalBalance)

				// anyone logged in can change their password
				users.POST("/change-password", ChangePassword)

				// only admins can initialize balance
				users.POST("/:id/initialize-balance", middleware.RequireRole(models.RoleAdmin), InitializeBalance)
			}

			// anyone logged in can send money
			v1.POST("/transfer", TransferCredits)
		}
	}
} 