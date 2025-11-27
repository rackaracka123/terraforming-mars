package dto

import (
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/board"
	"terraforming-mars-backend/internal/game/deck"
	"terraforming-mars-backend/internal/game/player"
)

// ToGameDto converts migration Game to GameDto
// Note: This returns basic game info without personalization (no viewing player)
// For personalized view, frontend should call with playerID parameter
func ToGameDto(g *game.Game) GameDto {
	players := g.GetAllPlayers()

	// For non-personalized view, just show basic data
	// Frontend will need to determine which player is "current" based on playerID
	var currentPlayer PlayerDto
	otherPlayers := make([]OtherPlayerDto, 0)

	// Just return first player as current for now (non-personalized)
	deck := g.Deck()
	if len(players) > 0 {
		currentPlayer = ToPlayerDto(players[0], deck)
		for i := 1; i < len(players); i++ {
			otherPlayers = append(otherPlayers, ToOtherPlayerDto(players[i], deck))
		}
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
			Type:    string(tile.Type),
			OwnerID: tile.OwnerID,
			Tags:    tile.Tags,
			Bonuses: convertTileBonuses(tile.Bonuses),
			Location: string(tile.Location),
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
		ViewingPlayerID:  currentPlayer.ID, // Set to first player for now
		CurrentTurn:      g.CurrentTurn(),
		Generation:       g.Generation(),
		TurnOrder:        []string{}, // Migration doesn't track turn order yet
		Board: BoardDto{
			Tiles: tileDtos,
		},
		PaymentConstants: paymentConstants,
	}
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

// getCorporationCard fetches the corporation card for a player from the deck
// TODO: Currently returns nil because card data is not available through Deck.
// Deck only stores card IDs. Full card data would require a card registry/repository.
// The player's corporationID is correctly stored and can be used for lookups when
// card data becomes available.
func getCorporationCard(p *player.Player, d *deck.Deck) *CardDto {
	if p.CorporationID() == "" {
		return nil
	}

	// TODO: Fetch actual card data once card registry is available
	// For now, return nil - corporation is correctly stored in player.PlayedCards()
	return nil
}

// ToPlayerDto converts migration Player to PlayerDto
func ToPlayerDto(p *player.Player, d *deck.Deck) PlayerDto {
	resourcesComponent := p.Resources()
	resources := resourcesComponent.Get()
	production := resourcesComponent.Production()

	// Get corporation card if player has one
	corporation := getCorporationCard(p, d)

	// Get played cards
	playedCards := p.PlayedCards().Cards()

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
		Cards:            []CardDto{},   // Migration doesn't have card details yet
		PlayedCards:      playedCards,
		Passed:           false,
		AvailableActions: 0,
		IsConnected:      true, // Assume connected
		Effects:          []PlayerEffectDto{},
		Actions:          []PlayerActionDto{},

		SelectStartingCardsPhase: nil,
		ProductionPhase:          nil,
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
func ToOtherPlayerDto(p *player.Player, d *deck.Deck) OtherPlayerDto {
	resourcesComponent := p.Resources()
	resources := resourcesComponent.Get()
	production := resourcesComponent.Production()

	// Get corporation card if player has one
	corporation := getCorporationCard(p, d)

	// Get played cards
	playedCards := p.PlayedCards().Cards()

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

		SelectStartingCardsPhase: nil,
		ProductionPhase:          nil,
		ResourceStorage:          map[string]int{},
		PaymentSubstitutes:       []PaymentSubstituteDto{},
	}
}
