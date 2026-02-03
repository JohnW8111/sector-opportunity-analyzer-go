// Package api provides HTTP handlers for the sector analyzer API.
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"sector-analyzer/analysis"
	"sector-analyzer/config"
	"sector-analyzer/data"
)

// AppState holds the application state including cached data.
type AppState struct {
	mu         sync.RWMutex
	cachedData *data.AllData
}

// NewAppState creates a new application state.
func NewAppState() *AppState {
	return &AppState{}
}

// GetData returns cached data or fetches fresh data.
func (s *AppState) GetData() *data.AllData {
	s.mu.RLock()
	if s.cachedData != nil {
		s.mu.RUnlock()
		return s.cachedData
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check after acquiring write lock
	if s.cachedData != nil {
		return s.cachedData
	}

	allData, _ := data.FetchAllData()
	s.cachedData = allData
	return s.cachedData
}

// RefreshData forces a data refresh.
func (s *AppState) RefreshData() *data.AllData {
	s.mu.Lock()
	defer s.mu.Unlock()

	allData, _ := data.FetchAllData()
	s.cachedData = allData
	return s.cachedData
}

// Global app state
var appState = NewAppState()

// JSON helper for writing responses
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// parseWeights extracts scoring weights from query parameters.
func parseWeights(r *http.Request) map[string]float64 {
	weights := make(map[string]float64)

	params := []string{"momentum", "valuation", "growth", "innovation", "macro"}
	hasAny := false

	for _, param := range params {
		if val := r.URL.Query().Get(param); val != "" {
			if f, err := strconv.ParseFloat(val, 64); err == nil && f >= 0 && f <= 1 {
				weights[param] = f
				hasAny = true
			}
		}
	}

	if !hasAny {
		return nil
	}

	// Fill in defaults for missing weights
	defaults := config.DefaultWeights
	for _, param := range params {
		if _, ok := weights[param]; !ok {
			weights[param] = defaults[param]
		}
	}

	// Normalize to sum to 1.0
	var sum float64
	for _, v := range weights {
		sum += v
	}
	if sum > 0 {
		for k, v := range weights {
			weights[k] = v / sum
		}
	}

	return weights
}

// HealthHandler handles GET /health
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, HealthResponse{
		Status:  "ok",
		Version: "1.0.0",
	})
}

// RootHandler handles GET /
func RootHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"name":    "Sector Opportunity Analyzer",
		"version": "1.0.0",
		"status":  "running",
	})
}

// GetScoresHandler handles GET /api/scores
func GetScoresHandler(w http.ResponseWriter, r *http.Request) {
	// Check for refresh flag
	refresh := r.URL.Query().Get("refresh") == "true"

	var allData *data.AllData
	if refresh {
		allData = appState.RefreshData()
	} else {
		allData = appState.GetData()
	}

	if allData == nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "data_unavailable",
			Message: "Failed to fetch sector data",
		})
		return
	}

	// Parse weights from query params
	weights := parseWeights(r)
	scorer := analysis.NewSectorScorer(weights)
	scores := scorer.CalculateScores(allData)

	// Convert to response format
	var scoreResponses []SectorScoreResponse
	for _, s := range scores {
		scoreResponses = append(scoreResponses, ToSectorScoreResponse(s))
	}

	writeJSON(w, http.StatusOK, ScoresResponse{
		Scores:      scoreResponses,
		WeightsUsed: scorer.Weights,
		Timestamp:   time.Now().Format(time.RFC3339),
	})
}

// GetSummaryHandler handles GET /api/scores/summary
func GetSummaryHandler(w http.ResponseWriter, r *http.Request) {
	allData := appState.GetData()

	if allData == nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "data_unavailable",
			Message: "Failed to fetch sector data",
		})
		return
	}

	weights := parseWeights(r)
	scores, summary := analysis.RunAnalysis(allData, weights)

	// The summary already has most fields, just use it
	_ = scores // Used by RunAnalysis to generate summary

	writeJSON(w, http.StatusOK, SummaryResponse{
		TopSectors:        summary.TopSectors,
		BottomSectors:     summary.BottomSectors,
		ScoreDistribution: summary.ScoreDistribution,
		TopSectorDrivers:  summary.TopSectorDrivers,
		WeightsUsed:       summary.WeightsUsed,
		Timestamp:         summary.Timestamp,
	})
}

// GetSectorScoreHandler handles GET /api/scores/{sector}
func GetSectorScoreHandler(w http.ResponseWriter, r *http.Request) {
	// Extract sector from URL path
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_request",
			Message: "Sector name required",
		})
		return
	}
	sectorName := parts[len(parts)-1]

	allData := appState.GetData()
	if allData == nil {
		writeJSON(w, http.StatusInternalServerError, ErrorResponse{
			Error:   "data_unavailable",
			Message: "Failed to fetch sector data",
		})
		return
	}

	scorer := analysis.NewSectorScorer(nil)
	scores := scorer.CalculateScores(allData)

	// Find the sector (case-insensitive)
	for _, s := range scores {
		if strings.EqualFold(s.Sector, sectorName) {
			writeJSON(w, http.StatusOK, ToSectorScoreResponse(s))
			return
		}
	}

	writeJSON(w, http.StatusNotFound, ErrorResponse{
		Error:   "not_found",
		Message: "Sector '" + sectorName + "' not found",
	})
}

