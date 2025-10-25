# Domain Models - Internal Backend Use Only

**CRITICAL**: Models in this directory are INTERNAL to the backend. They are NEVER exposed directly to the frontend.

## Core Principles

1. **Internal Use Only**: Models represent domain entities and business logic for backend use
2. **Must Convert to DTOs**: ALL models sent to frontend MUST be converted to DTOs in `internal/delivery/dto/`
3. **Never Export to TypeScript**: NEVER add models to `tygo.yaml` - only DTOs are exported

## Development Workflow

### When Adding/Modifying Models

1. **Add/Update model struct** with `json:` tags (for internal serialization)
2. **Create corresponding DTO** in `internal/delivery/dto/game_dto.go`
3. **Add DTO mapper function** in `internal/delivery/dto/mapper.go`
4. **Add `ts:` tags to DTO** (NOT to model) for TypeScript generation
5. **Run `make generate`** to create TypeScript types

### Example Pattern

```go
// internal/model/player.go - NO ts: tags
type Player struct {
    ID       string    `json:"id"`
    Credits  int       `json:"credits"`
}

// internal/delivery/dto/game_dto.go - WITH ts: tags
type PlayerDto struct {
    ID       string `json:"id" ts:"string"`
    Credits  int    `json:"credits" ts:"number"`
}

// internal/delivery/dto/mapper.go
func ToPlayerDto(player model.Player) PlayerDto {
    return PlayerDto{
        ID:      player.ID,
        Credits: player.Credits,
    }
}
```

## Common Mistakes to Avoid

❌ **WRONG**: Adding `ts:` tags to models
❌ **WRONG**: Exporting models in `tygo.yaml`
❌ **WRONG**: Returning models directly from HTTP/WebSocket handlers
❌ **WRONG**: Forgetting to update DTOs when models change

✅ **CORRECT**: Models have `json:` tags only
✅ **CORRECT**: DTOs have both `json:` and `ts:` tags
✅ **CORRECT**: Handlers always return DTOs, never models
✅ **CORRECT**: Update DTO + mapper when model changes

## Synchronization Checklist

When you modify a model:

- [ ] Update corresponding DTO structure
- [ ] Update mapper function to include new fields
- [ ] Run `make generate` to sync TypeScript types
- [ ] Verify frontend types updated in `frontend/src/types/generated/api-types.ts`

## Why This Separation?

- **Flexibility**: Models can change without breaking frontend contracts
- **Control**: DTOs explicitly define what data frontend receives
- **Type Safety**: TypeScript types generated only from stable DTO contracts
- **Encapsulation**: Internal implementation details stay in backend
