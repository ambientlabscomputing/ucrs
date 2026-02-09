package router

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ambientlabscomputing/ucrs/service"
	"github.com/ambientlabscomputing/ucrs/types"
	"github.com/ambientlabscomputing/ucrs/utils"
	"github.com/gin-gonic/gin"
)

// MockService implements the service.Service interface for testing
type MockService struct {
	capabilities *MockCapabilityService
	providers    *MockProviderService
	sync         *MockSyncService
}

func (m *MockService) Start(ctx context.Context) error {
	return nil
}

func (m *MockService) Stop(ctx context.Context) error {
	return nil
}

func (m *MockService) Capabilities() *service.CapabilityService {
	// This is a type mismatch - we'll need to use a different approach
	return nil
}

func (m *MockService) Providers() *service.ProviderService {
	return nil
}

func (m *MockService) Sync() *service.SyncService {
	return nil
}

type MockCapabilityService struct {
	CreateFunc func(ctx context.Context, req *types.CreateCapabilityRequest) (*types.Capability, error)
	GetFunc    func(ctx context.Context, id string) (*types.Capability, error)
	ListFunc   func(ctx context.Context, limit, offset int) ([]types.Capability, int, error)
}

type MockProviderService struct{}

type MockSyncService struct {
	GetSnapshotFunc func(ctx context.Context, ifNoneMatch string) (*types.RegistrySnapshot, bool, error)
	GetDeltaFunc    func(ctx context.Context, sinceVersion string) (*types.DeltaUpdate, error)
}

func setupRouter(t *testing.T) (*gin.Engine, *utils.Settings) {
	gin.SetMode(gin.TestMode)

	settings := &utils.Settings{
		Port:        "8083",
		Address:     "0.0.0.0",
		BasePath:    "/api/v1/registry",
		LogLevel:    "error",
		LogFormat:   "text",
		LogToStderr: true,
		LogToFile:   false,
	}

	utils.InitLogger(settings)

	engine := gin.New()
	return engine, settings
}

func TestHealthEndpoint(t *testing.T) {
	engine, settings := setupRouter(t)

	engine.GET(settings.BasePath+"/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/registry/health", nil)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("expected status ok, got %v", response["status"])
	}
}

func TestCORSMiddleware(t *testing.T) {
	engine, _ := setupRouter(t)

	engine.Use(CORSMiddleware())
	engine.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "ok"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("expected CORS header, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestTraceIDMiddleware(t *testing.T) {
	engine, _ := setupRouter(t)

	ctx := context.Background()
	engine.Use(TraceIDMiddleware(ctx))
	engine.GET("/test", func(c *gin.Context) {
		// Get trace ID from request context
		traceID := utils.GetTraceID(c.Request.Context())
		c.JSON(200, gin.H{"trace_id": traceID})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	traceID := w.Header().Get("X-Trace-ID")
	if traceID == "" {
		t.Error("expected X-Trace-ID header to be set")
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response["trace_id"] != traceID {
		t.Errorf("expected trace_id %s, got %v", traceID, response["trace_id"])
	}
}

func TestSplitScopes(t *testing.T) {
	tests := []struct {
		name     string
		scopes   string
		expected []string
	}{
		{
			name:     "single scope",
			scopes:   "read:registry",
			expected: []string{"read:registry"},
		},
		{
			name:     "multiple scopes",
			scopes:   "read:registry write:registry admin:registry",
			expected: []string{"read:registry", "write:registry", "admin:registry"},
		},
		{
			name:     "empty string",
			scopes:   "",
			expected: []string{},
		},
		{
			name:     "scopes with extra spaces",
			scopes:   "read:registry  write:registry",
			expected: []string{"read:registry", "write:registry"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitScopes(tt.scopes)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d scopes, got %d", len(tt.expected), len(result))
				return
			}
			for i, scope := range result {
				if scope != tt.expected[i] {
					t.Errorf("expected scope %s, got %s", tt.expected[i], scope)
				}
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		item     string
		expected bool
	}{
		{
			name:     "item exists",
			slice:    []string{"read:registry", "write:registry"},
			item:     "read:registry",
			expected: true,
		},
		{
			name:     "item does not exist",
			slice:    []string{"read:registry", "write:registry"},
			item:     "admin:registry",
			expected: false,
		},
		{
			name:     "empty slice",
			slice:    []string{},
			item:     "read:registry",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.slice, tt.item)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCreateCapabilityHandler_InvalidRequest(t *testing.T) {
	engine, settings := setupRouter(t)

	// Create a mock router with minimal setup
	router := &AppRouter{
		engine:   engine,
		settings: settings,
	}

	engine.POST("/capabilities", router.CreateCapabilityHandler())

	// Send invalid JSON
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/capabilities", bytes.NewBufferString("{invalid json}"))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestListResponse_Structure(t *testing.T) {
	response := types.ListResponse{
		Items:      []string{"item1", "item2"},
		TotalCount: 2,
		Limit:      10,
		Offset:     0,
	}

	if response.TotalCount != 2 {
		t.Errorf("expected total count 2, got %d", response.TotalCount)
	}
	if response.Limit != 10 {
		t.Errorf("expected limit 10, got %d", response.Limit)
	}
	if response.Offset != 0 {
		t.Errorf("expected offset 0, got %d", response.Offset)
	}
}

func TestSyncSnapshotResponse_Structure(t *testing.T) {
	snapshot := types.RegistrySnapshot{
		Version:   "12345",
		ETag:      "\"abc123\"",
		Timestamp: time.Now(),
	}

	response := types.SyncSnapshotResponse{
		Snapshot: snapshot,
	}

	if response.Snapshot.Version != "12345" {
		t.Errorf("expected version 12345, got %s", response.Snapshot.Version)
	}
	if response.Snapshot.ETag != "\"abc123\"" {
		t.Errorf("expected ETag \"abc123\", got %s", response.Snapshot.ETag)
	}
}

func TestSyncDeltaResponse_Structure(t *testing.T) {
	delta := types.DeltaUpdate{
		SinceVersion: "1000",
		ToVersion:    "2000",
	}

	response := types.SyncDeltaResponse{
		Delta: delta,
	}

	if response.Delta.SinceVersion != "1000" {
		t.Errorf("expected since version 1000, got %s", response.Delta.SinceVersion)
	}
	if response.Delta.ToVersion != "2000" {
		t.Errorf("expected to version 2000, got %s", response.Delta.ToVersion)
	}
}
