#!/usr/bin/env python3
"""
Exploration script for ReBEC (Brazilian Clinical Trials Registry) API
This script fetches and analyzes ReBEC XML data to understand its structure
and evaluate integration feasibility with our Trial model.
"""

import xml.etree.ElementTree as ET
import urllib.request
import urllib.error
import json
import sys
from datetime import datetime
from typing import Dict, Any, List, Optional, Tuple
from urllib.parse import urljoin

# ReBEC API endpoints (trying multiple variations)
REBEC_ENDPOINTS = [
    "http://www.ensaiosclinicos.gov.br/rg/all/xml/ictrp",
    "https://ensaiosclinicos.gov.br/rg/all/xml/ictrp",
    "http://ensaiosclinicos.gov.br/rg/all/xml/ictrp",
]
REBEC_INDIVIDUAL_BASE = "https://ensaiosclinicos.gov.br/xml_ictrp/downloadxmlictrp/"

class Colors:
    GREEN = '\033[0;32m'
    RED = '\033[0;31m'
    YELLOW = '\033[1;33m'
    BLUE = '\033[0;34m'
    CYAN = '\033[0;36m'
    NC = '\033[0m'

def print_header(text: str):
    print(f"\n{Colors.BLUE}{'='*70}{Colors.NC}")
    print(f"{Colors.BLUE}{text}{Colors.NC}")
    print(f"{Colors.BLUE}{'='*70}{Colors.NC}\n")

def print_section(text: str):
    print(f"\n{Colors.CYAN}{'─'*70}{Colors.NC}")
    print(f"{Colors.CYAN}{text}{Colors.NC}")
    print(f"{Colors.CYAN}{'─'*70}{Colors.NC}\n")

def print_success(message: str):
    print(f"{Colors.GREEN}✓ {message}{Colors.NC}")

def print_error(message: str):
    print(f"{Colors.RED}✗ {message}{Colors.NC}")

def print_info(message: str):
    print(f"  {message}")

def print_warning(message: str):
    print(f"{Colors.YELLOW}⚠ {message}{Colors.NC}")

def fetch_rebec_full_export(sample_size: int = 10) -> Tuple[Optional[ET.Element], Optional[str]]:
    """Fetch ReBEC full export XML (sample first few records)"""
    print_section("Fetching ReBEC Full Export")
    
    # Try multiple URL variations
    for url in REBEC_ENDPOINTS:
        print_info(f"Trying URL: {url}")
        try:
            req = urllib.request.Request(url)
            req.add_header('User-Agent', 'Mozilla/5.0 (Clinical Trials Microservice Explorer)')
            
            with urllib.request.urlopen(req, timeout=30) as response:
                status_code = response.getcode()
                content_type = response.headers.get('Content-Type', 'unknown')
                content = response.read()
                
                print_success(f"Successfully fetched XML (Status: {status_code})")
                print_info(f"Content-Type: {content_type}")
                print_info(f"Content-Length: {len(content)} bytes")
                
                # Parse XML
                root = ET.fromstring(content)
                print_success(f"XML parsed successfully")
                print_info(f"Root tag: {root.tag}")
                
                # Count trials
                trials = root.findall('.//trial')
                print_info(f"Total trials found: {len(trials)}")
                
                return root, url
        except urllib.error.HTTPError as e:
            print_warning(f"HTTP Error {e.code} - {e.reason}")
            continue
        except urllib.error.URLError as e:
            print_warning(f"URL Error: {e.reason}")
            continue
        except ET.ParseError as e:
            print_error(f"Failed to parse XML: {e}")
            return None, url
        except Exception as e:
            print_warning(f"Error: {e}")
            continue
    
    print_error("All ReBEC endpoint URLs failed")
    return None, None

def extract_trial_ids(root: ET.Element, limit: int = 5) -> List[str]:
    """Extract trial IDs from XML for testing individual endpoint"""
    print_section(f"Extracting Trial IDs (first {limit})")
    
    trials = root.findall('.//trial')
    trial_ids = []
    
    for i, trial in enumerate(trials[:limit]):
        # Try different possible ID fields
        trial_id = None
        for id_field in ['trial_id', 'id', 'primary_id', 'registration_number', 'unique_id']:
            elem = trial.find(id_field)
            if elem is not None and elem.text:
                trial_id = elem.text.strip()
                break
        
        # If no ID found, try getting first element that looks like an ID
        if not trial_id:
            for elem in trial.iter():
                if elem.text and len(elem.text.strip()) < 50:
                    # Check if it looks like an ID
                    if any(char.isdigit() for char in elem.text.strip()):
                        trial_id = elem.text.strip()
                        break
        
        if trial_id:
            trial_ids.append(trial_id)
            print_info(f"Trial {i+1} ID: {trial_id}")
        else:
            print_warning(f"Trial {i+1}: No ID found")
    
    return trial_ids

