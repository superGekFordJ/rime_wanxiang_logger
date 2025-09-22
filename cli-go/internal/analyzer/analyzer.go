// Package analyzer provides functions to process and analyze Rime logger data.
package analyzer

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

// LogEvent defines the structure of a single log entry in the JSONL file.
// Fields are tagged for JSON unmarshalling and match the Lua script output.
type LogEvent struct {
	EventType             string   `json:"event_type"`
	SelectedCandidateRank *int     `json:"selected_candidate_rank,omitempty"`
	CommittedText         string   `json:"committed_text,omitempty"`
	SourceFirstCandidate  string   `json:"source_first_candidate,omitempty"`
	InputSequenceAtCommit string   `json:"input_sequence_at_commit,omitempty"`
	SourceCandidatesList  []string `json:"source_candidates_list,omitempty"`
	SourceInputBuffer     string   `json:"source_input_buffer,omitempty"`
	SelectionMethod       string   `json:"selection_method,omitempty"`
	Timestamp             string   `json:"timestamp,omitempty"`
}

// AnalysisResult holds the calculated metrics from the log file analysis.
// This matches the Python version's analysis output.
type AnalysisResult struct {
	// Raw data counts
	TotalCommits    int
	TotalSelections int
	RawInputCommits int

	// Accuracy metrics
	FirstChoiceCount     int
	Top3Count            int
	FirstChoiceHitRate   float64
	Top3HitRate          float64
	AverageRank          float64
	OverallAccuracyScore float64
	DirectInputRate      float64

	// Validation flags
	HasValidSelections bool
	HasCommits         bool
}

// ReadLogFile parses a JSONL log file into a slice of LogEvent structs.
// This matches the Python version: pd.read_json(lines=True)
func ReadLogFile(filePath string) ([]LogEvent, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open log file %s: %w", filePath, err)
	}
	defer file.Close()

	var events []LogEvent
	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		if line == "" {
			continue
		}

		var event LogEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			// Skip invalid JSON lines but continue processing
			fmt.Printf("Warning: Skipping invalid JSON on line %d: %v\n", lineNumber, err)
			continue
		}

		// Only include text_committed events for analysis
		if event.EventType == "text_committed" {
			events = append(events, event)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading log file: %w", err)
	}

	return events, nil
}

// PerformAnalysis calculates comprehensive metrics from a list of log events.
// This exactly matches the Python version's analyze() method logic.
func PerformAnalysis(events []LogEvent) AnalysisResult {
	var result AnalysisResult

	// df_commit = df[df['event_type'] == 'text_committed'].copy()
	// This filtering is already done in ReadLogFile, so events are all text_committed

	result.TotalCommits = len(events)
	result.HasCommits = result.TotalCommits > 0

	if !result.HasCommits {
		return result
	}

	// df_selections = df_commit[df_commit['selected_candidate_rank'] >= 0].copy()
	var validSelections []LogEvent
	var rawInputCount int

	for _, event := range events {
		if event.SelectedCandidateRank == nil {
			// This shouldn't happen for text_committed events, but handle gracefully
			continue
		}

		rank := *event.SelectedCandidateRank

		if rank >= 0 {
			// Valid candidate selection
			validSelections = append(validSelections, event)
		} else if rank == -1 {
			// Direct input (raw text, no candidate selection)
			rawInputCount++
		}
	}

	result.TotalSelections = len(validSelections)
	result.RawInputCommits = rawInputCount
	result.HasValidSelections = result.TotalSelections > 0

	// Calculate direct input rate
	// raw_input_commits / total_commits
	if result.TotalCommits > 0 {
		result.DirectInputRate = (float64(result.RawInputCommits) / float64(result.TotalCommits)) * 100
	}

	if !result.HasValidSelections {
		return result
	}

	// Calculate accuracy metrics for valid selections
	var totalRank int
	var accuracySum float64

	for _, event := range validSelections {
		rank := *event.SelectedCandidateRank
		totalRank += rank

		// first_choice_count = (df_selections['selected_candidate_rank'] == 0).sum()
		if rank == 0 {
			result.FirstChoiceCount++
		}

		// top_3_count = (df_selections['selected_candidate_rank'] < 3).sum()
		if rank < 3 {
			result.Top3Count++
		}

		// df_selections['accuracy_score'] = 1 / (df_selections['selected_candidate_rank'] + 1)
		accuracySum += 1.0 / float64(rank+1)
	}

	// Calculate rates and averages
	totalSelectionsFloat := float64(result.TotalSelections)

	// first_choice_count / total_selections
	result.FirstChoiceHitRate = (float64(result.FirstChoiceCount) / totalSelectionsFloat) * 100

	// top_3_count / total_selections
	result.Top3HitRate = (float64(result.Top3Count) / totalSelectionsFloat) * 100

	// df_selections['selected_candidate_rank'].mean()
	result.AverageRank = float64(totalRank) / totalSelectionsFloat

	// overall_accuracy_score = df_selections['accuracy_score'].mean()
	result.OverallAccuracyScore = accuracySum / totalSelectionsFloat

	return result
}

