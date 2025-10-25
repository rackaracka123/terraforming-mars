# Data Transfer Objects (DTOs) - Frontend Communication Layer

**CRITICAL**: DTOs in this directory are the ONLY data structures exposed to the frontend. They define the contract between backend and frontend.

## Core Principles

1. **Frontend Contract**: DTOs are "artifacts" created from internal models before sending to frontend
2. **Bidirectional**: Frontend sends DTOs to backend, backend sends DTOs to frontend
3. **Must Generate TypeScript**: ALL DTOs MUST have `ts:` tags and be exported to TypeScript types
4. **NEVER Export Models**: Only DTOs are allowed in `tygo.yaml` - NEVER add model objects

## Development Workflow

### When Adding/Modifying DTOs

1. **Create/Update DTO struct** in `game_dto.go` with BOTH `json:` and `ts:` tags
2. **Add/Update mapper function** in `mapper.go` to convert model ↔ DTO
3. **Run `make generate`** to create TypeScript types (uses `tygo generate`)
4. **Verify TypeScript types** in `frontend/src/types/generated/api-types.ts`

### DTO Structure Template

```go
// game_dto.go
type PlayerDto struct {
    ID       string `json:"id" ts:"string"`
    Credits  int    `json:"credits" ts:"number"`
    IsActive bool   `json:"isActive" ts:"boolean"`
}

// mapper.go
func ToPlayerDto(player model.Player) PlayerDto {
    return PlayerDto{
        ID:       player.ID,
        Credits:  player.Resources.Credits,
        IsActive: player.IsActive,
    }
}
```

## TypeScript Tag Reference

```go
// Primitives
`ts:"string"`
`ts:"number"`
`ts:"boolean"`

// Nullable/Optional
`ts:"string | null"`
`ts:"number | undefined"`
`ts:"SomeDto | null | undefined"`

// Arrays
`ts:"string[]"`
`ts:"PlayerDto[]"`
`ts:"CardTag[] | undefined"`

// Maps/Records
`ts:"Record<string, number>"`
`ts:"Record<ResourceType, number> | undefined"`

// Enums
`ts:"GamePhase"`  // Uses enum from tygo.yaml config
```

## Synchronization Rules

**CRITICAL**: When model changes, DTO MUST be updated:

1. Model field added → Add to DTO with `ts:` tag
2. Model field removed → Remove from DTO
3. Model field renamed → Update DTO field name
4. Model field type changed → Update DTO type and `ts:` tag
5. **ALWAYS** update mapper function in `mapper.go`
6. **ALWAYS** run `make generate` after DTO changes

## Common Mistakes to Avoid

❌ **WRONG**: Forgetting `ts:` tags on DTO fields
❌ **WRONG**: Adding models to `tygo.yaml` (only DTOs allowed)
❌ **WRONG**: Using DTOs internally in backend (use models instead)
❌ **WRONG**: Updating DTO without updating mapper function
❌ **WRONG**: Not running `make generate` after DTO changes

✅ **CORRECT**: Every DTO field has both `json:` and `ts:` tags
✅ **CORRECT**: Only DTOs listed in `tygo.yaml` config
✅ **CORRECT**: DTOs used only at delivery layer boundaries
✅ **CORRECT**: Mapper functions updated with all DTO fields
✅ **CORRECT**: `make generate` run after every DTO change

## DTO Update Checklist

When you modify any DTO:

- [ ] Added/updated `json:` and `ts:` tags on all fields
- [ ] Updated corresponding mapper function in `mapper.go`
- [ ] Checked that new field is mapped from model correctly
- [ ] Ran `make generate` to regenerate TypeScript types
- [ ] Verified `frontend/src/types/generated/api-types.ts` updated
- [ ] Confirmed DTO NOT added to `tygo.yaml` if it's a new type (only add if frontend needs it)

## Why DTOs Exist

- **Stable Contract**: Frontend doesn't break when internal models change
- **Type Safety**: TypeScript types auto-generated ensure frontend/backend sync
- **API Versioning**: Can create v2 DTOs while keeping v1 for compatibility
- **Security**: Control exactly what data frontend receives (hide internal fields)
- **Flexibility**: DTOs can combine multiple models or transform data for frontend convenience

## File Organization

- `game_dto.go`: All DTO struct definitions
- `mapper.go`: All model ↔ DTO conversion functions
- Helper functions like `ToPlayerDto()`, `ToCardDto()`, etc.

## TypeScript Generation Command

```bash
# From project root
make generate

# Or directly (from backend/)
tygo generate
```

**Configuration**: See `backend/tygo.yaml` for TypeScript generation config. Only DTOs should be listed in `packages` section.
