// Package analysis provides signal calculations for sector opportunity scoring.
package analysis

import (
	"math"

	"gonum.org/v1/gonum/stat"

	"sector-analyzer/config"
	"sector-analyzer/data"
)

// NormalizeScore normalizes values to 0-100 scale using min-max normalization.
func NormalizeScore(values map[string]float64, higherIsBetter bool) map[string]float64 {
	if len(values) == 0 {
		return map[string]float64{}
	}

	// Find min and max
	var minVal, maxVal float64
	first := true
	for _, v := range values {
		if first {
			minVal, maxVal = v, v
			first = false
		} else {
			if v < minVal {
				minVal = v
			}
			if v > maxVal {
				maxVal = v
			}
		}
	}

	// Handle case where all values are the same
	if maxVal == minVal {
		result := make(map[string]float64)
		for k := range values {
			result[k] = 50.0
		}
		return result
	}

	result := make(map[string]float64)
	for sector, val := range values {
		score := ((val - minVal) / (maxVal - minVal)) * 100
		if !higherIsBetter {
			score = 100 - score
		}
		result[sector] = math.Round(score*100) / 100
	}

	return result
}

// NormalizeScoreZScore normalizes values using z-score, then converts to 0-100.
func NormalizeScoreZScore(values map[string]float64, higherIsBetter bool) map[string]float64 {
	if len(values) == 0 {
		return map[string]float64{}
	}

	// Calculate mean and standard deviation
	vals := make([]float64, 0, len(values))
	for _, v := range values {
		vals = append(vals, v)
	}

	mean := stat.Mean(vals, nil)
	std := stat.StdDev(vals, nil)

	// Handle case where std is 0
	if std == 0 {
		result := make(map[string]float64)
		for k := range values {
			result[k] = 50.0
		}
		return result
	}

	result := make(map[string]float64)
	for sector, val := range values {
		z := (val - mean) / std
		// Convert z-score to roughly 0-100 scale
		score := 50 + (z * 15)
		// Clamp to 0-100
		if score < 0 {
			score = 0
		}
		if score > 100 {
			score = 100
		}
		if !higherIsBetter {
			score = 100 - score
		}
		result[sector] = math.Round(score*100) / 100
	}

	return result
}

// CalculatePriceReturns calculates price returns for multiple periods.
func CalculatePriceReturns(prices data.SectorPrices) map[string]map[string]float64 {
	returns := make(map[string]map[string]float64)

	for sector, series := range prices {
		if sector == "_benchmark" || len(series) < 20 {
			continue
		}

		sectorReturns := make(map[string]float64)
		for _, months := range config.MomentumPeriods {
			tradingDays := months * 21
			if len(series) >= tradingDays {
				startPrice := series[len(series)-tradingDays].Close
				endPrice := series[len(series)-1].Close
				ret := ((endPrice - startPrice) / startPrice) * 100
				sectorReturns[periodKey(months)] = ret
			}
		}

		if len(sectorReturns) > 0 {
			returns[sector] = sectorReturns
		}
	}

	return returns
}

func periodKey(months int) string {
	switch months {
	case 3:
		return "3mo"
	case 6:
		return "6mo"
	case 12:
		return "12mo"
	default:
		return ""
	}
}

// CalculateRelativeStrength calculates relative strength vs benchmark (S&P 500).
func CalculateRelativeStrength(prices data.SectorPrices, periodMonths int) map[string]float64 {
	benchmarkSeries, ok := prices["_benchmark"]
	if !ok || len(benchmarkSeries) == 0 {
		return map[string]float64{}
	}

	tradingDays := periodMonths * 21
	if len(benchmarkSeries) < tradingDays {
		return map[string]float64{}
	}

	benchmarkReturn := (benchmarkSeries[len(benchmarkSeries)-1].Close/benchmarkSeries[len(benchmarkSeries)-tradingDays].Close - 1) * 100

	relStrength := make(map[string]float64)
	for sector, series := range prices {
		if sector == "_benchmark" || len(series) < tradingDays {
			continue
		}
		sectorReturn := (series[len(series)-1].Close/series[len(series)-tradingDays].Close - 1) * 100
		relStrength[sector] = sectorReturn - benchmarkReturn
	}

	return relStrength
}

