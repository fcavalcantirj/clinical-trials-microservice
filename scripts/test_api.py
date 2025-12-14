#!/usr/bin/env python3
"""
Comprehensive test script for Clinical Trials Microservice API
This script provides more detailed testing and analysis
"""

import json
import sys
import requests
from typing import Dict, Any, Optional
from datetime import datetime

BASE_URL = "http://localhost:8080"
API_BASE = f"{BASE_URL}/api/v1"

class Colors:
    GREEN = '\033[0;32m'
    RED = '\033[0;31m'
    YELLOW = '\033[1;33m'
    BLUE = '\033[0;34m'
    NC = '\033[0m'  # No Color

def print_header(text: str):
    print(f"\n{Colors.BLUE}{'='*60}{Colors.NC}")
    print(f"{Colors.BLUE}{text}{Colors.NC}")
    print(f"{Colors.BLUE}{'='*60}{Colors.NC}\n")

def print_test(name: str):
    print(f"{Colors.YELLOW}▶ {name}{Colors.NC}")

def print_success(message: str):
    print(f"{Colors.GREEN}✓ {message}{Colors.NC}")

def print_error(message: str):
    print(f"{Colors.RED}✗ {message}{Colors.NC}")

def print_info(message: str):
    print(f"  {message}")

def test_health():
    """Test health endpoint"""
    print_test("Health Check")
    try:
        response = requests.get(f"{BASE_URL}/health", timeout=5)
        if response.status_code == 200:
            print_success(f"Health check passed (Status: {response.status_code})")
            print_info(f"Response: {response.json()}")
            return True
        else:
            print_error(f"Health check failed (Status: {response.status_code})")
            return False
    except requests.exceptions.RequestException as e:
        print_error(f"Health check failed: {e}")
        return False

def test_search_basic():
    """Test basic search with default SCI terms"""
    print_test("Basic Search (Default SCI Terms)")
    try:
        response = requests.get(
            f"{API_BASE}/trials/search",
            params={"page_size": 5},
            timeout=30
        )
        if response.status_code == 200:
            data = response.json()
            total = data.get("total_count", 0)
            trials = data.get("trials", [])
            print_success(f"Basic search passed")
            print_info(f"Total trials found: {total}")
            print_info(f"Trials in response: {len(trials)}")
            if trials:
                print_info(f"First trial: {trials[0].get('title', 'N/A')[:80]}...")
                print_info(f"First trial NCT ID: {trials[0].get('nct_id', 'N/A')}")
            return True, data
        else:
            print_error(f"Search failed (Status: {response.status_code})")
            print_info(f"Response: {response.text[:200]}")
            return False, None
    except requests.exceptions.RequestException as e:
        print_error(f"Search failed: {e}")
        return False, None

def test_search_with_filters():
    """Test search with various filters"""
    print_test("Search with Filters (Status: RECRUITING, Phase: PHASE2)")
    try:
        response = requests.get(
            f"{API_BASE}/trials/search",
            params={
                "status": "RECRUITING",
                "phase": "PHASE2",
                "page_size": 3
            },
            timeout=30
        )
        if response.status_code == 200:
            data = response.json()
            trials = data.get("trials", [])
            print_success(f"Filtered search passed")
            print_info(f"Trials found: {len(trials)}")
            for i, trial in enumerate(trials[:2], 1):
                print_info(f"  {i}. {trial.get('title', 'N/A')[:70]}...")
                print_info(f"     Status: {trial.get('status', 'N/A')}, Phase: {trial.get('phase', 'N/A')}")
            return True
        else:
            print_error(f"Filtered search failed (Status: {response.status_code})")
            return False
    except requests.exceptions.RequestException as e:
        print_error(f"Filtered search failed: {e}")
        return False

def test_post_search():
    """Test POST search with JSON body"""
    print_test("POST Search with JSON Body")
    try:
        payload = {
            "conditions": ["spinal cord injury", "tetraplegia"],
            "status": ["RECRUITING", "NOT_YET_RECRUITING"],
            "page_size": 3
        }
        response = requests.post(
            f"{API_BASE}/trials/search",
            json=payload,
            timeout=30
        )
        if response.status_code == 200:
            data = response.json()
            trials = data.get("trials", [])
            print_success(f"POST search passed")
            print_info(f"Trials found: {len(trials)}")
            return True
        else:
            print_error(f"POST search failed (Status: {response.status_code})")
            return False
    except requests.exceptions.RequestException as e:
        print_error(f"POST search failed: {e}")
        return False

