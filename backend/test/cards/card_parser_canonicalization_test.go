package cards

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestCardParserCanonicalOutput verifies that the card parser produces deterministic,
// canonicalized output by running it multiple times and comparing the results.
// This ensures that any changes to the parser output can be reliably detected.
func TestCardParserCanonicalOutput(t *testing.T) {
	const numRuns = 5
	var outputs []string
	var hashes []string

	// Create temp directory for output files using mktemp
	mktempCmd := exec.Command("mktemp", "-d")
	tempDirBytes, err := mktempCmd.Output()
	if err != nil {
		t.Fatalf("Failed to create temp directory with mktemp: %v", err)
	}
	tempDir := strings.TrimSpace(string(tempDirBytes))

	// Clean up temp directory after test
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: failed to clean up temp directory %s: %v", tempDir, err)
		}
	}()

	// Run the card parser multiple times
	for i := 0; i < numRuns; i++ {
		outputFile := filepath.Join(tempDir, "cards_"+fmt.Sprintf("%d", i+1)+".json")

		// Run the card parser from the backend directory where CSV files are located
		// Use absolute path for output file to avoid path issues
		absOutputFile, err := filepath.Abs(outputFile)
		if err != nil {
			t.Fatalf("Failed to get absolute path for output file: %v", err)
		}

		// Get the backend directory (two levels up from test/cards/)
		backendDir, err := filepath.Abs("../..")
		if err != nil {
			t.Fatalf("Failed to get backend directory path: %v", err)
		}

		cmd := exec.Command("go", "run", "tools/parse_cards.go", absOutputFile)
		cmd.Dir = backendDir

		t.Logf("Running command: %s from dir: %s", cmd.String(), cmd.Dir)
		t.Logf("Output file: %s", absOutputFile)

		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to run card parser (run %d): %v\nOutput: %s", i+1, err, output)
		}

		t.Logf("Parser output (run %d): %s", i+1, string(output))

		// Debug: List temp directory contents
		if entries, err := os.ReadDir(tempDir); err == nil {
			t.Logf("Temp dir contents after run %d:", i+1)
			for _, entry := range entries {
				t.Logf("  %s", entry.Name())
			}
		}

		// Check if file exists before trying to read it
		if _, err := os.Stat(absOutputFile); os.IsNotExist(err) {
			t.Fatalf("Output file does not exist: %s", absOutputFile)
		}

		// Read the generated JSON file using absolute path
		jsonBytes, err := os.ReadFile(absOutputFile)
		if err != nil {
			t.Fatalf("Failed to read output file (run %d): %v", i+1, err)
		}

		// Verify it's valid JSON
		var cards []interface{}
		if err := json.Unmarshal(jsonBytes, &cards); err != nil {
			t.Fatalf("Generated output is not valid JSON (run %d): %v", i+1, err)
		}

		// Store the raw JSON string
		jsonString := string(jsonBytes)
		outputs = append(outputs, jsonString)

		// Calculate SHA256 hash of the output
		hash := sha256.Sum256(jsonBytes)
		hashString := hex.EncodeToString(hash[:])
		hashes = append(hashes, hashString)

		t.Logf("Run %d: Generated %d cards, hash: %s", i+1, len(cards), hashString)
	}

	// Compare all outputs to the first one
	firstOutput := outputs[0]
	firstHash := hashes[0]

	for i := 1; i < numRuns; i++ {
		if outputs[i] != firstOutput {
			t.Errorf("Output from run %d differs from run 1", i+1)
			t.Errorf("Run 1 hash: %s", firstHash)
			t.Errorf("Run %d hash: %s", i+1, hashes[i])

			// Save both outputs for manual inspection
			if err := os.WriteFile(filepath.Join(tempDir, "run1.json"), []byte(firstOutput), 0644); err == nil {
				if err := os.WriteFile(filepath.Join(tempDir, "run"+string(rune('1'+i))+".json"), []byte(outputs[i]), 0644); err == nil {
					t.Errorf("Outputs saved to %s for manual comparison", tempDir)
				}
			}

			return
		}
	}

	t.Logf("SUCCESS: All %d runs produced identical output (hash: %s)", numRuns, firstHash)
}
