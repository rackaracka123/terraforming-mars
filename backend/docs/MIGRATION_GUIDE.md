# Action Migration Guide

This guide explains how to migrate actions from the old architecture (`internal/action/`) to the new architecture (`internal/action_migration/`).

## Migration Status

### ✅ Phase 1: Type System and Core Entities (COMPLETED)

**Objective**: Create standalone `game_migration` package with complete type definitions and encapsulated entities.

**Completed Work**:
1. **Type Definitions** (`internal/game_migration/types.go`):
   - All basic game types (GamePhase, GameStatus, GameSettings)
   - Resource types (Resources, Production, ResourceType enum with 50+ constants)
   - All domain enums (CardTag, StandardProject, etc.)
   - Zero dependencies on old `internal/session/types/`

2. **Card Types** (`internal/game_migration/card_types.go`):
   - Complete card system types (Card, CardBehavior, Trigger, etc.)
   - Player effect and action types
   - Card payment and validation types
   - Zero dependencies on old `internal/session/game/card/`

3. **Player Component Architecture** (`internal/game_migration/player.go`):
   - Component accessor pattern matching existing migrated actions
   - 7 encapsulated components: Corporation, Hand, PlayerResources, Turn, Effects, Actions, Selection
   - Each component has own mutex for thread safety
   - All methods publish events automatically
   - Zero dependencies on old `internal/session/game/player/`

4. **Updated Existing Files**:
   - `game.go`: Updated to use local types
   - `board.go`: Updated to use local types
   - `repository.go`: Updated signatures
   - `global_parameters.go`: Already encapsulated

5. **Action Migration Package** (`internal/action_migration/`):
   - Fixed 26 action files to use new types
   - Removed all imports from old `internal/session/*` packages
   - All actions now compile successfully
   - BaseAction pattern established
   - Validation helpers created

**Verification**: ✅ All packages compile successfully
```bash
go build ./internal/game_migration/...
go build ./internal/action_migration/...
```

### ✅ Phase 2: Simple Actions Migration (COMPLETED)

**Objective**: Migrate all simple actions that don't depend on complex infrastructure.

**Migrated Actions (20 total)**:
- ✅ action_build_aquifer.go
- ✅ action_build_city.go
- ✅ action_build_power_plant.go
- ✅ action_confirm_card_draw.go
- ✅ action_confirm_production_cards.go
- ✅ action_confirm_sell_patents.go
- ✅ action_convert_heat.go
- ✅ action_convert_plants_to_greenery.go
- ✅ action_create_game.go **(newly migrated)**
- ✅ action_join_game.go
- ✅ action_launch_asteroid.go
- ✅ action_plant_greenery.go
- ✅ action_player_disconnected.go
- ✅ action_player_reconnected.go
- ✅ action_select_starting_cards.go
- ✅ action_sell_patents.go
- ✅ action_skip_action.go
- ✅ action_start_game.go
- ✅ admin/action_set_global_parameters.go
- ✅ admin/action_set_phase.go
- ✅ admin/action_set_production.go
- ✅ admin/action_set_resources.go

**Actions Blocked by Infrastructure Dependencies**:

These actions require infrastructure components that haven't been migrated yet:

1. **select_tile.go** - Blocked by:
   - `board.Processor` (tile processing logic)
   - `board.BonusCalculator` (bonus calculation)
   - Complex tile placement workflows

2. **play_card.go** - Blocked by:
   - `game.CardManager` (card validation and playing)
   - Card effect system integration
   - Complex card playing workflows

**Non-Actions (utilities)**:
- `workflows.go` - Helper functions (deprecated, not critical)
- `validation.go` - Validation helpers (already in action_migration/)
- `base.go` - Base action (already in action_migration/)

**Next Steps**: Migrate infrastructure components (TileProcessor, BonusCalculator, CardManager) before completing the remaining actions.

### ✅ Phase 3: Handler Integration (COMPLETE)

**Objective**: Wire migrated actions to WebSocket and HTTP handlers.

**Current Status**:
- ✅ All 21 migrated actions instantiated in `cmd/server/main.go`
- ✅ MigrationBroadcaster initialized and subscribed to BroadcastEvent
- ✅ DTO mapper for game_migration types complete
- ✅ 18 handlers created and registered with WebSocket hub
- ✅ All handlers compile and integrate with migrated actions
- ✅ Server compiles successfully with both old and new handlers active

