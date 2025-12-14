# ReBEC API Integration Evaluation

## Summary

This document evaluates the feasibility of integrating ReBEC (Brazilian Clinical Trials Registry) as a multi-source provider alongside ClinicalTrials.gov.

## Exploration Results

**Date:** 2025-12-14

### Endpoint Testing

**Tested Endpoints:**
- `http://www.ensaiosclinicos.gov.br/rg/all/xml/ictrp`
- `https://ensaiosclinicos.gov.br/rg/all/xml/ictrp`
- `http://ensaiosclinicos.gov.br/rg/all/xml/ictrp`

**Result:** All endpoints return HTTP 404 (Not Found)

### Findings

1. **API Accessibility:** ❌ Not Accessible
   - The XML export endpoint documented in the research guide appears to be unavailable
   - All tested URL variations return 404 errors
   - The endpoint may have been deprecated, moved, or requires authentication

2. **Research Document Reference:**
   - The research document (`research/trials_API_integration_guide_for_spinal_cord_injury.md`) references:
     - Full export: `http://www.ensaiosclinicos.gov.br/rg/all/xml/ictrp`
     - Individual trial: `https://ensaiosclinicos.gov.br/xml_ictrp/downloadxmlictrp/[TRIAL_ID]`

3. **Alternative Approaches:**
   - Contact ReBEC administrators for current API documentation
   - Check if authentication/API key is required
   - Investigate if the endpoint path has changed
   - Consider WHO ICTRP as alternative source for Brazilian trials
   - Evaluate web scraping as a fallback option

## Integration Feasibility Assessment

### Current Status: ⚠️ BLOCKED

**Blocking Issues:**
- API endpoint is not accessible
- Unable to evaluate data structure
- Cannot assess field mapping to our Trial model

### Prerequisites for Integration

Before proceeding with integration, we need:

1. **API Access:**
   - Verify correct endpoint URLs
   - Determine if authentication is required
   - Obtain API documentation if available

2. **Data Structure Analysis:**
   - Fetch sample XML data
   - Analyze XML schema/structure
   - Map fields to our `Trial` model

3. **Field Mapping:**
   - NCT ID equivalent (or different ID system)
   - Status, phase, conditions
   - Eligibility (age, gender)
   - Locations with coordinates
   - Sponsors, contacts
   - Dates (start, completion)

4. **Technical Considerations:**
   - XML parsing (Go's `encoding/xml` package)
   - Portuguese language handling
   - ID system differences
   - Duplicate detection (if same trials appear in both registries)

## Recommended Next Steps

1. **Immediate Actions:**
   - Contact ReBEC support to verify API availability
   - Check ReBEC website for updated API documentation
   - Investigate WHO ICTRP as alternative source

2. **If API Becomes Available:**
   - Re-run exploration script (`scripts/test_rebec_api.py`)
   - Analyze XML structure
   - Create field mapping document
   - Implement ReBEC client (`internal/api/rebec.go`)

3. **Integration Architecture:**
   - Add `registry` field to Trial model (already present: `registry: "clinicaltrials.gov"`)
   - Create unified interface for multiple data sources
   - Implement result aggregation and deduplication
   - Add source selection to search requests

## Alternatives

If ReBEC integration is not feasible:

1. **WHO ICTRP:**
   - Aggregates data from 18 primary registries globally
   - Includes Brazilian trials
   - Requires partnership agreement (may involve costs)

2. **Direct Web Scraping:**
   - More maintenance overhead
   - Potential legal/terms of service issues
   - Less reliable than API access

3. **Manual Data Import:**
   - Periodic CSV/XML exports
   - Lower real-time requirements
   - Higher maintenance burden

## Conclusion

**Current Recommendation:** ⚠️ **DEFER INTEGRATION**

The ReBEC API endpoint is currently inaccessible, blocking further evaluation. Integration should be deferred until:

1. API access is confirmed and working
2. Data structure can be analyzed
3. Field mapping feasibility is established

The exploration script (`scripts/test_rebec_api.py`) is ready to use once API access is available. The script will automatically analyze the XML structure and generate field mappings.

---

**Files Created:**
- `scripts/test_rebec_api.py` - Exploration script
- `rebec_findings.json` - Automated findings (if API accessible)
- `REBEC_EVALUATION.md` - This evaluation document

