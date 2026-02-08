package handlers

import (
	"context"
	"net/http"

	"github.com/eddie-wainaina1/maggiesb/internal/models"
	"github.com/gin-gonic/gin"
)

// AdminGetSummaryReport returns a summary report for a date range with daily breakdown
// Query parameters: startDate (YYYY-MM-DD), endDate (YYYY-MM-DD)
func AdminGetSummaryReport(c *gin.Context) {
	var query models.GetReportsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing or invalid date parameters (startDate, endDate in YYYY-MM-DD format)"})
		return
	}

	reportRepo := NewReportRepository
	if reportRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "report repository not initialized"})
		return
	}

	report, err := reportRepo.GetSummaryReport(context.Background(), query.StartDate, query.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// AdminGetDailyBreakdown returns detailed daily metrics for a date range
// Query parameters: startDate (YYYY-MM-DD), endDate (YYYY-MM-DD)
func AdminGetDailyBreakdown(c *gin.Context) {
	var query models.GetReportsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing or invalid date parameters (startDate, endDate in YYYY-MM-DD format)"})
		return
	}

	reportRepo := NewReportRepository
	if reportRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "report repository not initialized"})
		return
	}

	dailyBreakdown, err := reportRepo.GetDailyBreakdown(context.Background(), query.StartDate, query.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"dateRange": models.DateRange{
			StartDate: query.StartDate,
			EndDate:   query.EndDate,
		},
		"data": dailyBreakdown,
	})
}
