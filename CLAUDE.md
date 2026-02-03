# Sector Opportunity Analyzer - Go Edition

## Project Overview

This is a Go rewrite of the Python+Node Sector Opportunity Analyzer, designed for minimal resource usage on Replit.

## Key Improvements Over Python Version

1. **Single Binary** - No runtime dependencies to install
2. **15-30MB Memory** vs 500-700MB for Python+Node
3. **<1s Cold Start** vs 40-140s with pip/npm install
4. **Single Process** - No frontend/backend coordination needed

## Architecture

```
main.go           # HTTP server with chi router, embedded static files
config/config.go  # Sector definitions, weights, API configurations
data/cache.go     # In-memory cache with 12-hour TTL
data/types.go     # Data structures for prices, time series, API responses
data/fetchers.go  # HTTP clients for Yahoo Finance, FRED, BLS
analysis/signals.go  # Signal calculations (momentum, valuation, etc.)
analysis/scoring.go  # Weighted composite scoring engine
api/schemas.go    # JSON response types
api/handlers.go   # HTTP route handlers
static/           # Embedded frontend (React build)
```

## API Compatibility

This Go version maintains full API compatibility with the Python version:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/scores` | GET | All sector scores with optional weight params |
| `/api/scores/summary` | GET | Summary with top/bottom sectors |
| `/api/scores/{sector}` | GET | Single sector score |
| `/api/data/sectors` | GET | List of GICS sectors |
| `/api/cache/info` | GET | Cache statistics |
| `/api/cache/clear` | POST | Clear cache |
| `/health` | GET | Health check |

## Building

```bash
# On Replit, just run:
go build -o sector-analyzer .
./sector-analyzer

# With frontend:
# 1. Build React app in original project
# 2. Copy dist/* to static/
# 3. Rebuild Go binary
```

## Environment Variables

- `PORT` - Server port (default: 8000)
- `FRED_API_KEY` - Required for macro data
- `BLS_API_KEY` - Optional, for higher BLS rate limits

## Notes

- Yahoo Finance API doesn't require auth but has rate limits
- FRED API key is free at https://fred.stlouisfed.org/docs/api/api_key.html
- R&D data uses hardcoded Damodaran averages (Excel parsing not implemented)
- Cache is in-memory only (resets on restart)
