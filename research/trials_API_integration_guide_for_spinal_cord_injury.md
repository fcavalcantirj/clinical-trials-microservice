# Clinical trials API integration guide for spinal cord injury websites

**The ClinicalTrials.gov API v2 is the clear winner for somostetra.org integration** — it's free, requires no authentication, provides comprehensive global trial coverage, and offers robust query capabilities for spinal cord injury conditions. Combined with the AACT database for advanced analytics and ReBEC's direct XML export for Brazilian trials, a powerful multi-source aggregation system is achievable without licensing costs.

## ClinicalTrials.gov API v2 delivers everything needed

The modernized ClinicalTrials.gov API (launched 2024, replacing the retired classic API) provides the most practical path for integration. The REST API uses OpenAPI 3.0 specification with JSON as the primary response format.

**Base endpoint:** `https://clinicaltrials.gov/api/v2/studies`

The API requires **no authentication or API keys** — it's completely public. Rate limits allow approximately **50 requests per minute per IP address**, returning HTTP 429 when exceeded. For spinal cord injury queries, the recommended search combines multiple condition terms:

```
query.cond=spinal+cord+injury+OR+quadriplegia+OR+tetraplegia
filter.overallStatus=RECRUITING,NOT_YET_RECRUITING
format=json
countTotal=true
```

Key data fields returned include NCT ID, study title, status, phase, eligibility criteria, locations with coordinates, sponsor information, and contact details. The `pageToken` parameter enables cursor-based pagination through large result sets with up to **1,000 studies per request** for field subsets.

The official documentation at clinicaltrials.gov/data-api provides comprehensive guides including the data structure, search areas reference, and migration notes from the legacy API.

## International registries require different approaches

The **WHO ICTRP** aggregates data from 18 primary registries globally but presents a significant barrier: the XML Web Service API requires a formal agreement with the WHO Secretariat and involves costs. However, free alternatives exist — the search portal at trialsearch.who.int offers CSV and XML downloads, and organizations can request SharePoint access for bulk data containing all records. ICTRP updates weekly and includes data from registries across Europe, Asia, Latin America, and beyond.

| Registry | API Available | Best Access Method |
|----------|--------------|-------------------|
| WHO ICTRP | Paid partnership only | SharePoint bulk download request |
| EU Clinical Trials Register | No official API | R package `ctrdata` (scraping) |
| CTIS (new EU system) | No API | R/Python scrapers available |
| ReBEC (Brazil) | XML export | Direct URL: `ensaiosclinicos.gov.br/rg/all/xml/ictrp` |

The EU situation is notably challenging. Neither the legacy EU Clinical Trials Register (clinicaltrialsregister.eu) nor the new CTIS (euclinicaltrials.eu, mandatory since January 2023) offers public APIs. The EU-CTR limits downloads to 20 records in plain text format. The best workaround is the **`ctrdata` R package**, which scrapes both EU systems and stores data in local databases (PostgreSQL, SQLite, MongoDB supported).

## Brazilian trials accessible via direct XML export

**ReBEC** (Registro Brasileiro de Ensaios Clínicos) stands out among regional registries by offering direct programmatic access. As a WHO Primary Registry since 2011, it maintains over **8,600 registered trials** with nearly 4,800 actively recruiting.

The complete database is available at:
```
http://www.ensaiosclinicos.gov.br/rg/all/xml/ictrp
```

Individual trial records can be retrieved using:
```
https://ensaiosclinicos.gov.br/xml_ictrp/downloadxmlictrp/[TRIAL_ID]
```

For Portuguese search optimization, use terms like "lesão medular" (spinal cord injury), "tetraplegia," and "paraplegia" alongside English equivalents. ReBEC operates as a trilingual platform (Portuguese, English, Spanish) and recently launched Rebec@, an AI assistant for trial registration.

## AACT database enables powerful SQL analytics

The **Aggregate Analysis of ClinicalTrials.gov (AACT)** database offers a complementary approach — instead of API calls, it provides direct PostgreSQL access to a normalized **51-table relational database** refreshed daily. This enables complex SQL queries impossible through REST APIs.

**Connection details:**
- Host: `aact-db.ctti-clinicaltrials.org`
- Port: 5432
- Database: `aact`
- Requires free account registration at aact.ctti-clinicaltrials.org

Example SQL for SCI trials:
```sql
SELECT s.nct_id, s.brief_title, s.overall_status, c.name
FROM studies s
JOIN conditions c ON s.nct_id = c.nct_id
WHERE LOWER(c.name) LIKE '%spinal cord injury%'
AND s.overall_status = 'Recruiting';
```