// GetSectorsHandler handles GET /api/data/sectors
func GetSectorsHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, SectorListResponse{
		Sectors: config.SectorNames,
	})
}

// DataSourceStatus represents the status of a data source.
type DataSourceStatus struct {
	Name    string  `json:"name"`
	Status  string  `json:"status"`
	Message *string `json:"message,omitempty"`
}

// DataQualityResponse contains data quality info for all sources.
type DataQualityResponse struct {
	Sources       []DataSourceStatus `json:"sources"`
	OverallStatus string             `json:"overall_status"`
}

// GetDataQualityHandler handles GET /api/data/quality
func GetDataQualityHandler(w http.ResponseWriter, r *http.Request) {
	allData := appState.GetData()

	sources := []DataSourceStatus{
		{Name: "Yahoo Finance", Status: "pending"},
		{Name: "FRED", Status: "pending"},
		{Name: "BLS", Status: "pending"},
		{Name: "Damodaran", Status: "pending"},
	}

	if allData == nil {
		msg := "Data not loaded yet"
		for i := range sources {
			sources[i].Status = "error"
			sources[i].Message = &msg
		}
		writeJSON(w, http.StatusOK, DataQualityResponse{
			Sources:       sources,
			OverallStatus: "error",
		})
		return
	}

	// Check Yahoo Finance (prices)
	if len(allData.SectorPrices) >= 10 {
		sources[0].Status = "ok"
		msg := fmt.Sprintf("%d sectors loaded", len(allData.SectorPrices)-1) // -1 for benchmark
		sources[0].Message = &msg
	} else if len(allData.SectorPrices) > 0 {
		sources[0].Status = "warning"
		msg := fmt.Sprintf("Only %d sectors loaded", len(allData.SectorPrices))
		sources[0].Message = &msg
	} else {
		sources[0].Status = "error"
		msg := "No price data available"
		sources[0].Message = &msg
	}

	// Check FRED (macro data)
	if len(allData.MacroData) >= 3 {
		sources[1].Status = "ok"
		msg := fmt.Sprintf("%d series loaded", len(allData.MacroData))
		sources[1].Message = &msg
	} else if len(allData.MacroData) > 0 {
		sources[1].Status = "warning"
		msg := fmt.Sprintf("Only %d series loaded", len(allData.MacroData))
		sources[1].Message = &msg
	} else {
		sources[1].Status = "warning"
		msg := "FRED_API_KEY may not be set"
		sources[1].Message = &msg
	}

	// Check BLS (employment)
	if len(allData.EmploymentData) >= 8 {
		sources[2].Status = "ok"
		msg := fmt.Sprintf("%d sectors loaded", len(allData.EmploymentData))
		sources[2].Message = &msg
	} else if len(allData.EmploymentData) > 0 {
		sources[2].Status = "warning"
		msg := fmt.Sprintf("Only %d sectors loaded", len(allData.EmploymentData))
		sources[2].Message = &msg
	} else {
		sources[2].Status = "warning"
		msg := "No employment data"
		sources[2].Message = &msg
	}

	// Check Damodaran (R&D)
	nonZeroRD := 0
	for _, v := range allData.RDData {
		if v > 0 {
			nonZeroRD++
		}
	}
	if nonZeroRD >= 8 {
		sources[3].Status = "ok"
		msg := fmt.Sprintf("%d sectors with R&D data", nonZeroRD)
		sources[3].Message = &msg
	} else if nonZeroRD > 0 {
		sources[3].Status = "warning"
		msg := fmt.Sprintf("Only %d sectors with R&D data", nonZeroRD)
		sources[3].Message = &msg
	} else {
		sources[3].Status = "error"
		msg := "R&D data failed to load"
		sources[3].Message = &msg
	}

	// Determine overall status
	overall := "ok"
	for _, s := range sources {
		if s.Status == "error" {
			overall = "error"
			break
		}
		if s.Status == "warning" && overall == "ok" {
			overall = "warning"
		}
	}

	writeJSON(w, http.StatusOK, DataQualityResponse{
		Sources:       sources,
		OverallStatus: overall,
	})
}

// GetCacheInfoHandler handles GET /api/cache/info
func GetCacheInfoHandler(w http.ResponseWriter, r *http.Request) {
	info := data.GlobalCache.Info()
	writeJSON(w, http.StatusOK, CacheInfoResponse{
		TotalFiles:   info.TotalEntries,
		ValidFiles:   info.ValidEntries,
		ExpiredFiles: info.ExpiredEntries,
		TotalSizeMB:  0.0, // In-memory cache doesn't track size
	})
}

// ClearCacheHandler handles POST /api/cache/clear
func ClearCacheHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{
			Error:   "method_not_allowed",
			Message: "Use POST to clear cache",
		})
		return
	}

	count := data.GlobalCache.Clear()
	writeJSON(w, http.StatusOK, CacheClearResponse{
		FilesRemoved: count,
		Message:      "Cache cleared successfully",
	})
}
