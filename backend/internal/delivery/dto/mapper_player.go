package dto

import (
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

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

	// Get hand cards with playability state (Player-Scoped Card Architecture)
	handCards := mapPlayerCards(p)

	// Get standard projects with availability state (Player-Scoped Card Architecture)
	standardProjects := mapPlayerStandardProjects(p)

	// Only include turn-specific data if it's this player's turn
	var pendingTileSelection *PendingTileSelectionDto
	var forcedFirstAction *ForcedFirstActionDto
	currentTurn := g.CurrentTurn()
	if currentTurn != nil && currentTurn.PlayerID() == p.ID() {
		pendingTileSelection = convertPendingTileSelection(g.GetPendingTileSelection(p.ID()))
		forcedFirstAction = convertForcedFirstAction(g.GetForcedFirstAction(p.ID()))
	}

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
		Cards:            handCards, // PlayerCardDto[] with state
		PlayedCards:      playedCards,
		Passed:           p.HasPassed(),
		AvailableActions: getAvailableActionsForPlayer(g, p.ID()),
		IsConnected:      p.IsConnected(),
		Effects:          convertPlayerEffects(p.Effects().List()),
		Actions:          convertPlayerActions(p.Actions().List()),
		StandardProjects: standardProjects, // PlayerStandardProjectDto[] with state

		SelectStartingCardsPhase: convertSelectStartingCardsPhase(g.GetSelectStartingCardsPhase(p.ID()), cardRegistry),
		ProductionPhase:          convertProductionPhase(g.GetProductionPhase(p.ID()), cardRegistry),
		StartingCards:            []CardDto{},
		PendingTileSelection:     pendingTileSelection,
		PendingCardSelection:     convertPendingCardSelection(p.Selection().GetPendingCardSelection(), cardRegistry),
		PendingCardDrawSelection: convertPendingCardDrawSelection(p.Selection().GetPendingCardDrawSelection(), cardRegistry),
		ForcedFirstAction:        forcedFirstAction,
		ResourceStorage:          p.Resources().Storage(),
		PaymentSubstitutes:       convertPaymentSubstitutes(p.Resources().PaymentSubstitutes()),
		RequirementModifiers:     convertRequirementModifiers(p.Effects().RequirementModifiers()),
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
		Passed:           p.HasPassed(),
		AvailableActions: getAvailableActionsForPlayer(g, p.ID()),
		IsConnected:      p.IsConnected(),
		Effects:          convertPlayerEffects(p.Effects().List()),
		Actions:          convertPlayerActions(p.Actions().List()),

		SelectStartingCardsPhase: convertSelectStartingCardsPhaseForOtherPlayer(g.GetSelectStartingCardsPhase(p.ID())),
		ProductionPhase:          convertProductionPhaseForOtherPlayer(g.GetProductionPhase(p.ID())),
		ResourceStorage:          p.Resources().Storage(),
		PaymentSubstitutes:       convertPaymentSubstitutes(p.Resources().PaymentSubstitutes()),
	}
}

