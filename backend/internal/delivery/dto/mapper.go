package dto

import (
	"context"
	"terraforming-mars-backend/internal/model"
)

// ToGameDto converts a model Game to personalized GameDto
func ToGameDto(game model.Game, players []model.Player, viewingPlayerID string, playerCards ...map[string][]model.Card) GameDto {
	var cardMap map[string][]model.Card
	var startingCardMap map[string][]model.Card
	if len(playerCards) > 0 {
		cardMap = playerCards[0]
	}
	if len(playerCards) > 1 {
		startingCardMap = playerCards[1]
	}
	// Find the viewing player and other players
	var currentPlayer PlayerDto
	// Initialize as empty slice instead of nil to prevent interface conversion panics
	otherPlayers := []OtherPlayerDto{}

	for _, player := range players {
		if player.ID == viewingPlayerID {
			// Get resolved cards for this player if provided
			var cardDtos []CardDto
			if cardMap != nil {
				if cards, exists := cardMap[player.ID]; exists {
					cardDtos = ToCardDtoSlice(cards)
				}
			}
			// Get resolved starting cards for this player if provided
			var startingCardDtos []CardDto
			if startingCardMap != nil {
				if startingCards, exists := startingCardMap[player.ID]; exists {
					startingCardDtos = ToCardDtoSlice(startingCards)
				}
			}
			currentPlayer = ToPlayerDto(player, cardDtos, startingCardDtos)
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
		Board:            ToBoardDto(game.Board),
	}
}

// ToPlayerDtoWithCardService converts a model Player to PlayerDto with card lookup capability
// TODO: Implement card service integration when import cycle is resolved
func ToPlayerDtoWithCardService(ctx context.Context, player model.Player) PlayerDto {
	// Note: Card service integration commented out due to import cycle
	// This function should be moved to a higher-level service layer
	return PlayerDto{
		ID:                       player.ID,
		Name:                     player.Name,
		Corporation:              player.Corporation,
		Cards:                    []CardDto{}, // Empty until card service integration is resolved
		Resources:                ToResourcesDto(player.Resources),
		Production:               ToProductionDto(player.Production),
		TerraformRating:          player.TerraformRating,
		PlayedCards:              player.PlayedCards,
		Passed:                   player.Passed,
		AvailableActions:         player.AvailableActions,
		VictoryPoints:            player.VictoryPoints,
		IsConnected:              player.IsConnected,
		Effects:                  ToPlayerEffectDtoSlice(player.Effects),
		CardSelection:            ToProductionPhaseDto(player.ProductionSelection),
		StartingSelection:        []CardDto{}, // Empty until card service integration is resolved
		HasSelectedStartingCards: player.HasSelectedStartingCards,
	}
}

// ToPlayerDto converts a model Player to PlayerDto with resolved cards
func ToPlayerDto(player model.Player, playerCards []CardDto, startingCards ...[]CardDto) PlayerDto {
	var startingCardDtos []CardDto
	if len(startingCards) > 0 {
		startingCardDtos = startingCards[0]
	}
	return PlayerDto{
		ID:                       player.ID,
		Name:                     player.Name,
		Corporation:              player.Corporation,
		Cards:                    playerCards, // Use resolved cards from hand
		Resources:                ToResourcesDto(player.Resources),
		Production:               ToProductionDto(player.Production),
		TerraformRating:          player.TerraformRating,
		PlayedCards:              player.PlayedCards,
		Passed:                   player.Passed,
		AvailableActions:         player.AvailableActions,
		VictoryPoints:            player.VictoryPoints,
		IsConnected:              player.IsConnected,
		Effects:                  ToPlayerEffectDtoSlice(player.Effects),
		CardSelection:            ToProductionPhaseDto(player.ProductionSelection),
		StartingSelection:        startingCardDtos, // Resolved starting card selection
		HasSelectedStartingCards: player.HasSelectedStartingCards,
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
		PlayedCards:      otherPlayer.PlayedCards,
		Passed:           otherPlayer.Passed,
		AvailableActions: otherPlayer.AvailableActions,
		VictoryPoints:    otherPlayer.VictoryPoints,
		IsConnected:      otherPlayer.IsConnected,
		Effects:          ToPlayerEffectDtoSlice(otherPlayer.Effects),
		IsSelectingCards: otherPlayer.IsSelectingCards,
	}
}

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
		Effects:          ToPlayerEffectDtoSlice(player.Effects),                               // Effects are public information
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
		MaxPlayers:      settings.MaxPlayers,
		DevelopmentMode: settings.DevelopmentMode,
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
		Board:            ToBoardDto(game.Board),
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

// ToPlayerDtoSlice converts a slice of model Players to PlayerDto slice with empty cards
func ToPlayerDtoSlice(players []model.Player) []PlayerDto {
	dtos := make([]PlayerDto, len(players))
	for i, player := range players {
		dtos[i] = ToPlayerDto(player, []CardDto{}) // Empty cards for basic conversion
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

// ToPlayerEffectDto converts a model PlayerEffect to PlayerEffectDto
func ToPlayerEffectDto(effect model.PlayerEffect) PlayerEffectDto {
	// Convert model PlayerEffectType to DTO PlayerEffectType
	var dtoType PlayerEffectType
	switch effect.Type {
	case model.PlayerEffectDiscount:
		dtoType = PlayerEffectTypeDiscount
	case model.PlayerEffectGlobalParameterLenience:
		dtoType = PlayerEffectTypeGlobalParameterLenience
	case model.PlayerEffectDefense:
		dtoType = PlayerEffectTypeDefense
	case model.PlayerEffectValueModifier:
		dtoType = PlayerEffectTypeValueModifier
	default:
		dtoType = PlayerEffectType(effect.Type) // Fallback to string conversion
	}

	return PlayerEffectDto{
		Type:         dtoType,
		Amount:       effect.Amount,
		AffectedTags: ToCardTagDtoSlice(effect.AffectedTags),
	}
}

// ToPlayerEffectDtoSlice converts a slice of model PlayerEffects to PlayerEffectDto slice
func ToPlayerEffectDtoSlice(effects []model.PlayerEffect) []PlayerEffectDto {
	if effects == nil {
		return []PlayerEffectDto{}
	}

	result := make([]PlayerEffectDto, len(effects))
	for i, effect := range effects {
		result[i] = ToPlayerEffectDto(effect)
	}
	return result
}

// Board conversion functions

// ToBoardDto converts a model Board to BoardDto
func ToBoardDto(board model.Board) BoardDto {
	return BoardDto{
		Tiles: ToTileDtoSlice(board.Tiles),
	}
}

// ToTileDto converts a model Tile to TileDto
func ToTileDto(tile model.Tile) TileDto {
	return TileDto{
		Coordinates: HexPositionDto{
			Q: tile.Coordinates.Q,
			R: tile.Coordinates.R,
			S: tile.Coordinates.S,
		},
		Tags:        tile.Tags,
		Type:        string(tile.Type),
		Location:    string(tile.Location),
		DisplayName: tile.DisplayName,
		Bonuses:     ToTileBonusDtoSlice(tile.Bonuses),
		OccupiedBy:  ToTileOccupantDto(tile.OccupiedBy),
		OwnerID:     tile.OwnerID,
	}
}

// ToTileDtoSlice converts a slice of model Tiles to TileDto slice
func ToTileDtoSlice(tiles []model.Tile) []TileDto {
	if tiles == nil {
		return nil
	}

	result := make([]TileDto, len(tiles))
	for i, tile := range tiles {
		result[i] = ToTileDto(tile)
	}
	return result
}

// ToTileBonusDto converts a model TileBonus to TileBonusDto
func ToTileBonusDto(bonus model.TileBonus) TileBonusDto {
	return TileBonusDto{
		Type:   string(bonus.Type),
		Amount: bonus.Amount,
	}
}

// ToTileBonusDtoSlice converts a slice of model TileBonus to TileBonusDto slice
func ToTileBonusDtoSlice(bonuses []model.TileBonus) []TileBonusDto {
	if bonuses == nil {
		return []TileBonusDto{}
	}

	result := make([]TileBonusDto, len(bonuses))
	for i, bonus := range bonuses {
		result[i] = ToTileBonusDto(bonus)
	}
	return result
}

// ToTileOccupantDto converts a model TileOccupant pointer to TileOccupantDto pointer
func ToTileOccupantDto(occupant *model.TileOccupant) *TileOccupantDto {
	if occupant == nil {
		return nil
	}

	return &TileOccupantDto{
		Type: string(occupant.Type),
		Tags: occupant.Tags,
	}
}
