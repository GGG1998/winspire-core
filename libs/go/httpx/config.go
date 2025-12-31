package httpx

// Config holds configuration for HTTP middleware
type Config struct {
	// CORS settings
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int

	// Security headers
	EnableHSTS           bool
	HSTSMaxAge           int
	ContentSecurityPolicy string

	// Auth settings
	JWTSecret   string
	JWTIssuer   string
	JWTAudience string

	// Observability
	EnableTracing bool
	ServiceName   string
}

// DefaultConfig returns a configuration with sensible defaults for development
func DefaultConfig() Config {
	return Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Service"},
		ExposeHeaders:    []string{},
		AllowCredentials: false,
		MaxAge:           86400,

		EnableHSTS:            false,
		HSTSMaxAge:            31536000,
		ContentSecurityPolicy: "",

		EnableTracing: true,
	}
}

// ProductionConfig returns a configuration with strict security settings for production
func ProductionConfig() Config {
	return Config{
		AllowOrigins:     []string{}, // Must be explicitly set
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-ID", "X-Service"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           86400,

		EnableHSTS:            true,
		HSTSMaxAge:            31536000,
		ContentSecurityPolicy: "default-src 'self'",

		EnableTracing: true,
	}
}