def test_get_trial_by_id(nct_id: str = "NCT03003364"):
    """Test getting a specific trial by NCT ID"""
    print_test(f"Get Trial by ID: {nct_id}")
    try:
        response = requests.get(
            f"{API_BASE}/trials/{nct_id}",
            timeout=30
        )
        if response.status_code == 200:
            trial = response.json()
            print_success(f"Get trial by ID passed")
            print_info(f"Title: {trial.get('title', 'N/A')[:80]}...")
            print_info(f"Status: {trial.get('status', 'N/A')}")
            print_info(f"Conditions: {', '.join(trial.get('conditions', []))}")
            if trial.get('locations'):
                print_info(f"Locations: {len(trial['locations'])} location(s)")
            return True, trial
        else:
            print_error(f"Get trial failed (Status: {response.status_code})")
            print_info(f"Response: {response.text[:200]}")
            return False, None
    except requests.exceptions.RequestException as e:
        print_error(f"Get trial failed: {e}")
        return False, None

def test_location_search():
    """Test location-based search"""
    print_test("Location-Based Search (Los Angeles, 50 miles)")
    try:
        response = requests.get(
            f"{API_BASE}/trials/search",
            params={
                "latitude": 34.0522,
                "longitude": -118.2437,
                "distance": 50,
                "page_size": 3
            },
            timeout=30
        )
        if response.status_code == 200:
            data = response.json()
            trials = data.get("trials", [])
            print_success(f"Location search passed")
            print_info(f"Trials found: {len(trials)}")
            return True
        else:
            print_error(f"Location search failed (Status: {response.status_code})")
            return False
    except requests.exceptions.RequestException as e:
        print_error(f"Location search failed: {e}")
        return False

def test_direct_ct_api():
    """Test ClinicalTrials.gov API directly for comparison"""
    print_test("Direct ClinicalTrials.gov API Test")
    try:
        url = "https://clinicaltrials.gov/api/v2/studies"
        params = {
            "query.cond": "spinal cord injury OR tetraplegia",
            "filter.overallStatus": "RECRUITING",
            "format": "json",
            "pageSize": 3
        }
        response = requests.get(url, params=params, timeout=30)
        if response.status_code == 200:
            data = response.json()
            total = data.get("totalCount", 0)
            studies = data.get("studies", [])
            print_success(f"Direct API call successful")
            print_info(f"Total trials: {total}")
            print_info(f"Studies in response: {len(studies)}")
            return True
        else:
            print_error(f"Direct API call failed (Status: {response.status_code})")
            return False
    except requests.exceptions.RequestException as e:
        print_error(f"Direct API call failed: {e}")
        return False

def analyze_trial_structure(trial: Dict[str, Any]):
    """Analyze and display trial structure"""
    print_header("Trial Data Structure Analysis")
    print_info("Available fields:")
    for key in trial.keys():
        value = trial[key]
        if isinstance(value, list):
            print_info(f"  - {key}: list with {len(value)} items")
            if value and isinstance(value[0], dict):
                print_info(f"    First item keys: {list(value[0].keys())}")
        elif isinstance(value, dict):
            print_info(f"  - {key}: dict with keys: {list(value.keys())}")
        else:
            value_str = str(value)[:50] if value else "None"
            print_info(f"  - {key}: {type(value).__name__} = {value_str}")

def main():
    print_header("Clinical Trials Microservice API Comprehensive Tester")
    print_info(f"Base URL: {BASE_URL}")
    print_info(f"Start time: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    
    results = {}
    
    # Run tests
    results['health'] = test_health()
    
    if not results['health']:
        print_error("\nHealth check failed. Is the server running?")
        print_info("Start the server with: go run cmd/server/main.go")
        sys.exit(1)
    
    success, search_data = test_search_basic()
    results['basic_search'] = success
    
    # Get a trial ID from search results if available
    test_nct_id = None
    if search_data and search_data.get('trials'):
        test_nct_id = search_data['trials'][0].get('nct_id')
        if test_nct_id:
            print_info(f"\nUsing NCT ID from search results: {test_nct_id}")
    
    results['filtered_search'] = test_search_with_filters()
    results['post_search'] = test_post_search()
    
    # Use found NCT ID or fallback
    nct_id_to_test = test_nct_id or "NCT03003364"
    success, trial_data = test_get_trial_by_id(nct_id_to_test)
    results['get_trial'] = success
    
    if trial_data:
        analyze_trial_structure(trial_data)
    
    results['location_search'] = test_location_search()
    results['direct_api'] = test_direct_ct_api()
    
    # Summary
    print_header("Test Summary")
    passed = sum(1 for v in results.values() if v)
    total = len(results)
    print_info(f"Passed: {passed}/{total}")
    for test_name, result in results.items():
        status = "✓" if result else "✗"
        color = Colors.GREEN if result else Colors.RED
        print(f"  {color}{status}{Colors.NC} {test_name}")
    
    if passed == total:
        print_success("\nAll tests passed!")
        sys.exit(0)
    else:
        print_error(f"\n{total - passed} test(s) failed")
        sys.exit(1)

if __name__ == "__main__":
    main()
