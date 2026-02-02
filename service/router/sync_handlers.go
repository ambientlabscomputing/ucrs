package router

import (
	"net/http"

	"github.com/ambientlabscomputing/underleaf/capability_registry_service/types"
	"github.com/ambientlabscomputing/underleaf/capability_registry_service/utils"
	"github.com/gin-gonic/gin"
)

func (r *AppRouter) GetSnapshotHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := utils.GetLogger(c.Request.Context())

		// Check for If-None-Match header (ETag)
		ifNoneMatch := c.GetHeader("If-None-Match")

		snapshot, notModified, err := r.service.Sync().GetSnapshot(c.Request.Context(), ifNoneMatch)
		if err != nil {
			logger.Error("failed to get snapshot", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		if notModified {
			c.Status(http.StatusNotModified)
			return
		}

		if snapshot == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "no snapshot available"})
			return
		}

		// Set ETag header
		c.Header("ETag", snapshot.ETag)

		response := types.SyncSnapshotResponse{
			Snapshot: *snapshot,
		}

		c.JSON(http.StatusOK, response)
	}
}

func (r *AppRouter) GetDeltaHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := utils.GetLogger(c.Request.Context())

		// Get the "since" query parameter
		sinceVersion := c.Query("since")
		if sinceVersion == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing 'since' query parameter"})
			return
		}

		delta, err := r.service.Sync().GetDelta(c.Request.Context(), sinceVersion)
		if err != nil {
			logger.Error("failed to get delta", "error", err, "since_version", sinceVersion)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		if delta == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "no delta available"})
			return
		}

		response := types.SyncDeltaResponse{
			Delta: *delta,
		}

		c.JSON(http.StatusOK, response)
	}
}
