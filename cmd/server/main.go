package main

import (
	"flag"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/clinical-trials-microservice/internal/api"
	"github.com/clinical-trials-microservice/internal/cache"
	"github.com/clinical-trials-microservice/internal/handlers"
	"github.com/clinical-trials-microservice/internal/middleware"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// getEnv gets environment variable or returns default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	// Initialize structured logger
	initLogger()

	// Configuration flags
	port := flag.String("port", getEnv("PORT", "8080"), "Server port")
	cacheEnabled := flag.Bool("cache", true, "Enable caching")
	cacheTTL := flag.Duration("cache-ttl", 6*time.Hour, "Cache TTL duration")
	flag.Parse()

	// Initialize API client
	apiClient := api.NewClinicalTrialsClient()
	log.Info().Msg("ClinicalTrials.gov API client initialized")

	// Initialize cache
	var trialCache *cache.Cache
	if *cacheEnabled {
		trialCache = cache.NewCache(*cacheTTL)
		log.Info().Dur("ttl", *cacheTTL).Msg("Cache enabled")
	} else {
		trialCache = cache.NewCache(0) // Will use default
		log.Info().Msg("Cache disabled")
	}

	// Initialize handlers
	trialsHandler := handlers.NewTrialsHandler(apiClient, trialCache, *cacheEnabled)

	// Setup routes
	router := mux.NewRouter()

	// Add middleware (order matters - logging first to capture all requests)
	router.Use(middleware.LoggingMiddleware)
	router.Use(corsMiddleware)

	// Health check
	router.HandleFunc("/health", trialsHandler.Health).Methods("GET")

	// API routes
	apiRouter := router.PathPrefix("/api/v1").Subrouter()
	apiRouter.HandleFunc("/trials/search", trialsHandler.SearchTrials).Methods("GET")
	apiRouter.HandleFunc("/trials/search", trialsHandler.SearchTrialsPost).Methods("POST")
	apiRouter.HandleFunc("/trials/{nct_id}", trialsHandler.GetTrialByID).Methods("GET")

	// Start server
	addr := ":" + *port
	log.Info().
		Str("port", *port).
		Str("address", addr).
		Msg("Starting server")

	log.Info().Msg("API endpoints:")
	log.Info().Msg("  GET  /health")
	log.Info().Msg("  GET  /api/v1/trials/search")
	log.Info().Msg("  POST /api/v1/trials/search")
	log.Info().Msg("  GET  /api/v1/trials/{nct_id}")

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal().Err(err).Msg("Server failed to start")
	}
}

// initLogger initializes the structured logger
func initLogger() {
	// Set log level from environment variable
	logLevel := getEnv("LOG_LEVEL", "info")
	level, err := zerolog.ParseLevel(strings.ToLower(logLevel))
	if err != nil {
		level = zerolog.InfoLevel
		log.Warn().Str("level", logLevel).Msg("Invalid LOG_LEVEL, defaulting to info")
	}
	zerolog.SetGlobalLevel(level)

	// Set time format to RFC3339 for structured logs
	zerolog.TimeFieldFormat = time.RFC3339

	// In production, use JSON format. In development, use console format for readability
	logFormat := getEnv("LOG_FORMAT", "json")
	if logFormat == "console" || logFormat == "text" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	}

	log.Info().
		Str("level", level.String()).
		Str("format", logFormat).
		Msg("Logger initialized")
}

// corsMiddleware adds CORS headers to responses
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
