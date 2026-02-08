package models

// DailySalesReport represents sales metrics for a single day
type DailySalesReport struct {
	Date                string  `json:"date" bson:"date"`                     // YYYY-MM-DD
	OrderCount          int     `json:"orderCount" bson:"orderCount"`
	TotalSales          float64 `json:"totalSales" bson:"totalSales"`         // sum of totalCost for completed orders
	TotalDiscounts      float64 `json:"totalDiscounts" bson:"totalDiscounts"` // sum of discounts applied
	TotalPayments       float64 `json:"totalPayments" bson:"totalPayments"`   // sum of payments received
	TotalReversals      float64 `json:"totalReversals" bson:"totalReversals"` // sum of reversal amounts
	CancelledCount      int     `json:"cancelledCount" bson:"cancelledCount"`
	CompletedCount      int     `json:"completedCount" bson:"completedCount"`
	ProcessingCount     int     `json:"processingCount" bson:"processingCount"`
}

// SummaryReport represents aggregate metrics for a date range
type SummaryReport struct {
	DateRange              DateRange             `json:"dateRange"`
	TotalOrders            int                   `json:"totalOrders"`
	CompletedOrders        int                   `json:"completedOrders"`
	CancelledOrders        int                   `json:"cancelledOrders"`
	ProcessingOrders       int                   `json:"processingOrders"`
	TotalSalesAmount       float64               `json:"totalSalesAmount"`       // sum of all order totalCost
	TotalDiscountsGiven    float64               `json:"totalDiscountsGiven"`   // sum of all discounts
	TotalPaymentsReceived  float64               `json:"totalPaymentsReceived"` // sum of all payments
	TotalReversalsIssued   float64               `json:"totalReversalsIssued"`  // sum of all reversals
	AverageOrderValue      float64               `json:"averageOrderValue"`
	DailyBreakdown         []DailySalesReport    `json:"dailyBreakdown"`        // metrics for each day in range
}

// DateRange represents a start and end date
type DateRange struct {
	StartDate string `json:"startDate"` // YYYY-MM-DD
	EndDate   string `json:"endDate"`   // YYYY-MM-DD
}

// GetReportsQuery contains query parameters for fetching reports
type GetReportsQuery struct {
	StartDate string `form:"startDate" binding:"required"` // YYYY-MM-DD
	EndDate   string `form:"endDate" binding:"required"`   // YYYY-MM-DD
}
