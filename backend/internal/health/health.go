package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler is a liveness check: it reports that the process is up and serving HTTP.
func Handler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