// CalculateVolumeTrend calculates volume trend (short-term avg vs long-term avg).
func CalculateVolumeTrend(prices data.SectorPrices, shortPeriod, longPeriod int) map[string]float64 {
	trends := make(map[string]float64)

	for sector, series := range prices {
		if sector == "_benchmark" || len(series) < longPeriod {
			continue
		}

		// Calculate short-term average volume
		var shortSum int64
		for i := len(series) - shortPeriod; i < len(series); i++ {
			shortSum += series[i].Volume
		}
		shortAvg := float64(shortSum) / float64(shortPeriod)

		// Calculate long-term average volume
		var longSum int64
		for i := len(series) - longPeriod; i < len(series); i++ {
			longSum += series[i].Volume
		}
		longAvg := float64(longSum) / float64(longPeriod)

		if longAvg > 0 {
			trend := ((shortAvg - longAvg) / longAvg) * 100
			trends[sector] = trend
		}
	}

	return trends
}

// CalculateMomentumScore calculates combined momentum score.
func CalculateMomentumScore(prices data.SectorPrices) map[string]float64 {
	returns := CalculatePriceReturns(prices)
	relStrength := CalculateRelativeStrength(prices, 12)
	volumeTrend := CalculateVolumeTrend(prices, 20, 50)

	// Extract 12-month returns
	returns12mo := make(map[string]float64)
	for sector, rets := range returns {
		if ret, ok := rets["12mo"]; ok {
			returns12mo[sector] = ret
		}
	}

	// Normalize each component
	normReturns := NormalizeScoreZScore(returns12mo, true)
	normRelStrength := NormalizeScoreZScore(relStrength, true)
	normVolume := NormalizeScoreZScore(volumeTrend, true)

	// Combine with weights: 50% returns, 35% relative strength, 15% volume
	momentumScores := make(map[string]float64)
	for _, sector := range config.SectorNames {
		retScore := getOrDefault(normReturns, sector, 50.0)
		rsScore := getOrDefault(normRelStrength, sector, 50.0)
		volScore := getOrDefault(normVolume, sector, 50.0)

		combined := (0.50 * retScore) + (0.35 * rsScore) + (0.15 * volScore)
		momentumScores[sector] = math.Round(combined*100) / 100
	}

	return momentumScores
}

// CalculateValuationScore calculates valuation score based on P/E ratios.
func CalculateValuationScore(sectorPE map[string]float64, sectorInfo map[string]data.SectorInfo) map[string]float64 {
	// Build P/E map from available sources
	peMap := make(map[string]float64)

	// Try sectorPE first
	for sector, pe := range sectorPE {
		if pe > 0 {
			peMap[sector] = pe
		}
	}

	// Fall back to sectorInfo
	for sector, info := range sectorInfo {
		if _, exists := peMap[sector]; !exists && info.ForwardPE != nil && *info.ForwardPE > 0 {
			peMap[sector] = *info.ForwardPE
		}
	}

	if len(peMap) == 0 {
		return defaultScores()
	}

	// Lower P/E = better value = higher score
	scores := NormalizeScoreZScore(peMap, false)

	// Fill missing sectors
	for _, sector := range config.SectorNames {
		if _, exists := scores[sector]; !exists {
			scores[sector] = 50.0
		}
	}

	return scores
}

// CalculateEmploymentGrowth calculates YoY employment growth by sector.
func CalculateEmploymentGrowth(employment data.EmploymentData) map[string]float64 {
	growthRates := make(map[string]float64)

	for sector, ts := range employment {
		if len(ts.Values) < 13 {
			continue
		}

		current := ts.Values[len(ts.Values)-1]
		yearAgo := ts.Values[len(ts.Values)-13]

		if yearAgo > 0 {
			growth := ((current - yearAgo) / yearAgo) * 100
			growthRates[sector] = growth
		}
	}

	return growthRates
}

