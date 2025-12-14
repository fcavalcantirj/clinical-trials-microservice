package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/clinical-trials-microservice/internal/models"
	"github.com/rs/zerolog/log"
)

const (
	// ClinicalTrialsGovBaseURL is the base URL for the API v2
	ClinicalTrialsGovBaseURL = "https://clinicaltrials.gov/api/v2/studies"
	// DefaultRateLimitDelay is the delay between requests to respect rate limits
	DefaultRateLimitDelay = time.Second * 2 // 50 requests/min = ~1.2 sec per request, use 2 for safety
)

// ClinicalTrialsClient handles interactions with ClinicalTrials.gov API
type ClinicalTrialsClient struct {
	baseURL     string
	httpClient  *http.Client
	rateLimiter chan struct{}
	lastRequest time.Time
	minDelay    time.Duration
}

// NewClinicalTrialsClient creates a new client instance
func NewClinicalTrialsClient() *ClinicalTrialsClient {
	rateLimiter := make(chan struct{}, 1)
	rateLimiter <- struct{}{} // Allow first request immediately

	return &ClinicalTrialsClient{
		baseURL:     ClinicalTrialsGovBaseURL,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		rateLimiter: rateLimiter,
		minDelay:    DefaultRateLimitDelay,
		lastRequest: time.Now().Add(-DefaultRateLimitDelay),
	}
}

// rateLimit ensures we respect the API rate limits (50 requests/min)
func (c *ClinicalTrialsClient) rateLimit() {
	elapsed := time.Since(c.lastRequest)
	if elapsed < c.minDelay {
		time.Sleep(c.minDelay - elapsed)
	}
	c.lastRequest = time.Now()
}

