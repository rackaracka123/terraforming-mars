package dto

import (
	"terraforming-mars-backend/internal/model"
)

// resolveCards is a helper function to resolve card IDs to Card objects for multiple players
func resolveCards(cardIDs []string, resolvedMap map[string]model.Card) []CardDto {
	if resolvedMap == nil {
		return nil
	}

	cards := make([]CardDto, len(cardIDs))
	for i, cardID := range cardIDs {
		if card, exists := resolvedMap[cardID]; exists {
			cards[i] = ToCardDto(card)
		} else {
			// Fallback to a placeholder card if not found
			cards[i] = CardDto{
				ID:   cardID,
				Name: "Unknown Card",
			}
		}
	}

	return cards
}

// ToGameDto converts a model Game to personalized GameDto
func ToGameDto(game model.Game, players []model.Player, viewingPlayerID string, resolvedCards map[string]model.Card) GameDto {
	var currentPlayer PlayerDto
	otherPlayers := make([]OtherPlayerDto, 0)

	for _, player := range players {
		if player.ID == viewingPlayerID {
			currentPlayer = ToPlayerDto(player, resolvedCards)
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

// ToPlayerDto converts a model Player to PlayerDto with resolved cards
func ToPlayerDto(player model.Player, resolvedCards map[string]model.Card) PlayerDto {
	status := PlayerStatusActive
	if player.Passed {
		status = PlayerStatusWaiting
	} else if player.SelectStartingCardsPhase != nil {
		if player.SelectStartingCardsPhase.SelectionComplete {
			status = PlayerStatusActive
		} else {
			status = PlayerStatusSelectingStartingCards
		}
	} else if player.ProductionPhase != nil {
		if player.ProductionPhase.SelectionComplete {
			status = PlayerStatusActive
		} else {
			status = PlayerStatusSelectingProductionCards
		}
	}

	// Extract starting cards from SelectStartingCardsPhase if present
	var startingCards []CardDto
	if player.SelectStartingCardsPhase != nil && len(player.SelectStartingCardsPhase.AvailableCards) > 0 {
		startingCards = resolveCards(player.SelectStartingCardsPhase.AvailableCards, resolvedCards)
	} else {
		startingCards = []CardDto{}
	}

	// Convert corporation to CardDto if present
	var corporationDto *CardDto
	if player.Corporation != nil {
		dto := ToCardDto(*player.Corporation)
		corporationDto = &dto
	}

	return PlayerDto{
		ID:                       player.ID,
		Name:                     player.Name,
		Status:                   status,
		Corporation:              corporationDto,
		Cards:                    resolveCards(player.Cards, resolvedCards),
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
		SelectStartingCardsPhase: ToSelectStartingCardsPhaseDto(player.SelectStartingCardsPhase, resolvedCards),
		ProductionPhase:          ToProductionPhaseDto(player.ProductionPhase, resolvedCards),
		StartingCards:            startingCards,
		PendingTileSelection:     ToPendingTileSelectionDto(player.PendingTileSelection),
		PendingCardSelection:     ToPendingCardSelectionDto(player.PendingCardSelection, resolvedCards),
		ResourceStorage:          player.ResourceStorage,
	}
}

// PlayerToOtherPlayerDto converts a model.Player to OtherPlayerDto (limited view)
func PlayerToOtherPlayerDto(player model.Player) OtherPlayerDto {
	// Convert corporation to CardDto if present
	var corporationDto *CardDto
	if player.Corporation != nil {
		dto := ToCardDto(*player.Corporation)
		corporationDto = &dto
	}

	status := PlayerStatusActive
	if player.Passed {
		status = PlayerStatusWaiting
	} else if player.SelectStartingCardsPhase != nil {
		if player.SelectStartingCardsPhase.SelectionComplete {
			status = PlayerStatusActive
		} else {
			status = PlayerStatusSelectingStartingCards
		}
	} else if player.ProductionPhase != nil {
		if player.ProductionPhase.SelectionComplete {
			status = PlayerStatusActive
		} else {
			status = PlayerStatusSelectingProductionCards
		}
	}

	return OtherPlayerDto{
		ID:                       player.ID,
		Name:                     player.Name,
		Status:                   status,
		Corporation:              corporationDto,
		HandCardCount:            len(player.Cards), // Hide actual cards, show count only
		Resources:                ToResourcesDto(player.Resources),
		Production:               ToProductionDto(player.Production),
		TerraformRating:          player.TerraformRating,
		PlayedCards:              player.PlayedCards, // Played cards are public
		Passed:                   player.Passed,
		AvailableActions:         player.AvailableActions,
		VictoryPoints:            player.VictoryPoints,
		IsConnected:              player.IsConnected,
		Effects:                  ToPlayerEffectDtoSlice(player.Effects),
		Actions:                  ToPlayerActionDtoSlice(player.Actions),
		SelectStartingCardsPhase: ToSelectStartingCardsOtherPlayerDto(player.SelectStartingCardsPhase),
		ProductionPhase:          ToProductionPhaseOtherPlayerDto(player.ProductionPhase),
		ResourceStorage:          player.ResourceStorage, // Resource storage is public information
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

// ToGameDtoBasic provides a basic non-personalized game view (temporary compatibility)
// This is used for cases where personalization isn't needed (like game listings)
// TODO: Create a new model for this usecase. Or rename the other "Game" that contains player data,
func ToGameDtoBasic(game model.Game) GameDto {
	return GameDto{
		ID:               game.ID,
		Status:           GameStatus(game.Status),
		Settings:         ToGameSettingsDto(game.Settings),
		HostPlayerID:     game.HostPlayerID,
		CurrentPhase:     GamePhase(game.CurrentPhase),
		GlobalParameters: ToGlobalParametersDto(game.GlobalParameters),
		CurrentPlayer:    PlayerDto{},               // Empty for non-personalized view
		OtherPlayers:     make([]OtherPlayerDto, 0), // Empty for non-personalized view
		ViewingPlayerID:  "",                        // No viewing player for basic view
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
		dtos[i] = ToPlayerDto(player, nil) // Empty cards for basic conversion
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

	// Convert starting bonuses to DTO format (for corporations)
	var startingCredits *int
	var startingResources *ResourceSet
	var startingProduction *ResourceSet

	if card.StartingCredits != nil {
		startingCredits = card.StartingCredits
	}
	if card.StartingResources != nil {
		rs := ToResourceSetDto(*card.StartingResources)
		startingResources = &rs
	}
	if card.StartingProduction != nil {
		rp := ToResourceSetDto(*card.StartingProduction)
		startingProduction = &rp
	}

	return CardDto{
		ID:                 card.ID,
		Name:               card.Name,
		Type:               CardType(card.Type),
		Cost:               card.Cost,
		Description:        card.Description,
		Tags:               ToCardTagDtoSlice(card.Tags),
		Requirements:       card.Requirements,
		Behaviors:          behaviors,
		ResourceStorage:    resourceStorage,
		VPConditions:       card.VPConditions,
		StartingCredits:    startingCredits,
		StartingResources:  startingResources,
		StartingProduction: startingProduction,
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

// ToSelectStartingCardsPhaseDto converts model SelectStartingCardsPhase to SelectStartingCardsPhaseDto
func ToSelectStartingCardsPhaseDto(phase *model.SelectStartingCardsPhase, resolvedCards map[string]model.Card) *SelectStartingCardsPhaseDto {
	if phase == nil {
		return nil
	}

	return &SelectStartingCardsPhaseDto{
		AvailableCards:        resolveCards(phase.AvailableCards, resolvedCards),
		AvailableCorporations: phase.AvailableCorporations,
		SelectionComplete:     phase.SelectionComplete,
	}
}

func ToSelectStartingCardsOtherPlayerDto(phase *model.SelectStartingCardsPhase) *SelectStartingCardsOtherPlayerDto {
	if phase == nil {
		return nil
	}

	return &SelectStartingCardsOtherPlayerDto{
		SelectionComplete: phase.SelectionComplete,
	}
}

// ToProductionPhaseDto converts model ProductionPhase to ProductionPhaseDto
func ToProductionPhaseDto(phase *model.ProductionPhase, resolvedCards map[string]model.Card) *ProductionPhaseDto {
	if phase == nil {
		return nil
	}

	delta := model.Resources{
		Credits:  phase.AfterResources.Credits - phase.BeforeResources.Credits,
		Steel:    phase.AfterResources.Steel - phase.BeforeResources.Steel,
		Titanium: phase.AfterResources.Titanium - phase.BeforeResources.Titanium,
		Plants:   phase.AfterResources.Plants - phase.BeforeResources.Plants,
		Energy:   phase.AfterResources.Energy - phase.BeforeResources.Energy,
		Heat:     phase.AfterResources.Heat - phase.BeforeResources.Heat,
	}

	return &ProductionPhaseDto{
		AvailableCards:    resolveCards(phase.AvailableCards, resolvedCards),
		SelectionComplete: phase.SelectionComplete,
		BeforeResources:   ToResourcesDto(phase.BeforeResources),
		AfterResources:    ToResourcesDto(phase.AfterResources),
		ResourceDelta:     ToResourcesDto(delta),
		EnergyConverted:   phase.EnergyConverted,
		CreditsIncome:     phase.CreditsIncome,
	}
}

func ToProductionPhaseOtherPlayerDto(phase *model.ProductionPhase) *ProductionPhaseOtherPlayerDto {
	if phase == nil {
		return nil
	}

	delta := model.Resources{
		Credits:  phase.AfterResources.Credits - phase.BeforeResources.Credits,
		Steel:    phase.AfterResources.Steel - phase.BeforeResources.Steel,
		Titanium: phase.AfterResources.Titanium - phase.BeforeResources.Titanium,
		Plants:   phase.AfterResources.Plants - phase.BeforeResources.Plants,
		Energy:   phase.AfterResources.Energy - phase.BeforeResources.Energy,
		Heat:     phase.AfterResources.Heat - phase.BeforeResources.Heat,
	}

	return &ProductionPhaseOtherPlayerDto{
		SelectionComplete: phase.SelectionComplete,
		BeforeResources:   ToResourcesDto(phase.BeforeResources),
		AfterResources:    ToResourcesDto(phase.AfterResources),
		ResourceDelta:     ToResourcesDto(delta),
		EnergyConverted:   phase.EnergyConverted,
		CreditsIncome:     phase.CreditsIncome,
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
	return PlayerEffectDto{
		CardID:        effect.CardID,
		CardName:      effect.CardName,
		BehaviorIndex: effect.BehaviorIndex,
		Behavior:      ToCardBehaviorDto(effect.Behavior),
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
// Filters out initial actions that have already been used (PlayCount > 0)
func ToPlayerActionDtoSlice(actions []model.PlayerAction) []PlayerActionDto {
	if actions == nil {
		return []PlayerActionDto{}
	}
	result := make([]PlayerActionDto, 0, len(actions))
	for _, action := range actions {
		// Check if this is an initial action that has already been played
		isInitialAction := false
		if len(action.Behavior.Triggers) > 0 {
			isInitialAction = action.Behavior.Triggers[0].IsInitialAction
		}

		// Skip initial actions that have been used
		if isInitialAction && action.PlayCount > 0 {
			continue
		}

		result = append(result, ToPlayerActionDto(action))
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

// ToPendingTileSelectionDto converts a model PendingTileSelection pointer to PendingTileSelectionDto pointer
func ToPendingTileSelectionDto(selection *model.PendingTileSelection) *PendingTileSelectionDto {
	if selection == nil {
		return nil
	}

	return &PendingTileSelectionDto{
		TileType:       selection.TileType,
		AvailableHexes: selection.AvailableHexes,
		Source:         selection.Source,
	}
}

// ToPendingCardSelectionDto converts a model PendingCardSelection to PendingCardSelectionDto
func ToPendingCardSelectionDto(selection *model.PendingCardSelection, resolvedCards map[string]model.Card) *PendingCardSelectionDto {
	if selection == nil {
		return nil
	}

	// Resolve available cards from card IDs
	availableCards := resolveCards(selection.AvailableCards, resolvedCards)

	return &PendingCardSelectionDto{
		AvailableCards: availableCards,
		CardCosts:      selection.CardCosts,
		CardRewards:    selection.CardRewards,
		Source:         selection.Source,
		MinCards:       selection.MinCards,
		MaxCards:       selection.MaxCards,
	}
}

// ToResourceSetDto converts a model ResourceSet to ResourceSet
func ToResourceSetDto(rs model.ResourceSet) ResourceSet {
	return ResourceSet{
		Credits:  rs.Credits,
		Steel:    rs.Steel,
		Titanium: rs.Titanium,
		Plants:   rs.Plants,
		Energy:   rs.Energy,
		Heat:     rs.Heat,
	}
}