AACT integrates MeSH terminology, supports complex joins across intervention/outcome/eligibility tables, and maintains historical archives for longitudinal analysis. For websites requiring advanced filtering, faceted search, or analytics dashboards, AACT significantly outperforms REST API approaches.

## scitrials.org offers a model, not an API

**scitrials.org** is a specialized SCI trial finder operated by NASCIC (North American Spinal Cord Injury Consortium) with support from the Christopher & Dana Reeve Foundation, Craig H. Neilsen Foundation, and 20+ SCI organizations globally.

The platform **aggregates exclusively from ClinicalTrials.gov** and adds significant value through SCI-specific categorization unavailable elsewhere:

- **Injury level filters:** C1-C8 (cervical), T1-T12 (thoracic), L1-L5 (lumbar), S1-S5 (sacral)
- **Severity classification:** AIS-A through AIS-D (ASIA Impairment Scale)
- **Chronicity:** Acute (<72 hours), sub-acute (<6 months), chronic (>6 months)
- **Treatment categories:** Stem cells, epidural stimulation, exoskeletons, FES, drugs, rehabilitation (20 types total)

However, scitrials.org **offers no public API**. The underlying data can be replicated by querying ClinicalTrials.gov directly and implementing similar categorization logic. For partnership inquiries, contact contact@scitrials.org.

## OpenTrials is deprecated; use alternatives instead

The OpenTrials.net project, which aimed to aggregate structured clinical trial data across registries, is **no longer active** — all GitHub repositories are archived and the API at api.opentrials.net is non-functional. Organizations should not build dependencies on this platform.

**Recommended alternatives for multi-registry aggregation:**

| Resource | Status | Best Use Case |
|----------|--------|--------------|
| AACT Database | Active | SQL-based analysis, bulk queries |
| `ctrdata` R package | Active | Cross-registry aggregation (CT.gov, EU-CTR, CTIS, ISRCTN) |
| WHO ICTRP downloads | Active | Global registry coverage |

For SCI-specific research data (not clinical trials), the **Open Data Commons for Spinal Cord Injury (odc-sci.org)** provides preclinical research datasets from 52+ laboratories following FAIR principles. The **National Spinal Cord Injury Statistical Center** maintains patient outcomes data from 55,000+ individuals since 1973, available for download.

## Technical implementation recommendations

For somostetra.org, a hybrid architecture maximizes coverage while respecting rate limits:

**Phase 1 — MVP with ClinicalTrials.gov API:**
1. Query the v2 API for `spinal+cord+injury OR paraplegia OR tetraplegia OR quadriplegia`
2. Cache results in PostgreSQL with JSONB columns for flexible schema
3. Refresh every 6-12 hours via background job
4. Implement Redis for search result caching (1-hour TTL)

**Phase 2 — Multi-source aggregation:**
1. Add ReBEC via weekly XML import
2. Request WHO ICTRP SharePoint access for international coverage
3. Implement duplicate detection using NCT IDs and fuzzy matching on sponsor/title

**Search interface essentials:**
- Pre-filtered condition (SCI terms)
- Status filter (recruiting/not yet recruiting/completed)
- Distance-based location search using `filter.geo=distance(lat,lon,50mi)`
- Phase filter
- Age eligibility

**Data model normalization** should unify trials from different sources using NCT ID as primary key, with standardized status enumerations and location coordinates for mapping.

```javascript
{
  id: "NCT12345678",
  registry: "clinicaltrials.gov",
  title: "...",
  status: "RECRUITING",
  conditions: [{ term: "Spinal Cord Injury", meshId: "D013119" }],
  locations: [{ city: "...", country: "US", lat: 34.0, lng: -118.0 }]
}
```

## Conclusion

ClinicalTrials.gov API v2 provides the foundation — free, comprehensive, and well-documented. AACT extends capabilities with SQL analytics. ReBEC offers direct Brazilian access. The EU ecosystem requires scraping solutions due to missing APIs, while WHO ICTRP provides global reach for organizations willing to pursue formal agreements.

For a spinal cord injury focused website, the combination of **ClinicalTrials.gov API + AACT + ReBEC XML** covers the vast majority of relevant trials globally without licensing costs. Implementing scitrials.org-style injury level and AIS categorization adds substantial value for the SCI community, differentiating somostetra.org from generic trial finders.

Key technical considerations: implement aggressive caching (6-12 hour refresh for trial data), respect the 50 requests/minute rate limit with exponential backoff, use MeSH terms for standardized condition matching, and always link back to official sources with appropriate attribution to the U.S. National Library of Medicine.