// CalculateGrowthScore calculates growth score based on employment trends.
func CalculateGrowthScore(employment data.EmploymentData) map[string]float64 {
	growth := CalculateEmploymentGrowth(employment)

	if len(growth) == 0 {
		return defaultScores()
	}

	scores := NormalizeScoreZScore(growth, true)

	// Fill missing sectors
	for _, sector := range config.SectorNames {
		if _, exists := scores[sector]; !exists {
			scores[sector] = 50.0
		}
	}

	return scores
}

// CalculateInnovationScore calculates innovation score based on R&D intensity.
func CalculateInnovationScore(rdData data.RDData) map[string]float64 {
	if len(rdData) == 0 {
		return defaultScores()
	}

	// Filter out zeros
	validRD := make(map[string]float64)
	for sector, rd := range rdData {
		if rd > 0 {
			validRD[sector] = rd
		}
	}

	if len(validRD) == 0 {
		return defaultScores()
	}

	scores := NormalizeScoreZScore(validRD, true)

	// Fill missing sectors with below-average score
	for _, sector := range config.SectorNames {
		if _, exists := scores[sector]; !exists {
			scores[sector] = 30.0
		}
	}

	return scores
}

// CalculateRateSensitivity calculates sector sensitivity to interest rate changes.
func CalculateRateSensitivity(prices data.SectorPrices, interestRates data.TimeSeries) map[string]float64 {
	if len(interestRates.Values) == 0 {
		return map[string]float64{}
	}

	sensitivities := make(map[string]float64)

	// Calculate monthly rate changes
	rateChanges := monthlyChanges(interestRates)

	for sector, series := range prices {
		if sector == "_benchmark" || len(series) < 252 {
			continue
		}

		// Calculate monthly returns
		sectorReturns := monthlyReturnsFromPrices(series)

		// Align and calculate correlation
		if len(sectorReturns) < 12 || len(rateChanges) < 12 {
			continue
		}

		// Use the minimum length
		n := min(len(sectorReturns), len(rateChanges))
		if n < 12 {
			continue
		}

		// Take most recent n values
		returns := sectorReturns[len(sectorReturns)-n:]
		rates := rateChanges[len(rateChanges)-n:]

		corr := stat.Correlation(returns, rates, nil)
		if !math.IsNaN(corr) {
			sensitivities[sector] = corr
		}
	}

	return sensitivities
}

// monthlyChanges calculates month-over-month percentage changes.
func monthlyChanges(ts data.TimeSeries) []float64 {
	if len(ts.Values) < 2 {
		return nil
	}

	var changes []float64
	for i := 1; i < len(ts.Values); i++ {
		if ts.Values[i-1] != 0 {
			change := (ts.Values[i] - ts.Values[i-1]) / ts.Values[i-1]
			changes = append(changes, change)
		}
	}
	return changes
}

// monthlyReturnsFromPrices calculates monthly returns from daily prices.
func monthlyReturnsFromPrices(series data.PriceSeries) []float64 {
	if len(series) < 21 {
		return nil
	}

	var returns []float64
	// Approximate monthly by taking every 21 trading days
	for i := 21; i < len(series); i += 21 {
		prevPrice := series[i-21].Close
		currPrice := series[i].Close
		if prevPrice > 0 {
			ret := (currPrice - prevPrice) / prevPrice
			returns = append(returns, ret)
		}
	}
	return returns
}

// CalculateMacroScore calculates macro sensitivity score.
func CalculateMacroScore(prices data.SectorPrices, macroData data.MacroData) map[string]float64 {
	interestRates, ok := macroData["treasury_10y"]
	if !ok {
		return defaultScores()
	}

	sensitivity := CalculateRateSensitivity(prices, interestRates)

	if len(sensitivity) == 0 {
		return defaultScores()
	}

	// Lower correlation with rates = more resilient = higher score
	scores := NormalizeScoreZScore(sensitivity, false)

	// Fill missing sectors
	for _, sector := range config.SectorNames {
		if _, exists := scores[sector]; !exists {
			scores[sector] = 50.0
		}
	}

	return scores
}

// Helper functions

func getOrDefault(m map[string]float64, key string, def float64) float64 {
	if v, ok := m[key]; ok {
		return v
	}
	return def
}

func defaultScores() map[string]float64 {
	scores := make(map[string]float64)
	for _, sector := range config.SectorNames {
		scores[sector] = 50.0
	}
	return scores
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
