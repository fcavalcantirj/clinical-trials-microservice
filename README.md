# Clinical Trials Microservice

A fast, Go-based microservice for querying clinical trials data, specifically optimized for spinal cord injury (SCI) trials. This service integrates with the ClinicalTrials.gov API v2 to provide comprehensive trial search capabilities.

## Features

- 🔍 **Comprehensive Search**: Query clinical trials with multiple filters (conditions, status, phase, location, age)
- ⚡ **Fast & Efficient**: Built with Go for high performance and low latency
- 🔄 **Smart Caching**: In-memory caching to reduce API calls and improve response times
- 🛡️ **Rate Limiting**: Built-in rate limiting to respect ClinicalTrials.gov API limits (50 requests/min)
- 🌍 **Location Search**: Distance-based geographic search for finding nearby trials
- 📊 **Rich Data**: Returns comprehensive trial information including eligibility, locations, contacts, and more
- 🔌 **RESTful API**: Clean REST API with both GET and POST endpoints

## Architecture

```
clinical-trials-microservice/
├── cmd/
│   └── server/
│       └── main.go          # Application entry point
├── internal/
│   ├── api/
│   │   └── clinicaltrials.go  # ClinicalTrials.gov API client
│   ├── cache/
│   │   └── cache.go           # Caching layer
│   ├── handlers/
│   │   └── trials.go          # HTTP request handlers
│   └── models/
│       └── trial.go           # Data models
├── scripts/
│   ├── test_api.sh            # Bash test script
│   ├── test_api.py            # Python test script
│   └── test_direct_api.sh     # Direct API testing
└── research/
    └── trials_API_integration_guide_for_spinal_cord_injury.md
```

## Quick Reference

**Base URL:** `http://localhost:8080`

**Endpoints:**
- `GET /health` - Health check
- `GET /api/v1/trials/search` - Search trials with query parameters
- `POST /api/v1/trials/search` - Search trials with JSON body
- `GET /api/v1/trials/{nct_id}` - Get trial by NCT ID

**Common Filters:**
- `conditions` - Medical conditions (comma-separated)
- `status` - Trial status: `RECRUITING`, `NOT_YET_RECRUITING`, etc.
- `phase` - Trial phase: `PHASE1`, `PHASE2`, `PHASE3`, `PHASE4`
- `latitude` / `longitude` / `distance` - Location-based search
- `page_size` - Results per page (default: 100, max: 1000)

**Quick Example:**
```bash
curl "http://localhost:8080/api/v1/trials/search?status=RECRUITING&page_size=5"
```

## Quick Start

### Prerequisites

- Go 1.21 or higher (for local development)
- Docker (for containerized deployment)
- (Optional) Python 3.x for Python test scripts
- (Optional) `jq` for JSON formatting in bash scripts

### 🚀 Quick Deploy

