package database

import (
	"context"
	"fmt"
	"time"

	"github.com/eddie-wainaina1/maggiesb/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type ReportRepository struct {
	ordersCollection    *mongo.Collection
	paymentsCollection  *mongo.Collection
	reversalsCollection Collection
}

// NewReportRepository creates a new report repository
func NewReportRepository() *ReportRepository {
	return &ReportRepository{
		ordersCollection:    GetCollection(DBName, OrdersCollectionName),
		paymentsCollection:  GetCollection(DBName, PaymentRecordsCollectionName),
		reversalsCollection: NewMongoCollection(GetCollection(DBName, ReversalsCollectionName)),
	}
}

// GetSummaryReport returns aggregate metrics for a date range
func (rr *ReportRepository) GetSummaryReport(ctx context.Context, startDate, endDate string) (*models.SummaryReport, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Parse dates to time.Time for date range queries
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date format: %w", err)
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date format: %w", err)
	}

	// Include the entire end date by setting time to end of day
	end = end.AddDate(0, 0, 1).Add(-time.Second)

	// Aggregate orders by status and calculate totals
	pipeline := mongo.Pipeline{
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "createdAt", Value: bson.D{
					{Key: "$gte", Value: start},
					{Key: "$lte", Value: end},
				}},
			}},
		},
		bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$status"},
				{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
				{Key: "totalSales", Value: bson.D{{Key: "$sum", Value: "$totalCost"}}},
				{Key: "totalDiscounts", Value: bson.D{{Key: "$sum", Value: "$discount"}}},
			}},
		},
	}

	cursor, err := rr.ordersCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate orders: %w", err)
	}
	defer cursor.Close(ctx)

	var orderStats []struct {
		Status       string  `bson:"_id"`
		Count        int     `bson:"count"`
		TotalSales   float64 `bson:"totalSales"`
		TotalDiscount float64 `bson:"totalDiscounts"`
	}
	if err = cursor.All(ctx, &orderStats); err != nil {
		return nil, fmt.Errorf("failed to decode order stats: %w", err)
	}

	// Get reversal data for the date range
	reversalPipeline := mongo.Pipeline{
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "date", Value: bson.D{
					{Key: "$gte", Value: startDate},
					{Key: "$lte", Value: endDate},
				}},
			}},
		},
		bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: nil},
				{Key: "totalReversals", Value: bson.D{{Key: "$sum", Value: "$amount"}}},
				{Key: "reversalCount", Value: bson.D{{Key: "$sum", Value: 1}}},
			}},
		},
	}

	reversalCursor, err := rr.reversalsCollection.Aggregate(ctx, reversalPipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate reversals: %w", err)
	}
	defer reversalCursor.Close(ctx)

	var reversalStats []struct {
		TotalReversals float64 `bson:"totalReversals"`
		ReversalCount  int     `bson:"reversalCount"`
	}
	if err = reversalCursor.All(ctx, &reversalStats); err != nil {
		return nil, fmt.Errorf("failed to decode reversal stats: %w", err)
	}

	// Get payment data for the date range (payments completed)
	paymentPipeline := mongo.Pipeline{
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "status", Value: "completed"},
				{Key: "createdAt", Value: bson.D{
					{Key: "$gte", Value: start.Format("2006-01-02 15:04:05")},
					{Key: "$lte", Value: end.Format("2006-01-02 15:04:05")},
				}},
			}},
		},
		bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: nil},
				{Key: "totalPayments", Value: bson.D{{Key: "$sum", Value: "$amount"}}},
				{Key: "paymentCount", Value: bson.D{{Key: "$sum", Value: 1}}},
			}},
		},
	}

	paymentCursor, err := rr.paymentsCollection.Aggregate(ctx, paymentPipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate payments: %w", err)
	}
	defer paymentCursor.Close(ctx)

	var paymentStats []struct {
		TotalPayments float64 `bson:"totalPayments"`
		PaymentCount  int     `bson:"paymentCount"`
	}
	if err = paymentCursor.All(ctx, &paymentStats); err != nil {
		return nil, fmt.Errorf("failed to decode payment stats: %w", err)
	}

	// Build summary report
	summary := &models.SummaryReport{
		DateRange: models.DateRange{
			StartDate: startDate,
			EndDate:   endDate,
		},
	}

	// Process order stats
	for _, stat := range orderStats {
		summary.TotalOrders += stat.Count
		summary.TotalSalesAmount += stat.TotalSales
		summary.TotalDiscountsGiven += stat.TotalDiscount

		switch stat.Status {
		case models.OrderStatusComplete:
			summary.CompletedOrders = stat.Count
		case models.OrderStatusCancelled:
			summary.CancelledOrders = stat.Count
		case models.OrderStatusProcessing:
			summary.ProcessingOrders = stat.Count
		}
	}

	// Add payment stats
	if len(paymentStats) > 0 {
		summary.TotalPaymentsReceived = paymentStats[0].TotalPayments
	}

	// Add reversal stats
	if len(reversalStats) > 0 {
		summary.TotalReversalsIssued = reversalStats[0].TotalReversals
	}

	// Calculate average order value
	if summary.TotalOrders > 0 {
		summary.AverageOrderValue = summary.TotalSalesAmount / float64(summary.TotalOrders)
	}

	// Get daily breakdown
	dailyBreakdown, err := rr.GetDailyBreakdown(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily breakdown: %w", err)
	}
	summary.DailyBreakdown = dailyBreakdown

	return summary, nil
}

