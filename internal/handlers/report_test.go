package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eddie-wainaina1/maggiesb/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAdminGetSummaryReport_MissingDates(t *testing.T) {
	gin.SetMode(gin.TestMode)

	httpReq := httptest.NewRequest("GET", "/admin/reports/summary", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq

	AdminGetSummaryReport(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "missing or invalid date parameters")
}

func TestAdminGetSummaryReport_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	startDate := "2026-02-01"
	endDate := "2026-02-09"

	expectedReport := &models.SummaryReport{
		DateRange: models.DateRange{
			StartDate: startDate,
			EndDate:   endDate,
		},
		TotalOrders:           10,
		CompletedOrders:       8,
		CancelledOrders:       1,
		ProcessingOrders:      1,
		TotalSalesAmount:      5000.0,
		TotalDiscountsGiven:   500.0,
		TotalPaymentsReceived: 4500.0,
		TotalReversalsIssued:  100.0,
		AverageOrderValue:     500.0,
		DailyBreakdown: []models.DailySalesReport{
			{
				Date:            "2026-02-01",
				OrderCount:      2,
				TotalSales:      1000.0,
				TotalDiscounts:  100.0,
				TotalPayments:   900.0,
				TotalReversals:  0,
				CompletedCount:  2,
				CancelledCount:  0,
				ProcessingCount: 0,
			},
		},
	}

	mockReportRepo := new(MockReportRepository)
	mockReportRepo.On("GetSummaryReport", mock.Anything, startDate, endDate).Return(expectedReport, nil)

	oldReportRepo := NewReportRepository
	NewReportRepository = ReportRepository(mockReportRepo)
	defer func() {
		NewReportRepository = oldReportRepo
	}()

	httpReq := httptest.NewRequest("GET", "/admin/reports/summary?startDate=2026-02-01&endDate=2026-02-09", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq

	AdminGetSummaryReport(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "2026-02-01")
	assert.Contains(t, w.Body.String(), "totalOrders")
	mockReportRepo.AssertCalled(t, "GetSummaryReport", mock.Anything, startDate, endDate)
}

func TestAdminGetDailyBreakdown_MissingDates(t *testing.T) {
	gin.SetMode(gin.TestMode)

	httpReq := httptest.NewRequest("GET", "/admin/reports/daily", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq

	AdminGetDailyBreakdown(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "missing or invalid date parameters")
}

func TestAdminGetDailyBreakdown_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	startDate := "2026-02-01"
	endDate := "2026-02-09"

	expectedBreakdown := []models.DailySalesReport{
		{
			Date:            "2026-02-01",
			OrderCount:      2,
			TotalSales:      1000.0,
			TotalDiscounts:  100.0,
			TotalPayments:   900.0,
			TotalReversals:  0,
			CompletedCount:  2,
			CancelledCount:  0,
			ProcessingCount: 0,
		},
		{
			Date:            "2026-02-02",
			OrderCount:      3,
			TotalSales:      1500.0,
			TotalDiscounts:  150.0,
			TotalPayments:   1350.0,
			TotalReversals:  50.0,
			CompletedCount:  3,
			CancelledCount:  0,
			ProcessingCount: 0,
		},
	}

	mockReportRepo := new(MockReportRepository)
	mockReportRepo.On("GetDailyBreakdown", mock.Anything, startDate, endDate).Return(expectedBreakdown, nil)

	oldReportRepo := NewReportRepository
	NewReportRepository = ReportRepository(mockReportRepo)
	defer func() {
		NewReportRepository = oldReportRepo
	}()

	httpReq := httptest.NewRequest("GET", "/admin/reports/daily?startDate=2026-02-01&endDate=2026-02-09", nil)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httpReq

	AdminGetDailyBreakdown(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "2026-02-01")
	assert.Contains(t, w.Body.String(), "2026-02-02")
	mockReportRepo.AssertCalled(t, "GetDailyBreakdown", mock.Anything, startDate, endDate)
}
