package types

import (
	"fmt"
	"sort"
	"strings"
)

// ResultAggregator provides advanced result filtering and grouping
type ResultAggregator struct {
	results []ValidationResult
}

// NewResultAggregator creates a new result aggregator
func NewResultAggregator(results []ValidationResult) *ResultAggregator {
	return &ResultAggregator{
		results: results,
	}
}

// AggregationOptions defines options for result aggregation
type AggregationOptions struct {
	FilterBySeverity []string // Filter by severity levels
	FilterByType     []string // Filter by validation types
	FilterByFile     []string // Filter by file patterns
	FilterByResource []string // Filter by resource patterns
	GroupBy          string   // Group by: severity, type, file, resource
	SortBy           string   // Sort by: severity, type, file, resource, line
	SortOrder        string   // Sort order: asc, desc
	Limit            int      // Limit number of results
	IncludeStats     bool     // Include statistics in output
	ShowOnlyErrors   bool     // Show only error-level results
	ShowOnlyWarnings bool     // Show only warning-level results
	ShowOnlyInfo     bool     // Show only info-level results
}

// AggregatedResults represents aggregated validation results
type AggregatedResults struct {
	Results       []ValidationResult
	Statistics    ResultStatistics
	Groups        map[string][]ValidationResult
	FilteredCount int
	TotalCount    int
}

// ResultStatistics provides statistics about validation results
type ResultStatistics struct {
	TotalResults      int
	ErrorCount        int
	WarningCount      int
	InfoCount         int
	ByType            map[string]int
	BySeverity        map[string]int
	ByFile            map[string]int
	MostCommonTypes   []TypeCount
	MostCommonFiles   []FileCount
	SeverityBreakdown SeverityBreakdown
}

// TypeCount represents count of results by type
type TypeCount struct {
	Type  string
	Count int
}

// FileCount represents count of results by file
type FileCount struct {
	File  string
	Count int
}

// SeverityBreakdown provides detailed severity statistics
type SeverityBreakdown struct {
	Errors   int
	Warnings int
	Info     int
	Unknown  int
}

// Aggregate performs result aggregation based on options
func (ra *ResultAggregator) Aggregate(options AggregationOptions) *AggregatedResults {
	// Start with all results
	filteredResults := ra.results

	// Apply filters
	filteredResults = ra.applyFilters(filteredResults, options)

	// Group results if requested
	groups := make(map[string][]ValidationResult)
	if options.GroupBy != "" {
		groups = ra.groupResults(filteredResults, options.GroupBy)
	}

	// Sort results
	if options.SortBy != "" {
		filteredResults = ra.sortResults(filteredResults, options.SortBy, options.SortOrder)
	}

	// Limit results
	if options.Limit > 0 && options.Limit < len(filteredResults) {
		filteredResults = filteredResults[:options.Limit]
	}

	// Calculate statistics
	statistics := ra.calculateStatistics(ra.results)

	return &AggregatedResults{
		Results:       filteredResults,
		Statistics:    statistics,
		Groups:        groups,
		FilteredCount: len(filteredResults),
		TotalCount:    len(ra.results),
	}
}

// applyFilters applies filters to results
func (ra *ResultAggregator) applyFilters(results []ValidationResult, options AggregationOptions) []ValidationResult {
	var filtered []ValidationResult

	for _, result := range results {
		// Severity filter
		if len(options.FilterBySeverity) > 0 {
			if !ra.stringInSlice(result.Severity, options.FilterBySeverity) {
				continue
			}
		}

		// Type filter
		if len(options.FilterByType) > 0 {
			if !ra.stringInSlice(result.Type, options.FilterByType) {
				continue
			}
		}

		// File filter
		if len(options.FilterByFile) > 0 {
			if !ra.matchesPatterns(result.File, options.FilterByFile) {
				continue
			}
		}

		// Resource filter
		if len(options.FilterByResource) > 0 {
			if !ra.matchesPatterns(result.Resource, options.FilterByResource) {
				continue
			}
		}

		// Show only errors
		if options.ShowOnlyErrors && result.Severity != "error" {
			continue
		}

		// Show only warnings
		if options.ShowOnlyWarnings && result.Severity != "warning" {
			continue
		}

		// Show only info
		if options.ShowOnlyInfo && result.Severity != "info" {
			continue
		}

		filtered = append(filtered, result)
	}

	return filtered
}

// groupResults groups results by the specified field
func (ra *ResultAggregator) groupResults(results []ValidationResult, groupBy string) map[string][]ValidationResult {
	groups := make(map[string][]ValidationResult)

	for _, result := range results {
		var key string
		switch groupBy {
		case "severity":
			key = result.Severity
		case "type":
			key = result.Type
		case "file":
			key = result.File
		case "resource":
			key = result.Resource
		default:
			key = "unknown"
		}

		groups[key] = append(groups[key], result)
	}

	return groups
}

