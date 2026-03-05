package proofpack

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadScenariosFixture(t *testing.T) {
	scenarios, err := LoadScenarios()
	require.NoError(t, err)
	require.Len(t, scenarios, 3)

	for _, scenario := range scenarios {
		require.NotEmpty(t, scenario.Name)
		require.NotEmpty(t, scenario.VideoPath)
		require.NotEmpty(t, scenario.Query)
		require.NotEmpty(t, scenario.ExpectedEvidence)
		require.Greater(t, scenario.MinResults, 0)
	}
}

func TestLoadDomainProfileScenariosFixture(t *testing.T) {
	scenarios, err := LoadDomainProfileScenarios()
	require.NoError(t, err)
	require.Len(t, scenarios, 2)

	for _, scenario := range scenarios {
		require.NotEmpty(t, scenario.Name)
		require.NotEmpty(t, scenario.VideoPath)
		require.NotEmpty(t, scenario.Query)
		require.NotEmpty(t, scenario.ExpectedEvidence)
		require.Greater(t, scenario.MinResults, 0)
	}
}

func TestRunScenarioDeterministicReplay(t *testing.T) {
	scenario := Scenario{
		Name:             "deterministic",
		VideoPath:        "https://example.com/test.mp4",
		Query:            "budget",
		ExpectedEvidence: []string{"budget", "roadmap"},
		MinResults:       2,
	}

	indexCalls := 0
	searchCalls := 0
	indexFn := func(_ context.Context, videoPath string) error {
		indexCalls++
		require.Equal(t, scenario.VideoPath, videoPath)
		return nil
	}
	searchFn := func(_ context.Context, query string, limit int) ([]SearchHit, error) {
		searchCalls++
		require.Equal(t, scenario.Query, query)
		require.Equal(t, 2, limit)
		return []SearchHit{
			{Snippet: "roadmap and budget review", Type: "audio", Similarity: 0.9},
			{Snippet: "closing remarks and next actions", Type: "audio", Similarity: 0.8},
		}, nil
	}

	result, err := RunScenario(context.Background(), scenario, 0, indexFn, searchFn)
	require.NoError(t, err)
	require.Equal(t, 1, indexCalls)
	require.Equal(t, 2, searchCalls)
	require.Equal(t, scenario.Name, result.ScenarioName)
	require.Equal(t, 2, result.ResultCount)
	require.True(t, result.Deterministic)
	require.Equal(t, 2, result.EvidenceMatched)

	payload, err := MarshalResult(result)
	require.NoError(t, err)
	require.NotEmpty(t, payload)
}

func TestRunScenarioDetectsNonDeterministicOrdering(t *testing.T) {
	scenario := Scenario{
		Name:             "nondeterministic",
		VideoPath:        "https://example.com/test.mp4",
		Query:            "budget",
		ExpectedEvidence: []string{"budget"},
		MinResults:       2,
	}

	searchRun := 0
	searchFn := func(_ context.Context, _ string, _ int) ([]SearchHit, error) {
		searchRun++
		if searchRun == 1 {
			return []SearchHit{{Snippet: "budget", Type: "audio", Similarity: 0.9}, {Snippet: "roadmap", Type: "audio", Similarity: 0.8}}, nil
		}
		return []SearchHit{{Snippet: "roadmap", Type: "audio", Similarity: 0.8}, {Snippet: "budget", Type: "audio", Similarity: 0.9}}, nil
	}

	result, err := RunScenario(context.Background(), scenario, 2, func(context.Context, string) error { return nil }, searchFn)
	require.NoError(t, err)
	require.False(t, result.Deterministic)
}