// SearchTrials searches for clinical trials based on the provided criteria
func (c *ClinicalTrialsClient) SearchTrials(req models.SearchRequest) (*models.SearchResponse, error) {
	start := time.Now()
	c.rateLimit()

	queryParams := c.buildQueryParams(req)
	fullURL := fmt.Sprintf("%s?%s", c.baseURL, queryParams.Encode())

	// Log outbound API call
	baseLogger := log.With().
		Str("api", "clinicaltrials.gov").
		Str("method", "GET").
		Str("url", fullURL).
		Strs("conditions", req.Conditions).
		Strs("status", req.Status).
		Logger()

	resp, err := c.httpClient.Get(fullURL)
	duration := time.Since(start)

	if err != nil {
		baseLogger.Error().
			Err(err).
			Int64("duration_ms", duration.Milliseconds()).
			Msg("External API call failed")
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		baseLogger.Error().
			Int("status_code", resp.StatusCode).
			Int64("duration_ms", duration.Milliseconds()).
			Msg("Rate limit exceeded from external API")
		return nil, fmt.Errorf("rate limit exceeded: HTTP 429")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		baseLogger.Error().
			Int("status_code", resp.StatusCode).
			Int64("duration_ms", duration.Milliseconds()).
			Str("response_body", string(body)).
			Msg("External API returned error status")
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var apiResponse ClinicalTrialsGovResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		baseLogger.Error().
			Err(err).
			Int("status_code", resp.StatusCode).
			Int64("duration_ms", duration.Milliseconds()).
			Msg("Failed to decode external API response")
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	baseLogger.Info().
		Int("status_code", resp.StatusCode).
		Int64("duration_ms", duration.Milliseconds()).
		Int("total_count", apiResponse.TotalCount).
		Int("studies_returned", len(apiResponse.Studies)).
		Msg("External API call completed")

	return c.convertToSearchResponse(&apiResponse, req), nil
}

// buildQueryParams constructs query parameters for the API request
func (c *ClinicalTrialsClient) buildQueryParams(req models.SearchRequest) url.Values {
	params := url.Values{}
	params.Set("format", "json")
	params.Set("countTotal", "true")

	// Build condition query (default to SCI-related if not provided)
	if len(req.Conditions) > 0 {
		conditions := strings.Join(req.Conditions, " OR ")
		params.Set("query.cond", conditions)
	} else if req.Query != "" {
		params.Set("query.cond", req.Query)
	} else {
		// Default SCI search terms
		params.Set("query.cond", "spinal cord injury OR quadriplegia OR tetraplegia OR paraplegia")
	}

	// Status filter
	if len(req.Status) > 0 {
		statusFilter := strings.Join(req.Status, ",")
		params.Set("filter.overallStatus", statusFilter)
	} else {
		// Default to recruiting and not yet recruiting
		params.Set("filter.overallStatus", "RECRUITING,NOT_YET_RECRUITING")
	}

	// Phase filter: Note - API v2 doesn't support filter.phase parameter
	// Phase filtering is done client-side after receiving results

	// Location-based search
	if req.Latitude != 0 && req.Longitude != 0 {
		distance := req.Distance
		if distance == 0 {
			distance = 50 // Default 50 miles
		}
		geoFilter := fmt.Sprintf("distance(%f,%f,%dmi)", req.Latitude, req.Longitude, distance)
		params.Set("filter.geo", geoFilter)
	}

	// Pagination
	if req.PageSize > 0 {
		params.Set("pageSize", fmt.Sprintf("%d", req.PageSize))
	} else {
		params.Set("pageSize", "100") // Default page size
	}

	if req.PageToken != "" {
		params.Set("pageToken", req.PageToken)
	}

	return params
}

// ClinicalTrialsGovResponse represents the API response structure
type ClinicalTrialsGovResponse struct {
	Studies       []StudyData `json:"studies"`
	NextPageToken string      `json:"nextPageToken,omitempty"`
	TotalCount    int         `json:"totalCount"`
}

// StudyData represents a study in the API response
type StudyData struct {
	ProtocolSection ProtocolSection `json:"protocolSection"`
	DerivedSection  DerivedSection  `json:"derivedSection,omitempty"`
}

// ProtocolSection contains the main study information
type ProtocolSection struct {
	IdentificationModule       IdentificationModule       `json:"identificationModule"`
	StatusModule               StatusModule               `json:"statusModule"`
	DesignModule               DesignModule               `json:"designModule,omitempty"`
	ConditionsModule           ConditionsModule           `json:"conditionsModule,omitempty"`
	EligibilityModule          EligibilityModule          `json:"eligibilityModule,omitempty"`
	ContactsLocationsModule    ContactsLocationsModule    `json:"contactsLocationsModule,omitempty"`
	DescriptionModule          DescriptionModule          `json:"descriptionModule,omitempty"`
	SponsorCollaboratorsModule SponsorCollaboratorsModule `json:"sponsorCollaboratorsModule,omitempty"`
}

// IdentificationModule contains identification information
type IdentificationModule struct {
	NCTID         string `json:"nctId"`
	BriefTitle    string `json:"briefTitle,omitempty"`
	OfficialTitle string `json:"officialTitle,omitempty"`
}

// StatusModule contains status information
type StatusModule struct {
	OverallStatus        string               `json:"overallStatus,omitempty"`
	StartDateStruct      StartDateStruct      `json:"startDateStruct,omitempty"`
	CompletionDateStruct CompletionDateStruct `json:"completionDateStruct,omitempty"`
}

// StartDateStruct contains start date information
type StartDateStruct struct {
	Date string `json:"date,omitempty"`
}

// CompletionDateStruct contains completion date information
type CompletionDateStruct struct {
	Date string `json:"date,omitempty"`
}

// DesignModule contains design and phase information
type DesignModule struct {
	Phases []string `json:"phases,omitempty"`
}

// ConditionsModule contains condition information
type ConditionsModule struct {
	Conditions []string `json:"conditions,omitempty"`
}

// EligibilityModule contains eligibility criteria
type EligibilityModule struct {
	EligibilityCriteria string          `json:"eligibilityCriteria,omitempty"`
	HealthyVolunteers   json.RawMessage `json:"healthyVolunteers,omitempty"` // Can be bool or string
	Gender              string          `json:"sex,omitempty"`               // API uses "sex" not "gender"
	MinimumAge          string          `json:"minimumAge,omitempty"`
	MaximumAge          string          `json:"maximumAge,omitempty"`
}

// getHealthyVolunteersString converts the healthyVolunteers field to string
func (e *EligibilityModule) getHealthyVolunteersString() string {
	if len(e.HealthyVolunteers) == 0 {
		return ""
	}

	// Try to unmarshal as bool first
	var b bool
	if err := json.Unmarshal(e.HealthyVolunteers, &b); err == nil {
		return strconv.FormatBool(b)
	}

	// Try as string
	var s string
	if err := json.Unmarshal(e.HealthyVolunteers, &s); err == nil {
		return s
	}

	return ""
}

// ContactsLocationsModule contains contacts and locations
type ContactsLocationsModule struct {
	Contacts  Contacts       `json:"contacts,omitempty"`
	Locations []LocationData `json:"locations,omitempty"`
}

// Contacts contains contact information
type Contacts struct {
	CentralContacts []CentralContact `json:"centralContacts,omitempty"`
}

// CentralContact represents a central contact
type CentralContact struct {
	Name  string `json:"name,omitempty"`
	Phone string `json:"phone,omitempty"`
	Email string `json:"email,omitempty"`
}

// LocationData represents a location in the API
type LocationData struct {
	Facility string   `json:"facility,omitempty"` // Can be string or object
	City     string   `json:"city,omitempty"`
	State    string   `json:"state,omitempty"`
	Zip      string   `json:"zip,omitempty"` // API uses "zip" not "zipCode"
	Country  string   `json:"country,omitempty"`
	GeoPoint GeoPoint `json:"geoPoint,omitempty"` // API uses "geoPoint" not "geographic"
}

// GeoPoint contains geographic coordinates
type GeoPoint struct {
	Lat float64 `json:"lat,omitempty"` // API uses "lat" not "latitude"
	Lon float64 `json:"lon,omitempty"` // API uses "lon" not "longitude"
}

// DescriptionModule contains description information
type DescriptionModule struct {
	BriefSummary        string `json:"briefSummary,omitempty"`
	DetailedDescription string `json:"detailedDescription,omitempty"`
}

// DerivedSection contains derived/calculated data
type DerivedSection struct {
	MiscInfoModule MiscInfoModule `json:"miscInfoModule,omitempty"`
}

// MiscInfoModule contains miscellaneous information
type MiscInfoModule struct {
	SponsorCollaboratorsModule SponsorCollaboratorsModule `json:"sponsorCollaboratorsModule,omitempty"`
}

// SponsorCollaboratorsModule contains sponsor information
type SponsorCollaboratorsModule struct {
	LeadSponsor LeadSponsor `json:"leadSponsor,omitempty"`
}

// LeadSponsor represents the lead sponsor
type LeadSponsor struct {
	Name  string `json:"name,omitempty"`
	Class string `json:"class,omitempty"` // API uses "class" not "type" or "category"
}

// convertToSearchResponse converts the API response to our internal model
func (c *ClinicalTrialsClient) convertToSearchResponse(apiResp *ClinicalTrialsGovResponse, req models.SearchRequest) *models.SearchResponse {
	trials := make([]models.Trial, 0, len(apiResp.Studies))
	originalCount := len(apiResp.Studies)

	for _, study := range apiResp.Studies {
		trial := c.convertStudyToTrial(study)

		// Apply client-side phase filtering if requested
		if len(req.Phase) > 0 {
			if !c.matchesPhaseFilter(trial.Phase, req.Phase) {
				continue // Skip this trial if it doesn't match phase filter
			}
		}

		// Apply client-side age filtering if requested
		if req.MinimumAge != "" || req.MaximumAge != "" {
			if !c.matchesAgeFilter(trial.Eligibility.MinimumAge, trial.Eligibility.MaximumAge, req.MinimumAge, req.MaximumAge) {
				continue // Skip this trial if it doesn't match age filter
			}
		}

		trials = append(trials, trial)
	}

	// Track filtering for logging
	phaseFiltered := len(req.Phase) > 0
	ageFiltered := req.MinimumAge != "" || req.MaximumAge != ""
	filteredCount := len(trials)

	// Log if client-side phase filtering was applied
	if phaseFiltered && filteredCount != originalCount {
		log.Info().
			Strs("requested_phases", req.Phase).
			Int("original_count", originalCount).
			Int("filtered_count", filteredCount).
			Msg("Applied client-side phase filtering")
	}

	// Log if client-side age filtering was applied
	if ageFiltered && filteredCount != originalCount {
		log.Info().
			Str("requested_min_age", req.MinimumAge).
			Str("requested_max_age", req.MaximumAge).
			Int("original_count", originalCount).
			Int("filtered_count", filteredCount).
			Msg("Applied client-side age filtering")
	}

	return &models.SearchResponse{
		Trials:        trials,
		TotalCount:    len(trials), // Note: This is filtered count, not API total
		NextPageToken: apiResp.NextPageToken,
		PageSize:      len(trials),
	}
}

// matchesPhaseFilter checks if a trial's phases match any of the requested phases
func (c *ClinicalTrialsClient) matchesPhaseFilter(trialPhases []string, requestedPhases []string) bool {
	// If no phases in trial, it doesn't match (unless "NA" is requested)
	if len(trialPhases) == 0 {
		return containsPhase(requestedPhases, "NA")
	}

	// Check if any trial phase matches any requested phase (case-insensitive)
	for _, trialPhase := range trialPhases {
		for _, requestedPhase := range requestedPhases {
			if strings.EqualFold(trialPhase, requestedPhase) {
				return true
			}
		}
	}

	return false
}

// containsPhase checks if a phase exists in the slice (case-insensitive)
func containsPhase(phases []string, phase string) bool {
	for _, p := range phases {
		if strings.EqualFold(p, phase) {
			return true
		}
	}
	return false
}

// parseAgeYears parses an age string and returns the numeric value in years
// Handles formats like "18 Years", "18", "18Y", "18 Y", etc.
// Returns 0 if parsing fails
func parseAgeYears(ageStr string) int {
	if ageStr == "" {
		return 0
	}

	// Remove common words and whitespace
	ageStr = strings.TrimSpace(ageStr)
	ageStr = strings.ToLower(ageStr)
	ageStr = strings.TrimSuffix(ageStr, "years")
	ageStr = strings.TrimSuffix(ageStr, "year")
	ageStr = strings.TrimSuffix(ageStr, "y")
	ageStr = strings.TrimSpace(ageStr)

	// Extract first numeric value
	for i := 0; i < len(ageStr); i++ {
		if ageStr[i] >= '0' && ageStr[i] <= '9' {
			// Found start of number, extract it
			numStr := ""
			for j := i; j < len(ageStr) && ageStr[j] >= '0' && ageStr[j] <= '9'; j++ {
				numStr += string(ageStr[j])
			}
			if num, err := strconv.Atoi(numStr); err == nil {
				return num
			}
			break
		}
	}

	return 0
}

// matchesAgeFilter checks if a trial's age range matches the requested age filters
// Age matching rules:
// - If minimum_age specified: trial's maximum_age must be >= requested minimum_age (or trial has no upper limit)
// - If maximum_age specified: trial's minimum_age must be <= requested maximum_age (or trial has no lower limit)
// - If both specified: trial must overlap with requested range
// - If trial has no age data: include by default (don't exclude)
func (c *ClinicalTrialsClient) matchesAgeFilter(trialMinAge, trialMaxAge, requestedMinAge, requestedMaxAge string) bool {
	// Parse ages to integers
	reqMin := parseAgeYears(requestedMinAge)
	reqMax := parseAgeYears(requestedMaxAge)
	trialMin := parseAgeYears(trialMinAge)
	trialMax := parseAgeYears(trialMaxAge)

	// If no age filters requested, include all trials
	if reqMin == 0 && reqMax == 0 {
		return true
	}

	// If trial has no age data, include it by default (we can't exclude it)
	if trialMin == 0 && trialMax == 0 {
		return true
	}

	// Apply minimum age filter
	if reqMin > 0 {
		// Trial must accept people at least reqMin years old
		// This means trial's max age must be >= reqMin (or no upper limit)
		if trialMax > 0 && trialMax < reqMin {
			return false // Trial's upper limit is below requested minimum
		}
		// If trial has no upper limit (trialMax == 0) but has lower limit, check if it accepts reqMin
		// For example, if trial is "18+ Years" (min=18, max=0) and reqMin=20, it matches
		// If trial is "18+ Years" and reqMin=15, it matches too (18+ includes 18)
		// So if trialMin <= reqMin, it's fine (trial accepts from trialMin, and reqMin >= trialMin)
		if trialMin > 0 && trialMin > reqMin {
			return false // Trial's minimum age is above requested minimum
		}
	}

	// Apply maximum age filter
	if reqMax > 0 {
		// User wants trials that accept people up to reqMax years old
		// This means: trialMax must be >= reqMax (trial accepts people up to trialMax, where trialMax >= reqMax)
		// OR trial has no upper limit (trialMax == 0) - include those as they accept people of any age
		if trialMin > 0 && trialMin > reqMax {
			return false // Trial's lower limit is above requested maximum (e.g., trial min=60, user wants max=50)
		}
		// If trial has a max age limit, it must be >= requested max (trial accepts people up to trialMax, so if trialMax >= reqMax, it accepts reqMax-year-olds)
		// Example: user wants max=50, trial max=80 → matches (trial accepts up to 80, which includes 50-year-olds)
		// Example: user wants max=50, trial max=40 → doesn't match (trial only accepts up to 40, which doesn't include 50-year-olds)
		// Example: user wants max=50, trial max=0 (no limit) → matches (trial has no upper limit, so it accepts 50-year-olds)
		if trialMax > 0 {
			if trialMax < reqMax {
				return false // Trial's maximum age is below requested maximum
			}
		}
		// If trialMax == 0 (no upper limit), include it as it accepts people of any age including reqMax
	}

	// If we get here, the trial's age range overlaps with the requested range
	return true
}

// convertStudyToTrial converts a study from the API to our Trial model
func (c *ClinicalTrialsClient) convertStudyToTrial(study StudyData) models.Trial {
	protocol := study.ProtocolSection

	trial := models.Trial{
		NCTID:    protocol.IdentificationModule.NCTID,
		Title:    protocol.IdentificationModule.BriefTitle,
		Status:   protocol.StatusModule.OverallStatus,
		Registry: "clinicaltrials.gov",
		URL:      fmt.Sprintf("https://clinicaltrials.gov/study/%s", protocol.IdentificationModule.NCTID),
	}

	// Phase
	if protocol.DesignModule.Phases != nil {
		trial.Phase = protocol.DesignModule.Phases
	}

	// Conditions
	if protocol.ConditionsModule.Conditions != nil {
		trial.Conditions = protocol.ConditionsModule.Conditions
	}

	// Dates
	if protocol.StatusModule.StartDateStruct.Date != "" {
		trial.StartDate = protocol.StatusModule.StartDateStruct.Date
	}
	if protocol.StatusModule.CompletionDateStruct.Date != "" {
		trial.CompletionDate = protocol.StatusModule.CompletionDateStruct.Date
	}

	// Eligibility
	if protocol.EligibilityModule.EligibilityCriteria != "" {
		trial.Eligibility.Criteria = protocol.EligibilityModule.EligibilityCriteria
	}
	trial.Eligibility.MinimumAge = protocol.EligibilityModule.MinimumAge
	trial.Eligibility.MaximumAge = protocol.EligibilityModule.MaximumAge
	trial.Eligibility.Gender = protocol.EligibilityModule.Gender

	// Locations
	if protocol.ContactsLocationsModule.Locations != nil {
		trial.Locations = make([]models.Location, 0, len(protocol.ContactsLocationsModule.Locations))
		for _, loc := range protocol.ContactsLocationsModule.Locations {
			location := models.Location{
				City:    loc.City,
				State:   loc.State,
				Country: loc.Country,
				ZipCode: loc.Zip,
			}
			if loc.GeoPoint.Lat != 0 {
				location.Latitude = loc.GeoPoint.Lat
			}
			if loc.GeoPoint.Lon != 0 {
				location.Longitude = loc.GeoPoint.Lon
			}
			trial.Locations = append(trial.Locations, location)
		}
	}

	// Contacts
	if protocol.ContactsLocationsModule.Contacts.CentralContacts != nil {
		trial.Contacts = make([]models.Contact, 0, len(protocol.ContactsLocationsModule.Contacts.CentralContacts))
		for _, contact := range protocol.ContactsLocationsModule.Contacts.CentralContacts {
			trial.Contacts = append(trial.Contacts, models.Contact{
				Name:  contact.Name,
				Phone: contact.Phone,
				Email: contact.Email,
			})
		}
	}

	// Sponsor (from protocolSection, not derivedSection)
	if protocol.SponsorCollaboratorsModule.LeadSponsor.Name != "" {
		trial.Sponsor = models.Sponsor{
			Name:     protocol.SponsorCollaboratorsModule.LeadSponsor.Name,
			Type:     protocol.SponsorCollaboratorsModule.LeadSponsor.Class,
			Category: protocol.SponsorCollaboratorsModule.LeadSponsor.Class,
		}
	}

	// Description
	if protocol.DescriptionModule.BriefSummary != "" {
		trial.BriefSummary = protocol.DescriptionModule.BriefSummary
	}
	if protocol.DescriptionModule.DetailedDescription != "" {
		trial.DetailedSummary = protocol.DescriptionModule.DetailedDescription
	}

	return trial
}

// GetTrialDetails retrieves detailed information for a specific trial by NCT ID
func (c *ClinicalTrialsClient) GetTrialDetails(nctID string) (*models.Trial, error) {
	start := time.Now()
	c.rateLimit()

	fullURL := fmt.Sprintf("%s/%s", c.baseURL, nctID)
	params := url.Values{}
	params.Set("format", "json")
	fullURL = fmt.Sprintf("%s?%s", fullURL, params.Encode())

	// Log outbound API call
	baseLogger := log.With().
		Str("api", "clinicaltrials.gov").
		Str("method", "GET").
		Str("nct_id", nctID).
		Str("url", fullURL).
		Logger()

	resp, err := c.httpClient.Get(fullURL)
	duration := time.Since(start)

	if err != nil {
		baseLogger.Error().
			Err(err).
			Int64("duration_ms", duration.Milliseconds()).
			Msg("External API call failed")
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		baseLogger.Error().
			Int("status_code", resp.StatusCode).
			Int64("duration_ms", duration.Milliseconds()).
			Msg("Rate limit exceeded from external API")
		return nil, fmt.Errorf("rate limit exceeded: HTTP 429")
	}

	if resp.StatusCode != http.StatusOK {
		baseLogger.Warn().
			Int("status_code", resp.StatusCode).
			Int64("duration_ms", duration.Milliseconds()).
			Msg("Trial not found in external API")
		return nil, fmt.Errorf("trial not found: %s", nctID)
	}

	// Single trial endpoint returns the study directly, not wrapped in a response structure
	var studyData StudyData
	if err := json.NewDecoder(resp.Body).Decode(&studyData); err != nil {
		baseLogger.Error().
			Err(err).
			Int("status_code", resp.StatusCode).
			Int64("duration_ms", duration.Milliseconds()).
			Msg("Failed to decode external API response")
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	baseLogger.Info().
		Int("status_code", resp.StatusCode).
		Int64("duration_ms", duration.Milliseconds()).
		Msg("External API call completed")

	trial := c.convertStudyToTrial(studyData)
	return &trial, nil
}