// sortResults sorts results by the specified field
func (ra *ResultAggregator) sortResults(results []ValidationResult, sortBy, sortOrder string) []ValidationResult {
	sorted := make([]ValidationResult, len(results))
	copy(sorted, results)

	sort.Slice(sorted, func(i, j int) bool {
		var valueI, valueJ string

		switch sortBy {
		case "severity":
			valueI, valueJ = sorted[i].Severity, sorted[j].Severity
		case "type":
			valueI, valueJ = sorted[i].Type, sorted[j].Type
		case "file":
			valueI, valueJ = sorted[i].File, sorted[j].File
		case "resource":
			valueI, valueJ = sorted[i].Resource, sorted[j].Resource
		case "line":
			// Sort by line number numerically
			if sortOrder == "desc" {
				return sorted[i].Line > sorted[j].Line
			}
			return sorted[i].Line < sorted[j].Line
		default:
			return false
		}

		if sortOrder == "desc" {
			return valueI > valueJ
		}
		return valueI < valueJ
	})

	return sorted
}

// calculateStatistics calculates statistics for results
func (ra *ResultAggregator) calculateStatistics(results []ValidationResult) ResultStatistics {
	stats := ResultStatistics{
		TotalResults: len(results),
		ByType:       make(map[string]int),
		BySeverity:   make(map[string]int),
		ByFile:       make(map[string]int),
	}

	for _, result := range results {
		// Count by severity
		stats.BySeverity[result.Severity]++
		switch result.Severity {
		case "error":
			stats.ErrorCount++
		case "warning":
			stats.WarningCount++
		case "info":
			stats.InfoCount++
		default:
			stats.SeverityBreakdown.Unknown++
		}

		// Count by type
		stats.ByType[result.Type]++

		// Count by file
		stats.ByFile[result.File]++
	}

	// Calculate most common types
	stats.MostCommonTypes = ra.calculateMostCommon(stats.ByType, 10)

	// Calculate most common files
	stats.MostCommonFiles = ra.calculateMostCommonFiles(stats.ByFile, 10)

	// Set severity breakdown
	stats.SeverityBreakdown.Errors = stats.ErrorCount
	stats.SeverityBreakdown.Warnings = stats.WarningCount
	stats.SeverityBreakdown.Info = stats.InfoCount

	return stats
}

// calculateMostCommon calculates most common items from a count map
func (ra *ResultAggregator) calculateMostCommon(countMap map[string]int, limit int) []TypeCount {
	var items []TypeCount
	for item, count := range countMap {
		items = append(items, TypeCount{Type: item, Count: count})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Count > items[j].Count
	})

	if limit > 0 && limit < len(items) {
		return items[:limit]
	}
	return items
}

// calculateMostCommonFiles calculates most common files from a count map
func (ra *ResultAggregator) calculateMostCommonFiles(countMap map[string]int, limit int) []FileCount {
	var items []FileCount
	for file, count := range countMap {
		items = append(items, FileCount{File: file, Count: count})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Count > items[j].Count
	})

	if limit > 0 && limit < len(items) {
		return items[:limit]
	}
	return items
}

// Helper methods

func (ra *ResultAggregator) stringInSlice(str string, slice []string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func (ra *ResultAggregator) matchesPatterns(str string, patterns []string) bool {
	for _, pattern := range patterns {
		if strings.Contains(str, pattern) {
			return true
		}
	}
	return false
}

// GetSummary returns a summary of the aggregated results
func (ar *AggregatedResults) GetSummary() string {
	var summary strings.Builder

	summary.WriteString("Validation Summary:\n")
	summary.WriteString(fmt.Sprintf("  Total Results: %d\n", ar.TotalCount))
	summary.WriteString(fmt.Sprintf("  Filtered Results: %d\n", ar.FilteredCount))
	summary.WriteString(fmt.Sprintf("  Errors: %d\n", ar.Statistics.ErrorCount))
	summary.WriteString(fmt.Sprintf("  Warnings: %d\n", ar.Statistics.WarningCount))
	summary.WriteString(fmt.Sprintf("  Info: %d\n", ar.Statistics.InfoCount))

	if len(ar.Statistics.MostCommonTypes) > 0 {
		summary.WriteString("\nMost Common Issues:\n")
		for i, item := range ar.Statistics.MostCommonTypes {
			if i >= 5 { // Show top 5
				break
			}
			summary.WriteString(fmt.Sprintf("  %s: %d\n", item.Type, item.Count))
		}
	}

	return summary.String()
}
