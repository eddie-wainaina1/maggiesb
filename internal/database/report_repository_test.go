package database

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/eddie-wainaina1/maggiesb/internal/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func TestReportRepository_ValidateDateFormats(t *testing.T) {
	// Test date parsing logic without requiring MongoDB initialization
	tests := []struct {
		name      string
		date      string
		wantErr   bool
	}{
		{
			name:    "valid date",
			date:    "2026-02-01",
			wantErr: false,
		},
		{
			name:    "invalid date format 1",
			date:    "02-01-2026",
			wantErr: true,
		},
		{
			name:    "invalid date format 2",
			date:    "2026/02/01",
			wantErr: true,
		},
		{
			name:    "invalid date",
			date:    "invalid",
			wantErr: true,
		},
		{
			name:    "empty date",
			date:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := time.Parse("2006-01-02", tt.date)
			if (err != nil) != tt.wantErr {
				t.Errorf("date parse error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Integration tests - requires MONGO_TEST_URI environment variable
func TestReportRepository_GetSummaryReport_Integration(t *testing.T) {
	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		t.Skip("MONGO_TEST_URI not set; skipping report repository integration tests")
	}

	if err := InitMongo(uri); err != nil {
		t.Fatalf("InitMongo error: %v", err)
	}
	defer DisconnectMongo()

	repo := NewReportRepository()
	ctx := context.Background()

	// Cleanup
	ordersCol := GetCollection(DBName, OrdersCollectionName)
	paymentsCol := GetCollection(DBName, PaymentRecordsCollectionName)
	reversalsCol := GetCollection(DBName, ReversalsCollectionName)

	ordersCol.DeleteMany(ctx, bson.M{})
	paymentsCol.DeleteMany(ctx, bson.M{})
	reversalsCol.DeleteMany(ctx, bson.M{})

	// Insert test data
	now := time.Now()
	dateStr := now.Format("2006-01-02")

	orders := []interface{}{
		&models.Order{
			ID:        "order-1",
			Status:    models.OrderStatusComplete,
			Cost:      1000.0,
			Discount:  100.0,
			TotalCost: 900.0,
			UserID:    "user-1",
			Phone:     "1234567890",
			CreatedAt: now,
			UpdatedAt: now,
		},
		&models.Order{
			ID:        "order-2",
			Status:    models.OrderStatusCancelled,
			Cost:      500.0,
			Discount:  50.0,
			TotalCost: 450.0,
			UserID:    "user-2",
			Phone:     "0987654321",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
	_, err := ordersCol.InsertMany(ctx, orders)
	assert.NoError(t, err)

	// Insert payments
	payments := []interface{}{
		&models.PaymentRecord{
			ID:        "payment-1",
			InvoiceID: "inv-1",
			Amount:    900.0,
			Status:    "completed",
			CreatedAt: now.Format("2006-01-02 15:04:05"),
		},
	}
	_, err = paymentsCol.InsertMany(ctx, payments)
	assert.NoError(t, err)

	// Insert reversals
	reversals := []interface{}{
		&models.ReversalRecord{
			ID:        "reversal-1",
			InvoiceID: "inv-1",
			Amount:    50.0,
			Date:      dateStr,
			CreatedAt: now,
		},
	}
	_, err = reversalsCol.InsertMany(ctx, reversals)
	assert.NoError(t, err)

	// Get summary report
	startDate := now.AddDate(0, 0, -1).Format("2006-01-02")
	endDate := now.AddDate(0, 0, 1).Format("2006-01-02")

	report, err := repo.GetSummaryReport(ctx, startDate, endDate)
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, startDate, report.DateRange.StartDate)
	assert.Equal(t, endDate, report.DateRange.EndDate)
	assert.Greater(t, report.TotalOrders, 0)
	assert.Greater(t, report.TotalSalesAmount, 0.0)
	assert.Greater(t, report.TotalDiscountsGiven, 0.0)

	// Cleanup
	ordersCol.DeleteMany(ctx, bson.M{})
	paymentsCol.DeleteMany(ctx, bson.M{})
	reversalsCol.DeleteMany(ctx, bson.M{})
}

func TestReportRepository_GetDailyBreakdown_Integration(t *testing.T) {
	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		t.Skip("MONGO_TEST_URI not set; skipping report repository integration tests")
	}

	if err := InitMongo(uri); err != nil {
		t.Fatalf("InitMongo error: %v", err)
	}
	defer DisconnectMongo()

	repo := NewReportRepository()
	ctx := context.Background()

	// Cleanup
	ordersCol := GetCollection(DBName, OrdersCollectionName)
	paymentsCol := GetCollection(DBName, PaymentRecordsCollectionName)
	reversalsCol := GetCollection(DBName, ReversalsCollectionName)

	ordersCol.DeleteMany(ctx, bson.M{})
	paymentsCol.DeleteMany(ctx, bson.M{})
	reversalsCol.DeleteMany(ctx, bson.M{})

	// Insert test data for multiple days
	baseTime := time.Now().Truncate(24 * time.Hour)

	orders := []interface{}{
		&models.Order{
			ID:        "order-1",
			Status:    models.OrderStatusComplete,
			Cost:      1000.0,
			Discount:  100.0,
			TotalCost: 900.0,
			UserID:    "user-1",
			CreatedAt: baseTime,
			UpdatedAt: baseTime,
		},
		&models.Order{
			ID:        "order-2",
			Status:    models.OrderStatusComplete,
			Cost:      500.0,
			Discount:  50.0,
			TotalCost: 450.0,
			UserID:    "user-2",
			CreatedAt: baseTime.AddDate(0, 0, 1),
			UpdatedAt: baseTime.AddDate(0, 0, 1),
		},
	}
	_, err := ordersCol.InsertMany(ctx, orders)
	assert.NoError(t, err)

	// Get daily breakdown
	startDate := baseTime.Format("2006-01-02")
	endDate := baseTime.AddDate(0, 0, 2).Format("2006-01-02")

	breakdown, err := repo.GetDailyBreakdown(ctx, startDate, endDate)
	assert.NoError(t, err)
	assert.NotNil(t, breakdown)
	assert.Greater(t, len(breakdown), 0)

	// Verify daily data structure
	for _, daily := range breakdown {
		assert.NotEmpty(t, daily.Date)
		assert.GreaterOrEqual(t, daily.OrderCount, 0)
		assert.GreaterOrEqual(t, daily.TotalSales, 0.0)
		assert.GreaterOrEqual(t, daily.CompletedCount, 0)
	}

	// Cleanup
	ordersCol.DeleteMany(ctx, bson.M{})
	paymentsCol.DeleteMany(ctx, bson.M{})
	reversalsCol.DeleteMany(ctx, bson.M{})
}

func TestReportRepository_EmptyDateRange(t *testing.T) {
	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		t.Skip("MONGO_TEST_URI not set; skipping report repository integration tests")
	}

	if err := InitMongo(uri); err != nil {
		t.Fatalf("InitMongo error: %v", err)
	}
	defer DisconnectMongo()

	repo := NewReportRepository()
	ctx := context.Background()

	// Cleanup
	ordersCol := GetCollection(DBName, OrdersCollectionName)
	ordersCol.DeleteMany(ctx, bson.M{})

	// Query for date range with no orders
	pastDate := time.Now().AddDate(-1, 0, 0)
	startDate := pastDate.Format("2006-01-02")
	endDate := pastDate.AddDate(0, 0, 5).Format("2006-01-02")

	report, err := repo.GetSummaryReport(ctx, startDate, endDate)
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, 0, report.TotalOrders)
	assert.Equal(t, 0.0, report.TotalSalesAmount)

	breakdown, err := repo.GetDailyBreakdown(ctx, startDate, endDate)
	assert.NoError(t, err)
	assert.NotNil(t, breakdown)
	assert.Equal(t, 0, len(breakdown))
}

func TestReportRepository_StatusCounting(t *testing.T) {
	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		t.Skip("MONGO_TEST_URI not set; skipping report repository integration tests")
	}

	if err := InitMongo(uri); err != nil {
		t.Fatalf("InitMongo error: %v", err)
	}
	defer DisconnectMongo()

	repo := NewReportRepository()
	ctx := context.Background()

	// Cleanup
	ordersCol := GetCollection(DBName, OrdersCollectionName)
	ordersCol.DeleteMany(ctx, bson.M{})

	// Insert orders with different statuses
	now := time.Now()
	orders := []interface{}{
		&models.Order{
			ID:        "order-1",
			Status:    models.OrderStatusComplete,
			Cost:      1000.0,
			TotalCost: 1000.0,
			UserID:    "user-1",
			CreatedAt: now,
			UpdatedAt: now,
		},
		&models.Order{
			ID:        "order-2",
			Status:    models.OrderStatusComplete,
			Cost:      500.0,
			TotalCost: 500.0,
			UserID:    "user-2",
			CreatedAt: now,
			UpdatedAt: now,
		},
		&models.Order{
			ID:        "order-3",
			Status:    models.OrderStatusCancelled,
			Cost:      200.0,
			TotalCost: 200.0,
			UserID:    "user-3",
			CreatedAt: now,
			UpdatedAt: now,
		},
		&models.Order{
			ID:        "order-4",
			Status:    models.OrderStatusProcessing,
			Cost:      300.0,
			TotalCost: 300.0,
			UserID:    "user-4",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
	_, err := ordersCol.InsertMany(ctx, orders)
	assert.NoError(t, err)

	startDate := now.AddDate(0, 0, -1).Format("2006-01-02")
	endDate := now.AddDate(0, 0, 1).Format("2006-01-02")

	report, err := repo.GetSummaryReport(ctx, startDate, endDate)
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, 4, report.TotalOrders)
	assert.Equal(t, 2, report.CompletedOrders)
	assert.Equal(t, 1, report.CancelledOrders)
	assert.Equal(t, 1, report.ProcessingOrders)
	assert.Equal(t, 2000.0, report.TotalSalesAmount)

	// Cleanup
	ordersCol.DeleteMany(ctx, bson.M{})
}

func TestReportRepository_AverageOrderValue(t *testing.T) {
	uri := os.Getenv("MONGO_TEST_URI")
	if uri == "" {
		t.Skip("MONGO_TEST_URI not set; skipping report repository integration tests")
	}

	if err := InitMongo(uri); err != nil {
		t.Fatalf("InitMongo error: %v", err)
	}
	defer DisconnectMongo()

	repo := NewReportRepository()
	ctx := context.Background()

	// Cleanup
	ordersCol := GetCollection(DBName, OrdersCollectionName)
	ordersCol.DeleteMany(ctx, bson.M{})

	// Insert orders with specific amounts
	now := time.Now()
	orders := []interface{}{
		&models.Order{
			ID:        "order-1",
			Status:    models.OrderStatusComplete,
			Cost:      1000.0,
			TotalCost: 1000.0,
			UserID:    "user-1",
			CreatedAt: now,
			UpdatedAt: now,
		},
		&models.Order{
			ID:        "order-2",
			Status:    models.OrderStatusComplete,
			Cost:      500.0,
			TotalCost: 500.0,
			UserID:    "user-2",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
	_, err := ordersCol.InsertMany(ctx, orders)
	assert.NoError(t, err)

	startDate := now.AddDate(0, 0, -1).Format("2006-01-02")
	endDate := now.AddDate(0, 0, 1).Format("2006-01-02")

	report, err := repo.GetSummaryReport(ctx, startDate, endDate)
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, 2, report.TotalOrders)
	assert.Equal(t, 1500.0, report.TotalSalesAmount)
	assert.Equal(t, 750.0, report.AverageOrderValue)

	// Cleanup
	ordersCol.DeleteMany(ctx, bson.M{})
}
