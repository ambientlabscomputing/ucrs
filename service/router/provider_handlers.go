package router

import (
	"net/http"
	"strconv"

	"github.com/ambientlabscomputing/underleaf/capability_registry_service/types"
	"github.com/ambientlabscomputing/underleaf/capability_registry_service/utils"
	"github.com/gin-gonic/gin"
)

func (r *AppRouter) CreateProviderHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := utils.GetLogger(c.Request.Context())

		var req types.CreateProviderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Warn("invalid request", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		provider, err := r.service.Providers().Create(c.Request.Context(), &req)
		if err != nil {
			logger.Error("failed to create provider", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusCreated, provider)
	}
}

func (r *AppRouter) GetProviderHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := utils.GetLogger(c.Request.Context())
		providerID := c.Param("provider_id")
		version := c.Param("version")

		provider, err := r.service.Providers().Get(c.Request.Context(), providerID, version)
		if err != nil {
			logger.Error("failed to get provider", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		if provider == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
			return
		}

		c.JSON(http.StatusOK, provider)
	}
}

func (r *AppRouter) UpdateProviderHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := utils.GetLogger(c.Request.Context())
		providerID := c.Param("provider_id")
		version := c.Param("version")

		var req types.UpdateProviderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Warn("invalid request", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		provider, err := r.service.Providers().Update(c.Request.Context(), providerID, version, &req)
		if err != nil {
			logger.Error("failed to update provider", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		if provider == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
			return
		}

		c.JSON(http.StatusOK, provider)
	}
}

func (r *AppRouter) DeleteProviderHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := utils.GetLogger(c.Request.Context())
		providerID := c.Param("provider_id")
		version := c.Param("version")

		if err := r.service.Providers().Delete(c.Request.Context(), providerID, version); err != nil {
			logger.Error("failed to delete provider", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusNoContent, nil)
	}
}

func (r *AppRouter) ListProvidersHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := utils.GetLogger(c.Request.Context())

		// Parse query parameters
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		// Check if we have filters for Query vs List
		var capabilityID *string
		var trustTier *types.TrustTier
		var artifactType *types.ArtifactType

		if cid := c.Query("capability_id"); cid != "" {
			capabilityID = &cid
		}
		if tt := c.Query("trust_tier"); tt != "" {
			t := types.TrustTier(tt)
			trustTier = &t
		}
		if at := c.Query("artifact_type"); at != "" {
			a := types.ArtifactType(at)
			artifactType = &a
		}

		var providers []types.Provider
		var total int
		var err error

		if capabilityID != nil || trustTier != nil || artifactType != nil {
			// Use Query with filters
			req := &types.QueryProvidersRequest{
				CapabilityID: capabilityID,
				TrustTier:    trustTier,
				ArtifactType: artifactType,
				Limit:        limit,
				Offset:       offset,
			}
			providers, total, err = r.service.Providers().Query(c.Request.Context(), req)
		} else {
			// Use List (no filters)
			providers, total, err = r.service.Providers().List(c.Request.Context(), limit, offset)
		}

		if err != nil {
			logger.Error("failed to list providers", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		response := types.ListResponse{
			Items:      providers,
			TotalCount: total,
			Limit:      limit,
			Offset:     offset,
		}

		c.JSON(http.StatusOK, response)
	}
}
