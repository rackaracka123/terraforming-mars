package dto

import (
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"

	"go.uber.org/zap"
)

// ToGameDto converts a model Game to personalized GameDto
func ToGameDto(game model.Game, players []model.Player, viewingPlayerID string, cardRegistry *cards.CardRegistry) GameDto {
	// Find the viewing player and other players
	var currentPlayer PlayerDto
	// Initialize as empty slice instead of nil to prevent interface conversion panics
	otherPlayers := []OtherPlayerDto{}

	for _, player := range players {
		if player.ID == viewingPlayerID {
			currentPlayer = ToPlayerDto(player, cardRegistry)
		} else {
			otherPlayers = append(otherPlayers, PlayerToOtherPlayerDto(player))
		}
	}

	return GameDto{
		ID:               game.ID,
		Status:           GameStatus(game.Status),
		Settings:         ToGameSettingsDto(game.Settings),
		HostPlayerID:     game.HostPlayerID,
		CurrentPhase:     GamePhase(game.CurrentPhase),
		GlobalParameters: ToGlobalParametersDto(game.GlobalParameters),
		CurrentPlayer:    currentPlayer,
		OtherPlayers:     otherPlayers,
		ViewingPlayerID:  viewingPlayerID,
		CurrentTurn:      game.CurrentTurn,
		Generation:       game.Generation,
		TurnOrder:        game.PlayerIDs,
	}
}

// ToPlayerDto converts a model Player to PlayerDto
func ToPlayerDto(player model.Player, cardRegistry *cards.CardRegistry) PlayerDto {

	return PlayerDto{
		ID:                player.ID,
		Name:              player.Name,
		Corporation:       player.Corporation,
		Cards:             ToStartingSelectionDto(player.Cards, cardRegistry), // Convert card IDs to CardDto
		Resources:         ToResourcesDto(player.Resources),
		Production:        ToProductionDto(player.Production),
		TerraformRating:   player.TerraformRating,
		PlayedCards:       player.PlayedCards,
		Passed:            player.Passed,
		AvailableActions:  player.AvailableActions,
		VictoryPoints:     player.VictoryPoints,
		IsConnected:       player.IsConnected,
		CardSelection:     ToProductionPhaseDto(player.ProductionSelection),
		StartingSelection: ToStartingSelectionDto(player.StartingSelection, cardRegistry),
	}
}

// ToOtherPlayerDto removed - OtherPlayer model was deleted

// PlayerToOtherPlayerDto converts a model.Player to OtherPlayerDto (limited view)
func PlayerToOtherPlayerDto(player model.Player) OtherPlayerDto {
	corporationName := ""
	if player.Corporation != nil {
		corporationName = *player.Corporation
	}

	return OtherPlayerDto{
		ID:               player.ID,
		Name:             player.Name,
		Corporation:      corporationName,
		HandCardCount:    len(player.Cards), // Hide actual cards, show count only
		Resources:        ToResourcesDto(player.Resources),
		Production:       ToProductionDto(player.Production),
		TerraformRating:  player.TerraformRating,
		PlayedCards:      player.PlayedCards, // Played cards are public
		Passed:           player.Passed,
		AvailableActions: player.AvailableActions,
		VictoryPoints:    player.VictoryPoints,
		IsConnected:      player.IsConnected,
		IsSelectingCards: player.ProductionSelection != nil || player.StartingSelection != nil, // Whether player is currently selecting cards (production or starting)
	}
}

// ToResourcesDto converts model Resources to ResourcesDto
func ToResourcesDto(resources model.Resources) ResourcesDto {
	return ResourcesDto{
		Credits:  resources.Credits,
		Steel:    resources.Steel,
		Titanium: resources.Titanium,
		Plants:   resources.Plants,
		Energy:   resources.Energy,
		Heat:     resources.Heat,
	}
}

// ToProductionDto converts model Production to ProductionDto
func ToProductionDto(production model.Production) ProductionDto {
	return ProductionDto{
		Credits:  production.Credits,
		Steel:    production.Steel,
		Titanium: production.Titanium,
		Plants:   production.Plants,
		Energy:   production.Energy,
		Heat:     production.Heat,
	}
}

// ToGlobalParametersDto converts model GlobalParameters to GlobalParametersDto
func ToGlobalParametersDto(params model.GlobalParameters) GlobalParametersDto {
	return GlobalParametersDto{
		Temperature: params.Temperature,
		Oxygen:      params.Oxygen,
		Oceans:      params.Oceans,
	}
}

// ToGameSettingsDto converts model GameSettings to GameSettingsDto
func ToGameSettingsDto(settings model.GameSettings) GameSettingsDto {
	return GameSettingsDto{
		MaxPlayers: settings.MaxPlayers,
	}
}

// TODO: Create a new model for this usecase. Or rename the other "Game" that contains player data,
// ToGameDtoBasic provides a basic non-personalized game view (temporary compatibility)
// This is used for cases where personalization isn't needed (like game listings)
func ToGameDtoBasic(game model.Game) GameDto {
	return GameDto{
		ID:               game.ID,
		Status:           GameStatus(game.Status),
		Settings:         ToGameSettingsDto(game.Settings),
		HostPlayerID:     game.HostPlayerID,
		CurrentPhase:     GamePhase(game.CurrentPhase),
		GlobalParameters: ToGlobalParametersDto(game.GlobalParameters),
		CurrentPlayer:    PlayerDto{},        // Empty for non-personalized view
		OtherPlayers:     []OtherPlayerDto{}, // Empty for non-personalized view
		ViewingPlayerID:  "",                 // No viewing player for basic view
		CurrentTurn:      game.CurrentTurn,
		Generation:       game.Generation,
		RemainingActions: game.RemainingActions,
		TurnOrder:        game.PlayerIDs,
	}
}

