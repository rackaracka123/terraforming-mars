package typegen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// TypeScriptGenerator generates TypeScript interfaces from Go structs
type TypeScriptGenerator struct {
	output strings.Builder
}

// NewTypeScriptGenerator creates a new TypeScript generator
func NewTypeScriptGenerator() *TypeScriptGenerator {
	return &TypeScriptGenerator{}
}

// GenerateFromPackage generates TypeScript interfaces from all structs in a Go package
func (g *TypeScriptGenerator) GenerateFromPackage(packagePath string) (string, error) {
	g.output.Reset()
	
	// Write header
	g.output.WriteString("// Generated TypeScript interfaces from Go structs\n")
	g.output.WriteString("// DO NOT EDIT - This file is auto-generated\n\n")
	
	// Parse the package
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, packagePath, nil, parser.ParseComments)
	if err != nil {
		return "", fmt.Errorf("failed to parse package %s: %v", packagePath, err)
	}
	
	// Process each file in the package
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
			g.processFile(file)
		}
	}
	
	return g.output.String(), nil
}

// processFile processes a single Go file and extracts struct definitions
func (g *TypeScriptGenerator) processFile(file *ast.File) {
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.GenDecl:
			if node.Tok == token.TYPE {
				for _, spec := range node.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						if structType, ok := typeSpec.Type.(*ast.StructType); ok {
							g.processStruct(typeSpec.Name.Name, structType)
						} else if _, ok := typeSpec.Type.(*ast.Ident); ok {
							// Handle type aliases (like enums)
							g.processTypeAlias(typeSpec)
						}
					}
				}
			} else if node.Tok == token.CONST {
				// Handle constants (for enum values)
				g.processConstants(node)
			}
		}
		return true
	})
}

// processStruct converts a Go struct to a TypeScript interface
func (g *TypeScriptGenerator) processStruct(name string, structType *ast.StructType) {
	g.output.WriteString(fmt.Sprintf("export interface %s {\n", name))
	
	for _, field := range structType.Fields.List {
		g.processField(field)
	}
	
	g.output.WriteString("}\n\n")
}

// processField converts a Go struct field to a TypeScript interface property
func (g *TypeScriptGenerator) processField(field *ast.Field) {
	// Extract field name
	var fieldName string
	if len(field.Names) > 0 {
		fieldName = field.Names[0].Name
	} else {
		// Embedded field
		if ident, ok := field.Type.(*ast.Ident); ok {
			fieldName = ident.Name
		}
	}
	
	// Skip unexported fields
	if fieldName == "" || !ast.IsExported(fieldName) {
		return
	}
	
	// Extract TypeScript type from struct tag or infer from Go type
	tsType := g.extractTypeScriptType(field)
	
	// Extract JSON field name from tag
	jsonName := g.extractJSONFieldName(field.Tag)
	if jsonName == "" {
		jsonName = strings.ToLower(fieldName[:1]) + fieldName[1:] // camelCase
	}
	
	// Check if field is optional (pointer or omitempty)
	optional := g.isOptionalField(field)
	
	optionalMark := ""
	if optional {
		optionalMark = "?"
	}
	
	g.output.WriteString(fmt.Sprintf("  %s%s: %s;\n", jsonName, optionalMark, tsType))
}

// extractTypeScriptType converts Go type to TypeScript type
func (g *TypeScriptGenerator) extractTypeScriptType(field *ast.Field) string {
	// First check for explicit TypeScript type in struct tag
	if field.Tag != nil {
		tag := strings.Trim(field.Tag.Value, "`")
		if tsType := g.extractTSTag(tag); tsType != "" {
			return tsType
		}
	}
	
	// Infer from Go type
	return g.goTypeToTypeScript(field.Type)
}

// extractTSTag extracts TypeScript type from struct tag
func (g *TypeScriptGenerator) extractTSTag(tag string) string {
	parts := strings.Split(tag, " ")
	for _, part := range parts {
		if strings.HasPrefix(part, "ts:") {
			return strings.Trim(strings.TrimPrefix(part, "ts:"), "\"")
		}
	}
	return ""
}

// extractJSONFieldName extracts JSON field name from struct tag
func (g *TypeScriptGenerator) extractJSONFieldName(tag *ast.BasicLit) string {
	if tag == nil {
		return ""
	}
	
	tagValue := strings.Trim(tag.Value, "`")
	parts := strings.Split(tagValue, " ")
	
	for _, part := range parts {
		if strings.HasPrefix(part, "json:") {
			jsonTag := strings.Trim(strings.TrimPrefix(part, "json:"), "\"")
			jsonParts := strings.Split(jsonTag, ",")
			if len(jsonParts) > 0 && jsonParts[0] != "-" {
				return jsonParts[0]
			}
		}
	}
	
	return ""
}

// isOptionalField determines if a field should be optional in TypeScript
func (g *TypeScriptGenerator) isOptionalField(field *ast.Field) bool {
	// Check if it's a pointer type
	if _, ok := field.Type.(*ast.StarExpr); ok {
		return true
	}
	
	// Check for omitempty in JSON tag
	if field.Tag != nil {
		tag := strings.Trim(field.Tag.Value, "`")
		return strings.Contains(tag, "omitempty")
	}
	
	return false
}

// goTypeToTypeScript converts Go AST type to TypeScript type
func (g *TypeScriptGenerator) goTypeToTypeScript(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return g.identToTypeScript(t.Name)
	case *ast.ArrayType:
		elemType := g.goTypeToTypeScript(t.Elt)
		return elemType + "[]"
	case *ast.StarExpr:
		baseType := g.goTypeToTypeScript(t.X)
		return baseType + " | undefined"
	case *ast.MapType:
		keyType := g.goTypeToTypeScript(t.Key)
		valueType := g.goTypeToTypeScript(t.Value)
		return fmt.Sprintf("Record<%s, %s>", keyType, valueType)
	case *ast.InterfaceType:
		return "any"
	default:
		return "any"
	}
}

// identToTypeScript converts Go identifier to TypeScript type
func (g *TypeScriptGenerator) identToTypeScript(name string) string {
	switch name {
	case "string":
		return "string"
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "float32", "float64":
		return "number"
	case "bool":
		return "boolean"
	case "Time":
		return "string" // Time is serialized as ISO string
	default:
		// Custom types - assume they're defined elsewhere
		return name
	}
}

// processTypeAlias handles type aliases (like enum types)
func (g *TypeScriptGenerator) processTypeAlias(typeSpec *ast.TypeSpec) {
	if ident, ok := typeSpec.Type.(*ast.Ident); ok && ident.Name == "string" {
		// This looks like a string enum type
		g.output.WriteString(fmt.Sprintf("export type %s = string;\n\n", typeSpec.Name.Name))
	}
}

// processConstants handles constant declarations (enum values)
func (g *TypeScriptGenerator) processConstants(genDecl *ast.GenDecl) {
	// This could be enhanced to generate TypeScript enums from Go constants
	// For now, we'll skip this as the main focus is on struct types
}

// WriteToFile writes the generated TypeScript to a file
func (g *TypeScriptGenerator) WriteToFile(content, filePath string) error {
	// Ensure the directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", dir, err)
	}
	
	// Write the file
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %v", filePath, err)
	}
	
	return nil
}