**What's Initialized** (`cmd/server/main.go` lines 163-235):
```go
// Game lifecycle (2)
createGameActionMigrated := action_migration.NewCreateGameAction(...)
joinGameActionMigrated := action_migration.NewJoinGameAction(...)

// Resource conversions (2)
convertHeatActionMigrated := ...
convertPlantsActionMigrated := ...

// Standard projects (6)
buildPowerPlantActionMigrated, buildCityActionMigrated, ...

// + 11 more actions (confirmations, turn management, connection, admin)
```

**Architectural Status**:
1. **Event System**: ✅ Working
   - Migrated actions publish BroadcastEvent
   - MigrationBroadcaster subscribes and handles events

2. **Broadcasting**: ✅ Working
   - Event subscription working
   - DTO mapping implemented in `mapper_migration.go`
   - MigrationBroadcaster sends personalized game states
   - Full game state conversion: Game → Players → Board → GlobalParameters

3. **Handler Layer**: ✅ Complete
   - 18 handlers created in `handler_migration/` directory
   - All handlers registered with WebSocket hub
   - Old handlers remain active on same message types (parallel operation)

**What Was Completed**:
- ✅ Created `internal/delivery/dto/mapper_migration.go` (~250 lines)
  - `ToGameDtoFromMigration()` - Main game converter
  - `ToPlayerDtoFromMigration()` - Full player data for viewing player
  - `PlayerToOtherPlayerDtoFromMigration()` - Limited data for other players
  - `ToBoardDtoFromMigration()` - Board with tiles and occupancy
  - Helper converters for all sub-components

- ✅ Updated `broadcaster_migration.go`:
  - Replaced placeholder with real DTO mapping
  - Calls `ToGameDtoFromMigration(game, playerID)`
  - Sends personalized game state to each player

- ✅ All packages compile successfully:
  ```bash
  go build ./internal/delivery/dto/...
  go build ./internal/delivery/websocket/...
  go build ./cmd/server/...
  ```

**Handler Implementation**:
- ✅ Created `handler_migration/` directory structure:
  - `game/` - Game lifecycle (CreateGame, JoinGame)
  - `standard_project/` - Standard projects (6 handlers)
  - `resource_conversion/` - Resource conversions (2 handlers)
  - `turn_management/` - Turn management (3 handlers)
  - `confirmation/` - Confirmations (3 handlers)
  - `connection/` - Connection management (2 handlers)

- ✅ **18 Handlers Created**:
  1. CreateGameHandler - Game creation with settings
  2. JoinGameHandler - Player join with full game state
  3. LaunchAsteroidHandler - Raise temperature standard project
  4. BuildPowerPlantHandler - Energy production standard project
  5. BuildAquiferHandler - Place ocean standard project
  6. BuildCityHandler - Place city standard project
  7. PlantGreeneryHandler - Place greenery standard project
  8. SellPatentsHandler - Initiate patent selling
  9. ConvertHeatHandler - Convert heat to temperature
  10. ConvertPlantsHandler - Convert plants to greenery
  11. StartGameHandler - Start game from lobby
  12. SkipActionHandler - Pass turn/action
  13. SelectStartingCardsHandler - Choose corporation and cards
  14. ConfirmSellPatentsHandler - Confirm card selection for selling
  15. ConfirmProductionCardsHandler - Confirm production phase cards
  16. ConfirmCardDrawHandler - Confirm card draw selection
  17. PlayerReconnectedHandler - Handle player reconnection
  18. PlayerDisconnectedHandler - Handle player disconnection

- ✅ **Registry Integration** (`registry_migration.go`):
  - Expanded `RegisterMigrationHandlers()` to accept all 18 actions
  - All handlers registered on existing message types
  - Comprehensive logging of registered handlers

- ✅ **Main Server Integration** (`cmd/server/main.go:211-238`):
  - RegisterMigrationHandlers called with all 18 migrated actions
  - Unused variable warnings removed for registered actions
  - Admin actions remain unregistered (HTTP-only)

