package turn

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadFile parses a YAML file containing TurnEval test cases.
// Supports both "tests:" and "conversations:" as the top-level list key.
func LoadFile(path string) (*TurnEvalFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read eval file %s: %w", path, err)
	}
	return ParseBytes(data)
}

// ParseBytes parses raw YAML bytes into a TurnEvalFile.
func ParseBytes(data []byte) (*TurnEvalFile, error) {
	var f TurnEvalFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parse eval YAML: %w", err)
	}
	return &f, nil
}

// ApplyGlobalConfig merges global_config defaults into each TestCase.
// Currently merges default tags; extend as needed.
func ApplyGlobalConfig(f *TurnEvalFile) {
	tests := f.AllTests()
	for i := range tests {
		if len(f.GlobalConfig.Tags) > 0 {
			seen := make(map[string]bool, len(tests[i].Tags))
			for _, t := range tests[i].Tags {
				seen[t] = true
			}
			for _, t := range f.GlobalConfig.Tags {
				if !seen[t] {
					tests[i].Tags = append(tests[i].Tags, t)
				}
			}
		}
	}
}
