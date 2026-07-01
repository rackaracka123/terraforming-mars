package milestone

import (
	"encoding/json"
	"fmt"
	"os"
)

// LoadMilestonesFromJSON loads milestone definitions from a JSON file
func LoadMilestonesFromJSON(filepath string) ([]MilestoneDefinition, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read milestones file: %w", err)
	}

	var defs []MilestoneDefinition
	if err := json.Unmarshal(data, &defs); err != nil {
		return nil, fmt.Errorf("failed to parse milestones JSON: %w", err)
	}

	if len(defs) == 0 {
		return nil, fmt.Errorf("no milestones found in file: %s", filepath)
	}

	return defs, nil
}