**Complete Architecture Flow** (End-to-End):
```
Client sends WebSocket message
    ↓
handler_migration/JoinGameHandler.HandleMessage()
    ↓
action_migration/JoinGameAction.Execute(ctx, gameID, playerName)
    ↓
game_migration/Game.AddPlayer() [publishes BroadcastEvent]
    ↓
MigrationBroadcaster.OnBroadcastEvent() [subscribes to events]
    ↓
mapper_migration/ToGameDtoFromMigration(game, playerID)
    ↓
WebSocket sends personalized GameDto to all players
    ↓
Clients receive updated game state
```

**Status**: Phase 3 is **100% complete** - all handlers registered and ready for testing.

**Message Type Mapping**:
- `create-game` → CreateGameHandler
- `player-connect-v2` → JoinGameHandler (temporary, avoids conflict)
- `action.standard-project.*` → Standard project handlers (6)
- `action.resource-conversion.*` → Resource conversion handlers (2)
- `action.game-management.*` → Turn management handlers (2)
- `action.card.*` → Card-related handlers (4)
- `player-reconnected/disconnected` → Connection handlers (2)

**Testing & Migration Strategy**:
1. ✅ All handlers registered and active
2. Old handlers remain active on same message types (parallel operation)
3. Frontend can test migration handlers by sending appropriate message types
4. Once migration handlers proven stable, can replace old handlers
5. Admin actions remain HTTP-only for now (not critical for WebSocket)

### 🔄 Phase 4: Cleanup (PENDING)

**Objective**: Delete old packages and rename migration packages.
- Delete `internal/session/` package
- Rename `internal/game_migration/` → `internal/game/`
- Rename `internal/action_migration/` → `internal/action/`

---

## Architecture Overview

### Old Architecture
```
internal/action/
  ├── Old actions with 4-5 dependencies
  ├── BaseAction with SessionManagerFactory
  └── Manual broadcast calls

internal/session/
  ├── Session wrapper
  ├── Multiple repositories (game, board, card, deck)
  └── Public fields on domain types
```

### New Architecture
```
internal/action_migration/
  ├── New actions with 2 dependencies only
  ├── Direct GameRepository access
  └── Event-driven broadcasting (automatic)

internal/game_migration/
  ├── Encapsulated Game (private fields, public methods)
  ├── Single GameRepository
  └── Event publishing on all mutations
```

## Migration Pattern

### 1. Action Structure

**Old:**
```go
type OldAction struct {
    BaseAction
    gameRepo game.Repository
    sessionFactory session.SessionFactory
    // ... 2-3 more dependencies
}

func NewOldAction(
    gameRepo game.Repository,
    sessionFactory session.SessionFactory,
    sessionMgrFactory session.SessionManagerFactory,
    // ... more params
) *OldAction {
    return &OldAction{
        BaseAction: NewBaseAction(sessionMgrFactory),
        gameRepo: gameRepo,
        sessionFactory: sessionFactory,
    }
}

func (a *OldAction) Execute(ctx context.Context, sess *session.Session, playerID string) error {
    gameID := sess.GetGameID()
    log := a.InitLogger(gameID, playerID)
    // ... business logic ...
    a.BroadcastGameState(gameID, log)  // Manual broadcast!
    return nil
}
```

**New:**
```go
type NewAction struct {
    gameRepo game_migration.GameRepository
    logger   *zap.Logger
}

func NewNewAction(
    gameRepo game_migration.GameRepository,
    logger *zap.Logger,
) *NewAction {
    return &NewAction{
        gameRepo: gameRepo,
        logger:   logger,
    }
}

func (a *NewAction) Execute(
    ctx context.Context,
    gameID string,
    playerID string,
) error {
    log := a.logger.With(
        zap.String("game_id", gameID),
        zap.String("player_id", playerID),
    )
    // ... business logic ...
    // NO MANUAL BROADCAST - Events handle it automatically!
    return nil
}
```

### 2. Game Access

**Old:**
```go
// Through Session + Repository
sess := a.sessionFactory.Get(gameID)
player, exists := sess.GetPlayer(playerID)
g, err := a.gameRepo.GetByID(ctx, gameID)
```

**New:**
```go
// Direct via GameRepository
g, err := a.gameRepo.Get(ctx, gameID)
player, err := g.GetPlayer(playerID)
```

### 3. State Mutations

