package models

// Trial represents a clinical trial from ClinicalTrials.gov
type Trial struct {
	NCTID           string                 `json:"nct_id"`
	Title           string                 `json:"title"`
	Status          string                 `json:"status"`
	Phase           []string               `json:"phase,omitempty"`
	Conditions      []string               `json:"conditions,omitempty"`
	Locations       []Location             `json:"locations,omitempty"`
	Eligibility     Eligibility            `json:"eligibility,omitempty"`
	Sponsor         Sponsor                `json:"sponsor,omitempty"`
	Contacts        []Contact              `json:"contacts,omitempty"`
	StartDate       string                 `json:"start_date,omitempty"`
	CompletionDate  string                 `json:"completion_date,omitempty"`
	BriefSummary    string                 `json:"brief_summary,omitempty"`
	DetailedSummary string                 `json:"detailed_summary,omitempty"`
	URL             string                 `json:"url"`
	Registry        string                 `json:"registry"`
	AdditionalData  map[string]interface{} `json:"additional_data,omitempty"`
}

// Location represents a trial location
type Location struct {
	City      string  `json:"city,omitempty"`
	State     string  `json:"state,omitempty"`
	Country   string  `json:"country,omitempty"`
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
	ZipCode   string  `json:"zip_code,omitempty"`
}

// Eligibility represents trial eligibility criteria
type Eligibility struct {
	MinimumAge string `json:"minimum_age,omitempty"`
	MaximumAge string `json:"maximum_age,omitempty"`
	Gender     string `json:"gender,omitempty"`
	Criteria   string `json:"criteria,omitempty"`
}

// Sponsor represents trial sponsor information
type Sponsor struct {
	Name     string `json:"name,omitempty"`
	Type     string `json:"type,omitempty"`
	Category string `json:"category,omitempty"`
}

// Contact represents contact information
type Contact struct {
	Name  string `json:"name,omitempty"`
	Phone string `json:"phone,omitempty"`
	Email string `json:"email,omitempty"`
}

// SearchRequest represents a search request for trials
type SearchRequest struct {
	Query      string   `json:"query,omitempty"`
	Status     []string `json:"status,omitempty"`
	Phase      []string `json:"phase,omitempty"`
	Conditions []string `json:"conditions,omitempty"`
	Location   string   `json:"location,omitempty"` // "city, state" or "country"
	Latitude   float64  `json:"latitude,omitempty"`
	Longitude  float64  `json:"longitude,omitempty"`
	Distance   int      `json:"distance,omitempty"` // in miles
	MinimumAge string   `json:"minimum_age,omitempty"`
	MaximumAge string   `json:"maximum_age,omitempty"`
	PageSize   int      `json:"page_size,omitempty"`
	PageToken  string   `json:"page_token,omitempty"`
}

// SearchResponse represents the search results
type SearchResponse struct {
	Trials        []Trial `json:"trials"`
	TotalCount    int     `json:"total_count"`
	NextPageToken string  `json:"next_page_token,omitempty"`
	PageSize      int     `json:"page_size"`
}
