#!/bin/bash

# Direct API testing script for ClinicalTrials.gov API v2
# This helps understand the API structure and responses

echo "=========================================="
echo "ClinicalTrials.gov API v2 Direct Testing"
echo "=========================================="
echo ""

BASE_URL="https://clinicaltrials.gov/api/v2/studies"

# Test 1: Basic search with SCI terms
echo "1. Basic Search - Spinal Cord Injury"
echo "URL: $BASE_URL?query.cond=spinal+cord+injury&filter.overallStatus=RECRUITING&format=json&pageSize=3"
echo ""
curl -s "$BASE_URL?query.cond=spinal+cord+injury&filter.overallStatus=RECRUITING&format=json&pageSize=3&countTotal=true" | \
  jq '{totalCount, studyCount: (.studies | length), firstStudy: .studies[0].protocolSection.identificationModule}'
echo ""
echo ""

# Test 2: Multiple conditions with OR
echo "2. Multiple Conditions (OR logic)"
echo "URL: $BASE_URL?query.cond=spinal+cord+injury+OR+tetraplegia&format=json&pageSize=2"
echo ""
curl -s "$BASE_URL?query.cond=spinal+cord+injury+OR+tetraplegia&format=json&pageSize=2&countTotal=true" | \
  jq '{totalCount, studyCount: (.studies | length)}'
echo ""
echo ""

# Test 3: Phase filter
echo "3. Phase Filter"
echo "URL: $BASE_URL?query.cond=spinal+cord+injury&filter.phase=PHASE2&format=json&pageSize=2"
echo ""
curl -s "$BASE_URL?query.cond=spinal+cord+injury&filter.phase=PHASE2&format=json&pageSize=2" | \
  jq '{studyCount: (.studies | length), phases: [.studies[].protocolSection.designModule.phases // []]}'
echo ""
echo ""

# Test 4: Location-based search (Los Angeles)
echo "4. Location-Based Search (Los Angeles, 50 miles)"
echo "URL: $BASE_URL?query.cond=spinal+cord+injury&filter.geo=distance(34.0522,-118.2437,50mi)&format=json&pageSize=2"
echo ""
curl -s "$BASE_URL?query.cond=spinal+cord+injury&filter.geo=distance(34.0522,-118.2437,50mi)&format=json&pageSize=2" | \
  jq '{studyCount: (.studies | length), locations: [.studies[].protocolSection.contactsLocationsModule.locations[0].facility.city // "N/A"]}'
echo ""
echo ""

# Test 5: Get specific trial
echo "5. Get Specific Trial (NCT03003364)"
echo "URL: $BASE_URL/NCT03003364?format=json"
echo ""
curl -s "$BASE_URL/NCT03003364?format=json" | \
  jq '.studies[0].protocolSection | {
    nctId: .identificationModule.nctId,
    title: .identificationModule.briefTitle,
    status: .statusModule.overallStatus,
    conditions: .conditionsModule.conditions,
    phases: .designModule.phases
  }'
echo ""
echo ""

# Test 6: Response structure overview
echo "6. Full Response Structure (first study, truncated)"
echo "URL: $BASE_URL?query.cond=spinal+cord+injury&format=json&pageSize=1"
echo ""
curl -s "$BASE_URL?query.cond=spinal+cord+injury&format=json&pageSize=1" | \
  jq '.studies[0] | keys' | head -20
echo ""
echo ""

echo "=========================================="
echo "Testing Complete"
echo "=========================================="
echo ""
echo "Note: Install 'jq' for better JSON formatting:"
echo "  macOS: brew install jq"
echo "  Linux: apt-get install jq or yum install jq"