**Old:**
```go
// Direct field mutation + repository update
g.GlobalParameters.Temperature += 2
err = a.gameRepo.UpdateTemperature(ctx, gameID, newTemp)
a.BroadcastGameState(gameID, log)  // Manual!
```

**New:**
```go
// Encapsulated method (publishes event automatically)
stepsRaised, err := g.GlobalParameters().IncreaseTemperature(ctx, 1)
// ↑ Publishes TemperatureChangedEvent → SessionManager → Broadcast (automatic!)
```

### 4. Player State Changes

**Old:**
```go
resources := player.Resources().Get()
resources.Credits -= cost
player.Resources().Set(resources)
```

**New:**
```go
// Same! Player component is already encapsulated
player.Resources().Add(map[types.ResourceType]int{
    types.ResourceCredits: -cost,
})
// ↑ Publishes ResourcesChangedEvent automatically
```

## Complete Migration Examples

### Example 1: Simple Resource Action (ConvertHeat)

**Old Version:** `internal/action/convert_heat_to_temperature.go`
```go
type ConvertHeatToTemperatureAction struct {
    BaseAction
    gameRepo game.Repository
}

func (a *ConvertHeatToTemperatureAction) Execute(ctx context.Context, sess *session.Session, playerID string) error {
    gameID := sess.GetGameID()
    log := a.InitLogger(gameID, playerID)

    g, err := ValidateActiveGame(ctx, a.gameRepo, gameID, log)
    player, exists := sess.GetPlayer(playerID)

    // Business logic...
    player.Resources().Add(...)
    err = a.gameRepo.UpdateTemperature(ctx, gameID, newTemp)
    player.Resources().SetTerraformRating(newTR)
    player.Turn().SetAvailableActions(available - 1)

    a.BroadcastGameState(gameID, log)  // Manual!
    return nil
}
```

**New Version:** `internal/action_migration/action_convert_heat.go`
```go
type ConvertHeatToTemperatureAction struct {
    gameRepo game_migration.GameRepository
    logger   *zap.Logger
}

func (a *ConvertHeatToTemperatureAction) Execute(ctx context.Context, gameID string, playerID string) error {
    log := a.logger.With(zap.String("game_id", gameID), zap.String("player_id", playerID))

    // 1. Fetch game
    g, err := a.gameRepo.Get(ctx, gameID)

    // 2. Validate
    if g.Status() != types.GameStatusActive { return fmt.Errorf("game not active") }
    if g.CurrentTurn() == nil || *g.CurrentTurn() != playerID { return fmt.Errorf("not your turn") }

    // 3. Get player
    player, err := g.GetPlayer(playerID)

    // 4. Business logic
    player.Resources().Add(map[types.ResourceType]int{types.ResourceHeat: -requiredHeat})
    stepsRaised, _ := g.GlobalParameters().IncreaseTemperature(ctx, 1)
    if stepsRaised > 0 {
        player.Resources().SetTerraformRating(oldTR + 1)
    }
    player.Turn().ConsumeAction()

    // NO MANUAL BROADCAST - Events handle it!
    return nil
}
```

### Example 2: Player Creation (JoinGame)

**Old Version:** `internal/action/join_game.go`
```go
func (a *JoinGameAction) Execute(ctx context.Context, gameID string, playerName string, playerID ...string) (*JoinGameResult, error) {
    sess := a.sessionFactory.GetOrCreate(gameID)

    // Create player via session
    newPlayer := sess.CreateAndAddPlayer(playerName, pid)

    // Add to game via repository
    err = a.gameRepo.AddPlayer(ctx, gameID, newPlayer.ID())

    if isFirstPlayer {
        err = a.gameRepo.SetHostPlayer(ctx, gameID, newPlayer.ID())
    }

    return &JoinGameResult{...}, nil
}
```

**New Version:** `internal/action_migration/action_join_game.go`
```go
func (a *JoinGameAction) Execute(ctx context.Context, gameID string, playerName string, playerID ...string) (*JoinGameResult, error) {
    // 1. Get game
    g, err := a.gameRepo.Get(ctx, gameID)

    // 2. Validate lobby status
    if g.Status() != types.GameStatusLobby { return nil, fmt.Errorf("not in lobby") }

    // 3. Create player directly
    newPlayer := player.NewPlayer(a.eventBus, gameID, pid, playerName)

    // 4. Add to game (publishes PlayerJoinedEvent)
    err = g.AddPlayer(ctx, newPlayer)

    // 5. Set host if first
    if len(g.GetAllPlayers()) == 1 {
        err = g.SetHostPlayerID(ctx, newPlayer.ID())
    }

    return &JoinGameResult{...}, nil
}
```