// ToGameDtoSlice provides basic non-personalized game views (temporary compatibility)
func ToGameDtoSlice(games []model.Game) []GameDto {
	dtos := make([]GameDto, len(games))
	for i, game := range games {
		dtos[i] = ToGameDtoBasic(game)
	}
	return dtos
}

// ToPlayerDtoSlice converts a slice of model Players to PlayerDto slice
func ToPlayerDtoSlice(players []model.Player, cardRegistry *cards.CardRegistry) []PlayerDto {
	dtos := make([]PlayerDto, len(players))
	for i, player := range players {
		dtos[i] = ToPlayerDto(player, cardRegistry)
	}
	return dtos
}

// ToCardDto converts a model Card to CardDto
func ToCardDto(card model.Card) CardDto {

	// Convert behaviors to DTO format
	behaviors := ToCardBehaviorDtoSlice(card.Behaviors)

	// Convert resource storage to DTO format
	var resourceStorage *ResourceStorageDto
	if card.ResourceStorage != nil {
		resourceStorage = &ResourceStorageDto{
			Type:     card.ResourceStorage.Type,
			Capacity: card.ResourceStorage.Capacity,
			Starting: card.ResourceStorage.Starting,
		}
	}

	return CardDto{
		ID:              card.ID,
		Name:            card.Name,
		Type:            CardType(card.Type),
		Cost:            card.Cost,
		Description:     card.Description,
		Tags:            ToCardTagDtoSlice(card.Tags),
		Requirements:    card.Requirements,
		Behaviors:       behaviors,
		ResourceStorage: resourceStorage,
		VPConditions:    card.VPConditions,
	}
}

// ToCardDtoSlice converts a slice of model Cards to CardDto slice
func ToCardDtoSlice(cards []model.Card) []CardDto {
	if cards == nil {
		return nil
	}

	result := make([]CardDto, len(cards))
	for i, card := range cards {
		result[i] = ToCardDto(card)
	}
	return result
}

// ToCardTagDtoSlice converts a slice of model CardTags to CardTag slice
func ToCardTagDtoSlice(tags []model.CardTag) []CardTag {
	if tags == nil {
		return []CardTag{}
	}

	result := make([]CardTag, len(tags))
	for i, tag := range tags {
		result[i] = CardTag(tag)
	}
	return result
}

// ToProductionPhaseDto converts model ProductionPhase to ProductionPhaseDto
func ToProductionPhaseDto(phase *model.ProductionPhase) *ProductionPhaseDto {
	if phase == nil {
		return nil
	}

	return &ProductionPhaseDto{
		AvailableCards:    ToCardDtoSlice(phase.AvailableCards),
		SelectionComplete: phase.SelectionComplete,
	}
}

// ToStartingSelectionDto converts card IDs to real CardDto objects using card registry
func ToStartingSelectionDto(cardIDs []string, cardRegistry *cards.CardRegistry) []CardDto {
	if cardIDs == nil {
		return []CardDto{}
	}

	// DEBUG: Log what we're trying to convert
	logger.Debug("🔍 DEBUG: ToStartingSelectionDto called",
		zap.Strings("card_ids", cardIDs),
		zap.Bool("registry_available", cardRegistry != nil))

	result := make([]CardDto, len(cardIDs))
	for i, cardID := range cardIDs {
		// Try to get real card from registry
		if cardRegistry != nil {
			if card, exists := cardRegistry.GetCard(cardID); exists {
				logger.Debug("✅ DEBUG: Found card in registry",
					zap.String("card_id", cardID),
					zap.String("card_name", card.Name))
				result[i] = ToCardDto(*card)
				continue
			} else {
				logger.Debug("❌ DEBUG: Card NOT found in registry",
					zap.String("card_id", cardID))
			}
		} else {
			logger.Debug("❌ DEBUG: Card registry is nil")
		}

		// Fallback to placeholder if card not found
		logger.Debug("📝 DEBUG: Creating placeholder card",
			zap.String("card_id", cardID),
			zap.String("placeholder_name", "Starting Card "+string(rune('A'+i))))
		result[i] = CardDto{
			ID:          cardID,
			Name:        "Starting Card " + string(rune('A'+i)), // Placeholder names A, B, C, etc.
			Type:        CardTypeAutomated,                      // Default type
			Cost:        0,                                      // Placeholder cost
			Description: "Starting card option",                 // Placeholder description
			Tags:        []CardTag{},                            // Empty tags
			Requirements: nil,                                   // No requirements (nil slice)
			Behaviors:    nil,                                   // No behaviors
		}
	}
	return result
}

// ToCardBehaviorDtoSlice converts a slice of model CardBehaviors to CardBehaviorDto slice
func ToCardBehaviorDtoSlice(behaviors []model.CardBehavior) []CardBehaviorDto {
	if behaviors == nil {
		return nil
	}

	result := make([]CardBehaviorDto, len(behaviors))
	for i, behavior := range behaviors {
		result[i] = CardBehaviorDto{
			Triggers: behavior.Triggers,
			Inputs:   behavior.Inputs,
			Outputs:  behavior.Outputs,
			Choices:  behavior.Choices,
		}
	}
	return result
}
