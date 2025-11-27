package dto

import (
	"fmt"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/board"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
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
		TurnOrder:        []string{}, // Migration doesn't track turn order yet
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

// ToCardDto converts a Card to CardDto
func ToCardDto(card game.Card) CardDto {
	tags := make([]CardTag, len(card.Tags))
	for i, tag := range card.Tags {
		tags[i] = CardTag(tag)
	}

	return CardDto{
		ID:          card.ID,
		Name:        card.Name,
		Type:        CardType(card.Type),
		Cost:        card.Cost,
		Description: card.Description,
		Pack:        card.Pack,
		Tags:        tags,
	}
}

// getCorporationCard fetches the corporation card for a player using the card registry
func getCorporationCard(p *player.Player, cardRegistry cards.CardRegistry) *CardDto {
	if p.CorporationID() == "" {
		return nil
	}

	card, err := cardRegistry.GetByID(p.CorporationID())
	if err != nil {
		log := logger.Get()
		log.Warn("Failed to fetch corporation card",
			zap.String("player_id", p.ID()),
			zap.String("corporation_id", p.CorporationID()),
			zap.Error(err))
		return nil
	}

	cardDto := ToCardDto(*card)
	return &cardDto
}

// getPlayedCards converts a slice of card IDs to CardDto objects using the card registry
func getPlayedCards(cardIDs []string, cardRegistry cards.CardRegistry) []CardDto {
	cardDtos := make([]CardDto, 0, len(cardIDs))
	log := logger.Get()

	for _, cardID := range cardIDs {
		card, err := cardRegistry.GetByID(cardID)
		if err != nil {
			log.Warn("Failed to fetch played card",
				zap.String("card_id", cardID),
				zap.Error(err))
			continue // Skip cards that can't be found
		}
		cardDtos = append(cardDtos, ToCardDto(*card))
	}

	return cardDtos
}

// ToPlayerDto converts migration Player to PlayerDto
func ToPlayerDto(p *player.Player, g *game.Game, cardRegistry cards.CardRegistry) PlayerDto {
	resourcesComponent := p.Resources()
	resources := resourcesComponent.Get()
	production := resourcesComponent.Production()

	// Get corporation card if player has one
	corporation := getCorporationCard(p, cardRegistry)

	// Get played cards with full card details
	playedCardIDs := p.PlayedCards().Cards()
	playedCards := getPlayedCards(playedCardIDs, cardRegistry)

	// Get hand cards with full card details
	handCardIDs := p.Hand().Cards()
	handCards := getPlayedCards(handCardIDs, cardRegistry)

	return PlayerDto{
		ID:   p.ID(),
		Name: p.Name(),
		Resources: ResourcesDto{
			Credits:  resources.Credits,
			Steel:    resources.Steel,
			Titanium: resources.Titanium,
			Plants:   resources.Plants,
			Energy:   resources.Energy,
			Heat:     resources.Heat,
		},
		Production: ProductionDto{
			Credits:  production.Credits,
			Steel:    production.Steel,
			Titanium: production.Titanium,
			Plants:   production.Plants,
			Energy:   production.Energy,
			Heat:     production.Heat,
		},
		TerraformRating:  resourcesComponent.TerraformRating(),
		VictoryPoints:    resourcesComponent.VictoryPoints(),
		Status:           PlayerStatusWaiting, // Default status
		Corporation:      corporation,
		Cards:            handCards,
		PlayedCards:      playedCards,
		Passed:           false,
		AvailableActions: 0,
		IsConnected:      true, // Assume connected
		Effects:          []PlayerEffectDto{},
		Actions:          []PlayerActionDto{},

		SelectStartingCardsPhase: convertSelectStartingCardsPhase(g.GetSelectStartingCardsPhase(p.ID()), cardRegistry),
		ProductionPhase:          nil, // TODO: Implement production phase mapping
		StartingCards:            []CardDto{},
		PendingTileSelection:     nil,
		PendingCardSelection:     nil,
		PendingCardDrawSelection: nil,
		ForcedFirstAction:        nil,
		ResourceStorage:          map[string]int{},
		PaymentSubstitutes:       []PaymentSubstituteDto{},
		RequirementModifiers:     []RequirementModifierDto{},
	}
}

// ToOtherPlayerDto converts migration Player to OtherPlayerDto
func ToOtherPlayerDto(p *player.Player, g *game.Game, cardRegistry cards.CardRegistry) OtherPlayerDto {
	resourcesComponent := p.Resources()
	resources := resourcesComponent.Get()
	production := resourcesComponent.Production()

	// Get corporation card if player has one
	corporation := getCorporationCard(p, cardRegistry)

	// Get played cards with full card details
	playedCardIDs := p.PlayedCards().Cards()
	playedCards := getPlayedCards(playedCardIDs, cardRegistry)

	// Get hand card count
	handCardCount := len(p.Hand().Cards())

	return OtherPlayerDto{
		ID:   p.ID(),
		Name: p.Name(),
		Resources: ResourcesDto{
			Credits:  resources.Credits,
			Steel:    resources.Steel,
			Titanium: resources.Titanium,
			Plants:   resources.Plants,
			Energy:   resources.Energy,
			Heat:     resources.Heat,
		},
		Production: ProductionDto{
			Credits:  production.Credits,
			Steel:    production.Steel,
			Titanium: production.Titanium,
			Plants:   production.Plants,
			Energy:   production.Energy,
			Heat:     production.Heat,
		},
		TerraformRating:  resourcesComponent.TerraformRating(),
		VictoryPoints:    resourcesComponent.VictoryPoints(),
		Status:           PlayerStatusWaiting,
		Corporation:      corporation,
		HandCardCount:    handCardCount,
		PlayedCards:      playedCards,
		Passed:           false,
		AvailableActions: 0,
		IsConnected:      true,
		Effects:          []PlayerEffectDto{},
		Actions:          []PlayerActionDto{},

		SelectStartingCardsPhase: convertSelectStartingCardsPhaseForOtherPlayer(g.GetSelectStartingCardsPhase(p.ID())),
		ProductionPhase:          nil, // TODO: Implement production phase mapping
		ResourceStorage:          map[string]int{},
		PaymentSubstitutes:       []PaymentSubstituteDto{},
	}
}

// convertSelectStartingCardsPhase converts SelectStartingCardsPhase to DTO
func convertSelectStartingCardsPhase(phase *player.SelectStartingCardsPhase, cardRegistry cards.CardRegistry) *SelectStartingCardsPhaseDto {
	if phase == nil {
		fmt.Println("⚠️  convertSelectStartingCardsPhase: phase is nil")
		return nil
	}

	fmt.Printf("🔧 convertSelectStartingCardsPhase: cards=%d, corps=%v\n",
		len(phase.AvailableCards), phase.AvailableCorporations)

	// Get full card details for available cards
	availableCards := getPlayedCards(phase.AvailableCards, cardRegistry)

	// Get full card details for available corporations
	availableCorporations := getPlayedCards(phase.AvailableCorporations, cardRegistry)

	result := &SelectStartingCardsPhaseDto{
		AvailableCards:        availableCards,
		AvailableCorporations: availableCorporations,
	}

	fmt.Printf("🔧 convertSelectStartingCardsPhase result: cards=%d, corps=%d\n",
		len(result.AvailableCards), len(result.AvailableCorporations))

	return result
}

// convertSelectStartingCardsPhaseForOtherPlayer converts SelectStartingCardsPhase to DTO for other players (empty)
func convertSelectStartingCardsPhaseForOtherPlayer(phase *player.SelectStartingCardsPhase) *SelectStartingCardsOtherPlayerDto {
	if phase == nil {
		return nil
	}

	// Other players don't see selection details
	return &SelectStartingCardsOtherPlayerDto{}
}
