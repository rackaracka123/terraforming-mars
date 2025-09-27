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
		Actions:                  ToPlayerActionDtoSlice(player.Actions),
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
		Actions:                  ToPlayerActionDtoSlice(player.Actions),
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
		Actions:          ToPlayerActionDtoSlice(player.Actions),                               // Actions are public information
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

// ToResourceConditionDto converts a model ResourceCondition to ResourceConditionDto
func ToResourceConditionDto(rc model.ResourceCondition) ResourceConditionDto {
	return ResourceConditionDto{
		Type:              ResourceType(rc.Type),
		Amount:            rc.Amount,
		Target:            TargetType(rc.Target),
		AffectedResources: rc.AffectedResources,
		AffectedTags:      ToCardTagDtoSlice(rc.AffectedTags),
		MaxTrigger:        rc.MaxTrigger,
		Per:               ToPerConditionDto(rc.Per),
	}
}

// ToResourceConditionDtoSlice converts a slice of model ResourceCondition to ResourceConditionDto slice
func ToResourceConditionDtoSlice(conditions []model.ResourceCondition) []ResourceConditionDto {
	if conditions == nil {
		return nil
	}

	result := make([]ResourceConditionDto, len(conditions))
	for i, condition := range conditions {
		result[i] = ToResourceConditionDto(condition)
	}
	return result
}

// ToPerConditionDto converts a model PerCondition pointer to PerConditionDto pointer
func ToPerConditionDto(per *model.PerCondition) *PerConditionDto {
	if per == nil {
		return nil
	}

	return &PerConditionDto{
		Type:     ResourceType(per.Type),
		Amount:   per.Amount,
		Location: ToCardApplyLocationPointer(per.Location),
		Target:   ToTargetTypePointer(per.Target),
		Tag:      ToCardTagPointer(per.Tag),
	}
}

// ToChoiceDto converts a model Choice to ChoiceDto
func ToChoiceDto(choice model.Choice) ChoiceDto {
	return ChoiceDto{
		Inputs:  ToResourceConditionDtoSlice(choice.Inputs),
		Outputs: ToResourceConditionDtoSlice(choice.Outputs),
	}
}

// ToChoiceDtoSlice converts a slice of model Choice to ChoiceDto slice
func ToChoiceDtoSlice(choices []model.Choice) []ChoiceDto {
	if choices == nil {
		return nil
	}

	result := make([]ChoiceDto, len(choices))
	for i, choice := range choices {
		result[i] = ToChoiceDto(choice)
	}
	return result
}

// ToTriggerDto converts a model Trigger to TriggerDto
func ToTriggerDto(trigger model.Trigger) TriggerDto {
	return TriggerDto{
		Type:      ResourceTriggerType(trigger.Type),
		Condition: ToResourceTriggerConditionDto(trigger.Condition),
	}
}

// ToTriggerDtoSlice converts a slice of model Trigger to TriggerDto slice
func ToTriggerDtoSlice(triggers []model.Trigger) []TriggerDto {
	if triggers == nil {
		return nil
	}

	result := make([]TriggerDto, len(triggers))
	for i, trigger := range triggers {
		result[i] = ToTriggerDto(trigger)
	}
	return result
}

// ToResourceTriggerConditionDto converts a model ResourceTriggerCondition pointer to ResourceTriggerConditionDto pointer
func ToResourceTriggerConditionDto(condition *model.ResourceTriggerCondition) *ResourceTriggerConditionDto {
	if condition == nil {
		return nil
	}

	return &ResourceTriggerConditionDto{
		Type:         TriggerType(condition.Type),
		Location:     ToCardApplyLocationPointer(condition.Location),
		AffectedTags: ToCardTagDtoSlice(condition.AffectedTags),
	}
}

// ToCardBehaviorDto converts a model CardBehavior to CardBehaviorDto
func ToCardBehaviorDto(behavior model.CardBehavior) CardBehaviorDto {
	return CardBehaviorDto{
		Triggers: ToTriggerDtoSlice(behavior.Triggers),
		Inputs:   ToResourceConditionDtoSlice(behavior.Inputs),
		Outputs:  ToResourceConditionDtoSlice(behavior.Outputs),
		Choices:  ToChoiceDtoSlice(behavior.Choices),
	}
}

// ToCardBehaviorDtoSlice converts a slice of model CardBehaviors to CardBehaviorDto slice
func ToCardBehaviorDtoSlice(behaviors []model.CardBehavior) []CardBehaviorDto {
	if behaviors == nil {
		return nil
	}

	result := make([]CardBehaviorDto, len(behaviors))
	for i, behavior := range behaviors {
		result[i] = ToCardBehaviorDto(behavior)
	}
	return result
}

// Helper functions for type conversions

// ToCardApplyLocationPointer converts model CardApplyLocation pointer to DTO CardApplyLocation pointer
func ToCardApplyLocationPointer(ptr *model.CardApplyLocation) *CardApplyLocation {
	if ptr == nil {
		return nil
	}
	result := CardApplyLocation(*ptr)
	return &result
}

// ToTargetTypePointer converts model TargetType pointer to DTO TargetType pointer
func ToTargetTypePointer(ptr *model.TargetType) *TargetType {
	if ptr == nil {
		return nil
	}
	result := TargetType(*ptr)
	return &result
}

// ToCardTagPointer converts model CardTag pointer to DTO CardTag pointer
func ToCardTagPointer(ptr *model.CardTag) *CardTag {
	if ptr == nil {
		return nil
	}
	result := CardTag(*ptr)
	return &result
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

// ToPlayerActionDto converts a model PlayerAction to PlayerActionDto
func ToPlayerActionDto(action model.PlayerAction) PlayerActionDto {
	return PlayerActionDto{
		CardID:        action.CardID,
		CardName:      action.CardName,
		BehaviorIndex: action.BehaviorIndex,
		Behavior:      ToCardBehaviorDto(action.Behavior),
		PlayCount:     action.PlayCount,
	}
}

// ToPlayerActionDtoSlice converts a slice of model PlayerActions to PlayerActionDto slice
func ToPlayerActionDtoSlice(actions []model.PlayerAction) []PlayerActionDto {
	if actions == nil {
		return []PlayerActionDto{}
	}
	result := make([]PlayerActionDto, len(actions))
	for i, action := range actions {
		result[i] = ToPlayerActionDto(action)
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