### Example 3: Standard Project (BuildPowerPlant)

**Old Version:** `internal/action/build_power_plant.go`
```go
func (a *BuildPowerPlantAction) Execute(ctx context.Context, sess *session.Session, playerID string) error {
    gameID := sess.GetGameID()
    log := a.InitLogger(gameID, playerID)

    g, err := ValidateActiveGame(ctx, a.gameRepo, gameID, log)
    player, exists := sess.GetPlayer(playerID)

    // Deduct cost
    player.Resources().Add(map[types.ResourceType]int{
        types.ResourceCredits: -BuildPowerPlantCost,
    })

    // Increase production
    player.Resources().AddProduction(map[types.ResourceType]int{
        types.ResourceEnergy: 1,
    })

    player.Turn().ConsumeAction()
    a.BroadcastGameState(gameID, log)
    return nil
}
```

**New Version:** `internal/action_migration/action_build_power_plant.go`
```go
func (a *BuildPowerPlantAction) Execute(ctx context.Context, gameID string, playerID string) error {
    log := a.logger.With(zap.String("game_id", gameID), zap.String("player_id", playerID))

    // 1. Get game and validate
    g, err := a.gameRepo.Get(ctx, gameID)
    if g.Status() != types.GameStatusActive { return fmt.Errorf("not active") }
    if g.CurrentTurn() == nil || *g.CurrentTurn() != playerID { return fmt.Errorf("not your turn") }

    // 2. Get player
    player, err := g.GetPlayer(playerID)

    // 3. Business logic (same as old - player already encapsulated!)
    player.Resources().Add(map[types.ResourceType]int{
        types.ResourceCredits: -BuildPowerPlantCost,
    })

    player.Resources().AddProduction(map[types.ResourceType]int{
        types.ResourceEnergy: 1,
    })

    player.Turn().ConsumeAction()

    // NO MANUAL BROADCAST!
    return nil
}
```

## Migration Checklist

For each action:

- [ ] Create new file in `internal/action_migration/action_*.go`
- [ ] Update struct to have only `gameRepo` + `logger`
- [ ] Change constructor signature to accept `GameRepository` + `logger`
- [ ] Update `Execute` signature: `(ctx, gameID, playerID string)` instead of `(ctx, sess, playerID)`
- [ ] Replace `sess.GetGameID()` with direct `gameID` parameter
- [ ] Replace `a.InitLogger()` with inline `a.logger.With(...)`
- [ ] Replace `sess.GetPlayer()` with `g.GetPlayer()`
- [ ] Replace `a.gameRepo.GetByID()` with `a.gameRepo.Get()`
- [ ] Replace `a.gameRepo.UpdateX()` with `g.X().UpdateY()`
- [ ] Remove all `a.BroadcastGameState()` calls
- [ ] Update validation to check `g.Status()` and `g.CurrentTurn()`
- [ ] Add to `main.go` initialization (mark as unused with `_` for now)
- [ ] Verify compilation: `go build ./internal/action_migration/...`

## Event Publishing Reference

These domain methods automatically publish domain events AND BroadcastEvent:

