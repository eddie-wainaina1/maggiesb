package main

import (
	"log"
	"os"
	"time"

	"github.com/eddie-wainaina1/maggiesb/internal/auth"
	"github.com/eddie-wainaina1/maggiesb/internal/database"
	"github.com/eddie-wainaina1/maggiesb/internal/handlers"
	"github.com/eddie-wainaina1/maggiesb/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load env
	_ = godotenv.Load()

	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatalf("MONGODB_URI is required. Set it in the environment or in a .env file (see .env.example)")
	}

	// Optionally override DB name
	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		database.SetDBName(dbName)
	}

	// Initialize MongoDB
	if err := database.InitMongo(mongoURI); err != nil {
		log.Fatalf("Failed to initialize MongoDB: %v", err)
	}
	defer database.DisconnectMongo()

	// Create necessary indexes
	if err := database.CreateIndexes(); err != nil {
		log.Fatalf("Failed to create indexes: %v", err)
	}

	// Initialize DI repositories
	handlers.InitDependencies()

	auth.StartTokenCleanupRoutine(1 * time.Hour)

	// Initialize M-Pesa client (optional, only if credentials are provided)
	if err := handlers.InitMpesaClient(); err != nil {
		log.Printf("Warning: M-Pesa client not initialized: %v", err)
	}

	router := gin.Default()
	public := router.Group("/api/v1/auth")
	{
		public.POST("/register", handlers.Register)
		public.POST("/login", handlers.Login)
	}
	// Public product routes
	products := router.Group("/api/v1/products")
	{
		products.GET("", handlers.ListProducts)
		products.GET("/search", handlers.SearchProducts)
		products.GET("/:id", handlers.GetProduct)
	}

	// Protected routes
	protected := router.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/profile", handlers.GetProfile)
		protected.POST("/logout", handlers.Logout)

		// Orders (user)
		protected.POST("/orders", handlers.CreateOrder)
		protected.GET("/orders", handlers.ListOrders)
		protected.GET("/orders/:id", handlers.GetOrder)

		// Invoices (user)
		protected.GET("/invoices/:id", handlers.GetInvoice)
		protected.GET("/orders/:id/invoice", handlers.GetInvoiceByOrder)

		// Payments (user)
		protected.POST("/orders/:id/pay", handlers.InitiateMpesaPayment)
		protected.GET("/payments/:id/status", handlers.GetPaymentStatus)
	}

	// Admin product routes (protected + admin role)
	adminProducts := router.Group("/api/v1/admin/products")
	adminProducts.Use(middleware.AuthMiddleware(), middleware.RequireRole("admin"))
	{
		adminProducts.POST("", handlers.CreateProduct)
		adminProducts.PUT("/:id", handlers.UpdateProduct)
		adminProducts.DELETE("/:id", handlers.DeleteProduct)
	}

	// Admin order routes (protected + admin role)
	adminOrders := router.Group("/api/v1/admin/orders")
	adminOrders.Use(middleware.AuthMiddleware(), middleware.RequireRole("admin"))
	{
		adminOrders.GET("", handlers.AdminListOrders)
		adminOrders.PUT("/:id/status", handlers.AdminUpdateOrderStatus)
	}

	// Admin invoice routes (protected + admin role)
	adminInvoices := router.Group("/api/v1/admin/invoices")
	adminInvoices.Use(middleware.AuthMiddleware(), middleware.RequireRole("admin"))
	{
		adminInvoices.GET("", handlers.AdminListInvoices)
		adminInvoices.PUT("/:id/payment", handlers.AdminRecordPayment)
		adminInvoices.PUT("/:id/reverse", handlers.AdminReverseInvoice)
	}

	// M-Pesa callback route (public)
	router.POST("/api/v1/mpesa/callback", handlers.HandleMpesaCallback)

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	router.Run(":" + port)
}