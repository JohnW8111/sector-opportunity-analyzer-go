// Package data provides data fetching from various financial APIs.
package data

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/extrame/xls"
	"sector-analyzer/config"
)

// FetchSectorPrices retrieves historical price data for all sector ETFs.
func FetchSectorPrices(period string) (SectorPrices, error) {
	cacheKey := GenerateKey("yfinance", map[string]interface{}{"type": "sector_prices", "period": period})
	if cached, ok := GlobalCache.Get(cacheKey); ok {
		return cached.(SectorPrices), nil
	}

	prices := make(SectorPrices)

	// Fetch all sector ETFs
	for sector, ticker := range config.SectorETFs {
		series, err := fetchYahooHistory(ticker, period)
		if err != nil {
			fmt.Printf("Error fetching %s (%s): %v\n", sector, ticker, err)
			continue
		}
		prices[sector] = series
	}

	// Fetch benchmark
	benchmarkSeries, err := fetchYahooHistory(config.MarketBenchmark, period)
	if err == nil {
		prices["_benchmark"] = benchmarkSeries
	}

	GlobalCache.Set(cacheKey, prices)
	return prices, nil
}

// fetchYahooHistory retrieves historical data from Yahoo Finance.
func fetchYahooHistory(ticker string, period string) (PriceSeries, error) {
	// Calculate time range
	end := time.Now()
	var start time.Time
	switch period {
	case "1y":
		start = end.AddDate(-1, 0, 0)
	case "2y":
		start = end.AddDate(-2, 0, 0)
	case "5y":
		start = end.AddDate(-5, 0, 0)
	default:
		start = end.AddDate(-5, 0, 0)
	}

	// Build Yahoo Finance API URL
	apiURL := fmt.Sprintf(
		"https://query1.finance.yahoo.com/v8/finance/chart/%s?period1=%d&period2=%d&interval=1d&includePrePost=false",
		url.PathEscape(ticker),
		start.Unix(),
		end.Unix(),
	)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("yahoo finance returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var chartResp YahooFinanceChart
	if err := json.Unmarshal(body, &chartResp); err != nil {
		return nil, err
	}

	if len(chartResp.Chart.Result) == 0 {
		return nil, fmt.Errorf("no data returned for %s", ticker)
	}

	result := chartResp.Chart.Result[0]
	if len(result.Indicators.Quote) == 0 {
		return nil, fmt.Errorf("no quote data for %s", ticker)
	}

	quote := result.Indicators.Quote[0]
	timestamps := result.Timestamp

	var series PriceSeries
	for i, ts := range timestamps {
		if i >= len(quote.Close) || quote.Close[i] == 0 {
			continue
		}

		bar := PriceBar{
			Date:  time.Unix(ts, 0),
			Close: quote.Close[i],
		}
		if i < len(quote.Open) {
			bar.Open = quote.Open[i]
		}
		if i < len(quote.High) {
			bar.High = quote.High[i]
		}
		if i < len(quote.Low) {
			bar.Low = quote.Low[i]
		}
		if i < len(quote.Volume) {
			bar.Volume = quote.Volume[i]
		}

		series = append(series, bar)
	}

	return series, nil
}

// FetchSectorInfo retrieves current info (P/E, etc.) for all sector ETFs.
func FetchSectorInfo() (map[string]SectorInfo, error) {
	cacheKey := GenerateKey("yfinance", map[string]interface{}{"type": "sector_info"})
	if cached, ok := GlobalCache.Get(cacheKey); ok {
		return cached.(map[string]SectorInfo), nil
	}

	info := make(map[string]SectorInfo)

	for sector, ticker := range config.SectorETFs {
		sectorInfo, err := fetchYahooInfo(ticker)
		if err != nil {
			fmt.Printf("Error fetching info for %s: %v\n", ticker, err)
			info[sector] = SectorInfo{}
			continue
		}
		info[sector] = sectorInfo
	}

	GlobalCache.Set(cacheKey, info)
	return info, nil
}

// fetchYahooInfo retrieves ETF info from Yahoo Finance.
func fetchYahooInfo(ticker string) (SectorInfo, error) {
	apiURL := fmt.Sprintf(
		"https://query1.finance.yahoo.com/v10/finance/quoteSummary/%s?modules=summaryDetail,defaultKeyStatistics",
		url.PathEscape(ticker),
	)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return SectorInfo{}, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return SectorInfo{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return SectorInfo{}, fmt.Errorf("yahoo finance returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return SectorInfo{}, err
	}

	var quoteSummary YahooQuoteSummary
	if err := json.Unmarshal(body, &quoteSummary); err != nil {
		return SectorInfo{}, err
	}

	if len(quoteSummary.QuoteSummary.Result) == 0 {
		return SectorInfo{}, nil
	}

	result := quoteSummary.QuoteSummary.Result[0]
	info := SectorInfo{}

	if result.SummaryDetail.ForwardPE.Raw > 0 {
		pe := result.SummaryDetail.ForwardPE.Raw
		info.ForwardPE = &pe
	} else if result.DefaultKeyStatistics.ForwardPE.Raw > 0 {
		pe := result.DefaultKeyStatistics.ForwardPE.Raw
		info.ForwardPE = &pe
	}

	if result.SummaryDetail.TrailingPE.Raw > 0 {
		pe := result.SummaryDetail.TrailingPE.Raw
		info.TrailingPE = &pe
	}

	if result.SummaryDetail.DividendYield.Raw > 0 {
		dy := result.SummaryDetail.DividendYield.Raw
		info.DividendYield = &dy
	}

	return info, nil
}

// FetchFREDSeries retrieves a single FRED time series.
func FetchFREDSeries(seriesID string, startDate time.Time) (TimeSeries, error) {
	apiKey := os.Getenv("FRED_API_KEY")
	if apiKey == "" {
		return TimeSeries{}, fmt.Errorf("FRED_API_KEY not set")
	}

	cacheKey := GenerateKey("fred", map[string]interface{}{
		"series_id":  seriesID,
		"start_date": startDate.Format("2006-01-02"),
	})
	if cached, ok := GlobalCache.Get(cacheKey); ok {
		return cached.(TimeSeries), nil
	}

	apiURL := fmt.Sprintf(
		"https://api.stlouisfed.org/fred/series/observations?series_id=%s&api_key=%s&file_type=json&observation_start=%s",
		seriesID,
		apiKey,
		startDate.Format("2006-01-02"),
	)

	resp, err := http.Get(apiURL)
	if err != nil {
		return TimeSeries{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return TimeSeries{}, fmt.Errorf("FRED API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return TimeSeries{}, err
	}

	var fredResp FREDResponse
	if err := json.Unmarshal(body, &fredResp); err != nil {
		return TimeSeries{}, err
	}

	var ts TimeSeries
	for _, obs := range fredResp.Observations {
		if obs.Value == "." {
			continue // Skip missing values
		}

		date, err := time.Parse("2006-01-02", obs.Date)
		if err != nil {
			continue
		}

		value, err := strconv.ParseFloat(obs.Value, 64)
		if err != nil {
			continue
		}

		ts.Dates = append(ts.Dates, date)
		ts.Values = append(ts.Values, value)
	}

	GlobalCache.Set(cacheKey, ts)
	return ts, nil
}

// FetchMacroData retrieves all FRED macro series.
func FetchMacroData(yearsBack int) (MacroData, error) {
	startDate := time.Now().AddDate(-yearsBack, 0, 0)
	data := make(MacroData)

	for name, seriesID := range config.FREDSeries {
		ts, err := FetchFREDSeries(seriesID, startDate)
		if err != nil {
			fmt.Printf("Error fetching FRED series %s: %v\n", seriesID, err)
			continue
		}
		data[name] = ts
	}

	return data, nil
}

// FetchBLSEmployment retrieves employment data from BLS.
func FetchBLSEmployment(yearsBack int) (EmploymentData, error) {
	cacheKey := GenerateKey("bls", map[string]interface{}{"type": "employment", "years": yearsBack})
	if cached, ok := GlobalCache.Get(cacheKey); ok {
		return cached.(EmploymentData), nil
	}

	endYear := time.Now().Year()
	startYear := endYear - yearsBack

	// Build series IDs list
	var seriesIDs []string
	seriesIDToSector := make(map[string]string)
	for sector, seriesID := range config.BLSEmploymentSeries {
		seriesIDs = append(seriesIDs, seriesID)
		seriesIDToSector[seriesID] = sector
	}

	// BLS API request
	payload := map[string]interface{}{
		"seriesid":  seriesIDs,
		"startyear": strconv.Itoa(startYear),
		"endyear":   strconv.Itoa(endYear),
	}

	// Add API key if available
	if apiKey := os.Getenv("BLS_API_KEY"); apiKey != "" {
		payload["registrationkey"] = apiKey
	}

	payloadBytes, _ := json.Marshal(payload)
	resp, err := http.Post(
		"https://api.bls.gov/publicAPI/v2/timeseries/data/",
		"application/json",
		strings.NewReader(string(payloadBytes)),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var blsResp BLSResponse
	if err := json.Unmarshal(body, &blsResp); err != nil {
		return nil, err
	}

	if blsResp.Status != "REQUEST_SUCCEEDED" {
		return nil, fmt.Errorf("BLS API error: %v", blsResp.Message)
	}

	data := make(EmploymentData)
	for _, series := range blsResp.Results.Series {
		sector, ok := seriesIDToSector[series.SeriesID]
		if !ok {
			continue
		}

		var ts TimeSeries
		for _, item := range series.Data {
			year, _ := strconv.Atoi(item.Year)
			monthStr := strings.TrimPrefix(item.Period, "M")
			month, _ := strconv.Atoi(monthStr)

			if month < 1 || month > 12 {
				continue
			}

			date := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
			value, err := strconv.ParseFloat(item.Value, 64)
			if err != nil {
				continue
			}

			ts.Dates = append(ts.Dates, date)
			ts.Values = append(ts.Values, value)
		}

		// Sort by date (BLS returns newest first)
		sortTimeSeries(&ts)
		data[sector] = ts
	}

	GlobalCache.Set(cacheKey, data)
	return data, nil
}

// sortTimeSeries sorts a time series by date ascending.
func sortTimeSeries(ts *TimeSeries) {
	// Simple bubble sort for small arrays
	n := len(ts.Dates)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if ts.Dates[j].After(ts.Dates[j+1]) {
				ts.Dates[j], ts.Dates[j+1] = ts.Dates[j+1], ts.Dates[j]
				ts.Values[j], ts.Values[j+1] = ts.Values[j+1], ts.Values[j]
			}
		}
	}
}

// FetchDamodaranRD fetches R&D intensity data from Damodaran's Excel file.
func FetchDamodaranRD() (RDData, error) {
	cacheKey := GenerateKey("damodaran", map[string]interface{}{"type": "rd_intensity"})
	if cached, ok := GlobalCache.Get(cacheKey); ok {
		return cached.(RDData), nil
	}

	// Try to fetch and parse live data
	data, err := fetchDamodaranExcel()
	if err != nil {
		fmt.Printf("Warning: Could not fetch Damodaran data: %v. Using defaults.\n", err)
		// Fallback to defaults
		data = getDefaultRDData()
	}

	GlobalCache.Set(cacheKey, data)
	return data, nil
}

// fetchDamodaranExcel downloads and parses the Damodaran R&D Excel file (old .xls format).
func fetchDamodaranExcel() (RDData, error) {
	resp, err := http.Get(config.DamodaranRDURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d from Damodaran", resp.StatusCode)
	}

	// Read the Excel file into memory
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse with xls library (supports old .xls format)
	xlsFile, err := xls.OpenReader(bytes.NewReader(body), "utf-8")
	if err != nil {
		return nil, fmt.Errorf("failed to parse Excel: %w", err)
	}

	// Get sheet 1 ("Industry Averages") - sheet 0 is "Variables & FAQ"
	sheet := xlsFile.GetSheet(1)
	if sheet == nil {
		return nil, fmt.Errorf("Industry Averages sheet not found")
	}

	// Aggregate R&D by GICS sector
	sectorRD := make(map[string][]float64)
	for sector := range config.SectorETFs {
		sectorRD[sector] = []float64{}
	}

	// Column 0 = Industry Name, Column 5 = "Current R&D as % of Revenue"
	industryCol := 0
	rdSalesCol := 5
	dataStartRow := 8 // Row 7 is headers, data starts at row 8

	// Iterate through rows
	maxRow := int(sheet.MaxRow)
	for i := dataStartRow; i <= maxRow; i++ {
		row := sheet.Row(i)
		if row == nil {
			continue
		}

		// Get industry name
		industry := strings.TrimSpace(row.Col(industryCol))
		if industry == "" {
			continue
		}

		// Get R&D as % of Revenue value
		rdValueStr := strings.TrimSpace(row.Col(rdSalesCol))
		if rdValueStr == "" {
			continue
		}

		// Parse R&D value (already as decimal, e.g., 0.15 = 15%)
		rdValue, err := strconv.ParseFloat(rdValueStr, 64)
		if err != nil {
			continue
		}

		// Skip invalid values
		if rdValue < 0 || rdValue > 1 {
			continue
		}

		// Map to GICS sector
		if gicsSector, ok := config.DamodaranToGICS[industry]; ok {
			sectorRD[gicsSector] = append(sectorRD[gicsSector], rdValue)
		}
	}

	// Calculate averages
	result := make(RDData)
	for sector, values := range sectorRD {
		if len(values) > 0 {
			var sum float64
			for _, v := range values {
				sum += v
			}
			result[sector] = sum / float64(len(values))
		} else {
			result[sector] = 0.0
		}
	}

	// Verify we got meaningful data
	nonZeroCount := 0
	for sector, v := range result {
		if v > 0 {
			fmt.Printf("  %s: %.2f%%\n", sector, v*100)
			nonZeroCount++
		}
	}

	if nonZeroCount < 3 {
		return nil, fmt.Errorf("insufficient R&D data extracted (only %d sectors)", nonZeroCount)
	}

	fmt.Printf("Successfully parsed Damodaran R&D data for %d sectors\n", nonZeroCount)
	return result, nil
}

// getDefaultRDData returns fallback R&D values based on historical Damodaran averages.
func getDefaultRDData() RDData {
	return RDData{
		"Information Technology":   0.15,
		"Health Care":              0.12,
		"Communication Services":   0.08,
		"Consumer Discretionary":   0.04,
		"Industrials":              0.03,
		"Materials":                0.02,
		"Consumer Staples":         0.02,
		"Financials":               0.01,
		"Energy":                   0.01,
		"Utilities":                0.005,
		"Real Estate":              0.001,
	}
}

// FetchAllData retrieves all data needed for sector analysis.
func FetchAllData() (*AllData, error) {
	fmt.Println("Fetching sector price data...")
	sectorPrices, _ := FetchSectorPrices("5y")

	fmt.Println("Fetching sector info...")
	sectorInfo, _ := FetchSectorInfo()

	fmt.Println("Fetching macro data from FRED...")
	macroData, _ := FetchMacroData(config.MacroSensitivityYears)

	fmt.Println("Fetching employment data from BLS...")
	employmentData, _ := FetchBLSEmployment(5)

	fmt.Println("Fetching R&D data...")
	rdData, _ := FetchDamodaranRD()

	return &AllData{
		SectorPrices:   sectorPrices,
		SectorInfo:     sectorInfo,
		MacroData:      macroData,
		EmploymentData: employmentData,
		RDData:         rdData,
		FetchedAt:      time.Now(),
	}, nil
}