def fetch_individual_trial(trial_id: str) -> Optional[ET.Element]:
    """Fetch individual trial XML"""
    url = urljoin(REBEC_INDIVIDUAL_BASE, trial_id)
    print_info(f"Fetching: {url}")
    
    try:
        req = urllib.request.Request(url)
        req.add_header('User-Agent', 'Mozilla/5.0 (Clinical Trials Microservice Explorer)')
        
        with urllib.request.urlopen(req, timeout=10) as response:
            status_code = response.getcode()
            if status_code == 200:
                content = response.read()
                root = ET.fromstring(content)
                print_success(f"Successfully fetched trial {trial_id}")
                return root
            else:
                print_warning(f"Trial {trial_id} returned status {status_code}")
                return None
    except urllib.error.HTTPError as e:
        print_warning(f"Trial {trial_id} HTTP Error: {e.code}")
        return None
    except Exception as e:
        print_warning(f"Failed to fetch trial {trial_id}: {e}")
        return None

def analyze_trial_structure(trial: ET.Element, trial_num: int = 1) -> Dict[str, Any]:
    """Analyze a single trial's XML structure and extract fields"""
    print_section(f"Analyzing Trial Structure #{trial_num}")
    
    structure = {}
    
    def walk_element(elem: ET.Element, path: str = "", level: int = 0):
        """Recursively walk XML structure"""
        current_path = f"{path}/{elem.tag}" if path else elem.tag
        
        # Store element info
        if current_path not in structure:
            structure[current_path] = {
                'tag': elem.tag,
                'path': current_path,
                'has_text': bool(elem.text and elem.text.strip()),
                'text_sample': elem.text.strip()[:100] if elem.text and elem.text.strip() else None,
                'attributes': dict(elem.attrib) if elem.attrib else None,
                'child_count': len(list(elem)),
            }
        
        # Recurse into children
        for child in elem:
            walk_element(child, current_path, level + 1)
    
    walk_element(trial)
    
    return structure

def map_to_trial_model(trial: ET.Element) -> Dict[str, Any]:
    """Attempt to map ReBEC trial XML to our Trial model structure"""
    print_section("Mapping to Trial Model")
    
    mapped = {}
    
    # Common field mappings (based on typical clinical trial registries)
    field_mappings = {
        'nct_id': ['nct_id', 'nctid', 'primary_id', 'trial_id', 'id', 'registration_number'],
        'title': ['title', 'brief_title', 'public_title', 'study_title', 'official_title'],
        'status': ['status', 'recruitment_status', 'overall_status', 'trial_status'],
        'phase': ['phase', 'study_phase', 'trial_phase'],
        'conditions': ['condition', 'conditions', 'health_condition', 'disease'],
        'locations': ['location', 'locations', 'facility', 'site'],
        'sponsor': ['sponsor', 'lead_sponsor', 'primary_sponsor'],
        'contacts': ['contact', 'contacts', 'central_contact'],
        'eligibility': ['eligibility', 'eligibility_criteria'],
        'start_date': ['start_date', 'study_start_date'],
        'completion_date': ['completion_date', 'study_completion_date'],
    }
    
    def find_field(elem: ET.Element, field_names: List[str]) -> Optional[str]:
        """Find field value by trying multiple possible tag names"""
        for field_name in field_names:
            # Try direct match
            found = elem.find(field_name)
            if found is not None and found.text:
                return found.text.strip()
            
            # Try case-insensitive
            for child in elem.iter():
                if child.tag.lower() == field_name.lower() and child.text:
                    return child.text.strip()
        
        return None
    
    # Try to map each field
    for our_field, possible_names in field_mappings.items():
        value = find_field(trial, possible_names)
        if value:
            mapped[our_field] = value
            print_info(f"{our_field}: {value[:80] if len(value) > 80 else value}")
        else:
            print_warning(f"{our_field}: Not found")
    
    return mapped

def extract_all_fields(trial: ET.Element) -> Dict[str, Any]:
    """Extract all fields from a trial for analysis"""
    result = {}
    
    def extract_recursive(elem: ET.Element, path: str = ""):
        current_path = f"{path}.{elem.tag}" if path else elem.tag
        
        # Store value if text exists
        if elem.text and elem.text.strip():
            result[current_path] = elem.text.strip()
        
        # Store attributes
        if elem.attrib:
            for key, value in elem.attrib.items():
                result[f"{current_path}@{key}"] = value
        
        # Recurse
        for child in elem:
            extract_recursive(child, current_path)
    
    extract_recursive(trial)
    return result

