// Package api provides HTTP handlers and response schemas.
package api

import "sector-analyzer/analysis"

// SectorScoreResponse is the JSON response for a single sector score.
type SectorScoreResponse struct {
	Sector           string   `json:"sector"`
	OpportunityScore float64  `json:"opportunity_score"`
	Rank             int      `json:"rank"`
	MomentumScore    float64  `json:"momentum_score"`
	ValuationScore   float64  `json:"valuation_score"`
	GrowthScore      float64  `json:"growth_score"`
	InnovationScore  float64  `json:"innovation_score"`
	MacroScore       float64  `json:"macro_score"`
	PriceReturn3Mo   *float64 `json:"price_return_3mo"`
	PriceReturn6Mo   *float64 `json:"price_return_6mo"`
	PriceReturn12Mo  *float64 `json:"price_return_12mo"`
	RelativeStrength *float64 `json:"relative_strength"`
	ForwardPE        *float64 `json:"forward_pe"`
	EmploymentGrowth *float64 `json:"employment_growth"`
	RDIntensity      *float64 `json:"rd_intensity"`
}

// ScoresResponse is the JSON response for all sector scores.
type ScoresResponse struct {
	Scores      []SectorScoreResponse `json:"scores"`
	WeightsUsed map[string]float64    `json:"weights_used"`
	Timestamp   string                `json:"timestamp"`
}

// SummaryResponse is the JSON response for summary report.
type SummaryResponse struct {
	TopSectors        []analysis.SectorRank          `json:"top_sectors"`
	BottomSectors     []analysis.SectorRank          `json:"bottom_sectors"`
	ScoreDistribution analysis.ScoreDistribution     `json:"score_distribution"`
	TopSectorDrivers  []string                       `json:"top_sector_drivers"`
	WeightsUsed       map[string]float64             `json:"weights_used"`
	Timestamp         string                         `json:"timestamp"`
}

// SectorListResponse contains list of available sectors.
type SectorListResponse struct {
	Sectors []string `json:"sectors"`
}

// CacheInfoResponse contains cache statistics.
type CacheInfoResponse struct {
	TotalFiles   int     `json:"total_files"`
	ValidFiles   int     `json:"valid_files"`
	ExpiredFiles int     `json:"expired_files"`
	TotalSizeMB  float64 `json:"total_size_mb"`
}

// CacheClearResponse is the result of cache clear operation.
type CacheClearResponse struct {
	FilesRemoved int    `json:"files_removed"`
	Message      string `json:"message"`
}

// HealthResponse is the health check response.
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

// ErrorResponse is used for error responses.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// ToSectorScoreResponse converts analysis.SectorScore to API response.
func ToSectorScoreResponse(s analysis.SectorScore) SectorScoreResponse {
	return SectorScoreResponse{
		Sector:           s.Sector,
		OpportunityScore: s.OpportunityScore,
		Rank:             s.Rank,
		MomentumScore:    s.MomentumScore,
		ValuationScore:   s.ValuationScore,
		GrowthScore:      s.GrowthScore,
		InnovationScore:  s.InnovationScore,
		MacroScore:       s.MacroScore,
		PriceReturn3Mo:   s.PriceReturn3Mo,
		PriceReturn6Mo:   s.PriceReturn6Mo,
		PriceReturn12Mo:  s.PriceReturn12Mo,
		RelativeStrength: s.RelativeStrength,
		ForwardPE:        s.ForwardPE,
		EmploymentGrowth: s.EmploymentGrowth,
		RDIntensity:      s.RDIntensity,
	}
}
