package router

import (
	"log/slog"
	"net/http"

	"github.com/ambientlabscomputing/ucrs/sdk/types"
	"github.com/ambientlabscomputing/ucrs/service"
	"github.com/gin-gonic/gin"
)

// resolveSourceHandler is the handler for GET /sources/resolve.
// It is defined as a standalone function rather than a method on AppRouter so
// it can be constructed with the SourceService without changing the AppRouter struct.
func resolveSourceHandler(sourceSvc *service.SourceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := slog.Default()

		var req types.ResolveSourceRequest
		if err := c.ShouldBindQuery(&req); err != nil {
			logger.Warn("invalid resolve source request", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		resolved, err := sourceSvc.Resolve(c.Request.Context(), req.Source, req.Ref, req.Token)
		if err != nil {
			logger.Warn("source resolution failed", "source", req.Source, "error", err)
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, types.ResolveSourceResponse{Resolved: *resolved})
	}
}
