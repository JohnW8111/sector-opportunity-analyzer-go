// Package analysis provides the scoring engine for sector opportunity analysis.
package analysis

import (
	"math"
	"sort"
	"time"

	"sector-analyzer/config"
	"sector-analyzer/data"
)

// SectorScore contains a sector's complete scoring breakdown.
type SectorScore struct {
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

// SectorScorer calculates opportunity scores for all sectors.
type SectorScorer struct {
	Weights map[string]float64
}

// NewSectorScorer creates a new scorer with optional custom weights.
func NewSectorScorer(weights map[string]float64) *SectorScorer {
	if weights == nil {
		weights = config.DefaultWeights
	}

	// Normalize weights to sum to 1.0
	var sum float64
	for _, w := range weights {
		sum += w
	}
	if sum > 0 && math.Abs(sum-1.0) > 0.01 {
		for k, v := range weights {
			weights[k] = v / sum
		}
	}

	return &SectorScorer{Weights: weights}
}

// CalculateScores computes opportunity scores for all sectors.
func (s *SectorScorer) CalculateScores(allData *data.AllData) []SectorScore {
	// Calculate component scores
	momentumScores := CalculateMomentumScore(allData.SectorPrices)
	valuationScores := CalculateValuationScore(nil, allData.SectorInfo)
	growthScores := CalculateGrowthScore(allData.EmploymentData)
	innovationScores := CalculateInnovationScore(allData.RDData)
	macroScores := CalculateMacroScore(allData.SectorPrices, allData.MacroData)

	// Calculate raw metrics for display
	priceReturns := CalculatePriceReturns(allData.SectorPrices)
	relStrength := CalculateRelativeStrength(allData.SectorPrices, 12)
	employmentGrowth := CalculateEmploymentGrowth(allData.EmploymentData)

	// Build sector scores
	var scores []SectorScore

	for _, sector := range config.SectorNames {
		momentum := getOrDefault(momentumScores, sector, 50.0)
		valuation := getOrDefault(valuationScores, sector, 50.0)
		growth := getOrDefault(growthScores, sector, 50.0)
		innovation := getOrDefault(innovationScores, sector, 50.0)
		macro := getOrDefault(macroScores, sector, 50.0)

		// Calculate weighted opportunity score
		opportunity := (s.Weights["momentum"] * momentum) +
			(s.Weights["valuation"] * valuation) +
			(s.Weights["growth"] * growth) +
			(s.Weights["innovation"] * innovation) +
			(s.Weights["macro"] * macro)

		score := SectorScore{
			Sector:           sector,
			OpportunityScore: math.Round(opportunity*100) / 100,
			MomentumScore:    momentum,
			ValuationScore:   valuation,
			GrowthScore:      growth,
			InnovationScore:  innovation,
			MacroScore:       macro,
		}

		// Add raw metrics
		if returns, ok := priceReturns[sector]; ok {
			if ret, ok := returns["3mo"]; ok {
				score.PriceReturn3Mo = &ret
			}
			if ret, ok := returns["6mo"]; ok {
				score.PriceReturn6Mo = &ret
			}
			if ret, ok := returns["12mo"]; ok {
				score.PriceReturn12Mo = &ret
			}
		}

		if rs, ok := relStrength[sector]; ok {
			score.RelativeStrength = &rs
		}

		if info, ok := allData.SectorInfo[sector]; ok && info.ForwardPE != nil {
			score.ForwardPE = info.ForwardPE
		}

		if eg, ok := employmentGrowth[sector]; ok {
			score.EmploymentGrowth = &eg
		}

		if rd, ok := allData.RDData[sector]; ok {
			score.RDIntensity = &rd
		}

		scores = append(scores, score)
	}

	// Sort by opportunity score and assign ranks
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].OpportunityScore > scores[j].OpportunityScore
	})

	for i := range scores {
		scores[i].Rank = i + 1
	}

	return scores
}

// SummaryReport contains summary statistics and insights.
type SummaryReport struct {
	Timestamp         string                `json:"timestamp"`
	TopSectors        []SectorRank          `json:"top_sectors"`
	BottomSectors     []SectorRank          `json:"bottom_sectors"`
	ScoreDistribution ScoreDistribution     `json:"score_distribution"`
	TopSectorDrivers  []string              `json:"top_sector_drivers"`
	WeightsUsed       map[string]float64    `json:"weights_used"`
}

// SectorRank contains basic ranking info.
type SectorRank struct {
	Rank   int     `json:"rank"`
	Sector string  `json:"sector"`
	Score  float64 `json:"score"`
}

// ScoreDistribution contains score statistics.
type ScoreDistribution struct {
	Average float64 `json:"average"`
	Max     float64 `json:"max"`
	Min     float64 `json:"min"`
	Spread  float64 `json:"spread"`
}

// GetSummaryReport generates a summary report from sector scores.
func (s *SectorScorer) GetSummaryReport(scores []SectorScore) SummaryReport {
	if len(scores) == 0 {
		return SummaryReport{Timestamp: time.Now().Format(time.RFC3339)}
	}

	// Top 3 and bottom 3
	var topSectors, bottomSectors []SectorRank

	for i := 0; i < 3 && i < len(scores); i++ {
		topSectors = append(topSectors, SectorRank{
			Rank:   scores[i].Rank,
			Sector: scores[i].Sector,
			Score:  scores[i].OpportunityScore,
		})
	}

	for i := len(scores) - 3; i < len(scores); i++ {
		if i >= 0 {
			bottomSectors = append(bottomSectors, SectorRank{
				Rank:   scores[i].Rank,
				Sector: scores[i].Sector,
				Score:  scores[i].OpportunityScore,
			})
		}
	}

	// Calculate distribution
	var sum, maxScore, minScore float64
	minScore = 100
	for _, score := range scores {
		sum += score.OpportunityScore
		if score.OpportunityScore > maxScore {
			maxScore = score.OpportunityScore
		}
		if score.OpportunityScore < minScore {
			minScore = score.OpportunityScore
		}
	}
	avgScore := sum / float64(len(scores))

	// Identify top sector drivers
	var drivers []string
	topSector := scores[0]
	if topSector.MomentumScore >= 70 {
		drivers = append(drivers, "strong momentum")
	}
	if topSector.ValuationScore >= 70 {
		drivers = append(drivers, "attractive valuation")
	}
	if topSector.GrowthScore >= 70 {
		drivers = append(drivers, "employment growth")
	}
	if topSector.InnovationScore >= 70 {
		drivers = append(drivers, "high R&D investment")
	}
	if topSector.MacroScore >= 70 {
		drivers = append(drivers, "favorable macro positioning")
	}

	return SummaryReport{
		Timestamp:     time.Now().Format(time.RFC3339),
		TopSectors:    topSectors,
		BottomSectors: bottomSectors,
		ScoreDistribution: ScoreDistribution{
			Average: math.Round(avgScore*100) / 100,
			Max:     math.Round(maxScore*100) / 100,
			Min:     math.Round(minScore*100) / 100,
			Spread:  math.Round((maxScore-minScore)*100) / 100,
		},
		TopSectorDrivers: drivers,
		WeightsUsed:      s.Weights,
	}
}

// RunAnalysis is a convenience function to run full analysis.
func RunAnalysis(allData *data.AllData, weights map[string]float64) ([]SectorScore, SummaryReport) {
	scorer := NewSectorScorer(weights)
	scores := scorer.CalculateScores(allData)
	summary := scorer.GetSummaryReport(scores)
	return scores, summary
}
