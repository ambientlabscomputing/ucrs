package router

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/ambientlabscomputing/ucrs/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// SlogLoggerMiddleware replaces gin's default stdout logger with the app slog logger.
func SlogLoggerMiddleware(appCtx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		if raw := c.Request.URL.RawQuery; raw != "" {
			path = path + "?" + raw
		}

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		logger := utils.GetLogger(appCtx)

		attrs := []slog.Attr{
			slog.String("method", c.Request.Method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.Duration("latency", latency),
			slog.String("client_ip", c.ClientIP()),
			slog.Int("bytes", c.Writer.Size()),
		}
		if errs := c.Errors.ByType(gin.ErrorTypePrivate).String(); errs != "" {
			attrs = append(attrs, slog.String("errors", errs))
		}

		switch {
		case status >= 500:
			logger.LogAttrs(appCtx, slog.LevelError, "request", attrs...)
		case status >= 400:
			logger.LogAttrs(appCtx, slog.LevelWarn, "request", attrs...)
		default:
			logger.LogAttrs(appCtx, slog.LevelInfo, "request", attrs...)
		}
	}
}

// SlogRecoveryMiddleware replaces gin.Recovery() so panics are captured via slog.
func SlogRecoveryMiddleware(appCtx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				utils.GetLogger(appCtx).Error("panic recovered",
					"error", err,
					"stack", string(debug.Stack()),
					"path", c.Request.URL.Path,
				)
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}

var jwks keyfunc.Keyfunc

// TraceIDMiddleware adds a trace ID to each request's context
func TraceIDMiddleware(appCtx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for existing trace ID in header
		traceID := c.GetHeader("X-Trace-ID")
		// If not found, generate a new one
		if traceID == "" {
			traceID = uuid.New().String()
		}
		ctx := utils.SetTraceID(c.Request.Context(), traceID)
		logger := utils.GetLogger(appCtx).With("trace_id", traceID)
		ctx = utils.WithLogger(ctx, logger)

		c.Request = c.Request.WithContext(ctx)
		c.Header("X-Trace-ID", traceID)

		c.Next()
	}
}

// CORSMiddleware handles CORS headers for cross-origin requests
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Organization-ID, X-Trace-ID")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// StartMiddleware initializes middleware components
func StartMiddleware(ctx context.Context, settings *utils.Settings) error {
	// Initialize JWKS for JWT validation
	if err := initJWKS(ctx, settings); err != nil {
		return fmt.Errorf("failed to initialize JWKS: %w", err)
	}

	return nil
}

func initJWKS(ctx context.Context, settings *utils.Settings) error {
	logger := utils.GetLogger(ctx)

	jwksURL := "https://" + settings.Auth.AuthDomain + "/.well-known/jwks.json"

	var err error
	jwks, err = keyfunc.NewDefaultCtx(ctx, []string{jwksURL})
	if err != nil {
		logger.Error("JWKS initialization error", "error", err)
		return err
	}

	return nil
}

// VerifyToken parses and validates a JWT string using the JWKS
func VerifyToken(tokenString string, settings *utils.Settings) (*jwt.Token, error) {
	if jwks == nil {
		return nil, fmt.Errorf("jwks is not initialized")
	}

	token, err := jwt.Parse(
		tokenString,
		jwks.Keyfunc,
		jwt.WithValidMethods([]string{"RS256"}),
	)

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is invalid")
	}

	return token, nil
}

// AuthMiddleware validates JWT tokens and checks required permissions
func AuthMiddleware(ctx context.Context, settings *utils.Settings, requiredPermissions []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := utils.GetLogger(ctx)

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Debug("missing auth header")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing Authorization header"})
			c.Abort()
			return
		}

		var tokenString string
		fmt.Sscanf(authHeader, "Bearer %s", &tokenString)

		// Verify JWT token
		token, err := VerifyToken(tokenString, settings)
		if err != nil {
			logger.Warn("token verification failed", "error", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			logger.Error("failed to extract claims")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			c.Abort()
			return
		}

		// Check permissions (scopes)
		scopes, ok := claims["scope"].(string)
		if !ok {
			logger.Warn("no scopes in token")
			c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			c.Abort()
			return
		}

		// Verify required permissions
		tokenScopes := splitScopes(scopes)
		for _, required := range requiredPermissions {
			if !contains(tokenScopes, required) {
				logger.Warn("missing required permission", "required", required, "scopes", tokenScopes)
				c.JSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("missing required permission: %s", required)})
				c.Abort()
				return
			}
		}

		// Store claims in context
		c.Set("claims", claims)

		c.Next()
	}
}

// AdminAuthMiddleware is a stricter auth middleware requiring admin:registry scope
func AdminAuthMiddleware(ctx context.Context, settings *utils.Settings) gin.HandlerFunc {
	return AuthMiddleware(ctx, settings, []string{"admin:registry"})
}

// ReadAuthMiddleware requires read:registry scope
func ReadAuthMiddleware(ctx context.Context, settings *utils.Settings) gin.HandlerFunc {
	return AuthMiddleware(ctx, settings, []string{"read:registry"})
}

// WriteAuthMiddleware requires write:registry scope
func WriteAuthMiddleware(ctx context.Context, settings *utils.Settings) gin.HandlerFunc {
	return AuthMiddleware(ctx, settings, []string{"write:registry"})
}

// Helper functions
func splitScopes(scopes string) []string {
	if scopes == "" {
		return []string{}
	}

	var result []string
	for _, scope := range splitString(scopes, " ") {
		if scope != "" {
			result = append(result, scope)
		}
	}

	return result
}

func splitString(s, sep string) []string {
	var parts []string
	start := 0

	for i := 0; i < len(s); i++ {
		if s[i:i+len(sep)] == sep {
			parts = append(parts, s[start:i])
			start = i + len(sep)
		}
	}

	parts = append(parts, s[start:])
	return parts
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