// convertSelectStartingCardsPhase converts SelectStartingCardsPhase to DTO
func convertSelectStartingCardsPhase(phase *player.SelectStartingCardsPhase, cardRegistry cards.CardRegistry) *SelectStartingCardsPhaseDto {
	if phase == nil {
		return nil
	}

	// Get full card details for available cards
	availableCards := getPlayedCards(phase.AvailableCards, cardRegistry)

	// Get full card details for available corporations
	availableCorporations := getPlayedCards(phase.AvailableCorporations, cardRegistry)

	result := &SelectStartingCardsPhaseDto{
		AvailableCards:        availableCards,
		AvailableCorporations: availableCorporations,
	}

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

// convertProductionPhase converts production phase data to DTO for current player
func convertProductionPhase(phase *player.ProductionPhase, cardRegistry cards.CardRegistry) *ProductionPhaseDto {
	if phase == nil {
		return nil
	}

	// Get full card details for available cards
	availableCards := getPlayedCards(phase.AvailableCards, cardRegistry)

	// Convert resources
	beforeResources := ResourcesDto{
		Credits:  phase.BeforeResources.Credits,
		Steel:    phase.BeforeResources.Steel,
		Titanium: phase.BeforeResources.Titanium,
		Plants:   phase.BeforeResources.Plants,
		Energy:   phase.BeforeResources.Energy,
		Heat:     phase.BeforeResources.Heat,
	}

	afterResources := ResourcesDto{
		Credits:  phase.AfterResources.Credits,
		Steel:    phase.AfterResources.Steel,
		Titanium: phase.AfterResources.Titanium,
		Plants:   phase.AfterResources.Plants,
		Energy:   phase.AfterResources.Energy,
		Heat:     phase.AfterResources.Heat,
	}

	// Calculate resource delta
	resourceDelta := ResourcesDto{
		Credits:  phase.AfterResources.Credits - phase.BeforeResources.Credits,
		Steel:    phase.AfterResources.Steel - phase.BeforeResources.Steel,
		Titanium: phase.AfterResources.Titanium - phase.BeforeResources.Titanium,
		Plants:   phase.AfterResources.Plants - phase.BeforeResources.Plants,
		Energy:   phase.AfterResources.Energy - phase.BeforeResources.Energy,
		Heat:     phase.AfterResources.Heat - phase.BeforeResources.Heat,
	}

	return &ProductionPhaseDto{
		AvailableCards:    availableCards,
		SelectionComplete: phase.SelectionComplete,
		BeforeResources:   beforeResources,
		AfterResources:    afterResources,
		ResourceDelta:     resourceDelta,
		EnergyConverted:   phase.EnergyConverted,
		CreditsIncome:     phase.CreditsIncome,
	}
}

// convertProductionPhaseForOtherPlayer converts production phase data to DTO for other players
func convertProductionPhaseForOtherPlayer(phase *player.ProductionPhase) *ProductionPhaseOtherPlayerDto {
	if phase == nil {
		return nil
	}

	// Convert resources
	beforeResources := ResourcesDto{
		Credits:  phase.BeforeResources.Credits,
		Steel:    phase.BeforeResources.Steel,
		Titanium: phase.BeforeResources.Titanium,
		Plants:   phase.BeforeResources.Plants,
		Energy:   phase.BeforeResources.Energy,
		Heat:     phase.BeforeResources.Heat,
	}

	afterResources := ResourcesDto{
		Credits:  phase.AfterResources.Credits,
		Steel:    phase.AfterResources.Steel,
		Titanium: phase.AfterResources.Titanium,
		Plants:   phase.AfterResources.Plants,
		Energy:   phase.AfterResources.Energy,
		Heat:     phase.AfterResources.Heat,
	}

	// Calculate resource delta
	resourceDelta := ResourcesDto{
		Credits:  phase.AfterResources.Credits - phase.BeforeResources.Credits,
		Steel:    phase.AfterResources.Steel - phase.BeforeResources.Steel,
		Titanium: phase.AfterResources.Titanium - phase.BeforeResources.Titanium,
		Plants:   phase.AfterResources.Plants - phase.BeforeResources.Plants,
		Energy:   phase.AfterResources.Energy - phase.BeforeResources.Energy,
		Heat:     phase.AfterResources.Heat - phase.BeforeResources.Heat,
	}

	// Other players don't see available cards
	return &ProductionPhaseOtherPlayerDto{
		SelectionComplete: phase.SelectionComplete,
		BeforeResources:   beforeResources,
		AfterResources:    afterResources,
		ResourceDelta:     resourceDelta,
		EnergyConverted:   phase.EnergyConverted,
		CreditsIncome:     phase.CreditsIncome,
	}
}

// convertPlayerEffects converts CardEffect slice to PlayerEffectDto slice
func convertPlayerEffects(effects []player.CardEffect) []PlayerEffectDto {
	if len(effects) == 0 {
		return []PlayerEffectDto{}
	}

	dtos := make([]PlayerEffectDto, len(effects))
	for i, effect := range effects {
		dtos[i] = PlayerEffectDto{
			CardID:        effect.CardID,
			CardName:      effect.CardName,
			BehaviorIndex: effect.BehaviorIndex,
			Behavior:      toCardBehaviorDto(effect.Behavior),
		}
	}
	return dtos
}

// convertPlayerActions converts CardAction slice to PlayerActionDto slice
func convertPlayerActions(actions []player.CardAction) []PlayerActionDto {
	if len(actions) == 0 {
		return []PlayerActionDto{}
	}

	dtos := make([]PlayerActionDto, len(actions))
	for i, action := range actions {
		dtos[i] = PlayerActionDto{
			CardID:        action.CardID,
			CardName:      action.CardName,
			BehaviorIndex: action.BehaviorIndex,
			Behavior:      toCardBehaviorDto(action.Behavior),
			PlayCount:     action.PlayCount,
		}
	}
	return dtos
}

// convertPaymentSubstitutes converts PaymentSubstitute slice to PaymentSubstituteDto slice
func convertPaymentSubstitutes(substitutes []shared.PaymentSubstitute) []PaymentSubstituteDto {
	if len(substitutes) == 0 {
		return []PaymentSubstituteDto{}
	}

	dtos := make([]PaymentSubstituteDto, len(substitutes))
	for i, sub := range substitutes {
		dtos[i] = PaymentSubstituteDto{
			ResourceType:   ResourceType(sub.ResourceType),
			ConversionRate: sub.ConversionRate,
		}
	}
	return dtos
}

// convertRequirementModifiers converts RequirementModifier slice to RequirementModifierDto slice
func convertRequirementModifiers(modifiers []shared.RequirementModifier) []RequirementModifierDto {
	if len(modifiers) == 0 {
		return []RequirementModifierDto{}
	}

	dtos := make([]RequirementModifierDto, len(modifiers))
	for i, mod := range modifiers {
		// Convert resource types
		affectedResources := make([]ResourceType, len(mod.AffectedResources))
		for j, res := range mod.AffectedResources {
			affectedResources[j] = ResourceType(res)
		}

		// Convert standard project pointer
		var standardProjectTarget *StandardProject
		if mod.StandardProjectTarget != nil {
			sp := StandardProject(*mod.StandardProjectTarget)
			standardProjectTarget = &sp
		}

		dtos[i] = RequirementModifierDto{
			Amount:                mod.Amount,
			AffectedResources:     affectedResources,
			CardTarget:            mod.CardTarget,
			StandardProjectTarget: standardProjectTarget,
		}
	}
	return dtos
}

// convertPendingCardSelection converts PendingCardSelection to DTO
func convertPendingCardSelection(selection *player.PendingCardSelection, cardRegistry cards.CardRegistry) *PendingCardSelectionDto {
	if selection == nil {
		return nil
	}

	// Convert card IDs to full CardDtos
	availableCards := getPlayedCards(selection.AvailableCards, cardRegistry)

	return &PendingCardSelectionDto{
		AvailableCards: availableCards,
		CardCosts:      selection.CardCosts,
		CardRewards:    selection.CardRewards,
		Source:         selection.Source,
		MinCards:       selection.MinCards,
		MaxCards:       selection.MaxCards,
	}
}

// convertPendingCardDrawSelection converts PendingCardDrawSelection to DTO
func convertPendingCardDrawSelection(selection *player.PendingCardDrawSelection, cardRegistry cards.CardRegistry) *PendingCardDrawSelectionDto {
	if selection == nil {
		return nil
	}

	// Convert card IDs to full CardDtos
	availableCards := getPlayedCards(selection.AvailableCards, cardRegistry)

	return &PendingCardDrawSelectionDto{
		AvailableCards: availableCards,
		FreeTakeCount:  selection.FreeTakeCount,
		MaxBuyCount:    selection.MaxBuyCount,
		CardBuyCost:    selection.CardBuyCost,
		Source:         selection.Source,
	}
}

// convertForcedFirstAction converts ForcedFirstAction to DTO
func convertForcedFirstAction(action *player.ForcedFirstAction) *ForcedFirstActionDto {
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

// convertPendingTileSelection converts PendingTileSelection to DTO
func convertPendingTileSelection(selection *player.PendingTileSelection) *PendingTileSelectionDto {
	if selection == nil {
		return nil
	}

	return &PendingTileSelectionDto{
		TileType:       selection.TileType,
		AvailableHexes: selection.AvailableHexes,
		Source:         selection.Source,
	}
}

// getAvailableActionsForPlayer returns the available actions for a player
// Actions are now at game level, so only the current player has actions
func getAvailableActionsForPlayer(g *game.Game, playerID string) int {
	currentTurn := g.CurrentTurn()
	if currentTurn == nil {
		return 0 // No current turn set
	}

	// Only return actions if this is the current player
	if currentTurn.PlayerID() == playerID {
		return currentTurn.ActionsRemaining()
	}

	// Other players don't have actions (actions are per-turn, not per-player)
	return 0
}

// ========================================
// Player-Scoped Card Architecture Mappers
// ========================================

// convertStateErrors converts EntityState errors to DTOs
func convertStateErrors(errors []player.StateError) []StateErrorDto {
	result := make([]StateErrorDto, len(errors))
	for i, err := range errors {
		result[i] = StateErrorDto{
			Code:     err.Code,
			Category: err.Category,
			Message:  err.Message,
		}
	}
	return result
}

// ToPlayerCardDto converts a PlayerCard to PlayerCardDto with state information
func ToPlayerCardDto(pc *player.PlayerCard) PlayerCardDto {
	state := pc.State()

	// Type assert card from any to *gamecards.Card
	cardAny := pc.Card()
	card, ok := cardAny.(*gamecards.Card)
	if !ok {
		// Defensive: return empty DTO if card type is wrong (should not happen)
		return PlayerCardDto{
			Available: false,
			Errors:    []StateErrorDto{{Code: "INVALID_CARD_TYPE", Category: "internal", Message: "Invalid card type"}},
		}
	}

	// Extract discounts from metadata
	discounts := make(map[string]int)
	if discountData, ok := state.Metadata["discounts"].(map[shared.CardTag]int); ok {
		for tag, amount := range discountData {
			discounts[string(tag)] = amount
		}
	}

	// Convert tags
	tags := make([]CardTag, len(card.Tags))
	for i, tag := range card.Tags {
		tags[i] = CardTag(tag)
	}

	effectiveCost := 0
	if state.Cost != nil {
		effectiveCost = *state.Cost
	}

	return PlayerCardDto{
		ID:            card.ID,
		Name:          card.Name,
		Type:          string(card.Type),
		Cost:          card.Cost,
		Description:   card.Description,
		Tags:          tags,
		Available:     state.Available(),
		Errors:        convertStateErrors(state.Errors),
		EffectiveCost: effectiveCost,
		Discounts:     discounts,
	}
}

// mapPlayerCards converts cached PlayerCard instances from hand to DTOs
func mapPlayerCards(p *player.Player) []PlayerCardDto {
	handCardIDs := p.Hand().Cards()
	result := make([]PlayerCardDto, 0, len(handCardIDs))

	for _, cardID := range handCardIDs {
		// Get cached PlayerCard from hand
		pc, exists := p.Hand().GetPlayerCard(cardID)
		if !exists {
			// PlayerCard not cached - skip (should not happen if architecture is working correctly)
			continue
		}

		result = append(result, ToPlayerCardDto(pc))
	}

	return result
}

// mapPlayerStandardProjects converts PlayerStandardProject instances to DTOs
// TODO: Implement when standard projects are integrated into Player
func mapPlayerStandardProjects(p *player.Player) []PlayerStandardProjectDto {
	// Placeholder - will be implemented in Phase 7
	return []PlayerStandardProjectDto{}
}
