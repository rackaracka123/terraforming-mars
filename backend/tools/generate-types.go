package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"terraforming-mars-backend/pkg/typegen"
)

func main() {
	// Get the current working directory
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}
	
	// Initialize the TypeScript generator
	generator := typegen.NewTypeScriptGenerator()
	
	// Generate types from domain package
	domainPath := filepath.Join(wd, "internal", "domain")
	content, err := generator.GenerateFromPackage(domainPath)
	if err != nil {
		log.Fatalf("Failed to generate types from domain package: %v", err)
	}
	
	// Write to frontend types directory
	outputPath := filepath.Join(wd, "..", "frontend", "src", "types", "generated", "api-types.ts")
	if err := generator.WriteToFile(content, outputPath); err != nil {
		log.Fatalf("Failed to write generated types: %v", err)
	}
	
	fmt.Printf("Successfully generated TypeScript types at: %s\n", outputPath)
	
	// Also generate types documentation
	docsPath := filepath.Join(wd, "docs", "generated-types.md")
	docsContent := generateTypesDocs(content)
	if err := generator.WriteToFile(docsContent, docsPath); err != nil {
		log.Printf("Warning: Failed to write types documentation: %v", err)
	} else {
		fmt.Printf("Generated types documentation at: %s\n", docsPath)
	}
}

// generateTypesDocs creates markdown documentation for the generated types
func generateTypesDocs(tsContent string) string {
	return fmt.Sprintf(`# Generated TypeScript Types

This file documents the automatically generated TypeScript interfaces from Go structs.

**DO NOT EDIT** - This file is auto-generated from Go domain models.

## Generated Interfaces

%s

## Usage

Import the types in your TypeScript/React code:

%s

## Regeneration

To regenerate these types, run:

%s
`,
		"```typescript\n"+tsContent+"\n```",
		"```typescript\nimport { GameState, Player, Corporation } from '../types/generated/api-types';\n```",
		"```bash\ncd backend && go run tools/generate-types.go\n```",
	)
}