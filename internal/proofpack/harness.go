package proofpack

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"
)

type SearchHit struct {
	Snippet    string
	Type       string
	Similarity float64
}

type IndexFn func(ctx context.Context, videoPath string) error

type SearchFn func(ctx context.Context, query string, limit int) ([]SearchHit, error)

type ScenarioRunResult struct {
	ScenarioName         string        `json:"scenarioName"`
	Query                string        `json:"query"`
	IndexDuration        time.Duration `json:"indexDuration"`
	FirstSearchDuration  time.Duration `json:"firstSearchDuration"`
	SecondSearchDuration time.Duration `json:"secondSearchDuration"`
	ResultCount          int           `json:"resultCount"`
	Deterministic        bool          `json:"deterministic"`
	EvidenceMatched      int           `json:"evidenceMatched"`
}

func RunScenario(ctx context.Context, scenario Scenario, limit int, indexFn IndexFn, searchFn SearchFn) (ScenarioRunResult, error) {
	if limit <= 0 {
		limit = scenario.MinResults
	}

	indexStart := time.Now()
	if err := indexFn(ctx, scenario.VideoPath); err != nil {
		return ScenarioRunResult{}, fmt.Errorf("index scenario video: %w", err)
	}
	indexDuration := time.Since(indexStart)

	firstStart := time.Now()
	firstResults, err := searchFn(ctx, scenario.Query, limit)
	if err != nil {
		return ScenarioRunResult{}, fmt.Errorf("first search run: %w", err)
	}
	firstDuration := time.Since(firstStart)

	secondStart := time.Now()
	secondResults, err := searchFn(ctx, scenario.Query, limit)
	if err != nil {
		return ScenarioRunResult{}, fmt.Errorf("second search run: %w", err)
	}
	secondDuration := time.Since(secondStart)

	evidenceMatched := countEvidenceMatches(firstResults, scenario.ExpectedEvidence)

	return ScenarioRunResult{
		ScenarioName:         scenario.Name,
		Query:                scenario.Query,
		IndexDuration:        indexDuration,
		FirstSearchDuration:  firstDuration,
		SecondSearchDuration: secondDuration,
		ResultCount:          len(firstResults),
		Deterministic:        areResultsEquivalent(firstResults, secondResults),
		EvidenceMatched:      evidenceMatched,
	}, nil
}

func MarshalResult(result ScenarioRunResult) ([]byte, error) {
	payload, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("marshal scenario run result: %w", err)
	}
	return payload, nil
}

func areResultsEquivalent(left []SearchHit, right []SearchHit) bool {
	if len(left) != len(right) {
		return false
	}
	for idx := range left {
		if left[idx] != right[idx] {
			return false
		}
	}
	return true
}

func countEvidenceMatches(results []SearchHit, expectedEvidence []string) int {
	snippets := make([]string, 0, len(results))
	for _, result := range results {
		snippets = append(snippets, strings.ToLower(result.Snippet))
	}

	matches := 0
	for _, evidence := range expectedEvidence {
		needle := strings.ToLower(strings.TrimSpace(evidence))
		if needle == "" {
			continue
		}
		if slices.ContainsFunc(snippets, func(snippet string) bool {
			return strings.Contains(snippet, needle)
		}) {
			matches++
		}
	}
	return matches
}
