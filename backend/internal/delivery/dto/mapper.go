package dto

import "terraforming-mars-backend/internal/model"

// ToGameDto converts a model Game to personalized GameDto
func ToGameDto(game model.Game, players []model.Player, viewingPlayerID string) GameDto {
	// Find the viewing player and other players
	var currentPlayer PlayerDto
	// Initialize as empty slice instead of nil to prevent interface conversion panics
	otherPlayers := []OtherPlayerDto{}

	for _, player := range players {
		if player.ID == viewingPlayerID {
			currentPlayer = ToPlayerDto(player)
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
func ToPlayerDto(player model.Player) PlayerDto {
	return PlayerDto{
		ID:                player.ID,
		Name:              player.Name,
		Corporation:       player.Corporation,
		Cards:             player.Cards,
		Resources:         ToResourcesDto(player.Resources),
		Production:        ToProductionDto(player.Production),
		TerraformRating:   player.TerraformRating,
		IsActive:          player.IsActive,
		IsReady:           player.IsReady,
		PlayedCards:       player.PlayedCards,
		Passed:            player.Passed,
		AvailableActions:  player.AvailableActions,
		VictoryPoints:     player.VictoryPoints,
		MilestoneIcon:     player.MilestoneIcon,
		ConnectionStatus:  player.ConnectionStatus,
		CardSelection:     ToProductionPhaseDto(player.ProductionSelection),
		StartingSelection: ToCardDtoSlice(player.StartingSelection),
	}
}

// ToOtherPlayerDto converts a model OtherPlayer to OtherPlayerDto
func ToOtherPlayerDto(otherPlayer model.OtherPlayer) OtherPlayerDto {
	return OtherPlayerDto{
		ID:               otherPlayer.ID,
		Name:             otherPlayer.Name,
		Corporation:      otherPlayer.Corporation,
		HandCardCount:    otherPlayer.HandCardCount,
		Resources:        ToResourcesDto(otherPlayer.Resources),
		Production:       ToProductionDto(otherPlayer.Production),
		TerraformRating:  otherPlayer.TerraformRating,
		IsActive:         otherPlayer.IsActive,
		IsReady:          otherPlayer.IsReady,
		PlayedCards:      otherPlayer.PlayedCards,
		Passed:           otherPlayer.Passed,
		AvailableActions: otherPlayer.AvailableActions,
		VictoryPoints:    otherPlayer.VictoryPoints,
		MilestoneIcon:    otherPlayer.MilestoneIcon,
		ConnectionStatus: otherPlayer.ConnectionStatus,
	}
}

// PlayerToOtherPlayerDto converts a model.Player to OtherPlayerDto (limited view)
func PlayerToOtherPlayerDto(player model.Player) OtherPlayerDto {
	return OtherPlayerDto{
		ID:               player.ID,
		Name:             player.Name,
		Corporation:      player.Corporation,
		HandCardCount:    len(player.Cards), // Hide actual cards, show count only
		Resources:        ToResourcesDto(player.Resources),
		Production:       ToProductionDto(player.Production),
		TerraformRating:  player.TerraformRating,
		IsActive:         player.IsActive,
		IsReady:          player.IsReady,
		PlayedCards:      player.PlayedCards, // Played cards are public
		Passed:           player.Passed,
		AvailableActions: player.AvailableActions,
		VictoryPoints:    player.VictoryPoints,
		MilestoneIcon:    player.MilestoneIcon,
		ConnectionStatus: player.ConnectionStatus,
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
func ToPlayerDtoSlice(players []model.Player) []PlayerDto {
	dtos := make([]PlayerDto, len(players))
	for i, player := range players {
		dtos[i] = ToPlayerDto(player)
	}
	return dtos
}

// ToCardDto converts a model Card to CardDto
func ToCardDto(card model.Card) CardDto {
	// Convert production effects if present
	var productionEffects *ProductionEffects
	if card.ProductionEffects != nil {
		productionEffects = &ProductionEffects{
			Credits:  card.ProductionEffects.Credits,
			Steel:    card.ProductionEffects.Steel,
			Titanium: card.ProductionEffects.Titanium,
			Plants:   card.ProductionEffects.Plants,
			Energy:   card.ProductionEffects.Energy,
			Heat:     card.ProductionEffects.Heat,
		}
	}

	// Convert requirements
	requirements := CardRequirements{
		MinTemperature: card.Requirements.MinTemperature,
		MaxTemperature: card.Requirements.MaxTemperature,
		MinOxygen:      card.Requirements.MinOxygen,
		MaxOxygen:      card.Requirements.MaxOxygen,
		MinOceans:      card.Requirements.MinOceans,
		MaxOceans:      card.Requirements.MaxOceans,
		RequiredTags:   ToCardTagDtoSlice(card.Requirements.RequiredTags),
	}

	// Convert required production if present
	if card.Requirements.RequiredProduction != nil {
		requirements.RequiredProduction = &ResourceSet{
			Credits:  card.Requirements.RequiredProduction.Credits,
			Steel:    card.Requirements.RequiredProduction.Steel,
			Titanium: card.Requirements.RequiredProduction.Titanium,
			Plants:   card.Requirements.RequiredProduction.Plants,
			Energy:   card.Requirements.RequiredProduction.Energy,
			Heat:     card.Requirements.RequiredProduction.Heat,
		}
	}

	return CardDto{
		ID:                card.ID,
		Name:              card.Name,
		Type:              CardType(card.Type),
		Cost:              card.Cost,
		Description:       card.Description,
		Tags:              ToCardTagDtoSlice(card.Tags),
		Requirements:      requirements,
		VictoryPoints:     card.VictoryPoints,
		Number:            card.Number,
		ProductionEffects: productionEffects,
	}
}

// ToCardDtoSlice converts a slice of model Cards to CardDto slice
func ToCardDtoSlice(cards []model.Card) []CardDto {
	if cards == nil {
		return []CardDto{}
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
