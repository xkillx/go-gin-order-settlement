package settlement

import (
	"fmt"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

// RegisterRoutes registers settlement-related routes like CSV downloads.
func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	// avoid unused param warnings for now
	_ = injector

	// Dynamic route: GET /downloads/:job_id -> serves /tmp/settlements/<job_id>.csv
	server.GET("/downloads/:job_id", func(c *gin.Context) {
		jobID := c.Param("job_id")
		basePath := "/tmp/settlements"
		filename := jobID
		if filepath.Ext(filename) == "" {
			filename += ".csv"
		}
		fullPath := filepath.Join(basePath, filename)

		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(fullPath)))
		c.File(fullPath)
	})

	// Static mapping also allows direct access to /downloads/<job_id>.csv
	server.Static("/downloads", "/tmp/settlements")
}
