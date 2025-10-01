package settlement

import (
    "errors"
    "fmt"
    "net/http"
    "os"
    "path/filepath"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/samber/do"
    jobrepo "github.com/xkillx/go-gin-order-settlement/modules/job/repository"
    settlementService "github.com/xkillx/go-gin-order-settlement/modules/settlement/service"
    "github.com/xkillx/go-gin-order-settlement/pkg/constants"
    "gorm.io/gorm"
)

// RegisterRoutes registers settlement-related routes like CSV downloads.
func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	// Resolve dependencies
	db := do.MustInvokeNamed[*gorm.DB](injector, constants.DB)
	jobRepository := jobrepo.NewJobRepository(db)
	jobManager := do.MustInvoke[*settlementService.JobManager](injector)

	// 1) POST /jobs/settlement
	server.POST("/jobs/settlement", func(c *gin.Context) {
		var req struct {
			From string `json:"from" binding:"required"`
			To   string `json:"to" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		const layout = "2006-01-02"
		fromDate, err := time.Parse(layout, req.From)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid 'from' date format, expected YYYY-MM-DD"})
			return
		}
		toDate, err := time.Parse(layout, req.To)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid 'to' date format, expected YYYY-MM-DD"})
			return
		}
		if toDate.Before(fromDate) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "'to' must be on or after 'from'"})
			return
		}

		jobID, err := jobManager.StartSettlementJob(c.Request.Context(), fromDate, toDate)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{
			"job_id": jobID,
			"status": "QUEUED",
		})
	})

	// 2) GET /jobs/:id
	server.GET("/jobs/:id", func(c *gin.Context) {
		id := c.Param("id")
		j, err := jobRepository.Get(c.Request.Context(), id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "job not found"})
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		payload := gin.H{
			"job_id":   j.ID,
			"status":   j.Status,
			"progress": j.Progress,
			"processed": j.Processed,
			"total":    j.Total,
		}
		if j.Status == "COMPLETED" && j.ResultPath != "" {
			payload["download_url"] = "/downloads/" + j.ID + ".csv"
		}
		c.JSON(http.StatusOK, payload)
	})

	// 3) POST /jobs/:id/cancel
	server.POST("/jobs/:id/cancel", func(c *gin.Context) {
		id := c.Param("id")
		// Ensure job exists
		if _, err := jobRepository.Get(c.Request.Context(), id); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "job not found"})
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err := jobRepository.RequestCancel(c.Request.Context(), id); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// Try to cancel in-memory running job
		wasRunning := jobManager.Cancel(id)
		if wasRunning {
			_ = jobRepository.SetStatus(c.Request.Context(), id, "CANCELLING")
			c.JSON(http.StatusAccepted, gin.H{"job_id": id, "status": "CANCELLING"})
			return
		}
		_ = jobRepository.SetStatus(c.Request.Context(), id, "CANCELLED")
		c.JSON(http.StatusAccepted, gin.H{"job_id": id, "status": "CANCELLED"})
	})

	// 4) GET /downloads/:job_id.csv -> serves /tmp/settlements/<job_id>.csv
	server.GET("/downloads/:job_id.csv", func(c *gin.Context) {
		jobID := c.Param("job_id")
		fullPath := filepath.Join("/tmp/settlements", jobID+".csv")

		if _, err := os.Stat(fullPath); err != nil {
			if os.IsNotExist(err) {
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "file not found"})
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(fullPath)))
		c.File(fullPath)
	})
}
