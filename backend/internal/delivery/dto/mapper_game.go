package dto

import (
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/board"
	"terraforming-mars-backend/internal/game/player"
)

// ToGameDto converts migration Game to GameDto with personalized view
// The playerID parameter determines which player is "currentPlayer" vs "otherPlayers"
func ToGameDto(g *game.Game, cardRegistry cards.CardRegistry, playerID string) GameDto {
	players := g.GetAllPlayers()

	// Create personalized view: viewing player is currentPlayer, others are otherPlayers
	var currentPlayer PlayerDto
	otherPlayers := make([]OtherPlayerDto, 0)

	var viewingPlayer *player.Player
	for _, p := range players {
		if p.ID() == playerID {
			viewingPlayer = p
			currentPlayer = ToPlayerDto(p, g, cardRegistry)
		} else {
			otherPlayers = append(otherPlayers, ToOtherPlayerDto(p, g, cardRegistry))
		}
	}

	// If viewing player not found, use first player as fallback
	if viewingPlayer == nil && len(players) > 0 {
		currentPlayer = ToPlayerDto(players[0], g, cardRegistry)
		for i := 1; i < len(players); i++ {
			otherPlayers = append(otherPlayers, ToOtherPlayerDto(players[i], g, cardRegistry))
		}
		playerID = players[0].ID()
	}

	// Get settings
	settings := g.Settings()
	settingsDto := GameSettingsDto{
		MaxPlayers: settings.MaxPlayers,
		CardPacks:  settings.CardPacks,
	}

	// Convert global parameters
	globalParams := g.GlobalParameters()
	globalParamsDto := GlobalParametersDto{
		Temperature: globalParams.Temperature(),
		Oxygen:      globalParams.Oxygen(),
		Oceans:      globalParams.Oceans(),
	}

	// Get board tiles
	board := g.Board()
	tiles := board.Tiles()
	tileDtos := make([]TileDto, len(tiles))
	for i, tile := range tiles {
		tileDtos[i] = TileDto{
			Coordinates: HexPositionDto{
				Q: tile.Coordinates.Q,
				R: tile.Coordinates.R,
				S: tile.Coordinates.S,
			},
			Type:        string(tile.Type),
			OwnerID:     tile.OwnerID,
			Tags:        tile.Tags,
			Bonuses:     convertTileBonuses(tile.Bonuses),
			Location:    string(tile.Location),
			DisplayName: tile.DisplayName,
		}
		if tile.OccupiedBy != nil {
			occupant := &TileOccupantDto{
				Type: string(tile.OccupiedBy.Type),
				Tags: tile.OccupiedBy.Tags,
			}
			tileDtos[i].OccupiedBy = occupant
		}
	}

	// Payment constants (default values)
	paymentConstants := PaymentConstantsDto{
		SteelValue:    2, // Default steel value
		TitaniumValue: 3, // Default titanium value
	}

	return GameDto{
		ID:               g.ID(),
		Status:           GameStatus(g.Status()),
		Settings:         settingsDto,
		HostPlayerID:     g.HostPlayerID(),
		CurrentPhase:     GamePhase(g.CurrentPhase()),
		GlobalParameters: globalParamsDto,
		CurrentPlayer:    currentPlayer,
		OtherPlayers:     otherPlayers,
		ViewingPlayerID:  playerID, // The player viewing this game state
		CurrentTurn:      getCurrentTurnPlayerID(g),
		Generation:       g.Generation(),
		TurnOrder:        g.TurnOrder(),
		Board: BoardDto{
			Tiles: tileDtos,
		},
		PaymentConstants: paymentConstants,
	}
}

// getCurrentTurnPlayerID extracts the player ID from the current turn
func getCurrentTurnPlayerID(g *game.Game) *string {
	turn := g.CurrentTurn()
	if turn == nil {
		return nil
	}
	playerID := turn.PlayerID()
	return &playerID
}

// convertTileBonuses converts migration TileBonus to DTO
func convertTileBonuses(bonuses []board.TileBonus) []TileBonusDto {
	dtos := make([]TileBonusDto, len(bonuses))
	for i, bonus := range bonuses {
		dtos[i] = TileBonusDto{
			Type:   string(bonus.Type),
			Amount: bonus.Amount,
		}
	}
	return dtos
}
