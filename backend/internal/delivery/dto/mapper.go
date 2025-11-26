package dto

import (
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/game/board"
	"terraforming-mars-backend/internal/session/game/card"
	"terraforming-mars-backend/internal/session/game/player"
	"terraforming-mars-backend/internal/session/game/player/actions"
	"terraforming-mars-backend/internal/session/game/player/effects"
	"terraforming-mars-backend/internal/session/game/player/selection"
	"terraforming-mars-backend/internal/session/types"

	"go.uber.org/zap"
)

// getPlayerIDs extracts player IDs from Game.Players map
// Note: Order is not guaranteed - TODO: implement proper turn order
func getPlayerIDs(g game.Game) []string {
	if g.Players == nil {
		return []string{}
	}

	playerIDs := make([]string, 0, len(g.Players))
	for id := range g.Players {
		playerIDs = append(playerIDs, id)
	}
	return playerIDs
}

// ToCardPayment converts a CardPaymentDto to domain model CardPayment
func ToCardPayment(dto CardPaymentDto) card.CardPayment {
	payment := card.CardPayment{
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
func ToGameDto(g game.Game, players []player.Player, viewingPlayerID string, resolvedCards map[string]card.Card, paymentConstants PaymentConstantsDto) GameDto {
	log := logger.Get().With(
		zap.String("game_id", g.ID),
		zap.String("viewing_player_id", viewingPlayerID),
		zap.Int("total_players", len(players)))

	log.Debug("🎯 ToGameDto called")

	var currentPlayer PlayerDto
	otherPlayers := make([]OtherPlayerDto, 0)

	for _, player := range players {
		log.Debug("🔍 Comparing player ID",
			zap.String("player_id", player.ID()),
			zap.String("player_name", player.Name()),
			zap.Bool("matches", player.ID() == viewingPlayerID))

		if player.ID() == viewingPlayerID {
			log.Debug("✅ Match found - setting as currentPlayer")
			currentPlayer = ToPlayerDto(g, player, resolvedCards)
		} else {
			otherPlayers = append(otherPlayers, PlayerToOtherPlayerDto(g, player))
		}
	}

	log.Debug("🎯 ToGameDto result",
		zap.String("current_player_id", currentPlayer.ID),
		zap.String("current_player_name", currentPlayer.Name),
		zap.Int("other_players_count", len(otherPlayers)))

	return GameDto{
		ID:               g.ID,
		Status:           GameStatus(g.Status),
		Settings:         ToGameSettingsDto(g.Settings),
		HostPlayerID:     g.HostPlayerID,
		CurrentPhase:     GamePhase(g.CurrentPhase),
		GlobalParameters: ToGlobalParametersDto(g.GlobalParameters),
		CurrentPlayer:    currentPlayer,
		OtherPlayers:     otherPlayers,
		ViewingPlayerID:  viewingPlayerID,
		CurrentTurn:      g.CurrentTurn,
		Generation:       g.Generation,
		TurnOrder:        getPlayerIDs(g),
		Board:            ToBoardDto(g.Board),
		PaymentConstants: paymentConstants,
	}
}

// ToPlayerDto converts a model Player to PlayerDto with resolved cards
// Requires Game to access phase state (phase state moved from Player to Game)
func ToPlayerDto(g game.Game, player player.Player, resolvedCards map[string]card.Card) PlayerDto {
	playerID := player.ID()

	// Get card selection phase state from Player (card selection phase state owned by Player)
	selectStartingCardsPhase := player.Selection().GetSelectStartingCardsPhase()
	productionPhase := player.Selection().GetProductionPhase()

	status := PlayerStatusActive
	if player.Turn().Passed() {
		status = PlayerStatusWaiting
	} else if selectStartingCardsPhase != nil {
		// Phase exists means selection is in progress (phase is set to nil when complete)
		status = PlayerStatusSelectingStartingCards
	} else if productionPhase != nil {
		if productionPhase.SelectionComplete {
			status = PlayerStatusActive
		} else {
			status = PlayerStatusSelectingProductionCards
		}
	}

	// Extract starting cards from SelectStartingCardsPhase if present
	var startingCards []CardDto
	if selectStartingCardsPhase != nil && len(selectStartingCardsPhase.AvailableCards) > 0 {
		startingCards = resolveCards(selectStartingCardsPhase.AvailableCards, resolvedCards)
	} else {
		startingCards = []CardDto{}
	}

	// Convert corporation to CardDto if present
	var corporationDto *CardDto
	if player.Corp().HasCorporation() {
		dto := ToCardDto(*player.Corp().Card())
		corporationDto = &dto
	}

	return PlayerDto{
		ID:                       player.ID(),
		Name:                     player.Name(),
		Status:                   status,
		Corporation:              corporationDto,
		Cards:                    resolveCards(player.Hand().Cards(), resolvedCards),
		Resources:                ToResourcesDto(player.Resources().Get()),
		Production:               ToProductionDto(player.Resources().Production()),
		TerraformRating:          player.Resources().TerraformRating(),
		PlayedCards:              player.Hand().PlayedCards(),
		Passed:                   player.Turn().Passed(),
		AvailableActions:         player.Turn().AvailableActions(),
		VictoryPoints:            player.Resources().VictoryPoints(),
		IsConnected:              player.Turn().IsConnected(),
		Effects:                  ToPlayerEffectDtoSlice(player.Effects().List()),
		Actions:                  ToPlayerActionDtoSlice(player.Actions().List()),
		SelectStartingCardsPhase: ToSelectStartingCardsPhaseDto(selectStartingCardsPhase, resolvedCards),
		ProductionPhase:          ToProductionPhaseDto(productionPhase, resolvedCards),
		StartingCards:            startingCards,
		PendingTileSelection:     ToPendingTileSelectionDto(g.GetPendingTileSelection(playerID)),
		PendingCardSelection:     ToPendingCardSelectionDto(player.Selection().GetPendingCardSelection(), resolvedCards),
		PendingCardDrawSelection: ToPendingCardDrawSelectionDto(player.Selection().GetPendingCardDrawSelection(), resolvedCards),
		ForcedFirstAction:        ToForcedFirstActionDto(g.GetForcedFirstAction(playerID)),
		ResourceStorage:          player.Resources().Storage(),
		PaymentSubstitutes:       ToPaymentSubstituteDtoSlice(player.Resources().PaymentSubstitutes()),
		RequirementModifiers:     ToRequirementModifierDtoSlice(player.Effects().RequirementModifiers()),
	}
}

// PlayerToOtherPlayerDto converts a player.Player to OtherPlayerDto (limited view)
// Requires Game to access phase state for tile/forced action (card selection phase on Player)
func PlayerToOtherPlayerDto(g game.Game, player player.Player) OtherPlayerDto {
	// Get card selection phase state from Player (card selection phase state owned by Player)
	selectStartingCardsPhase := player.Selection().GetSelectStartingCardsPhase()
	productionPhase := player.Selection().GetProductionPhase()

	// Convert corporation to CardDto if present
	var corporationDto *CardDto
	if player.Corp().HasCorporation() {
		dto := ToCardDto(*player.Corp().Card())
		corporationDto = &dto
	}

	status := PlayerStatusActive
	if player.Turn().Passed() {
		status = PlayerStatusWaiting
	} else if selectStartingCardsPhase != nil {
		// Phase exists means selection is in progress (phase is set to nil when complete)
		status = PlayerStatusSelectingStartingCards
	} else if productionPhase != nil {
		if productionPhase.SelectionComplete {
			status = PlayerStatusActive
		} else {
			status = PlayerStatusSelectingProductionCards
		}
	}

	return OtherPlayerDto{
		ID:                       player.ID(),
		Name:                     player.Name(),
		Status:                   status,
		Corporation:              corporationDto,
		HandCardCount:            len(player.Hand().Cards()), // Hide actual cards, show count only
		Resources:                ToResourcesDto(player.Resources().Get()),
		Production:               ToProductionDto(player.Resources().Production()),
		TerraformRating:          player.Resources().TerraformRating(),
		PlayedCards:              player.Hand().PlayedCards(), // Played cards are public
		Passed:                   player.Turn().Passed(),
		AvailableActions:         player.Turn().AvailableActions(),
		VictoryPoints:            player.Resources().VictoryPoints(),
		IsConnected:              player.Turn().IsConnected(),
		Effects:                  ToPlayerEffectDtoSlice(player.Effects().List()),
		Actions:                  ToPlayerActionDtoSlice(player.Actions().List()),
		SelectStartingCardsPhase: ToSelectStartingCardsOtherPlayerDto(selectStartingCardsPhase),
		ProductionPhase:          ToProductionPhaseOtherPlayerDto(productionPhase),
		ResourceStorage:          player.Resources().Storage(), // Resource storage is public information
		PaymentSubstitutes:       ToPaymentSubstituteDtoSlice(player.Resources().PaymentSubstitutes()),
	}
}

// ToResourcesDto converts model Resources to ResourcesDto
func ToResourcesDto(resources types.Resources) ResourcesDto {
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
func ToProductionDto(production types.Production) ProductionDto {
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
func ToPaymentSubstituteDto(substitute card.PaymentSubstitute) PaymentSubstituteDto {
	return PaymentSubstituteDto{
		ResourceType:   ResourceType(substitute.ResourceType),
		ConversionRate: substitute.ConversionRate,
	}
}

// ToPaymentSubstituteDtoSlice converts a slice of model PaymentSubstitute to PaymentSubstituteDto slice
func ToPaymentSubstituteDtoSlice(substitutes []card.PaymentSubstitute) []PaymentSubstituteDto {
	if substitutes == nil {
		return []PaymentSubstituteDto{}
	}

	result := make([]PaymentSubstituteDto, len(substitutes))
	for i, substitute := range substitutes {
		result[i] = ToPaymentSubstituteDto(substitute)
	}
	return result
}

// ToRequirementModifierDto converts model RequirementModifier to RequirementModifierDto
func ToRequirementModifierDto(modifier effects.RequirementModifier) RequirementModifierDto {
	// Convert affected resources slice
	affectedResources := make([]ResourceType, len(modifier.AffectedResources))
	for i, res := range modifier.AffectedResources {
		affectedResources[i] = ResourceType(res)
	}

	// Convert card target pointer if exists
	var cardTarget *string
	if modifier.CardTarget != nil {
		val := *modifier.CardTarget
		cardTarget = &val
	}

	// Convert standard project target pointer if exists
	var standardProjectTarget *StandardProject
	if modifier.StandardProjectTarget != nil {
		val := StandardProject(*modifier.StandardProjectTarget)
		standardProjectTarget = &val
	}

	return RequirementModifierDto{
		Amount:                modifier.Amount,
		AffectedResources:     affectedResources,
		CardTarget:            cardTarget,
		StandardProjectTarget: standardProjectTarget,
	}
}

// ToRequirementModifierDtoSlice converts a slice of model RequirementModifier to RequirementModifierDto slice
func ToRequirementModifierDtoSlice(modifiers []effects.RequirementModifier) []RequirementModifierDto {
	if modifiers == nil {
		return []RequirementModifierDto{}
	}

	result := make([]RequirementModifierDto, len(modifiers))
	for i, modifier := range modifiers {
		result[i] = ToRequirementModifierDto(modifier)
	}
	return result
}

// ToGlobalParametersDto converts model GlobalParameters to GlobalParametersDto
func ToGlobalParametersDto(params types.GlobalParameters) GlobalParametersDto {
	return GlobalParametersDto{
		Temperature: params.Temperature,
		Oxygen:      params.Oxygen,
		Oceans:      params.Oceans,
	}
}

// ToGameSettingsDto converts model GameSettings to GameSettingsDto
func ToGameSettingsDto(settings types.GameSettings) GameSettingsDto {
	return GameSettingsDto{
		MaxPlayers:      settings.MaxPlayers,
		DevelopmentMode: settings.DevelopmentMode,
		CardPacks:       settings.CardPacks,
	}
}

// ToGameDtoBasic provides a basic non-personalized game view (temporary compatibility)
// This is used for cases where personalization isn't needed (like game listings)
// TODO: Architectural cleanup needed:
// Option 1: Create separate DTOs - GameListingDto (basic) and GameDetailDto (personalized)
// Option 2: Rename existing PersonalizedGameDto and use GameDto only for basic views
// Current approach reuses GameDto with empty player fields, which is suboptimal
func ToGameDtoBasic(g game.Game, paymentConstants PaymentConstantsDto) GameDto {
	return GameDto{
		ID:               g.ID,
		Status:           GameStatus(g.Status),
		Settings:         ToGameSettingsDto(g.Settings),
		HostPlayerID:     g.HostPlayerID,
		CurrentPhase:     GamePhase(g.CurrentPhase),
		GlobalParameters: ToGlobalParametersDto(g.GlobalParameters),
		CurrentPlayer:    PlayerDto{},               // Empty for non-personalized view
		OtherPlayers:     make([]OtherPlayerDto, 0), // Empty for non-personalized view
		ViewingPlayerID:  "",                        // No viewing player for basic view
		CurrentTurn:      g.CurrentTurn,
		Generation:       g.Generation,
		TurnOrder:        getPlayerIDs(g),
		Board:            ToBoardDto(g.Board),
		PaymentConstants: paymentConstants,
	}
}

// ToGameDtoSlice provides basic non-personalized game views (temporary compatibility)
func ToGameDtoSlice(games []game.Game, paymentConstants PaymentConstantsDto) []GameDto {
	dtos := make([]GameDto, len(games))
	for i, g := range games {
		dtos[i] = ToGameDtoBasic(g, paymentConstants)
	}
	return dtos
}

// ToPlayerDtoSlice removed - function not used anywhere and requires Game parameter
// If needed in future, signature should be: ToPlayerDtoSlice(g game.Game, players []player.Player) []PlayerDto

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
func ToCardTagDtoSlice(tags []types.CardTag) []CardTag {
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
func ToSelectStartingCardsPhaseDto(phase *selection.SelectStartingCardsPhase, resolvedCards map[string]card.Card) *SelectStartingCardsPhaseDto {
	if phase == nil {
		return nil
	}

	return &SelectStartingCardsPhaseDto{
		AvailableCards:        resolveCards(phase.AvailableCards, resolvedCards),
		AvailableCorporations: phase.AvailableCorporations,
	}
}

func ToSelectStartingCardsOtherPlayerDto(phase *selection.SelectStartingCardsPhase) *SelectStartingCardsOtherPlayerDto {
	if phase == nil {
		return nil
	}

	// Empty struct - other players don't see selection details
	return &SelectStartingCardsOtherPlayerDto{}
}

// ToProductionPhaseDto converts model ProductionPhase to ProductionPhaseDto
func ToProductionPhaseDto(phase *selection.ProductionPhase, resolvedCards map[string]card.Card) *ProductionPhaseDto {
	if phase == nil {
		return nil
	}

	delta := types.Resources{
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

func ToProductionPhaseOtherPlayerDto(phase *selection.ProductionPhase) *ProductionPhaseOtherPlayerDto {
	if phase == nil {
		return nil
	}

	delta := types.Resources{
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
func ToCardTagPointer(ptr *types.CardTag) *CardTag {
	if ptr == nil {
		return nil
	}
	result := CardTag(*ptr)
	return &result
}

// ToPlayerEffectDto converts a model PlayerEffect to PlayerEffectDto
func ToPlayerEffectDto(effect card.PlayerEffect) PlayerEffectDto {
	return PlayerEffectDto{
		CardID:        effect.CardID,
		CardName:      effect.CardName,
		BehaviorIndex: effect.BehaviorIndex,
		Behavior:      ToCardBehaviorDto(effect.Behavior),
	}
}

// ToPlayerEffectDtoSlice converts a slice of model PlayerEffects to PlayerEffectDto slice
func ToPlayerEffectDtoSlice(effects []card.PlayerEffect) []PlayerEffectDto {
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
func ToPlayerActionDto(action actions.PlayerAction) PlayerActionDto {
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
func ToPlayerActionDtoSlice(actions []actions.PlayerAction) []PlayerActionDto {
	if actions == nil {
		return []PlayerActionDto{}
	}
	result := make([]PlayerActionDto, 0, len(actions))
	for _, action := range actions {
		// Check if this is an auto-first-action that has already been played
		isAutoFirstAction := false
		if len(action.Behavior.Triggers) > 0 {
			isAutoFirstAction = action.Behavior.Triggers[0].Type == card.ResourceTriggerAutoCorporationFirstAction
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
func ToBoardDto(b board.Board) BoardDto {
	return BoardDto{
		Tiles: ToTileDtoSlice(b.Tiles),
	}
}

// ToTileDto converts a model Tile to TileDto
func ToTileDto(tile board.Tile) TileDto {
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
func ToTileDtoSlice(tiles []board.Tile) []TileDto {
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
func ToTileBonusDto(bonus board.TileBonus) TileBonusDto {
	return TileBonusDto{
		Type:   string(bonus.Type),
		Amount: bonus.Amount,
	}
}

// ToTileBonusDtoSlice converts a slice of model TileBonus to TileBonusDto slice
func ToTileBonusDtoSlice(bonuses []board.TileBonus) []TileBonusDto {
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
func ToTileOccupantDto(occupant *board.TileOccupant) *TileOccupantDto {
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

func ToPendingTileSelectionDto(selection *player.PendingTileSelection) *PendingTileSelectionDto {
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
func ToPendingCardSelectionDto(selection *selection.PendingCardSelection, resolvedCards map[string]card.Card) *PendingCardSelectionDto {
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
func ToPendingCardDrawSelectionDto(selection *selection.PendingCardDrawSelection, resolvedCards map[string]card.Card) *PendingCardDrawSelectionDto {
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
func ToResourceSetDto(rs types.ResourceSet) ResourceSet {
	return ResourceSet{
		Credits:  rs.Credits,
		Steel:    rs.Steel,
		Titanium: rs.Titanium,
		Plants:   rs.Plants,
		Energy:   rs.Energy,
		Heat:     rs.Heat,
	}
}