// ExportMisses filters for mispredictions (rank > 0) and writes them to a CSV file.
// The output format matches the Python version's CSV structure.
func ExportMisses(events []LogEvent, outputCsvPath string) error {
	// If no output path specified, use default like Python version
	if outputCsvPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			outputCsvPath = "rime_mispredictions_report.csv"
		} else {
			outputCsvPath = fmt.Sprintf("%s/rime_mispredictions_report.csv", homeDir)
		}
	}

	file, err := os.Create(outputCsvPath)
	if err != nil {
		return fmt.Errorf("could not create CSV file %s: %w", outputCsvPath, err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header (matching Python version column names)
	header := []string{"用户输入", "实际选择", "程序预测", "选择排名"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Collect mispredictions for sorting (matching Python logic)
	type MissRecord struct {
		UserInput    string
		ActualChoice string
		ProgramPred  string
		SelectedRank int
		Frequency    int
	}

	var misses []MissRecord
	missFrequency := make(map[string]int)

	// First pass: collect misses and count frequency
	// df_misses = df_commit[df_commit['selected_candidate_rank'] > 0].copy()
	for _, event := range events {
		if event.SelectedCandidateRank != nil && *event.SelectedCandidateRank > 0 {
			miss := MissRecord{
				UserInput:    event.SourceInputBuffer,
				ActualChoice: event.CommittedText,
				ProgramPred:  event.SourceFirstCandidate,
				SelectedRank: *event.SelectedCandidateRank,
			}
			misses = append(misses, miss)
			missFrequency[event.CommittedText]++
		}
	}

	// Second pass: add frequency information
	for i := range misses {
		misses[i].Frequency = missFrequency[misses[i].ActualChoice]
	}

	// Sort by frequency (descending) then by user input (ascending)
	// Python: sort_values(by=['错误频率', '用户输入'], ascending=[False, True])
	for i := 0; i < len(misses)-1; i++ {
		for j := 0; j < len(misses)-i-1; j++ {
			if misses[j].Frequency < misses[j+1].Frequency ||
				(misses[j].Frequency == misses[j+1].Frequency && misses[j].UserInput > misses[j+1].UserInput) {
				misses[j], misses[j+1] = misses[j+1], misses[j]
			}
		}
	}

	// Write data rows
	for _, miss := range misses {
		row := []string{
			miss.UserInput,
			miss.ActualChoice,
			miss.ProgramPred,
			strconv.Itoa(miss.SelectedRank),
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

// GetMissCount returns the number of mispredictions in the events.
func GetMissCount(events []LogEvent) int {
	count := 0
	for _, event := range events {
		if event.SelectedCandidateRank != nil && *event.SelectedCandidateRank > 0 {
			count++
		}
	}
	return count
}