// GetDailyBreakdown returns daily metrics for each day in the date range
func (rr *ReportRepository) GetDailyBreakdown(ctx context.Context, startDate, endDate string) ([]models.DailySalesReport, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Parse dates
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date format: %w", err)
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date format: %w", err)
	}

	// Include the entire end date
	end = end.AddDate(0, 0, 1).Add(-time.Second)

	// Aggregate orders by date
	pipeline := mongo.Pipeline{
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "createdAt", Value: bson.D{
					{Key: "$gte", Value: start},
					{Key: "$lte", Value: end},
				}},
			}},
		},
		bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: bson.D{
					{Key: "$dateToString", Value: bson.D{
						{Key: "format", Value: "%Y-%m-%d"},
						{Key: "date", Value: "$createdAt"},
					}},
				}},
				{Key: "orderCount", Value: bson.D{{Key: "$sum", Value: 1}}},
				{Key: "totalSales", Value: bson.D{{Key: "$sum", Value: "$totalCost"}}},
				{Key: "totalDiscounts", Value: bson.D{{Key: "$sum", Value: "$discount"}}},
				{Key: "statuses", Value: bson.D{{Key: "$push", Value: "$status"}}},
			}},
		},
		bson.D{
			{Key: "$sort", Value: bson.D{{Key: "_id", Value: 1}}},
		},
	}

	cursor, err := rr.ordersCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate daily orders: %w", err)
	}
	defer cursor.Close(ctx)

	var dailyStats []struct {
		Date           string   `bson:"_id"`
		OrderCount     int      `bson:"orderCount"`
		TotalSales     float64  `bson:"totalSales"`
		TotalDiscounts float64  `bson:"totalDiscounts"`
		Statuses       []string `bson:"statuses"`
	}
	if err = cursor.All(ctx, &dailyStats); err != nil {
		return nil, fmt.Errorf("failed to decode daily stats: %w", err)
	}

	// Create daily reports with status breakdown
	dailyReports := make([]models.DailySalesReport, 0, len(dailyStats))

	for _, stat := range dailyStats {
		daily := models.DailySalesReport{
			Date:           stat.Date,
			OrderCount:     stat.OrderCount,
			TotalSales:     stat.TotalSales,
			TotalDiscounts: stat.TotalDiscounts,
		}

		// Count by status
		for _, status := range stat.Statuses {
			switch status {
			case models.OrderStatusComplete:
				daily.CompletedCount++
			case models.OrderStatusCancelled:
				daily.CancelledCount++
			case models.OrderStatusProcessing:
				daily.ProcessingCount++
			}
		}

		// Get payments for this day
		dayStart, _ := time.Parse("2006-01-02", stat.Date)
		dayEnd := dayStart.AddDate(0, 0, 1).Add(-time.Second)

		paymentPipeline := mongo.Pipeline{
			bson.D{
				{Key: "$match", Value: bson.D{
					{Key: "status", Value: "completed"},
					{Key: "createdAt", Value: bson.D{
						{Key: "$gte", Value: dayStart.Format("2006-01-02 15:04:05")},
						{Key: "$lte", Value: dayEnd.Format("2006-01-02 15:04:05")},
					}},
				}},
			},
			bson.D{
				{Key: "$group", Value: bson.D{
					{Key: "_id", Value: nil},
					{Key: "totalPayments", Value: bson.D{{Key: "$sum", Value: "$amount"}}},
				}},
			},
		}

		paymentCursor, err := rr.paymentsCollection.Aggregate(ctx, paymentPipeline)
		if err == nil {
			defer paymentCursor.Close(ctx)
			var paymentData []struct {
				TotalPayments float64 `bson:"totalPayments"`
			}
			if paymentCursor.All(ctx, &paymentData) == nil && len(paymentData) > 0 {
				daily.TotalPayments = paymentData[0].TotalPayments
			}
		}

		// Get reversals for this day
		reversalPipeline := mongo.Pipeline{
			bson.D{
				{Key: "$match", Value: bson.D{
					{Key: "date", Value: stat.Date},
				}},
			},
			bson.D{
				{Key: "$group", Value: bson.D{
					{Key: "_id", Value: nil},
					{Key: "totalReversals", Value: bson.D{{Key: "$sum", Value: "$amount"}}},
				}},
			},
		}

		reversalCursor, err := rr.reversalsCollection.Aggregate(ctx, reversalPipeline)
		if err == nil {
			defer reversalCursor.Close(ctx)
			var reversalData []struct {
				TotalReversals float64 `bson:"totalReversals"`
			}
			if reversalCursor.All(ctx, &reversalData) == nil && len(reversalData) > 0 {
				daily.TotalReversals = reversalData[0].TotalReversals
			}
		}

		dailyReports = append(dailyReports, daily)
	}

	return dailyReports, nil
}