def create_findings_document(working_url: Optional[str] = None, accessible: bool = True, root: Optional[ET.Element] = None):
    """Create a findings document for evaluation"""
    findings = {
        'date': datetime.now().isoformat(),
        'endpoints_tested': REBEC_ENDPOINTS,
        'accessible': accessible,
        'working_url': working_url,
        'recommendations': []
    }
    
    if not accessible:
        findings['recommendations'] = [
            'ReBEC XML export endpoint appears to be unavailable (404 errors)',
            'May need to contact ReBEC administrators for current API access',
            'Alternative: Use web scraping or manual data export',
            'Consider WHO ICTRP as alternative source for Brazilian trials',
        ]
    elif root:
        findings['xml_structure'] = 'Available'
        findings['recommendations'] = [
            'XML structure is accessible',
            'Proceed with XML parsing implementation',
            'Map fields to Trial model',
        ]
    
    findings_file = 'rebec_findings.json'
    with open(findings_file, 'w', encoding='utf-8') as f:
        json.dump(findings, f, indent=2, ensure_ascii=False)
    
    print_info(f"Findings document created: {findings_file}")
    return findings_file

def main():
    print_header("ReBEC API Exploration Script")
    print_info("This script explores the ReBEC XML structure to evaluate integration feasibility")
    print()
    
    # Step 1: Fetch full export
    root, working_url = fetch_rebec_full_export()
    if not root:
        print_error("Cannot proceed without full export data")
        print_section("Documenting Findings")
        create_findings_document(working_url=None, accessible=False)
        sys.exit(1)
    
    # Step 2: Analyze structure of first trial
    trials = root.findall('.//trial')
    if not trials:
        print_error("No trials found in XML")
        sys.exit(1)
    
    first_trial = trials[0]
    structure = analyze_trial_structure(first_trial, trial_num=1)
    
    # Step 3: Print structure summary
    print_section("XML Structure Summary")
    print_info("Top-level elements found:")
    for path, info in sorted(structure.items())[:30]:  # Show first 30
        indent = "  " * path.count('/')
        text_preview = f" = '{info['text_sample']}'" if info['text_sample'] else ""
        print(f"{indent}{path}{text_preview}")
    
    if len(structure) > 30:
        print_info(f"... and {len(structure) - 30} more fields")
    
    # Step 4: Extract all fields from first trial
    print_section("Complete Field Extraction (First Trial)")
    all_fields = extract_all_fields(first_trial)
    print_info(f"Total fields extracted: {len(all_fields)}")
    print_info("Sample fields:")
    for i, (key, value) in enumerate(list(all_fields.items())[:20]):
        print(f"  {key}: {value[:100] if len(value) > 100 else value}")
    
    # Step 5: Try mapping to our model
    mapped = map_to_trial_model(first_trial)
    
    # Step 6: Save detailed analysis to file
    output_file = "rebec_analysis.json"
    print_section(f"Saving Analysis to {output_file}")
    
    analysis = {
        'total_trials_in_export': len(trials),
        'structure_fields': {k: v for k, v in list(structure.items())[:100]},  # Limit size
        'all_fields_sample': {k: v for k, v in list(all_fields.items())[:50]},
        'mapped_fields': mapped,
        'sample_trial_xml': ET.tostring(first_trial, encoding='unicode')[:2000],  # First 2000 chars
    }
    
    with open(output_file, 'w', encoding='utf-8') as f:
        json.dump(analysis, f, indent=2, ensure_ascii=False)
    
    print_success(f"Analysis saved to {output_file}")
    
    # Step 7: Try individual trial endpoint if we have IDs
    trial_ids = extract_trial_ids(root, limit=3)
    if trial_ids:
        print_section("Testing Individual Trial Endpoint")
        for trial_id in trial_ids[:2]:  # Try first 2
            individual = fetch_individual_trial(trial_id)
            if individual:
                print_success(f"Individual trial endpoint works for {trial_id}")
                break
    
    # Step 8: Summary
    print_header("Summary")
    print_info(f"Total trials in export: {len(trials)}")
    print_info(f"Unique XML paths found: {len(structure)}")
    print_info(f"Fields mapped to our model: {len([k for k, v in mapped.items() if v])}/{len(mapped)}")
    
    print_section("Integration Feasibility Assessment")
    if mapped.get('title'):
        print_success("Title field found - Good")
    else:
        print_warning("Title field not found - Needs investigation")
    
    if mapped.get('status'):
        print_success("Status field found - Good")
    else:
        print_warning("Status field not found - Needs investigation")
    
    print_info("\nNext steps:")
    print_info("1. Review rebec_analysis.json for detailed structure")
    print_info("2. Compare field mappings with internal/models/trial.go")
    print_info("3. Determine if XML parsing library is needed (encoding/xml in Go)")
    print_info("4. Evaluate if model extensions are needed for ReBEC-specific fields")
    
    # Create findings document
    create_findings_document(working_url=working_url, accessible=True, root=root)

if __name__ == "__main__":
    main()

