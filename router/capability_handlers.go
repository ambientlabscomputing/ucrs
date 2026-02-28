package router

import (
	"net/http"
	"strconv"

	"github.com/ambientlabscomputing/ucrs/sdk/types"
	"github.com/ambientlabscomputing/ucrs/utils"
	"github.com/gin-gonic/gin"
)

func (r *AppRouter) CreateCapabilityHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := utils.GetLogger(c.Request.Context())

		var req types.CreateCapabilityRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Warn("invalid request", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		capability, err := r.service.Capabilities().Create(c.Request.Context(), &req)
		if err != nil {
			logger.Error("failed to create capability", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusCreated, capability)
	}
}

func (r *AppRouter) GetCapabilityHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := utils.GetLogger(c.Request.Context())
		id := c.Param("id")

		capability, err := r.service.Capabilities().Get(c.Request.Context(), id)
		if err != nil {
			logger.Error("failed to get capability", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		if capability == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "capability not found"})
			return
		}

		c.JSON(http.StatusOK, capability)
	}
}

func (r *AppRouter) UpdateCapabilityHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := utils.GetLogger(c.Request.Context())
		id := c.Param("id")

		var req types.UpdateCapabilityRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Warn("invalid request", "error", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		capability, err := r.service.Capabilities().Update(c.Request.Context(), id, &req)
		if err != nil {
			logger.Error("failed to update capability", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		if capability == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "capability not found"})
			return
		}

		c.JSON(http.StatusOK, capability)
	}
}

func (r *AppRouter) DeleteCapabilityHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := utils.GetLogger(c.Request.Context())
		id := c.Param("id")

		if err := r.service.Capabilities().Delete(c.Request.Context(), id); err != nil {
			logger.Error("failed to delete capability", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		c.JSON(http.StatusNoContent, nil)
	}
}

func (r *AppRouter) ListCapabilitiesHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := utils.GetLogger(c.Request.Context())

		// Parse query parameters
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		// Check if we have filters for Query vs List
		var riskClass *types.RiskClass
		var permissionClass *string
		var verb *string

		if rc := c.Query("risk_class"); rc != "" {
			r := types.RiskClass(rc)
			riskClass = &r
		}
		if pc := c.Query("permission_class"); pc != "" {
			permissionClass = &pc
		}
		if v := c.Query("verb"); v != "" {
			verb = &v
		}

		var capabilities []types.Capability
		var total int
		var err error

		if riskClass != nil || permissionClass != nil || verb != nil {
			// Use Query with filters
			req := &types.QueryCapabilitiesRequest{
				RiskClass:       riskClass,
				PermissionClass: permissionClass,
				Verb:            verb,
				Limit:           limit,
				Offset:          offset,
			}
			capabilities, total, err = r.service.Capabilities().Query(c.Request.Context(), req)
		} else {
			// Use List (no filters)
			capabilities, total, err = r.service.Capabilities().List(c.Request.Context(), limit, offset)
		}

		if err != nil {
			logger.Error("failed to list capabilities", "error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		response := types.ListResponse{
			Items:      capabilities,
			TotalCount: total,
			Limit:      limit,
			Offset:     offset,
		}

		c.JSON(http.StatusOK, response)
	}
}
