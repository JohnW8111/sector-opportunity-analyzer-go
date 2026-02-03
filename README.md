# Sector Opportunity Analyzer - Go Edition

A high-performance sector analysis tool that calculates opportunity scores for GICS sectors based on momentum, valuation, growth, innovation, and macro sensitivity signals.

## Performance Comparison

| Metric | Python + Node | Go |
|--------|--------------|-----|
| Memory | 500-700MB | 15-30MB |
| Cold Start | 40-140s | <1s |
| Dependencies | ~50 packages | 3 Go modules |
| Processes | 2 | 1 |

## Quick Start

### Build
```bash
# Download dependencies
go mod download

# Build the binary
go build -o sector-analyzer .

# Run
./sector-analyzer
```

### With Frontend

```bash
# Build the React frontend (from the original project)
cd ../sector-opportunity-analyzer/frontend
npm install
npm run build

# Copy the build to static/
cp -r dist/* ../sector-opportunity-analyzer-go/static/

# Rebuild Go binary with embedded frontend
cd ../sector-opportunity-analyzer-go
go build -o sector-analyzer .

# Run - serves both API and frontend
./sector-analyzer
```

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `PORT` | No | Server port (default: 8000) |
| `FRED_API_KEY` | Yes* | FRED API key for macro data |
| `BLS_API_KEY` | No | BLS API key (higher rate limits) |

*Without FRED API key, macro data will be unavailable.

Get a free FRED API key at: https://fred.stlouisfed.org/docs/api/api_key.html

## API Endpoints

### Scores

```
GET /api/scores
  Query params:
    - momentum (0-1): Weight for momentum signal
    - valuation (0-1): Weight for valuation signal
    - growth (0-1): Weight for growth signal
    - innovation (0-1): Weight for innovation signal
    - macro (0-1): Weight for macro signal
    - refresh (bool): Force data refresh

GET /api/scores/summary
  Returns top/bottom sectors and score distribution

GET /api/scores/{sector}
  Returns score for a specific sector
```

### Data

```
GET /api/data/sectors
  Returns list of all GICS sectors
```

### Cache

```
GET /api/cache/info
  Returns cache statistics

POST /api/cache/clear
  Clears all cached data
```

### Health

```
GET /health
  Returns service health status
```

## Architecture

```
sector-analyzer-go/
├── main.go              # Entry point, HTTP server, static file serving
├── config/
│   └── config.go        # Sector definitions, weights, API configs
├── data/
│   ├── cache.go         # In-memory cache with TTL
│   ├── types.go         # Data structures
│   └── fetchers.go      # Yahoo Finance, FRED, BLS API clients
├── analysis/
│   ├── signals.go       # Signal calculations (momentum, valuation, etc.)
│   └── scoring.go       # Weighted composite scoring
├── api/
│   ├── schemas.go       # JSON response types
│   └── handlers.go      # HTTP route handlers
└── static/              # Embedded frontend (built React app)
```

## Data Sources

1. **Yahoo Finance** (yfinance alternative)
   - Historical prices for sector ETFs (XLK, XLF, XLE, etc.)
   - ETF info (P/E ratios, yields)

2. **FRED** (Federal Reserve Economic Data)
   - Treasury rates (DGS10, DGS2)
   - CPI inflation data
   - GDP

3. **BLS** (Bureau of Labor Statistics)
   - Employment by sector
   - Used for growth signals

4. **Damodaran** (NYU)
   - R&D intensity by industry
   - Currently uses hardcoded averages (Excel parsing TODO)

## Signal Calculations

### Momentum (25% default)
- 12-month price returns (50%)
- Relative strength vs S&P 500 (35%)
- Volume trend (15%)

### Valuation (20% default)
- Forward P/E relative to other sectors
- Lower P/E = higher score

### Growth (20% default)
- Year-over-year employment growth

### Innovation (20% default)
- R&D intensity (R&D/Revenue)

### Macro (15% default)
- Interest rate sensitivity
- Lower correlation with rates = higher score

## Deployment on Replit

```nix
# replit.nix
{ pkgs }: {
  deps = [
    pkgs.go
  ];
}
```

```toml
# .replit
run = "./sector-analyzer"

[nix]
channel = "stable-23_11"
```

Build command:
```bash
go build -o sector-analyzer .
```

The binary serves both the API and frontend from a single process, eliminating the startup race conditions and multi-process overhead of the Python+Node version.

## License

MIT
