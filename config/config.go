// Package config contains all configuration constants and mappings.
package config

import "time"

// SectorETFs maps GICS sector names to their SPDR ETF tickers.
var SectorETFs = map[string]string{
	"Information Technology":   "XLK",
	"Financials":               "XLF",
	"Energy":                   "XLE",
	"Health Care":              "XLV",
	"Consumer Discretionary":   "XLY",
	"Consumer Staples":         "XLP",
	"Industrials":              "XLI",
	"Materials":                "XLB",
	"Utilities":                "XLU",
	"Real Estate":              "XLRE",
	"Communication Services":   "XLC",
}

// SectorNames is an ordered list of all sector names.
var SectorNames = []string{
	"Information Technology",
	"Financials",
	"Energy",
	"Health Care",
	"Consumer Discretionary",
	"Consumer Staples",
	"Industrials",
	"Materials",
	"Utilities",
	"Real Estate",
	"Communication Services",
}

// MarketBenchmark is the S&P 500 ETF for relative strength calculations.
const MarketBenchmark = "SPY"

// CacheDuration is how long cached data remains valid.
const CacheDuration = 12 * time.Hour

// DefaultWeights for scoring categories.
var DefaultWeights = map[string]float64{
	"momentum":   0.25,
	"valuation":  0.20,
	"growth":     0.20,
	"innovation": 0.20,
	"macro":      0.15,
}

// MomentumPeriods in months for return calculations.
var MomentumPeriods = []int{3, 6, 12}

// MacroSensitivityYears is the historical period for macro calculations.
const MacroSensitivityYears = 5

// PEHistoricalYears is the period for P/E comparison.
const PEHistoricalYears = 5

// BLSEmploymentSeries maps sectors to BLS CES series IDs.
var BLSEmploymentSeries = map[string]string{
	"Information Technology":   "CES6000000001",
	"Financials":               "CES5500000001",
	"Energy":                   "CES1021000001",
	"Health Care":              "CES6562000001",
	"Consumer Discretionary":   "CES4200000001",
	"Consumer Staples":         "CES3100000001",
	"Industrials":              "CES3000000001",
	"Materials":                "CES1021200001",
	"Utilities":                "CES4422000001",
	"Real Estate":              "CES5553000001",
	"Communication Services":   "CES5000000001",
}

// FREDSeries maps names to FRED series identifiers.
var FREDSeries = map[string]string{
	"treasury_10y": "DGS10",
	"treasury_2y":  "DGS2",
	"fed_funds":    "FEDFUNDS",
	"cpi":          "CPIAUCSL",
	"core_cpi":     "CPILFESL",
	"gdp":          "GDP",
}

// DamodaranRDURL is the URL for R&D intensity data.
const DamodaranRDURL = "https://pages.stern.nyu.edu/~adamodar/pc/datasets/R&D.xls"

// DamodaranToGICS maps Damodaran industry names to GICS sectors.
var DamodaranToGICS = map[string]string{
	// Information Technology
	"Software (System & Application)": "Information Technology",
	"Software (Entertainment)":        "Information Technology",
	"Software (Internet)":             "Information Technology",
	"Semiconductor":                   "Information Technology",
	"Semiconductor Equip":             "Information Technology",
	"Computer Services":               "Information Technology",
	"Computers/Peripherals":           "Information Technology",
	"Electronics (Consumer & Office)": "Information Technology",
	"Electronics (General)":           "Information Technology",
	// Financials
	"Banks (Regional)":                        "Financials",
	"Banks (Money Center)":                    "Financials",
	"Financial Svcs. (Non-bank & Insurance)":  "Financials",
	"Insurance (General)":                     "Financials",
	"Insurance (Life)":                        "Financials",
	"Insurance (Prop/Cas.)":                   "Financials",
	"Brokerage & Investment Banking":          "Financials",
	// Energy
	"Oil/Gas (Production and Exploration)": "Energy",
	"Oil/Gas (Integrated)":                 "Energy",
	"Oil/Gas Distribution":                 "Energy",
	"Oilfield Svcs/Equip.":                 "Energy",
	// Health Care
	"Healthcare Products":                    "Health Care",
	"Healthcare Support Services":            "Health Care",
	"Healthcare Information and Technology":  "Health Care",
	"Hospitals/Healthcare Facilities":        "Health Care",
	"Drugs (Pharmaceutical)":                 "Health Care",
	"Drugs (Biotechnology)":                  "Health Care",
	"Medical Supplies":                       "Health Care",
	// Consumer Discretionary
	"Retail (General)":       "Consumer Discretionary",
	"Retail (Online)":        "Consumer Discretionary",
	"Retail (Special Lines)": "Consumer Discretionary",
	"Auto & Truck":           "Consumer Discretionary",
	"Auto Parts":             "Consumer Discretionary",
	"Apparel":                "Consumer Discretionary",
	"Restaurant/Dining":      "Consumer Discretionary",
	"Hotel/Gaming":           "Consumer Discretionary",
	// Consumer Staples
	"Household Products":    "Consumer Staples",
	"Food Processing":       "Consumer Staples",
	"Beverage (Alcoholic)":  "Consumer Staples",
	"Beverage (Soft)":       "Consumer Staples",
	"Tobacco":               "Consumer Staples",
	// Industrials
	"Aerospace/Defense":         "Industrials",
	"Air Transport":             "Industrials",
	"Trucking":                  "Industrials",
	"Transportation":            "Industrials",
	"Machinery":                 "Industrials",
	"Industrial Services":       "Industrials",
	"Building Materials":        "Industrials",
	"Engineering/Construction":  "Industrials",
	// Materials
	"Metals & Mining":          "Materials",
	"Steel":                    "Materials",
	"Chemical (Basic)":         "Materials",
	"Chemical (Diversified)":   "Materials",
	"Chemical (Specialty)":     "Materials",
	"Paper/Forest Products":    "Materials",
	"Packaging & Container":    "Materials",
	// Utilities
	"Utility (General)": "Utilities",
	"Utility (Water)":   "Utilities",
	"Power":             "Utilities",
	// Real Estate
	"R.E.I.T.":                               "Real Estate",
	"Real Estate (General/Diversified)":      "Real Estate",
	"Real Estate (Development)":              "Real Estate",
	"Real Estate (Operations & Services)":    "Real Estate",
	// Communication Services
	"Telecom Services":          "Communication Services",
	"Telecom. Equipment":        "Communication Services",
	"Broadcasting":              "Communication Services",
	"Cable TV":                  "Communication Services",
	"Entertainment":             "Communication Services",
	"Publishing & Newspapers":   "Communication Services",
	"Advertising":               "Communication Services",
}
