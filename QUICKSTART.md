# Quick Start Guide

## Start the Server

```bash
# 1. Install dependencies
go mod download

# 2. Run the server
go run cmd/server/main.go
```

The server will start on `http://localhost:8080`

## Quick Test

### Test 1: Health Check
```bash
curl http://localhost:8080/health
```

### Test 2: Search for SCI Trials
```bash
curl "http://localhost:8080/api/v1/trials/search?page_size=5"
```

### Test 3: Search with Filters
```bash
curl "http://localhost:8080/api/v1/trials/search?status=RECRUITING&page_size=3"
```

### Test 4: Get Specific Trial
```bash
curl "http://localhost:8080/api/v1/trials/NCT03003364"
```

## Run Test Scripts

### Bash Test Script
```bash
./scripts/test_api.sh
```

### Python Test Script (requires requests)
```bash
pip install requests  # if needed
python3 scripts/test_api.py
```

### Test Direct API (requires jq)
```bash
# Install jq first: brew install jq (macOS) or apt-get install jq (Linux)
./scripts/test_direct_api.sh
```

## Example API Calls

### Find Recruiting Trials
```bash
curl "http://localhost:8080/api/v1/trials/search?status=RECRUITING&page_size=10"
```

### Search Near Location (Los Angeles)
```bash
curl "http://localhost:8080/api/v1/trials/search?latitude=34.0522&longitude=-118.2437&distance=50&page_size=5"
```

### POST Request with JSON
```bash
curl -X POST http://localhost:8080/api/v1/trials/search \
  -H "Content-Type: application/json" \
  -d '{
    "conditions": ["spinal cord injury"],
    "status": ["RECRUITING"],
    "page_size": 5
  }'
```

## Build for Production

```bash
go build -o bin/trials-service cmd/server/main.go
./bin/trials-service -port 8080 -cache=true -cache-ttl=6h
```

## Configuration Options

```bash
# Custom port
go run cmd/server/main.go -port 3000

# Disable cache
go run cmd/server/main.go -cache=false

# Custom cache TTL (6 hours, 12 hours, etc.)
go run cmd/server/main.go -cache-ttl=12h
```

## No API Keys Required!

The ClinicalTrials.gov API v2 is completely public and free. No authentication or API keys needed!
