package router

import (
	"encoding/base64"
	"net/http"

	"github.com/ambientlabscomputing/underleaf/capability_registry_service/types"
	"github.com/ambientlabscomputing/underleaf/capability_registry_service/utils"
	"github.com/gin-gonic/gin"
)

func (r *AppRouter) GetSnapshotHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := utils.GetLogger(c.Request.Context())

		// Note: ETag/304 handling disabled for better frontend compatibility
		// Always return full snapshot with 200 OK
		snapshot, _, err := r.service.Sync().GetSnapshot(c.Request.Context(), "")
		if err != nil {
			logger.Error("failed to get snapshot", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		if snapshot == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "no snapshot available"})
			return
		}

		// Set ETag header for future caching (informational only)
		c.Header("ETag", snapshot.ETag)
		// Set cache headers for browser caching
		c.Header("Cache-Control", "public, max-age=300") // 5 minutes

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

// GetPublicKeyHandler returns the UCRS public key for signature verification
func (r *AppRouter) GetPublicKeyHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get public key from service
		publicKey := r.service.GetPublicKey()
		if publicKey == nil || len(publicKey) == 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "public key not available"})
			return
		}

		// Return as base64-encoded string
		publicKeyBase64 := base64.StdEncoding.EncodeToString(publicKey)

		c.JSON(http.StatusOK, gin.H{
			"public_key": publicKeyBase64,
			"algorithm":  "Ed25519",
		})
	}
}
