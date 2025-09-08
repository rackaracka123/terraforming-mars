package dto

import "terraforming-mars-backend/internal/model"

// ToGameDto converts a model Game to GameDto
func ToGameDto(game model.Game) GameDto {
	// Legacy players field for compatibility
	players := make([]PlayerDto, len(game.Players))
	for i, player := range game.Players {
		players[i] = ToPlayerDto(player)
	}

	// Convert CurrentPlayer to DTO
	var currentPlayerDto *PlayerDto
	if game.CurrentPlayer != nil {
		dto := ToPlayerDto(*game.CurrentPlayer)
		currentPlayerDto = &dto
	}

	// Convert OtherPlayers to DTOs
	otherPlayers := make([]OtherPlayerDto, len(game.OtherPlayers))
	for i, otherPlayer := range game.OtherPlayers {
		otherPlayers[i] = ToOtherPlayerDto(otherPlayer)
	}

	return GameDto{
		ID:               game.ID,
		Status:           GameStatus(game.Status),
		Settings:         ToGameSettingsDto(game.Settings),
		CurrentPlayer:    currentPlayerDto,
		OtherPlayers:     otherPlayers,
		Players:          players, // Legacy field for compatibility
		HostPlayerID:     game.HostPlayerID,
		CurrentPhase:     GamePhase(game.CurrentPhase),
		GlobalParameters: ToGlobalParametersDto(game.GlobalParameters),
		CurrentPlayerID:  game.CurrentPlayerID,
		Generation:       game.Generation,
		RemainingActions: game.RemainingActions,
	}
}

// ToPlayerDto converts a model Player to PlayerDto
func ToPlayerDto(player model.Player) PlayerDto {
	return PlayerDto{
		ID:               player.ID,
		Name:             player.Name,
		Corporation:      player.Corporation,
		Cards:            player.Cards,
		Resources:        ToResourcesDto(player.Resources),
		Production:       ToProductionDto(player.Production),
		TerraformRating:  player.TerraformRating,
		IsActive:         player.IsActive,
		PlayedCards:      player.PlayedCards,
		Passed:           player.Passed,
		AvailableActions: player.AvailableActions,
		VictoryPoints:    player.VictoryPoints,
		MilestoneIcon:    player.MilestoneIcon,
		ConnectionStatus: player.ConnectionStatus,
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
		PlayedCards:      otherPlayer.PlayedCards,
		Passed:           otherPlayer.Passed,
		AvailableActions: otherPlayer.AvailableActions,
		VictoryPoints:    otherPlayer.VictoryPoints,
		MilestoneIcon:    otherPlayer.MilestoneIcon,
		ConnectionStatus: otherPlayer.ConnectionStatus,
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

// ToCardDto converts a model Card to CardDto
func ToCardDto(card model.Card) CardDto {
	return CardDto{
		ID:          card.ID,
		Name:        card.Name,
		Type:        CardType(card.Type),
		Cost:        card.Cost,
		Description: card.Description,
	}
}

// ToCardDtoSlice converts a slice of model Cards to CardDto slice
func ToCardDtoSlice(cards []model.Card) []CardDto {
	dtos := make([]CardDto, len(cards))
	for i, card := range cards {
		dtos[i] = ToCardDto(card)
	}
	return dtos
}

// ToGameDtoSlice converts a slice of model Games to GameDto slice
func ToGameDtoSlice(games []model.Game) []GameDto {
	dtos := make([]GameDto, len(games))
	for i, game := range games {
		dtos[i] = ToGameDto(game)
	}
	return dtos
}
