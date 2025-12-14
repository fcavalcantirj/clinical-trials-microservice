#!/bin/bash

# Test script for Clinical Trials Microservice API
# This script tests various endpoints to understand API behavior

BASE_URL="${BASE_URL:-http://localhost:8080}"
API_BASE="${API_BASE:-$BASE_URL/api/v1}"

echo "=========================================="
echo "Clinical Trials Microservice API Tester"
echo "=========================================="
echo "Base URL: $BASE_URL"
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test health endpoint
echo -e "${YELLOW}1. Testing Health Endpoint${NC}"
echo "GET $BASE_URL/health"
response=$(curl -s -w "\nHTTP_CODE:%{http_code}" "$BASE_URL/health")
http_code=$(echo "$response" | grep "HTTP_CODE" | cut -d: -f2)
body=$(echo "$response" | sed '/HTTP_CODE/d')
echo "Response: $body"
echo "HTTP Status: $http_code"
if [ "$http_code" = "200" ]; then
    echo -e "${GREEN}✓ Health check passed${NC}"
else
    echo -e "${RED}✗ Health check failed${NC}"
fi
echo ""

# Test basic search (default SCI terms)
echo -e "${YELLOW}2. Testing Basic Search (Default SCI Terms)${NC}"
echo "GET $API_BASE/trials/search"
response=$(curl -s -w "\nHTTP_CODE:%{http_code}" "$API_BASE/trials/search")
http_code=$(echo "$response" | grep "HTTP_CODE" | cut -d: -f2)
body=$(echo "$response" | sed '/HTTP_CODE/d')
echo "Response preview (first 500 chars):"
echo "$body" | head -c 500
echo "..."
echo "HTTP Status: $http_code"
if [ "$http_code" = "200" ]; then
    echo -e "${GREEN}✓ Basic search passed${NC}"
    # Extract and show total count if available
    total=$(echo "$body" | grep -o '"total_count":[0-9]*' | cut -d: -f2 | head -1)
    if [ -n "$total" ]; then
        echo "Total trials found: $total"
    fi
else
    echo -e "${RED}✗ Basic search failed${NC}"
fi
echo ""

# Test search with status filter
echo -e "${YELLOW}3. Testing Search with Status Filter${NC}"
echo "GET $API_BASE/trials/search?status=RECRUITING"
response=$(curl -s -w "\nHTTP_CODE:%{http_code}" "$API_BASE/trials/search?status=RECRUITING&page_size=5")
http_code=$(echo "$response" | grep "HTTP_CODE" | cut -d: -f2)
body=$(echo "$response" | sed '/HTTP_CODE/d')
echo "HTTP Status: $http_code"
if [ "$http_code" = "200" ]; then
    echo -e "${GREEN}✓ Status filter search passed${NC}"
else
    echo -e "${RED}✗ Status filter search failed${NC}"
fi
echo ""

# Test search with custom conditions
echo -e "${YELLOW}4. Testing Search with Custom Conditions${NC}"
echo "GET $API_BASE/trials/search?conditions=spinal+cord+injury&page_size=3"
response=$(curl -s -w "\nHTTP_CODE:%{http_code}" "$API_BASE/trials/search?conditions=spinal+cord+injury&page_size=3")
http_code=$(echo "$response" | grep "HTTP_CODE" | cut -d: -f2)
body=$(echo "$response" | sed '/HTTP_CODE/d')
echo "HTTP Status: $http_code"
if [ "$http_code" = "200" ]; then
    echo -e "${GREEN}✓ Custom conditions search passed${NC}"
    # Show first trial title if available
    title=$(echo "$body" | grep -o '"title":"[^"]*"' | head -1 | cut -d'"' -f4)
    if [ -n "$title" ]; then
        echo "First trial: $title"
    fi
else
    echo -e "${RED}✗ Custom conditions search failed${NC}"
fi
echo ""

# Test POST search with JSON body
echo -e "${YELLOW}5. Testing POST Search with JSON Body${NC}"
echo "POST $API_BASE/trials/search"
json_body='{"conditions":["spinal cord injury"],"status":["RECRUITING"],"page_size":2}'
response=$(curl -s -w "\nHTTP_CODE:%{http_code}" \
    -X POST \
    -H "Content-Type: application/json" \
    -d "$json_body" \
    "$API_BASE/trials/search")
http_code=$(echo "$response" | grep "HTTP_CODE" | cut -d: -f2)
body=$(echo "$response" | sed '/HTTP_CODE/d')
echo "HTTP Status: $http_code"
if [ "$http_code" = "200" ]; then
    echo -e "${GREEN}✓ POST search passed${NC}"
else
    echo -e "${RED}✗ POST search failed${NC}"
fi
echo ""

# Test getting a specific trial (will use first trial from previous search if available)
echo -e "${YELLOW}6. Testing Get Trial by ID${NC}"
# Try a well-known NCT ID for spinal cord injury
test_nct_id="NCT03003364"  # Example NCT ID - may need to be updated
echo "GET $API_BASE/trials/$test_nct_id"
response=$(curl -s -w "\nHTTP_CODE:%{http_code}" "$API_BASE/trials/$test_nct_id")
http_code=$(echo "$response" | grep "HTTP_CODE" | cut -d: -f2)
body=$(echo "$response" | sed '/HTTP_CODE/d')
echo "HTTP Status: $http_code"
if [ "$http_code" = "200" ]; then
    echo -e "${GREEN}✓ Get trial by ID passed${NC}"
    title=$(echo "$body" | grep -o '"title":"[^"]*"' | head -1 | cut -d'"' -f4)
    if [ -n "$title" ]; then
        echo "Trial title: $title"
    fi
else
    echo -e "${YELLOW}⚠ Trial not found (this is okay if the NCT ID doesn't exist)${NC}"
fi
echo ""

# Test location-based search (Los Angeles coordinates)
echo -e "${YELLOW}7. Testing Location-Based Search${NC}"
echo "GET $API_BASE/trials/search?latitude=34.0522&longitude=-118.2437&distance=50&page_size=3"
response=$(curl -s -w "\nHTTP_CODE:%{http_code}" "$API_BASE/trials/search?latitude=34.0522&longitude=-118.2437&distance=50&page_size=3")
http_code=$(echo "$response" | grep "HTTP_CODE" | cut -d: -f2)
body=$(echo "$response" | sed '/HTTP_CODE/d')
echo "HTTP Status: $http_code"
if [ "$http_code" = "200" ]; then
    echo -e "${GREEN}✓ Location-based search passed${NC}"
else
    echo -e "${RED}✗ Location-based search failed${NC}"
fi
echo ""

echo "=========================================="
echo "Testing Complete"
echo "=========================================="
echo ""
echo "To test the ClinicalTrials.gov API directly, use:"
echo "  curl 'https://clinicaltrials.gov/api/v2/studies?query.cond=spinal+cord+injury&filter.overallStatus=RECRUITING&format=json&pageSize=5'"
