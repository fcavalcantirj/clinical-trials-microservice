package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/clinical-trials-microservice/internal/api"
	"github.com/clinical-trials-microservice/internal/cache"
	"github.com/clinical-trials-microservice/internal/middleware"
	"github.com/clinical-trials-microservice/internal/models"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// TrialsHandler handles trial-related HTTP requests
type TrialsHandler struct {
	apiClient    *api.ClinicalTrialsClient
	cache        *cache.Cache
	cacheEnabled bool
}

// NewTrialsHandler creates a new trials handler
func NewTrialsHandler(apiClient *api.ClinicalTrialsClient, cache *cache.Cache, cacheEnabled bool) *TrialsHandler {
	return &TrialsHandler{
		apiClient:    apiClient,
		cache:        cache,
		cacheEnabled: cacheEnabled,
	}
}

// SearchTrials handles GET /api/v1/trials/search
func (h *TrialsHandler) SearchTrials(w http.ResponseWriter, r *http.Request) {
	req := h.parseSearchRequest(r)
	ctx := r.Context()
	logger := getLogger(ctx)

	// Log search parameters
	logger.Info().
		Strs("conditions", req.Conditions).
		Strs("status", req.Status).
		Strs("phase", req.Phase).
		Int("page_size", req.PageSize).
		Msg("Search trials request")

	// Check cache if enabled
	var response *models.SearchResponse
	var err error
	cacheHit := false

	if h.cacheEnabled {
		cacheKey := h.generateCacheKey("search", req)
		if cached, found := h.cache.Get(cacheKey); found {
			if cachedResp, ok := cached.(*models.SearchResponse); ok {
				cacheHit = true
				logger.Info().
					Str("cache_key", cacheKey).
					Int("total_count", cachedResp.TotalCount).
					Msg("Cache hit")
				h.writeJSON(w, http.StatusOK, cachedResp)
				return
			}
		}
	}

	// Make API call
	response, err = h.apiClient.SearchTrials(req)
	if err != nil {
		logger.Error().
			Err(err).
			Bool("cache_hit", cacheHit).
			Msg("Error searching trials")
		h.writeError(w, http.StatusInternalServerError, "Failed to search trials: "+err.Error())
		return
	}

	// Store in cache if enabled
	if h.cacheEnabled {
		cacheKey := h.generateCacheKey("search", req)
		h.cache.Set(cacheKey, response)
	}

	// Log successful response
	logger.Info().
		Bool("cache_hit", cacheHit).
		Int("total_count", response.TotalCount).
		Int("trials_returned", len(response.Trials)).
		Msg("Search trials completed")

	h.writeJSON(w, http.StatusOK, response)
}

// GetTrialByID handles GET /api/v1/trials/{nct_id}
func (h *TrialsHandler) GetTrialByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nctID := vars["nct_id"]
	ctx := r.Context()
	logger := getLogger(ctx)

	if nctID == "" {
		logger.Warn().Msg("NCT ID is required")
		h.writeError(w, http.StatusBadRequest, "NCT ID is required")
		return
	}

	logger.Info().Str("nct_id", nctID).Msg("Get trial by ID request")

	// Check cache if enabled
	var trial *models.Trial
	var err error
	cacheHit := false

	if h.cacheEnabled {
		cacheKey := "trial:" + nctID
		if cached, found := h.cache.Get(cacheKey); found {
			if cachedTrial, ok := cached.(*models.Trial); ok {
				cacheHit = true
				logger.Info().
					Str("nct_id", nctID).
					Str("cache_key", cacheKey).
					Msg("Cache hit")
				h.writeJSON(w, http.StatusOK, cachedTrial)
				return
			}
		}
	}

	// Make API call
	trial, err = h.apiClient.GetTrialDetails(nctID)
	if err != nil {
		logger.Error().
			Err(err).
			Str("nct_id", nctID).
			Bool("cache_hit", cacheHit).
			Msg("Error getting trial details")
		h.writeError(w, http.StatusNotFound, "Trial not found: "+err.Error())
		return
	}

	// Store in cache if enabled
	if h.cacheEnabled {
		cacheKey := "trial:" + nctID
		h.cache.Set(cacheKey, trial)
	}

	logger.Info().
		Str("nct_id", nctID).
		Bool("cache_hit", cacheHit).
		Str("title", trial.Title).
		Msg("Get trial completed")

	h.writeJSON(w, http.StatusOK, trial)
}

