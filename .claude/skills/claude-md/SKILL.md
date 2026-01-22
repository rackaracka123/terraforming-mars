---
name: claude-md
description: Create or update CLAUDE.md files that provide efficient context for Claude Code. This skill should be used when the user asks to create a new CLAUDE.md file, update an existing one, or asks about CLAUDE.md best practices. Ensures files contain only relevant, actionable information without redundancy.
---

# CLAUDE.md File Creator

## Overview

Create and maintain CLAUDE.md files that provide efficient, actionable context for Claude Code when working in a codebase. Focus on locality of information - details belong in the closest parent CLAUDE.md to where they apply.

## Core Principles

### 1. Hierarchy and Locality

CLAUDE.md files form a hierarchy. Each file should only contain information specific to its scope:

- **Root CLAUDE.md**: Project overview, how to run/build/test, cross-cutting workflows
- **Subdirectory CLAUDE.md**: Details specific to that directory and its contents
- **Never duplicate**: If a parent covers it, don't repeat in children

When a subdirectory has its own CLAUDE.md, the parent should only provide a high-level overview of that subdirectory (1-2 sentences), not details.

### 2. No Specific Filenames

Never reference specific files like `extractor.go` or `utils.ts`. Instead:

- Reference folders: `internal/action/`, `src/components/`
- Reference patterns: "action files", "React components"
- Reference concepts: "the card system", "event handlers"

### 3. Context Efficiency

Include only what helps Claude work effectively:

- **Include**: Commands to run, architectural patterns, critical rules, workflows
- **Exclude**: Obvious information, things Claude can infer, documentation duplicated elsewhere

## Workflow

### Creating a New CLAUDE.md

1. **Analyze the directory scope**
   - What code/functionality lives here?
   - Are there subdirectories with their own CLAUDE.md files?
   - What does a developer need to know to work here effectively?

2. **Check parent CLAUDE.md files**
   - What's already documented at higher levels?
   - Avoid duplicating existing content

3. **Determine essential content**
   - For root: run/build/test commands, project overview, cross-cutting workflows
   - For subdirectories: architecture specific to this area, patterns, local conventions

4. **Write with locality in mind**
   - If subdirectory has own CLAUDE.md, provide only 1-2 sentence overview
   - Details belong in the most specific applicable location

### Updating an Existing CLAUDE.md

1. **Read the current file completely**
2. **Identify what needs to change** (add, remove, or modify)
3. **Check for hierarchy conflicts**
   - Is this information better placed in a parent or child CLAUDE.md?
   - Does this duplicate information elsewhere?
4. **Make minimal, targeted edits**

## CLAUDE.md Structure Template

```markdown
# [Directory Name] - [Brief Purpose]

[1-2 sentence overview of what this directory/project contains]

## Overview (if needed for complex areas)

[Architecture description using folder names, not file names]

## Commands (root level only)

```bash
make run          # Start development servers
make test         # Run test suite
make lint         # Check code quality
```

## Architecture / Structure

[Describe patterns and organization using folder references]

## Key Patterns

[Document conventions and approaches specific to this scope]

## Important Notes

[Critical information, common pitfalls, must-know rules]
```

## Anti-Patterns to Avoid

### Wrong: Referencing specific files
```markdown
The main logic is in `calculator.go` and `validator.go`
```

### Correct: Reference folders and concepts
```markdown
Calculator logic lives in `internal/calculator/`, with separate validation
```

### Wrong: Duplicating parent information
```markdown
## Testing
Run `make test` from project root...
[Already covered in root CLAUDE.md]
```

### Correct: Reference parent or add only new details
```markdown
## Testing
Tests mirror the `internal/` structure. See root CLAUDE.md for commands.
```

### Wrong: Documenting obvious patterns
```markdown
## How to Add a File
Create a new .go file in the appropriate directory...
```

### Correct: Document non-obvious patterns
```markdown
## Adding New Actions
Actions must extend BaseAction and implement Execute() method.
Wire in `cmd/server/` dependency injection.
```

## Content Guidelines

### For Root CLAUDE.md
- Project overview (1-2 sentences)
- Quick start commands
- Essential commands (run, build, test, lint, format)
- Cross-cutting workflows that span multiple directories
- High-level architecture overview
- Important project-wide rules and reminders

### For Subdirectory CLAUDE.md
- Focused overview of this directory's purpose
- Architecture patterns specific to this area
- Key development patterns and conventions
- Directory structure using folder names only
- Important notes and pitfalls for this scope
- References to related CLAUDE.md files (parent, siblings)

### Referencing Subdirectories with Own CLAUDE.md

When a subdirectory has its own CLAUDE.md, the parent should mention it briefly:

```markdown
## Architecture

**Backend** (`backend/`)
Go-based API server. See `backend/CLAUDE.md` for details.

**Frontend** (`frontend/`)
React application with Three.js. See `frontend/CLAUDE.md` for details.
```

Not:
```markdown
## Backend Architecture
[Detailed explanation that belongs in backend/CLAUDE.md]
```
