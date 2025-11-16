# Bugs Found During Testing

This document tracks bugs discovered by the test suite after backend restructuring.

## Critical Bugs

### Session Management: Reconnection Failure

**Status**: ✅ Fixed
**Severity**: Critical (was)
**Found By**: `TestGameFlow_ReconnectPlayer`
**Date Found**: 2025-11-16
**Date Fixed**: 2025-11-16

**Description**:
When a player disconnects and attempts to reconnect to a game using their player ID, the session manager fails to find the session in the session repository, causing broadcast failures.

**Error Message**:
```
Failed to get session for broadcast: session not found for game c0b52557-bc0d-4c1b-83b1-c8097db5e631
```

**Location**:
- `internal/delivery/websocket/session/manager.go:194`
- Called from `internal/lobby/service.go:546`

**Root Cause**:
The session is not being properly persisted or retrieved from the session repository during player reconnection flow. When `SessionManager.Broadcast()` is called after a reconnection, it cannot find the session.

**Impact**:
- Players cannot reconnect to games after disconnection
- Complete loss of game state for disconnected players
- Multiplayer games become unrecoverable if any player disconnects

**Reproduction**:
```bash
go test -v ./test/integration/ -run TestGameFlow_ReconnectPlayer
```

**Expected Behavior**:
1. Player connects to game
2. Player disconnects
3. Player reconnects with same player ID
4. Session manager finds existing session
5. Game state is broadcast to reconnected player

**Actual Behavior**:
1. Player connects to game ✅
2. Player disconnects ✅
3. Player reconnects with same player ID ✅
4. Session manager fails to find session ❌
5. Broadcast fails, player receives no state ❌

**Fix Applied**:
1. **Root Cause**: SessionManager was checking for session existence, but sessions don't exist for lobby games (only created when game starts).
2. **Solution**: Modified `SessionManager.broadcastGameStateInternal` to check `game.Status` instead of session existence.
3. **Changes Made**:
   - Check `game.Status == GameStatusLobby` to determine if using feature services
   - Pass `nil` for all feature services (ParametersService, BoardService, TurnOrderService) when in lobby
   - Modified `ToGameDto` to handle `nil` services gracefully
   - Updated test setup to use `nil` for BoardService instead of `NewBoardService(nil)` to avoid interface type issues

**Files Modified**:
- `internal/delivery/websocket/session/manager.go` (lines 191-240)
- `internal/delivery/dto/mapper.go` (lines 87-117)
- `test/integration/websocket_game_flow_test.go` (line 494, 202-234)

**Tests Passing**:
- ✅ `TestGameFlow_CreateLobbyAndJoin`
- ✅ `TestGameFlow_StartGame`
- ✅ `TestGameFlow_ReconnectPlayer`

---

## Test Coverage

### Integration Tests Created

**WebSocket Game Flow Tests** (`test/integration/websocket_game_flow_test.go`):
- ✅ TestGameFlow_CreateLobbyAndJoin (PASSING)
- ✅ TestGameFlow_StartGame (PASSING)
- ⏭️ TestGameFlow_SelectStartingCards (SKIPPED - card service interface issue)
- ✅ TestGameFlow_ReconnectPlayer (PASSING - bug fixed)

**Overall Test Stats**:
- Total: 95 tests
- Passing: 94
- Skipped: 1 (TestGameFlow_SelectStartingCards - requires card.GameRepository interface fix)
- Coverage: Repository layer, domain layer, actions, game rules, WebSocket flow

---

## Notes

This bug was found during the creation of comprehensive integration tests after the backend restructuring from layer-based to domain-based architecture. The tests are working as intended - exposing real architectural issues that need to be addressed.