// SearchTrialsPost handles POST /api/v1/trials/search (with JSON body)
func (h *TrialsHandler) SearchTrialsPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := getLogger(ctx)

	var req models.SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn().Err(err).Msg("Invalid request body")
		h.writeError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Log search parameters
	logger.Info().
		Strs("conditions", req.Conditions).
		Strs("status", req.Status).
		Strs("phase", req.Phase).
		Int("page_size", req.PageSize).
		Msg("POST search trials request")

	// Use same logic as GET handler (without cache for POST - can add later if needed)
	response, err := h.apiClient.SearchTrials(req)
	if err != nil {
		logger.Error().Err(err).Msg("Error searching trials")
		h.writeError(w, http.StatusInternalServerError, "Failed to search trials: "+err.Error())
		return
	}

	logger.Info().
		Int("total_count", response.TotalCount).
		Int("trials_returned", len(response.Trials)).
		Msg("POST search trials completed")

	h.writeJSON(w, http.StatusOK, response)
}

// Health handles GET /health
func (h *TrialsHandler) Health(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "healthy"})
}

// parseSearchRequest parses query parameters into a SearchRequest
func (h *TrialsHandler) parseSearchRequest(r *http.Request) models.SearchRequest {
	req := models.SearchRequest{
		PageSize: 100, // Default page size
	}

	// Query/Conditions
	if query := r.URL.Query().Get("query"); query != "" {
		req.Query = query
	}
	if conditions := r.URL.Query().Get("conditions"); conditions != "" {
		req.Conditions = strings.Split(conditions, ",")
		for i := range req.Conditions {
			req.Conditions[i] = strings.TrimSpace(req.Conditions[i])
		}
	}

	// Status
	if status := r.URL.Query().Get("status"); status != "" {
		req.Status = strings.Split(status, ",")
		for i := range req.Status {
			req.Status[i] = strings.TrimSpace(req.Status[i])
		}
	}

	// Phase
	if phase := r.URL.Query().Get("phase"); phase != "" {
		req.Phase = strings.Split(phase, ",")
		for i := range req.Phase {
			req.Phase[i] = strings.TrimSpace(req.Phase[i])
		}
	}

	// Location (latitude/longitude)
	if latStr := r.URL.Query().Get("latitude"); latStr != "" {
		if lat, err := strconv.ParseFloat(latStr, 64); err == nil {
			req.Latitude = lat
		}
	}
	if lonStr := r.URL.Query().Get("longitude"); lonStr != "" {
		if lon, err := strconv.ParseFloat(lonStr, 64); err == nil {
			req.Longitude = lon
		}
	}
	if distStr := r.URL.Query().Get("distance"); distStr != "" {
		if dist, err := strconv.Atoi(distStr); err == nil {
			req.Distance = dist
		}
	}

	// Age filters
	if minAge := r.URL.Query().Get("minimum_age"); minAge != "" {
		req.MinimumAge = minAge
	}
	if maxAge := r.URL.Query().Get("maximum_age"); maxAge != "" {
		req.MaximumAge = maxAge
	}

	// Pagination
	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 {
			req.PageSize = pageSize
		}
	}
	if pageToken := r.URL.Query().Get("page_token"); pageToken != "" {
		req.PageToken = pageToken
	}

	return req
}

// generateCacheKey generates a cache key from search request
func (h *TrialsHandler) generateCacheKey(prefix string, req models.SearchRequest) string {
	params := map[string]interface{}{
		"query":      req.Query,
		"conditions": req.Conditions,
		"status":     req.Status,
		"phase":      req.Phase,
		"page_token": req.PageToken,
		"page_size":  req.PageSize,
	}
	if req.Latitude != 0 {
		params["lat"] = req.Latitude
	}
	if req.Longitude != 0 {
		params["lon"] = req.Longitude
	}
	if req.Distance != 0 {
		params["distance"] = req.Distance
	}
	return cache.GenerateCacheKey(prefix, params)
}

// writeJSON writes a JSON response
func (h *TrialsHandler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Error().Err(err).Msg("Error encoding JSON response")
	}
}

// getLogger extracts logger from context with request ID
func getLogger(ctx context.Context) zerolog.Logger {
	requestID := ctx.Value(middleware.RequestIDKey{})
	if requestID != nil {
		if id, ok := requestID.(string); ok {
			return log.With().Str("request_id", id).Logger()
		}
	}
	return log.Logger
}

// writeError writes an error response
func (h *TrialsHandler) writeError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