| Method | Domain Event Published | BroadcastEvent | WebSocket Broadcast |
|--------|----------------------|----------------|-------------------|
| `g.AddPlayer()` | `PlayerJoinedEvent` | ✅ (all) | ✅ |
| `g.UpdateStatus()` | `GameStatusChangedEvent` | ✅ (all) | ✅ |
| `g.UpdatePhase()` | `GamePhaseChangedEvent` | ✅ (all) | ✅ |
| `g.AdvanceGeneration()` | `GenerationAdvancedEvent` | ✅ (all) | ✅ |
| `g.SetCurrentTurn()` | None | ✅ (all) | ✅ |
| `g.SetPendingTileSelection()` | None | ✅ (player) | ✅ |
| `g.SetPendingTileSelectionQueue()` | None | ✅ (player) | ✅ |
| `g.SetForcedFirstAction()` | None | ✅ (player) | ✅ |
| `g.SetProductionPhase()` | None | ✅ (player) | ✅ |
| `g.ProcessNextTile()` | None | ✅ (player) | ✅ |
| `g.GlobalParameters().IncreaseTemperature()` | `TemperatureChangedEvent` | ✅ (all) | ✅ |
| `g.GlobalParameters().IncreaseOxygen()` | `OxygenChangedEvent` | ✅ (all) | ✅ |
| `g.GlobalParameters().PlaceOcean()` | `OceansChangedEvent` | ✅ (all) | ✅ |
| `g.GlobalParameters().SetTemperature()` | `TemperatureChangedEvent` | ✅ (all) | ✅ |
| `g.GlobalParameters().SetOxygen()` | `OxygenChangedEvent` | ✅ (all) | ✅ |
| `g.GlobalParameters().SetOceans()` | `OceansChangedEvent` | ✅ (all) | ✅ |
| `g.Board().UpdateTileOccupancy()` | `TilePlacedEvent` | ✅ (all) | ✅ |
| `player.Resources().Set()` | `ResourcesChangedEvent` | (via old arch) | ✅ |
| `player.Resources().Add()` | `ResourcesChangedEvent` | (via old arch) | ✅ |
| `player.Resources().SetTerraformRating()` | `TerraformRatingChangedEvent` | (via old arch) | ✅ |
| `player.Resources().AddProduction()` | `ProductionChangedEvent` | (via old arch) | ✅ |

**Legend:**
- **Domain Event**: Traditional domain event (e.g., TemperatureChangedEvent, ResourcesChangedEvent)
- **BroadcastEvent**: Meta-event that triggers WebSocket broadcasts
  - **(all)**: Broadcasts to all players in the game (`PlayerIDs: nil`)
  - **(player)**: Broadcasts only to the specific player (`PlayerIDs: [playerID]`)
- **WebSocket Broadcast**: MigrationBroadcaster subscribes to BroadcastEvent and sends personalized game state to clients

**Phase 2 Complete (Event-Driven Broadcasting):**
- ✅ BroadcastEvent defined in `internal/events/domain_events.go`
- ✅ All game_migration components publish BroadcastEvent after state mutations
- ✅ MigrationBroadcaster subscribes to BroadcastEvent and handles WebSocket updates
- ✅ No manual broadcast calls needed in actions!

## Common Patterns

### Pattern: Validation

**Old:**
```go
g, err := ValidateActiveGame(ctx, a.gameRepo, gameID, log)
if err := ValidateCurrentTurn(g, playerID, log); err != nil {
    return err
}
```

**New:**
```go
g, err := a.gameRepo.Get(ctx, gameID)
if g.Status() != types.GameStatusActive {
    return fmt.Errorf("game is not active")
}
currentTurn := g.CurrentTurn()
if currentTurn == nil || *currentTurn != playerID {
    return fmt.Errorf("not your turn")
}
```

### Pattern: Resource Cost

**Both old and new (Player component already encapsulated):**
```go
// Validate
resources := player.Resources().Get()
if resources.Credits < cost {
    return fmt.Errorf("insufficient credits")
}

// Deduct
player.Resources().Add(map[types.ResourceType]int{
    types.ResourceCredits: -cost,
})
```

### Pattern: Global Parameter Changes

**Old:**
```go
newTemp := g.GlobalParameters.Temperature + 2
err = a.gameRepo.UpdateTemperature(ctx, gameID, newTemp)
a.BroadcastGameState(gameID, log)
```

**New:**
```go
stepsRaised, err := g.GlobalParameters().IncreaseTemperature(ctx, 1)
// Automatic broadcast via TemperatureChangedEvent
```

## Benefits of New Architecture

1. **Simpler Dependencies**: 2 instead of 4-5
2. **Type Safety**: Encapsulation prevents invalid state
3. **Event-Driven**: No manual broadcast calls to forget
4. **Testability**: Mock 1 repository instead of many
5. **Thread Safety**: Proper lock patterns throughout
6. **Single Source of Truth**: One GameRepository, one Game entity

## Next Steps

After migrating actions:
1. Update tests to use `game_migration.GameRepository`
2. Create bridge layer for WebSocket/HTTP handlers
3. Delete `internal/session/` package
4. Rename `game_migration` → `game` and `action_migration` → `action`
