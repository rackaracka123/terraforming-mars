package dto

import (
	"terraforming-mars-backend/internal/features/card"
	"terraforming-mars-backend/internal/features/production"
	"terraforming-mars-backend/internal/features/resources"
	"terraforming-mars-backend/internal/features/tiles"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/player"
	"terraforming-mars-backend/internal/shared/types"
)

// ToCardPayment converts a CardPaymentDto to domain model CardPayment
func ToCardPayment(dto CardPaymentDto) types.CardPayment {
	payment := types.CardPayment{
		Credits:  dto.Credits,
		Steel:    dto.Steel,
		Titanium: dto.Titanium,
	}

	// Convert substitutes map from string keys to ResourceType keys
	if dto.Substitutes != nil && len(dto.Substitutes) > 0 {
		payment.Substitutes = make(map[types.ResourceType]int, len(dto.Substitutes))
		for resourceStr, amount := range dto.Substitutes {
			payment.Substitutes[types.ResourceType(resourceStr)] = amount
		}
	}

	return payment
}

// resolveCards is a helper function to resolve card IDs to Card objects for multiple players
func resolveCards(cardIDs []string, resolvedMap map[string]card.Card) []CardDto {
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
func ToGameDto(game game.Game, players []player.Player, viewingPlayerID string, resolvedCards map[string]card.Card, paymentConstants PaymentConstantsDto) (GameDto, error) {
	var currentPlayer PlayerDto
	otherPlayers := make([]OtherPlayerDto, 0)

	for _, player := range players {
		if player.ID == viewingPlayerID {
			playerDto, err := ToPlayerDto(player, resolvedCards)
			if err != nil {
				return GameDto{}, err
			}
			currentPlayer = playerDto
		} else {
			otherPlayerDto, err := PlayerToOtherPlayerDto(player)
			if err != nil {
				return GameDto{}, err
			}
			otherPlayers = append(otherPlayers, otherPlayerDto)
		}
	}

	// Get global parameters via service
	globalParams, err := game.GetGlobalParameters()
	if err != nil {
		return GameDto{}, err
	}

	// Get current turn via service
	currentTurn, err := game.GetCurrentTurn()
	if err != nil {
		return GameDto{}, err
	}

	// Get board via service
	board, err := game.GetBoard()
	if err != nil {
		return GameDto{}, err
	}

	return GameDto{
		ID:               game.ID,
		Status:           GameStatus(game.Status),
		Settings:         ToGameSettingsDto(game.Settings),
		HostPlayerID:     game.HostPlayerID,
		CurrentPhase:     GamePhase(game.CurrentPhase),
		GlobalParameters: ToGlobalParametersDto(globalParams),
		CurrentPlayer:    currentPlayer,
		OtherPlayers:     otherPlayers,
		ViewingPlayerID:  viewingPlayerID,
		CurrentTurn:      currentTurn,
		Generation:       game.Generation,
		TurnOrder:        game.PlayerIDs,
		Board:            ToBoardDto(board),
		PaymentConstants: paymentConstants,
	}, nil
}

// ToPlayerDto converts a model Player to PlayerDto with resolved cards
func ToPlayerDto(player player.Player, resolvedCards map[string]card.Card) (PlayerDto, error) {
	// Get resources via service
	resources, err := player.GetResources()
	if err != nil {
		return PlayerDto{}, err
	}

	// Get production via service
	production, err := player.GetProduction()
	if err != nil {
		return PlayerDto{}, err
	}

	// Get passed state via service
	passed, err := player.GetPassed()
	if err != nil {
		return PlayerDto{}, err
	}

	// Get available actions via service
	availableActions, err := player.GetAvailableActions()
	if err != nil {
		return PlayerDto{}, err
	}

	// Get production phase state via service
	productionPhaseState, err := player.GetProductionPhaseState()
	if err != nil {
		return PlayerDto{}, err
	}

	status := PlayerStatusActive
	if passed {
		status = PlayerStatusWaiting
	} else if player.SelectStartingCardsPhase != nil {
		if player.SelectStartingCardsPhase.SelectionComplete {
			status = PlayerStatusActive
		} else {
			status = PlayerStatusSelectingStartingCards
		}
	} else if productionPhaseState != nil {
		if productionPhaseState.SelectionComplete {
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

	// Get pending tile selection via service (returns tiles.PendingTileSelection)
	pendingTileSelection, err := player.GetPendingTileSelection()
	if err != nil {
		return PlayerDto{}, err
	}

	return PlayerDto{
		ID:                       player.ID,
		Name:                     player.Name,
		Status:                   status,
		Corporation:              corporationDto,
		Cards:                    resolveCards(player.Cards, resolvedCards),
		Resources:                ToResourcesDto(resources),
		Production:               ToProductionDto(production),
		TerraformRating:          player.TerraformRating,
		PlayedCards:              player.PlayedCards,
		Passed:                   passed,
		AvailableActions:         availableActions,
		VictoryPoints:            player.VictoryPoints,
		IsConnected:              player.IsConnected,
		Effects:                  ToPlayerEffectDtoSlice(player.Effects),
		Actions:                  ToPlayerActionDtoSlice(player.Actions),
		SelectStartingCardsPhase: ToSelectStartingCardsPhaseDto(player.SelectStartingCardsPhase, resolvedCards),
		ProductionPhase:          ToProductionPhaseStateDto(productionPhaseState, resolvedCards),
		StartingCards:            startingCards,
		PendingTileSelection:     ToPendingTileSelectionDto(pendingTileSelection),
		PendingCardSelection:     ToPendingCardSelectionDto(player.PendingCardSelection, resolvedCards),
		PendingCardDrawSelection: ToPendingCardDrawSelectionDto(player.PendingCardDrawSelection, resolvedCards),
		ForcedFirstAction:        ToForcedFirstActionDto(player.ForcedFirstAction),
		ResourceStorage:          player.ResourceStorage,
		PaymentSubstitutes:       ToPaymentSubstituteDtoSlice(player.PaymentSubstitutes),
	}, nil
}

// PlayerToOtherPlayerDto converts a player.Player to OtherPlayerDto (limited view)
func PlayerToOtherPlayerDto(player player.Player) (OtherPlayerDto, error) {
	// Get passed state via service
	passed, err := player.GetPassed()
	if err != nil {
		return OtherPlayerDto{}, err
	}

	// Get production phase state via service
	productionPhaseState, err := player.GetProductionPhaseState()
	if err != nil {
		return OtherPlayerDto{}, err
	}

	// Convert corporation to CardDto if present
	var corporationDto *CardDto
	if player.Corporation != nil {
		dto := ToCardDto(*player.Corporation)
		corporationDto = &dto
	}

	status := PlayerStatusActive
	if passed {
		status = PlayerStatusWaiting
	} else if player.SelectStartingCardsPhase != nil {
		if player.SelectStartingCardsPhase.SelectionComplete {
			status = PlayerStatusActive
		} else {
			status = PlayerStatusSelectingStartingCards
		}
	} else if productionPhaseState != nil {
		if productionPhaseState.SelectionComplete {
			status = PlayerStatusActive
		} else {
			status = PlayerStatusSelectingProductionCards
		}
	}

	// Get resources via service
	resources, err := player.GetResources()
	if err != nil {
		return OtherPlayerDto{}, err
	}

	// Get production via service
	production, err := player.GetProduction()
	if err != nil {
		return OtherPlayerDto{}, err
	}

	// Get available actions via service
	availableActions, err := player.GetAvailableActions()
	if err != nil {
		return OtherPlayerDto{}, err
	}

	return OtherPlayerDto{
		ID:                       player.ID,
		Name:                     player.Name,
		Status:                   status,
		Corporation:              corporationDto,
		HandCardCount:            len(player.Cards), // Hide actual cards, show count only
		Resources:                ToResourcesDto(resources),
		Production:               ToProductionDto(production),
		TerraformRating:          player.TerraformRating,
		PlayedCards:              player.PlayedCards, // Played cards are public
		Passed:                   passed,
		AvailableActions:         availableActions,
		VictoryPoints:            player.VictoryPoints,
		IsConnected:              player.IsConnected,
		Effects:                  ToPlayerEffectDtoSlice(player.Effects),
		Actions:                  ToPlayerActionDtoSlice(player.Actions),
		SelectStartingCardsPhase: ToSelectStartingCardsOtherPlayerDto(player.SelectStartingCardsPhase),
		ProductionPhase:          ToProductionPhaseStateOtherPlayerDto(productionPhaseState),
		ResourceStorage:          player.ResourceStorage, // Resource storage is public information
		PaymentSubstitutes:       ToPaymentSubstituteDtoSlice(player.PaymentSubstitutes),
	}, nil
}

// ToResourcesDto converts model Resources to ResourcesDto
func ToResourcesDto(resources resources.Resources) ResourcesDto {
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
func ToProductionDto(production resources.Production) ProductionDto {
	return ProductionDto{
		Credits:  production.Credits,
		Steel:    production.Steel,
		Titanium: production.Titanium,
		Plants:   production.Plants,
		Energy:   production.Energy,
		Heat:     production.Heat,
	}
}

// ToPaymentSubstituteDto converts model PaymentSubstitute to PaymentSubstituteDto
func ToPaymentSubstituteDto(substitute types.PaymentSubstitute) PaymentSubstituteDto {
	return PaymentSubstituteDto{
		ResourceType:   ResourceType(substitute.ResourceType),
		ConversionRate: substitute.ConversionRate,
	}
}

// ToPaymentSubstituteDtoSlice converts a slice of model PaymentSubstitute to PaymentSubstituteDto slice
func ToPaymentSubstituteDtoSlice(substitutes []types.PaymentSubstitute) []PaymentSubstituteDto {
	if substitutes == nil {
		return []PaymentSubstituteDto{}
	}

	result := make([]PaymentSubstituteDto, len(substitutes))
	for i, substitute := range substitutes {
		result[i] = ToPaymentSubstituteDto(substitute)
	}
	return result
}

// ToGlobalParametersDto converts model GlobalParameters to GlobalParametersDto
func ToGlobalParametersDto(params game.GlobalParameters) GlobalParametersDto {
	return GlobalParametersDto{
		Temperature: params.Temperature,
		Oxygen:      params.Oxygen,
		Oceans:      params.Oceans,
	}
}

// ToGameSettingsDto converts model GameSettings to GameSettingsDto
func ToGameSettingsDto(settings game.GameSettings) GameSettingsDto {
	return GameSettingsDto{
		MaxPlayers:      settings.MaxPlayers,
		DevelopmentMode: settings.DevelopmentMode,
		CardPacks:       settings.CardPacks,
	}
}

// ToGameDtoBasic provides a basic non-personalized game view (temporary compatibility)
// This is used for cases where personalization isn't needed (like game listings)
// TODO: Create a new model for this usecase. Or rename the other "Game" that contains player data,
func ToGameDtoBasic(game game.Game, paymentConstants PaymentConstantsDto) GameDto {
	// Get global parameters via service
	globalParams, _ := game.GetGlobalParameters()

	// Get current turn via service
	currentTurn, _ := game.GetCurrentTurn()

	// Get board via service
	board, _ := game.GetBoard()

	return GameDto{
		ID:               game.ID,
		Status:           GameStatus(game.Status),
		Settings:         ToGameSettingsDto(game.Settings),
		HostPlayerID:     game.HostPlayerID,
		CurrentPhase:     GamePhase(game.CurrentPhase),
		GlobalParameters: ToGlobalParametersDto(globalParams),
		CurrentPlayer:    PlayerDto{},               // Empty for non-personalized view
		OtherPlayers:     make([]OtherPlayerDto, 0), // Empty for non-personalized view
		ViewingPlayerID:  "",                        // No viewing player for basic view
		CurrentTurn:      currentTurn,
		Generation:       game.Generation,
		TurnOrder:        game.PlayerIDs,
		Board:            ToBoardDto(board),
		PaymentConstants: paymentConstants,
	}
}

// ToGameDtoSlice provides basic non-personalized game views (temporary compatibility)
func ToGameDtoSlice(games []game.Game, paymentConstants PaymentConstantsDto) []GameDto {
	dtos := make([]GameDto, len(games))
	for i, game := range games {
		dtos[i] = ToGameDtoBasic(game, paymentConstants)
	}
	return dtos
}

// ToPlayerDtoSlice converts a slice of model Players to PlayerDto slice with empty cards
func ToPlayerDtoSlice(players []player.Player) []PlayerDto {
	dtos := make([]PlayerDto, 0, len(players))
	for _, player := range players {
		dto, err := ToPlayerDto(player, nil) // Empty cards for basic conversion
		if err != nil {
			continue // Skip players with errors
		}
		dtos = append(dtos, dto)
	}
	return dtos
}

// ToCardDto converts a model Card to CardDto
func ToCardDto(card card.Card) CardDto {

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
	var startingResources *ResourceSet
	var startingProduction *ResourceSet

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
		Pack:               card.Pack,
		Tags:               ToCardTagDtoSlice(card.Tags),
		Requirements:       card.Requirements,
		Behaviors:          behaviors,
		ResourceStorage:    resourceStorage,
		VPConditions:       card.VPConditions,
		StartingResources:  startingResources,
		StartingProduction: startingProduction,
	}
}

// ToCardDtoSlice converts a slice of model Cards to CardDto slice
func ToCardDtoSlice(cards []card.Card) []CardDto {
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
func ToCardTagDtoSlice(tags []card.CardTag) []CardTag {
	if tags == nil || len(tags) == 0 {
		return nil
	}

	result := make([]CardTag, len(tags))
	for i, tag := range tags {
		result[i] = CardTag(tag)
	}
	return result
}

// ToCardTypeDtoSlice converts a slice of model CardType to CardType slice
func ToCardTypeDtoSlice(cardTypes []card.CardType) []CardType {
	if cardTypes == nil || len(cardTypes) == 0 {
		return nil
	}

	result := make([]CardType, len(cardTypes))
	for i, cardType := range cardTypes {
		result[i] = CardType(cardType)
	}
	return result
}

// ToStandardProjectDtoSlice converts a slice of model StandardProjects to StandardProject slice
func ToStandardProjectDtoSlice(projects []types.StandardProject) []StandardProject {
	if projects == nil || len(projects) == 0 {
		return nil
	}

	result := make([]StandardProject, len(projects))
	for i, project := range projects {
		result[i] = StandardProject(project)
	}
	return result
}

// ToSelectStartingCardsPhaseDto converts model SelectStartingCardsPhase to SelectStartingCardsPhaseDto
func ToSelectStartingCardsPhaseDto(phase *player.SelectStartingCardsPhase, resolvedCards map[string]card.Card) *SelectStartingCardsPhaseDto {
	if phase == nil {
		return nil
	}

	return &SelectStartingCardsPhaseDto{
		AvailableCards:        resolveCards(phase.AvailableCards, resolvedCards),
		AvailableCorporations: phase.AvailableCorporations,
		SelectionComplete:     phase.SelectionComplete,
	}
}

func ToSelectStartingCardsOtherPlayerDto(phase *player.SelectStartingCardsPhase) *SelectStartingCardsOtherPlayerDto {
	if phase == nil {
		return nil
	}

	return &SelectStartingCardsOtherPlayerDto{
		SelectionComplete: phase.SelectionComplete,
	}
}

// ToProductionPhaseDto converts model ProductionPhase to ProductionPhaseDto
func ToProductionPhaseDto(phase *production.ProductionPhaseState, resolvedCards map[string]card.Card) *ProductionPhaseDto {
	if phase == nil {
		return nil
	}

	return &ProductionPhaseDto{
		AvailableCards:    resolveCards(phase.AvailableCards, resolvedCards),
		SelectionComplete: phase.SelectionComplete,
		BeforeResources:   ResourcesDto{}, // Not available in ProductionPhaseState
		AfterResources:    ResourcesDto{}, // Not available in ProductionPhaseState
		ResourceDelta:     ResourcesDto{}, // Not available in ProductionPhaseState
		EnergyConverted:   0,   // Not available in ProductionPhaseState
		CreditsIncome:     0,   // Not available in ProductionPhaseState
	}
}

// ToProductionPhaseStateDto converts production.ProductionPhaseState to ProductionPhaseDto
// Note: This only includes card selection state, resource tracking fields are nil
func ToProductionPhaseStateDto(state *production.ProductionPhaseState, resolvedCards map[string]card.Card) *ProductionPhaseDto {
	if state == nil {
		return nil
	}

	return &ProductionPhaseDto{
		AvailableCards:    resolveCards(state.AvailableCards, resolvedCards),
		SelectionComplete: state.SelectionComplete,
		// Resource tracking fields are not available from ProductionPhaseState
		BeforeResources: ResourcesDto{},
		AfterResources:  ResourcesDto{},
		ResourceDelta:   ResourcesDto{},
		EnergyConverted: 0,
		CreditsIncome:   0,
	}
}

// ToProductionPhaseStateOtherPlayerDto converts production.ProductionPhaseState to ProductionPhaseOtherPlayerDto
func ToProductionPhaseStateOtherPlayerDto(state *production.ProductionPhaseState) *ProductionPhaseOtherPlayerDto {
	if state == nil {
		return nil
	}

	return &ProductionPhaseOtherPlayerDto{
		SelectionComplete: state.SelectionComplete,
	}
}

func ToProductionPhaseOtherPlayerDto(phase *production.ProductionPhaseState) *ProductionPhaseOtherPlayerDto {
	if phase == nil {
		return nil
	}

	return &ProductionPhaseOtherPlayerDto{
		SelectionComplete: phase.SelectionComplete,
		BeforeResources:   ResourcesDto{}, // Not available in ProductionPhaseState
		AfterResources:    ResourcesDto{}, // Not available in ProductionPhaseState
		ResourceDelta:     ResourcesDto{}, // Not available in ProductionPhaseState
		EnergyConverted:   0,   // Not available in ProductionPhaseState
		CreditsIncome:     0,   // Not available in ProductionPhaseState
	}
}

// ToResourceConditionDto converts a model ResourceCondition to ResourceConditionDto
func ToResourceConditionDto(rc card.ResourceCondition) ResourceConditionDto {
	return ResourceConditionDto{
		Type:                     ResourceType(rc.Type),
		Amount:                   rc.Amount,
		Target:                   TargetType(rc.Target),
		AffectedResources:        rc.AffectedResources,
		AffectedTags:             ToCardTagDtoSlice(rc.AffectedTags),
		AffectedStandardProjects: ToStandardProjectDtoSlice(rc.AffectedStandardProjects),
		MaxTrigger:               rc.MaxTrigger,
		Per:                      ToPerConditionDto(rc.Per),
	}
}

// ToResourceConditionDtoSlice converts a slice of model ResourceCondition to ResourceConditionDto slice
func ToResourceConditionDtoSlice(conditions []card.ResourceCondition) []ResourceConditionDto {
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
func ToPerConditionDto(per *card.PerCondition) *PerConditionDto {
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
func ToChoiceDto(choice card.Choice) ChoiceDto {
	return ChoiceDto{
		Inputs:  ToResourceConditionDtoSlice(choice.Inputs),
		Outputs: ToResourceConditionDtoSlice(choice.Outputs),
	}
}

// ToChoiceDtoSlice converts a slice of model Choice to ChoiceDto slice
func ToChoiceDtoSlice(choices []card.Choice) []ChoiceDto {
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
func ToTriggerDto(trigger card.Trigger) TriggerDto {
	return TriggerDto{
		Type:      ResourceTriggerType(trigger.Type),
		Condition: ToResourceTriggerConditionDto(trigger.Condition),
	}
}

// ToTriggerDtoSlice converts a slice of model Trigger to TriggerDto slice
func ToTriggerDtoSlice(triggers []card.Trigger) []TriggerDto {
	if triggers == nil {
		return nil
	}

	result := make([]TriggerDto, len(triggers))
	for i, trigger := range triggers {
		result[i] = ToTriggerDto(trigger)
	}
	return result
}

// ToMinMaxValueDto converts a model MinMaxValue pointer to MinMaxValueDto pointer
func ToMinMaxValueDto(value *card.MinMaxValue) *MinMaxValueDto {
	if value == nil {
		return nil
	}
	return &MinMaxValueDto{
		Min: value.Min,
		Max: value.Max,
	}
}

// ToResourceChangeMapDto converts a model RequiredResourceChange map to DTO map
func ToResourceChangeMapDto(changeMap map[types.ResourceType]card.MinMaxValue) map[ResourceType]MinMaxValueDto {
	if changeMap == nil {
		return nil
	}

	result := make(map[ResourceType]MinMaxValueDto)
	for k, v := range changeMap {
		result[ResourceType(k)] = MinMaxValueDto{
			Min: v.Min,
			Max: v.Max,
		}
	}
	return result
}

// ToResourceTriggerConditionDto converts a model ResourceTriggerCondition pointer to ResourceTriggerConditionDto pointer
func ToResourceTriggerConditionDto(condition *card.ResourceTriggerCondition) *ResourceTriggerConditionDto {
	if condition == nil {
		return nil
	}

	return &ResourceTriggerConditionDto{
		Type:                   TriggerType(condition.Type),
		Location:               ToCardApplyLocationPointer(condition.Location),
		AffectedTags:           ToCardTagDtoSlice(condition.AffectedTags),
		AffectedResources:      condition.AffectedResources,
		AffectedCardTypes:      ToCardTypeDtoSlice(condition.AffectedCardTypes),
		Target:                 ToTargetTypePointer(condition.Target),
		RequiredOriginalCost:   ToMinMaxValueDto(condition.RequiredOriginalCost),
		RequiredResourceChange: ToResourceChangeMapDto(condition.RequiredResourceChange),
	}
}

// ToCardBehaviorDto converts a model CardBehavior to CardBehaviorDto
func ToCardBehaviorDto(behavior card.CardBehavior) CardBehaviorDto {
	return CardBehaviorDto{
		Triggers: ToTriggerDtoSlice(behavior.Triggers),
		Inputs:   ToResourceConditionDtoSlice(behavior.Inputs),
		Outputs:  ToResourceConditionDtoSlice(behavior.Outputs),
		Choices:  ToChoiceDtoSlice(behavior.Choices),
	}
}

// ToCardBehaviorDtoSlice converts a slice of model CardBehaviors to CardBehaviorDto slice
func ToCardBehaviorDtoSlice(behaviors []card.CardBehavior) []CardBehaviorDto {
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
func ToCardApplyLocationPointer(ptr *card.CardApplyLocation) *CardApplyLocation {
	if ptr == nil {
		return nil
	}
	result := CardApplyLocation(*ptr)
	return &result
}

// ToTargetTypePointer converts model TargetType pointer to DTO TargetType pointer
func ToTargetTypePointer(ptr *card.TargetType) *TargetType {
	if ptr == nil {
		return nil
	}
	result := TargetType(*ptr)
	return &result
}

// ToCardTagPointer converts model CardTag pointer to DTO CardTag pointer
func ToCardTagPointer(ptr *card.CardTag) *CardTag {
	if ptr == nil {
		return nil
	}
	result := CardTag(*ptr)
	return &result
}

// ToPlayerEffectDto converts a model PlayerEffect to PlayerEffectDto
func ToPlayerEffectDto(effect player.PlayerEffect) PlayerEffectDto {
	return PlayerEffectDto{
		CardID:        effect.CardID,
		CardName:      effect.CardName,
		BehaviorIndex: effect.BehaviorIndex,
		Behavior:      ToCardBehaviorDto(effect.Behavior),
	}
}

// ToPlayerEffectDtoSlice converts a slice of model PlayerEffects to PlayerEffectDto slice
func ToPlayerEffectDtoSlice(effects []player.PlayerEffect) []PlayerEffectDto {
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
func ToPlayerActionDto(action player.PlayerAction) PlayerActionDto {
	return PlayerActionDto{
		CardID:        action.CardID,
		CardName:      action.CardName,
		BehaviorIndex: action.BehaviorIndex,
		Behavior:      ToCardBehaviorDto(action.Behavior),
		PlayCount:     action.PlayCount,
	}
}

// ToPlayerActionDtoSlice converts a slice of model PlayerActions to PlayerActionDto slice
// Filters out auto-first-action triggers that have already been used (PlayCount > 0)
func ToPlayerActionDtoSlice(actions []player.PlayerAction) []PlayerActionDto {
	if actions == nil {
		return []PlayerActionDto{}
	}
	result := make([]PlayerActionDto, 0, len(actions))
	for _, action := range actions {
		// Check if this is an auto-first-action that has already been played
		isAutoFirstAction := false
		if len(action.Behavior.Triggers) > 0 {
			isAutoFirstAction = action.Behavior.Triggers[0].Type == card.ResourceTriggerAutoFirstAction
		}

		// Skip auto-first-actions that have been used
		if isAutoFirstAction && action.PlayCount > 0 {
			continue
		}

		result = append(result, ToPlayerActionDto(action))
	}
	return result
}

// Board conversion functions

// ToBoardDto converts a model Board to BoardDto
func ToBoardDto(board tiles.Board) BoardDto {
	return BoardDto{
		Tiles: ToTileDtoSlice(board.Tiles),
	}
}

// ToTileDto converts a model Tile to TileDto
func ToTileDto(tile tiles.Tile) TileDto {
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
func ToTileDtoSlice(tiles []tiles.Tile) []TileDto {
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
func ToTileBonusDto(bonus tiles.TileBonus) TileBonusDto {
	return TileBonusDto{
		Type:   string(bonus.Type),
		Amount: bonus.Amount,
	}
}

// ToTileBonusDtoSlice converts a slice of model TileBonus to TileBonusDto slice
func ToTileBonusDtoSlice(bonuses []tiles.TileBonus) []TileBonusDto {
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
func ToTileOccupantDto(occupant *tiles.TileOccupant) *TileOccupantDto {
	if occupant == nil {
		return nil
	}

	return &TileOccupantDto{
		Type: string(occupant.Type),
		Tags: occupant.Tags,
	}
}

// ToPendingTileSelectionDto converts a model PendingTileSelection pointer to PendingTileSelectionDto pointer
// ToForcedFirstActionDto converts a model ForcedFirstAction to ForcedFirstActionDto
func ToForcedFirstActionDto(action *player.ForcedFirstAction) *ForcedFirstActionDto {
	if action == nil {
		return nil
	}

	return &ForcedFirstActionDto{
		ActionType:    action.ActionType,
		CorporationID: action.CorporationID,
		Completed:     action.Completed,
		Description:   action.Description,
	}
}

func ToPendingTileSelectionDto(selection *tiles.PendingTileSelection) *PendingTileSelectionDto {
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
func ToPendingCardSelectionDto(selection *player.PendingCardSelection, resolvedCards map[string]card.Card) *PendingCardSelectionDto {
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

// ToPendingCardDrawSelectionDto converts a model PendingCardDrawSelection to PendingCardDrawSelectionDto
func ToPendingCardDrawSelectionDto(selection *player.PendingCardDrawSelection, resolvedCards map[string]card.Card) *PendingCardDrawSelectionDto {
	if selection == nil {
		return nil
	}

	// Resolve available cards from card IDs
	availableCards := resolveCards(selection.AvailableCards, resolvedCards)

	return &PendingCardDrawSelectionDto{
		AvailableCards: availableCards,
		FreeTakeCount:  selection.FreeTakeCount,
		MaxBuyCount:    selection.MaxBuyCount,
		CardBuyCost:    selection.CardBuyCost,
		Source:         selection.Source,
	}
}

// ToResourceSetDto converts a model ResourceSet to ResourceSet
func ToResourceSetDto(rs resources.ResourceSet) ResourceSet {
	return ResourceSet{
		Credits:  rs.Credits,
		Steel:    rs.Steel,
		Titanium: rs.Titanium,
		Plants:   rs.Plants,
		Energy:   rs.Energy,
		Heat:     rs.Heat,
	}
}
