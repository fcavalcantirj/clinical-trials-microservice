package api

import (
	"strings"
	"testing"
	"time"

	"github.com/clinical-trials-microservice/internal/models"
)

func TestRateLimiting(t *testing.T) {
	client := NewClinicalTrialsClient()

	// Test that rate limiting respects delays
	start := time.Now()
	client.rateLimit()
	client.rateLimit()
	elapsed := time.Since(start)

	// Should have at least the minDelay between calls
	if elapsed < client.minDelay {
		t.Errorf("Rate limiting not working properly, elapsed: %v, expected at least: %v", elapsed, client.minDelay)
	}
}

func TestBuildQueryParams(t *testing.T) {
	client := NewClinicalTrialsClient()

	tests := []struct {
		name     string
		req      models.SearchRequest
		expected string
	}{
		{
			name: "default SCI search",
			req:  models.SearchRequest{},
		},
		{
			name: "custom conditions",
			req: models.SearchRequest{
				Conditions: []string{"spinal cord injury", "tetraplegia"},
			},
		},
		{
			name: "status filter",
			req: models.SearchRequest{
				Status: []string{"RECRUITING", "NOT_YET_RECRUITING"},
			},
		},
		{
			name: "location search",
			req: models.SearchRequest{
				Latitude:  34.0522,
				Longitude: -118.2437,
				Distance:  50,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := client.buildQueryParams(tt.req)

			// Check that format is always json
			if params.Get("format") != "json" {
				t.Errorf("Expected format=json, got %s", params.Get("format"))
			}

			// Check that countTotal is always true
			if params.Get("countTotal") != "true" {
				t.Errorf("Expected countTotal=true, got %s", params.Get("countTotal"))
			}
		})
	}
}

func TestBuildQueryParamsDefaultSCI(t *testing.T) {
	client := NewClinicalTrialsClient()
	req := models.SearchRequest{} // Empty request should default to SCI terms

	params := client.buildQueryParams(req)
	cond := params.Get("query.cond")

	if cond == "" {
		t.Errorf("Expected default SCI search terms, got empty condition")
		return
	}

	// Check that it contains at least one SCI-related term (case-insensitive check)
	expectedTerms := []string{"spinal cord injury", "quadriplegia", "tetraplegia", "paraplegia"}
	found := false
	condLower := strings.ToLower(cond)
	for _, term := range expectedTerms {
		// Check if the condition contains any of the expected terms (case-insensitive)
		if strings.Contains(condLower, strings.ToLower(term)) {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected SCI search terms in condition query, got: %s", cond)
	}
}

// Note: Integration tests that actually call the API should be in a separate file
// and can be run with: go test -tags=integration
