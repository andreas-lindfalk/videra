package proofpack

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

type Scenario struct {
	Name             string   `json:"name"`
	VideoPath        string   `json:"videoPath"`
	Query            string   `json:"query"`
	ExpectedEvidence []string `json:"expectedEvidence"`
	MinResults       int      `json:"minResults"`
}

//go:embed fixtures/scenarios.json
var scenariosFixture []byte

func LoadScenarios() ([]Scenario, error) {
	var scenarios []Scenario
	if err := json.Unmarshal(scenariosFixture, &scenarios); err != nil {
		return nil, fmt.Errorf("decode scenarios fixture: %w", err)
	}
	return scenarios, nil
}
