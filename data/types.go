// Package data contains type definitions for market data.
package data

import "time"

// PriceBar represents a single price record.
type PriceBar struct {
	Date   time.Time `json:"date"`
	Open   float64   `json:"open"`
	High   float64   `json:"high"`
	Low    float64   `json:"low"`
	Close  float64   `json:"close"`
	Volume int64     `json:"volume"`
}

// PriceSeries is a slice of price bars ordered by date.
type PriceSeries []PriceBar

// SectorPrices maps sector names to their price series.
type SectorPrices map[string]PriceSeries

// SectorInfo contains metadata about a sector ETF.
type SectorInfo struct {
	ForwardPE     *float64 `json:"forward_pe"`
	TrailingPE    *float64 `json:"trailing_pe"`
	DividendYield *float64 `json:"dividend_yield"`
	AvgVolume     *int64   `json:"avg_volume"`
	MarketCap     *float64 `json:"market_cap"`
}

// TimeSeries represents a time-indexed series of float values.
type TimeSeries struct {
	Dates  []time.Time `json:"dates"`
	Values []float64   `json:"values"`
}

// MacroData contains FRED time series data.
type MacroData map[string]TimeSeries

// EmploymentData maps sectors to employment time series.
type EmploymentData map[string]TimeSeries

// RDData maps sectors to R&D intensity values.
type RDData map[string]float64

// AllData aggregates all fetched data sources.
type AllData struct {
	SectorPrices   SectorPrices           `json:"sector_prices"`
	SectorInfo     map[string]SectorInfo  `json:"sector_info"`
	MacroData      MacroData              `json:"macro_data"`
	EmploymentData EmploymentData         `json:"employment_data"`
	RDData         RDData                 `json:"rd_data"`
	FetchedAt      time.Time              `json:"fetched_at"`
}

// YahooFinanceResponse structures for parsing Yahoo Finance API responses.
type YahooFinanceChart struct {
	Chart struct {
		Result []struct {
			Meta struct {
				Symbol             string  `json:"symbol"`
				RegularMarketPrice float64 `json:"regularMarketPrice"`
				PreviousClose      float64 `json:"previousClose"`
			} `json:"meta"`
			Timestamp  []int64 `json:"timestamp"`
			Indicators struct {
				Quote []struct {
					Open   []float64 `json:"open"`
					High   []float64 `json:"high"`
					Low    []float64 `json:"low"`
					Close  []float64 `json:"close"`
					Volume []int64   `json:"volume"`
				} `json:"quote"`
			} `json:"indicators"`
		} `json:"result"`
		Error interface{} `json:"error"`
	} `json:"chart"`
}

// YahooQuoteSummary for getting ETF info.
type YahooQuoteSummary struct {
	QuoteSummary struct {
		Result []struct {
			SummaryDetail struct {
				ForwardPE     YahooValue `json:"forwardPE"`
				TrailingPE    YahooValue `json:"trailingPE"`
				DividendYield YahooValue `json:"dividendYield"`
			} `json:"summaryDetail"`
			DefaultKeyStatistics struct {
				ForwardPE YahooValue `json:"forwardPE"`
			} `json:"defaultKeyStatistics"`
		} `json:"result"`
	} `json:"quoteSummary"`
}

// YahooValue wraps numeric values from Yahoo API.
type YahooValue struct {
	Raw float64 `json:"raw"`
	Fmt string  `json:"fmt"`
}

// FREDResponse for parsing FRED API responses.
type FREDResponse struct {
	Observations []struct {
		Date  string `json:"date"`
		Value string `json:"value"`
	} `json:"observations"`
}

// BLSResponse for parsing BLS API responses.
type BLSResponse struct {
	Status  string `json:"status"`
	Message []string `json:"message"`
	Results struct {
		Series []struct {
			SeriesID string `json:"seriesID"`
			Data     []struct {
				Year   string `json:"year"`
				Period string `json:"period"`
				Value  string `json:"value"`
			} `json:"data"`
		} `json:"series"`
	} `json:"Results"`
}
