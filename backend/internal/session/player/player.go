package player

import (
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/session/types"
)

// Player wraps a types.Player with its associated sub-repositories
// This allows calling player.Resources.Update() instead of playerRepo.UpdateResources(gameID, playerID, ...)
type Player struct {
	*types.Player // Embedded player data

	// Sub-repositories for operations on this specific player
	Resources   *ResourceRepository
	Hand        *HandRepository
	Selection   *SelectionRepository
	Corporation *CorporationRepository
	Turn        *TurnRepository
	Effects     *EffectRepository
	TileQueue   *TileQueueRepository
}

// NewPlayer creates a new player with wired sub-repositories
func NewPlayer(player *types.Player, eventBus *events.EventBusImpl) *Player {
	p := &Player{
		Player: player,
	}

	// Wire up sub-repositories with reference to this player
	p.Resources = NewResourceRepository(p, eventBus)
	p.Hand = NewHandRepository(p)
	p.Selection = NewSelectionRepository(p)
	p.Corporation = NewCorporationRepository(p)
	p.Turn = NewTurnRepository(p)
	p.Effects = NewEffectRepository(p)
	p.TileQueue = NewTileQueueRepository(p, eventBus)

	return p
}