**Deploy to Render (Recommended - Free tier available):**
1. Push your code to GitHub/GitLab/Bitbucket
2. Go to [render.com](https://render.com) → Sign up (free)
3. New → Blueprint → Connect your repository
4. Render auto-detects `render.yaml` → Click "Apply"
5. Done! Your service will be live at `*.onrender.com`

**Or deploy to Easypanel:**
1. Push your code to GitHub/GitLab
2. In Easypanel: New Project → Connect Git Repository
3. Select "Dockerfile" build method
4. Set port: `8080`, health check: `/health`
5. Deploy!

See [DEPLOYMENT.md](./DEPLOYMENT.md) for detailed instructions for both platforms.

### Installation

1. **Clone or navigate to the directory:**
   ```bash
   cd clinical-trials-microservice
   ```

2. **Download dependencies:**
   ```bash
   go mod download
   ```

3. **Run the server:**
   ```bash
   go run cmd/server/main.go
   ```

   Or with custom options:
   ```bash
   go run cmd/server/main.go -port 8080 -cache=true -cache-ttl=6h
   ```

4. **Server will start on `http://localhost:8080`**

### Build for Production

```bash
go build -o bin/trials-service cmd/server/main.go
./bin/trials-service
```

### Deploy with Docker

```bash
# Build Docker image
docker build -t clinical-trials-service .

# Run container
docker run -p 8080:8080 clinical-trials-service

# Or use docker-compose
docker-compose up -d
```

### Deploy to Cloud

**Recommended platforms:**
- **Render** (Free tier available) - Auto-deploys from `render.yaml`, see [DEPLOYMENT.md](./DEPLOYMENT.md)
- **Easypanel** - Docker-based deployment, see [DEPLOYMENT.md](./DEPLOYMENT.md)
- **Railway** - Docker support with free tier
- **Fly.io** - Global edge deployment

See [DEPLOYMENT.md](./DEPLOYMENT.md) for complete deployment guides.

## API Endpoints

### Health Check
```http
GET /health
```

**Response:**
```json
{
  "status": "healthy"
}
```

## How to Use

### Search Trials (GET)

The search endpoint accepts multiple query parameters to filter clinical trials.

**Endpoint:**
```http
GET /api/v1/trials/search
```

#### All Available Filters

| Parameter | Type | Description | Example | Default |
|-----------|------|-------------|---------|---------|
| `query` | string | Free text search query (alternative to conditions) | `spinal+cord+injury` | None (uses default SCI terms if no conditions) |
| `conditions` | string | Comma-separated list of medical conditions | `spinal+cord+injury,tetraplegia` | `spinal cord injury OR quadriplegia OR tetraplegia OR paraplegia` |
| `status` | string | Comma-separated trial status values | `RECRUITING,NOT_YET_RECRUITING` | `RECRUITING,NOT_YET_RECRUITING` |
| `phase` | string | Comma-separated trial phases | `PHASE2,PHASE3` | None (all phases) |
| `latitude` | float | Latitude for location-based search (requires longitude) | `34.0522` | None |
| `longitude` | float | Longitude for location-based search (requires latitude) | `-118.2437` | None |
| `distance` | integer | Distance in miles for location search | `50` | `50` |
| `minimum_age` | string | Minimum age requirement (client-side filter) | `18 Years` or `18` | None |
| `maximum_age` | string | Maximum age requirement (client-side filter) | `65 Years` or `65` | None |
| `page_size` | integer | Number of results per page | `10` | `100` (max: 1000) |
| `page_token` | string | Token for pagination (from previous response) | `eyJ...` | None |

#### Valid Status Values

- `RECRUITING` - Currently recruiting participants
- `NOT_YET_RECRUITING` - Not yet open for participant recruitment
- `ACTIVE_NOT_RECRUITING` - Study is ongoing but not recruiting
- `COMPLETED` - Study has completed
- `SUSPENDED` - Study has been stopped early
- `TERMINATED` - Study has been stopped early and will not resume
- `WITHDRAWN` - Study has been withdrawn prior to enrollment
- `UNKNOWN` - Status is unknown

#### Valid Phase Values

- `PHASE1` - Phase 1 clinical trials
- `PHASE2` - Phase 2 clinical trials
- `PHASE3` - Phase 3 clinical trials
- `PHASE4` - Phase 4 clinical trials
- `NA` - Not applicable (e.g., observational studies)
- `EARLY_PHASE1` - Early Phase 1 trials

#### Filter Implementation Details

Filters are applied in two ways depending on ClinicalTrials.gov API support:

**Server-Side Filters** (applied via API query parameters):
- `status` - Filtered by the API using `filter.overallStatus`
- `conditions` - Filtered by the API using `query.cond`
- `latitude`/`longitude`/`distance` - Filtered by the API using `filter.geo`
- `page_size`/`page_token` - Handled by the API for pagination

**Client-Side Filters** (applied after receiving API results):
- `phase` - The API v2 doesn't support phase filtering, so results are filtered client-side after fetching
- `minimum_age` / `maximum_age` - Age filtering is done client-side by parsing trial eligibility criteria

**Age Filter Semantics:**
- `minimum_age`: Returns trials where the trial's minimum age requirement is ≤ the requested minimum age (or the trial has no lower age limit)
- `maximum_age`: Returns trials where the trial's maximum age limit is ≥ the requested maximum age, or the trial has no upper age limit
- Age values are parsed from strings like `"18 Years"`, `"65"`, etc. - any numeric age value will work
- Both filters can be combined to find trials within a specific age range

**Note:** Client-side filtering means the API may return more results than the final filtered count, but ensures accurate filtering based on trial eligibility criteria.

#### Examples

**1. Basic Search (Default SCI Terms)**
```bash
curl "http://localhost:8080/api/v1/trials/search?page_size=5"
```

**2. Search by Specific Conditions**
```bash
curl "http://localhost:8080/api/v1/trials/search?conditions=spinal+cord+injury,tetraplegia&page_size=10"
```

**3. Find Only Recruiting Trials**
```bash
curl "http://localhost:8080/api/v1/trials/search?status=RECRUITING&page_size=20"
```

**4. Filter by Phase** (client-side filtering)
```bash
curl "http://localhost:8080/api/v1/trials/search?phase=PHASE2,PHASE3&status=RECRUITING"
```

**5. Filter by Age** (client-side filtering)
```bash
# Find trials for people aged 18-65
curl "http://localhost:8080/api/v1/trials/search?minimum_age=18&maximum_age=65&status=RECRUITING"

# Find trials for people 50 or younger
curl "http://localhost:8080/api/v1/trials/search?maximum_age=50&conditions=spinal+cord+injury"

# Find trials for people 21 or older
curl "http://localhost:8080/api/v1/trials/search?minimum_age=21&status=RECRUITING"
```

**6. Location-Based Search (within 50 miles of Los Angeles)**
```bash
curl "http://localhost:8080/api/v1/trials/search?latitude=34.0522&longitude=-118.2437&distance=50&page_size=10"
```

**7. Combined Filters** (server-side + client-side)
```bash
# Combining server-side filters (status, conditions, location) with client-side filters (phase, age)
curl "http://localhost:8080/api/v1/trials/search?conditions=spinal+cord+injury&status=RECRUITING&phase=PHASE2&minimum_age=18&maximum_age=65&latitude=40.7128&longitude=-74.0060&distance=25&page_size=15"
```

**8. Pagination (Get Next Page)**
```bash
# First request
curl "http://localhost:8080/api/v1/trials/search?page_size=10"

# Use next_page_token from response
curl "http://localhost:8080/api/v1/trials/search?page_size=10&page_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

#### Complete JSON Response Structure

**Search Response:**
```json
{
  "trials": [
    {
      "nct_id": "NCT06511934",
      "title": "Feasibility of the BrainGate2 Neural Interface System in Persons With Tetraplegia (BG-Tablet-01)",
      "status": "RECRUITING",
      "phase": ["NA"],
      "conditions": [
        "Brainstem Stroke",
        "ALS",
        "Tetraplegia",
        "Spinal Cord Injuries",
        "Cervical Spinal Cord Injury"
      ],
      "locations": [
        {
          "city": "Boston",
          "state": "Massachusetts",
          "country": "United States",
          "latitude": 42.35843,
          "longitude": -71.05977,
          "zip_code": "02114"
        }
      ],
      "eligibility": {
        "minimum_age": "18 Years",
        "maximum_age": "80 Years",
        "gender": "ALL",
        "criteria": "Inclusion Criteria:\n\n* Clinical diagnosis of spinal cord injury...\n\nExclusion Criteria:\n\n* Visual impairment..."
      },
      "sponsor": {
        "name": "Leigh R. Hochberg, MD, PhD.",
        "type": "OTHER",
        "category": "OTHER"
      },
      "contacts": [
        {
          "name": "Contact Name",
          "phone": "555-1234",
          "email": "contact@example.com"
        }
      ],
      "start_date": "2024-07-22",
      "completion_date": "2027-07-30",
      "brief_summary": "People with brainstem stroke, advanced amyotrophic lateral sclerosis...",
      "detailed_summary": "The goal of this project is to advance the methods...",
      "url": "https://clinicaltrials.gov/study/NCT06511934",
      "registry": "clinicaltrials.gov"
    }
  ],
  "total_count": 499,
  "next_page_token": "ZVNj7o2Elu8o3lpvDsvyv72tmpOQJJxuYPGl2Pg",
  "page_size": 10
}
```

**Response Fields Explained:**

- `trials` - Array of trial objects matching the search criteria
- `total_count` - Total number of trials matching the search (across all pages)
- `next_page_token` - Token to retrieve the next page of results (if available)
- `page_size` - Number of trials returned in this response

**Trial Object Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `nct_id` | string | Unique ClinicalTrials.gov identifier (e.g., "NCT06511934") |
| `title` | string | Brief title of the clinical trial |
| `status` | string | Current recruitment status (see valid status values above) |
| `phase` | array[string] | Trial phases (may be empty for observational studies) |
| `conditions` | array[string] | Medical conditions being studied |
| `locations` | array[Location] | Trial locations (may be empty) |
| `eligibility` | Eligibility | Eligibility criteria (age, gender, criteria) |
| `sponsor` | Sponsor | Lead sponsor information |
| `contacts` | array[Contact] | Contact information for the trial |
| `start_date` | string | Trial start date (YYYY-MM-DD format) |
| `completion_date` | string | Expected or actual completion date |
| `brief_summary` | string | Brief description of the trial |
| `detailed_summary` | string | Detailed description of the trial |
| `url` | string | Link to the trial on ClinicalTrials.gov |
| `registry` | string | Always "clinicaltrials.gov" |

**Location Object:**
```json
{
  "city": "Boston",
  "state": "Massachusetts",
  "country": "United States",
  "latitude": 42.35843,
  "longitude": -71.05977,
  "zip_code": "02114"
}
```

**Eligibility Object:**
```json
{
  "minimum_age": "18 Years",
  "maximum_age": "80 Years",
  "gender": "ALL",
  "criteria": "Inclusion Criteria:\n\n* Clinical diagnosis..."
}
```

**Sponsor Object:**
```json
{
  "name": "Sponsor Name",
  "type": "INDUSTRY",
  "category": "INDUSTRY"
}
```

Valid sponsor types: `INDUSTRY`, `NIH`, `US_FED`, `OTHER`, `NETWORK`, `AMERICAN_INDIAN_OR_ALASKAN_NATIVE_TRIBE`, `INDIVIDUAL`, `AMBIGUOUS`

**Contact Object:**
```json
{
  "name": "John Doe",
  "phone": "555-123-4567",
  "email": "john.doe@example.com"
}
```

### Search Trials (POST)

For complex searches, you can use POST with a JSON body. This is especially useful when you have many filters.

**Endpoint:**
```http
POST /api/v1/trials/search
Content-Type: application/json
```

**Request Body:**
```json
{
  "conditions": ["spinal cord injury", "tetraplegia"],
  "status": ["RECRUITING", "NOT_YET_RECRUITING"],
  "phase": ["PHASE2", "PHASE3"],
  "latitude": 34.0522,
  "longitude": -118.2437,
  "distance": 50,
  "minimum_age": "18 Years",
  "maximum_age": "65 Years",
  "page_size": 10,
  "page_token": "optional_token_from_previous_response"
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/api/v1/trials/search \
  -H "Content-Type: application/json" \
  -d '{
    "conditions": ["spinal cord injury"],
    "status": ["RECRUITING"],
    "phase": ["PHASE2", "PHASE3"],
    "latitude": 40.7128,
    "longitude": -74.0060,
    "distance": 25,
    "page_size": 10
  }'
```

The response format is identical to the GET endpoint.

### Get Trial by ID

Retrieve detailed information for a specific trial using its NCT ID.

**Endpoint:**
```http
GET /api/v1/trials/{nct_id}
```

**Example:**
```bash
curl "http://localhost:8080/api/v1/trials/NCT06511934"
```

**Response:**
Returns a single `Trial` object (same structure as items in the `trials` array from search results).

**Example Response:**
```json
{
  "nct_id": "NCT06511934",
  "title": "Feasibility of the BrainGate2 Neural Interface System in Persons With Tetraplegia (BG-Tablet-01)",
  "status": "RECRUITING",
  "phase": ["NA"],
  "conditions": ["Brainstem Stroke", "ALS", "Tetraplegia", "Spinal Cord Injuries"],
  "locations": [...],
  "eligibility": {...},
  "sponsor": {...},
  "contacts": [...],
  "start_date": "2024-07-22",
  "completion_date": "2027-07-30",
  "brief_summary": "...",
  "detailed_summary": "...",
  "url": "https://clinicaltrials.gov/study/NCT06511934",
  "registry": "clinicaltrials.gov"
}
```

## Configuration

### Command Line Flags

- `-port` - Server port (default: 8080)
- `-cache` - Enable caching (default: true)
- `-cache-ttl` - Cache TTL duration (default: 6h)
  - Examples: `6h`, `12h`, `1h30m`

### Environment Variables

Currently, the service uses default values. Future versions may support:
- `CLINICAL_TRIALS_PORT` - Server port
- `CLINICAL_TRIALS_CACHE_ENABLED` - Enable/disable cache
- `CLINICAL_TRIALS_CACHE_TTL` - Cache TTL

## Testing

### Bash Test Script

```bash
# Make sure server is running first
./scripts/test_api.sh

# Or with custom base URL
BASE_URL=http://localhost:8080 ./scripts/test_api.sh
```

### Python Test Script

```bash
# Install dependencies (if needed)
pip install requests

# Run comprehensive tests
python3 scripts/test_api.py
```

### Direct API Testing

Test the ClinicalTrials.gov API directly to understand its structure:

```bash
./scripts/test_direct_api.sh
```

Note: Requires `jq` for JSON formatting. Install with:
- macOS: `brew install jq`
- Linux: `apt-get install jq` or `yum install jq`

### Go Unit Tests

```bash
go test ./internal/api/...
```

## Default Search Behavior

**Important:** When no `conditions` or `query` parameters are provided, the service automatically searches for:
- `spinal cord injury OR quadriplegia OR tetraplegia OR paraplegia`

This ensures that SCI-related trials are found even without explicit search terms. To search for other conditions, always provide the `conditions` or `query` parameter.

**Examples:**
- No parameters: Searches for SCI terms by default
- `?conditions=cancer`: Searches for cancer trials only
- `?query=diabetes`: Searches using free-text query for diabetes

## Rate Limiting

The service implements rate limiting to respect ClinicalTrials.gov API limits:
- **Limit**: ~50 requests per minute per IP
- **Implementation**: 2-second delay between requests (conservative)
- **Error Handling**: Returns HTTP 429 if rate limit is exceeded

## Caching Strategy

- **Default TTL**: 6 hours (configurable)
- **Cache Keys**: Based on search parameters
- **Cache Type**: In-memory (can be extended to Redis)
- **Cache Invalidation**: Automatic based on TTL

## Data Sources

### Primary: ClinicalTrials.gov API v2
- **Base URL**: `https://clinicaltrials.gov/api/v2/studies`
- **Authentication**: None required (public API)
- **Rate Limit**: ~50 requests/minute
- **Documentation**: https://clinicaltrials.gov/data-api

### Future Integrations (from research doc)
- ReBEC (Brazilian registry) - XML export available
- AACT Database - PostgreSQL access for advanced analytics
- WHO ICTRP - Requires partnership agreement

## API Keys & Authentication

**No API keys required!** The ClinicalTrials.gov API v2 is completely public and free to use. No authentication is needed.

## Practical Examples

### Example 1: Find All Recruiting SCI Trials
```bash
curl "http://localhost:8080/api/v1/trials/search?status=RECRUITING&page_size=20"
```

### Example 2: Phase 2 or 3 Trials Near Los Angeles
```bash
curl "http://localhost:8080/api/v1/trials/search?phase=PHASE2,PHASE3&latitude=34.0522&longitude=-118.2437&distance=50&status=RECRUITING"
```

### Example 3: Search Multiple Conditions
```bash
curl "http://localhost:8080/api/v1/trials/search?conditions=multiple+sclerosis,spinal+cord+injury,ALS"
```

### Example 4: Age-Restricted Search (client-side filtering)
```bash
# Find recruiting trials for people aged 18-65
curl "http://localhost:8080/api/v1/trials/search?minimum_age=18&maximum_age=65&status=RECRUITING"

# Age values can be numeric (18, 65) or with units (18 Years, 65 Years)
curl "http://localhost:8080/api/v1/trials/search?minimum_age=18+Years&maximum_age=65+Years&status=RECRUITING"
```

### Example 5: Get Detailed Trial Information
```bash
curl "http://localhost:8080/api/v1/trials/NCT06511934" | jq
```

### Example 6: Complex Search with POST (combining server-side and client-side filters)
```bash
curl -X POST http://localhost:8080/api/v1/trials/search \
  -H "Content-Type: application/json" \
  -d '{
    "conditions": ["spinal cord injury", "tetraplegia"],
    "status": ["RECRUITING", "NOT_YET_RECRUITING"],
    "phase": ["PHASE2", "PHASE3"],
    "latitude": 40.7128,
    "longitude": -74.0060,
    "distance": 30,
    "minimum_age": "18",
    "maximum_age": "70",
    "page_size": 25
  }'
```

Note: `phase`, `minimum_age`, and `maximum_age` are applied client-side after receiving results from the API.

### Example 7: Pagination Workflow
```bash
# Get first page
RESPONSE=$(curl -s "http://localhost:8080/api/v1/trials/search?page_size=10")
echo "$RESPONSE" | jq '.trials[] | {nct_id, title}'

# Extract next_page_token (using jq)
NEXT_TOKEN=$(echo "$RESPONSE" | jq -r '.next_page_token')

# Get next page
curl -s "http://localhost:8080/api/v1/trials/search?page_size=10&page_token=$NEXT_TOKEN" | jq '.trials[] | {nct_id, title}'
```

## Development

### Project Structure

- `cmd/server/` - Application entry point
- `internal/api/` - External API clients
- `internal/cache/` - Caching implementation
- `internal/handlers/` - HTTP request handlers
- `internal/models/` - Data models and types
- `scripts/` - Test and utility scripts

### Adding New Features

1. **New API Integration**: Add client in `internal/api/`
2. **New Endpoint**: Add handler in `internal/handlers/`
3. **New Data Model**: Add type in `internal/models/`

## Performance Considerations

- **Response Time**: Typically < 1s for cached requests, 2-5s for API calls
- **Concurrent Requests**: Go's goroutines handle concurrency efficiently
- **Memory Usage**: Minimal due to efficient data structures
- **Rate Limiting**: Built-in delays prevent API throttling

## Troubleshooting

### Server won't start
- Check if port 8080 is already in use
- Verify Go version: `go version` (need 1.21+)

### API calls failing
- Check internet connection
- Verify ClinicalTrials.gov API is accessible
- Check rate limiting (may need to wait if hitting limits)

### Cache not working
- Verify cache is enabled: `-cache=true`
- Check cache TTL is appropriate
- Clear cache by restarting server

## License

This project is provided as-is for integration with clinical trials data.

## References

- [ClinicalTrials.gov API v2 Documentation](https://clinicaltrials.gov/data-api)
- [Research Guide](./research/trials_API_integration_guide_for_spinal_cord_injury.md)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## Support

For issues or questions, please check the research document in the `research/` directory for detailed API integration information.
