package dto

import (
	"fmt"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/board"
	"terraforming-mars-backend/internal/game/cardtypes"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/internal/logger"
)

// ToGameDto converts migration Game to GameDto with personalized view
// The playerID parameter determines which player is "currentPlayer" vs "otherPlayers"
func ToGameDto(g *game.Game, cardRegistry cards.CardRegistry, playerID string) GameDto {
	players := g.GetAllPlayers()

	// Create personalized view: viewing player is currentPlayer, others are otherPlayers
	var currentPlayer PlayerDto
	otherPlayers := make([]OtherPlayerDto, 0)

	var viewingPlayer *player.Player
	for _, p := range players {
		if p.ID() == playerID {
			viewingPlayer = p
			currentPlayer = ToPlayerDto(p, g, cardRegistry)
		} else {
			otherPlayers = append(otherPlayers, ToOtherPlayerDto(p, g, cardRegistry))
		}
	}

	// If viewing player not found, use first player as fallback
	if viewingPlayer == nil && len(players) > 0 {
		currentPlayer = ToPlayerDto(players[0], g, cardRegistry)
		for i := 1; i < len(players); i++ {
			otherPlayers = append(otherPlayers, ToOtherPlayerDto(players[i], g, cardRegistry))
		}
		playerID = players[0].ID()
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
			Type:        string(tile.Type),
			OwnerID:     tile.OwnerID,
			Tags:        tile.Tags,
			Bonuses:     convertTileBonuses(tile.Bonuses),
			Location:    string(tile.Location),
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
		ViewingPlayerID:  playerID, // The player viewing this game state
		CurrentTurn:      getCurrentTurnPlayerID(g),
		Generation:       g.Generation(),
		TurnOrder:        g.TurnOrder(),
		Board: BoardDto{
			Tiles: tileDtos,
		},
		PaymentConstants: paymentConstants,
	}
}

// getCurrentTurnPlayerID extracts the player ID from the current turn
func getCurrentTurnPlayerID(g *game.Game) *string {
	turn := g.CurrentTurn()
	if turn == nil {
		return nil
	}
	playerID := turn.PlayerID()
	return &playerID
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
func ToCardDto(card cardtypes.Card) CardDto {
	// Convert tags
	tags := make([]CardTag, len(card.Tags))
	for i, tag := range card.Tags {
		tags[i] = CardTag(tag)
	}

	// Convert requirements
	var requirements []RequirementDto
	if len(card.Requirements) > 0 {
		requirements = make([]RequirementDto, len(card.Requirements))
		for i, req := range card.Requirements {
			requirements[i] = toRequirementDto(req)
		}
	}

	// Convert behaviors
	var behaviors []CardBehaviorDto
	if len(card.Behaviors) > 0 {
		behaviors = make([]CardBehaviorDto, len(card.Behaviors))
		for i, behavior := range card.Behaviors {
			behaviors[i] = toCardBehaviorDto(behavior)
		}
	}

	// Convert resource storage
	var resourceStorage *ResourceStorageDto
	if card.ResourceStorage != nil {
		storage := toResourceStorageDto(*card.ResourceStorage)
		resourceStorage = &storage
	}

	// Convert VP conditions
	var vpConditions []VPConditionDto
	if len(card.VPConditions) > 0 {
		vpConditions = make([]VPConditionDto, len(card.VPConditions))
		for i, vp := range card.VPConditions {
			vpConditions[i] = toVPConditionDto(vp)
		}
	}

	return CardDto{
		ID:              card.ID,
		Name:            card.Name,
		Type:            CardType(card.Type),
		Cost:            card.Cost,
		Description:     card.Description,
		Pack:            card.Pack,
		Tags:            tags,
		Requirements:    requirements,
		Behaviors:       behaviors,
		ResourceStorage: resourceStorage,
		VPConditions:    vpConditions,
	}
}

// Helper functions for nested DTO conversions

func toRequirementDto(req cardtypes.Requirement) RequirementDto {
	var location *CardApplyLocation
	if req.Location != nil {
		loc := CardApplyLocation(*req.Location)
		location = &loc
	}

	var tag *CardTag
	if req.Tag != nil {
		t := CardTag(*req.Tag)
		tag = &t
	}

	var resource *ResourceType
	if req.Resource != nil {
		r := ResourceType(*req.Resource)
		resource = &r
	}

	return RequirementDto{
		Type:     RequirementType(req.Type),
		Min:      req.Min,
		Max:      req.Max,
		Location: location,
		Tag:      tag,
		Resource: resource,
	}
}

func toCardBehaviorDto(behavior cardtypes.CardBehavior) CardBehaviorDto {
	var triggers []TriggerDto
	if len(behavior.Triggers) > 0 {
		triggers = make([]TriggerDto, len(behavior.Triggers))
		for i, trigger := range behavior.Triggers {
			triggers[i] = toTriggerDto(trigger)
		}
	}

	var inputs []ResourceConditionDto
	if len(behavior.Inputs) > 0 {
		inputs = make([]ResourceConditionDto, len(behavior.Inputs))
		for i, input := range behavior.Inputs {
			inputs[i] = toResourceConditionDto(input)
		}
	}

	var outputs []ResourceConditionDto
	if len(behavior.Outputs) > 0 {
		outputs = make([]ResourceConditionDto, len(behavior.Outputs))
		for i, output := range behavior.Outputs {
			outputs[i] = toResourceConditionDto(output)
		}
	}

	var choices []ChoiceDto
	if len(behavior.Choices) > 0 {
		choices = make([]ChoiceDto, len(behavior.Choices))
		for i, choice := range behavior.Choices {
			choices[i] = toChoiceDto(choice)
		}
	}

	return CardBehaviorDto{
		Triggers: triggers,
		Inputs:   inputs,
		Outputs:  outputs,
		Choices:  choices,
	}
}

func toTriggerDto(trigger cardtypes.Trigger) TriggerDto {
	var condition *ResourceTriggerConditionDto
	if trigger.Condition != nil {
		cond := toResourceTriggerConditionDto(*trigger.Condition)
		condition = &cond
	}

	return TriggerDto{
		Type:      ResourceTriggerType(trigger.Type),
		Condition: condition,
	}
}

func toResourceTriggerConditionDto(cond cardtypes.ResourceTriggerCondition) ResourceTriggerConditionDto {
	var location *CardApplyLocation
	if cond.Location != nil {
		loc := CardApplyLocation(*cond.Location)
		location = &loc
	}

	var affectedTags []CardTag
	if len(cond.AffectedTags) > 0 {
		affectedTags = make([]CardTag, len(cond.AffectedTags))
		for i, tag := range cond.AffectedTags {
			affectedTags[i] = CardTag(tag)
		}
	}

	var affectedCardTypes []CardType
	if len(cond.AffectedCardTypes) > 0 {
		affectedCardTypes = make([]CardType, len(cond.AffectedCardTypes))
		for i, ct := range cond.AffectedCardTypes {
			affectedCardTypes[i] = CardType(ct)
		}
	}

	var target *TargetType
	if cond.Target != nil {
		t := TargetType(*cond.Target)
		target = &t
	}

	var requiredOriginalCost *MinMaxValueDto
	if cond.RequiredOriginalCost != nil {
		cost := MinMaxValueDto{
			Min: cond.RequiredOriginalCost.Min,
			Max: cond.RequiredOriginalCost.Max,
		}
		requiredOriginalCost = &cost
	}

	var requiredResourceChange map[ResourceType]MinMaxValueDto
	if len(cond.RequiredResourceChange) > 0 {
		requiredResourceChange = make(map[ResourceType]MinMaxValueDto)
		for k, v := range cond.RequiredResourceChange {
			requiredResourceChange[ResourceType(k)] = MinMaxValueDto{
				Min: v.Min,
				Max: v.Max,
			}
		}
	}

	return ResourceTriggerConditionDto{
		Type:                   TriggerType(cond.Type),
		Location:               location,
		AffectedTags:           affectedTags,
		AffectedResources:      cond.AffectedResources,
		AffectedCardTypes:      affectedCardTypes,
		Target:                 target,
		RequiredOriginalCost:   requiredOriginalCost,
		RequiredResourceChange: requiredResourceChange,
	}
}

func toResourceConditionDto(rc cardtypes.ResourceCondition) ResourceConditionDto {
	var affectedTags []CardTag
	if len(rc.AffectedTags) > 0 {
		affectedTags = make([]CardTag, len(rc.AffectedTags))
		for i, tag := range rc.AffectedTags {
			affectedTags[i] = CardTag(tag)
		}
	}

	var affectedCardTypes []CardType
	if len(rc.AffectedCardTypes) > 0 {
		affectedCardTypes = make([]CardType, len(rc.AffectedCardTypes))
		for i, ct := range rc.AffectedCardTypes {
			affectedCardTypes[i] = CardType(ct)
		}
	}

	var affectedStandardProjects []StandardProject
	if len(rc.AffectedStandardProjects) > 0 {
		affectedStandardProjects = make([]StandardProject, len(rc.AffectedStandardProjects))
		for i, sp := range rc.AffectedStandardProjects {
			affectedStandardProjects[i] = StandardProject(sp)
		}
	}

	var per *PerConditionDto
	if rc.Per != nil {
		p := toPerConditionDto(*rc.Per)
		per = &p
	}

	return ResourceConditionDto{
		Type:                     ResourceType(rc.Type),
		Amount:                   rc.Amount,
		Target:                   TargetType(rc.Target),
		AffectedResources:        rc.AffectedResources,
		AffectedTags:             affectedTags,
		AffectedStandardProjects: affectedStandardProjects,
		MaxTrigger:               rc.MaxTrigger,
		Per:                      per,
	}
}

func toChoiceDto(choice cardtypes.Choice) ChoiceDto {
	var inputs []ResourceConditionDto
	if len(choice.Inputs) > 0 {
		inputs = make([]ResourceConditionDto, len(choice.Inputs))
		for i, input := range choice.Inputs {
			inputs[i] = toResourceConditionDto(input)
		}
	}

	var outputs []ResourceConditionDto
	if len(choice.Outputs) > 0 {
		outputs = make([]ResourceConditionDto, len(choice.Outputs))
		for i, output := range choice.Outputs {
			outputs[i] = toResourceConditionDto(output)
		}
	}

	return ChoiceDto{
		Inputs:  inputs,
		Outputs: outputs,
	}
}

func toPerConditionDto(pc cardtypes.PerCondition) PerConditionDto {
	var location *CardApplyLocation
	if pc.Location != nil {
		loc := CardApplyLocation(*pc.Location)
		location = &loc
	}

	var target *TargetType
	if pc.Target != nil {
		t := TargetType(*pc.Target)
		target = &t
	}

	var tag *CardTag
	if pc.Tag != nil {
		t := CardTag(*pc.Tag)
		tag = &t
	}

	return PerConditionDto{
		Type:     ResourceType(pc.Type),
		Amount:   pc.Amount,
		Location: location,
		Target:   target,
		Tag:      tag,
	}
}

func toResourceStorageDto(storage cardtypes.ResourceStorage) ResourceStorageDto {
	return ResourceStorageDto{
		Type:     ResourceType(storage.Type),
		Capacity: storage.Capacity,
		Starting: storage.Starting,
	}
}

func toVPConditionDto(vp cardtypes.VictoryPointCondition) VPConditionDto {
	var per *PerConditionDto
	if vp.Per != nil {
		p := toPerConditionDto(*vp.Per)
		per = &p
	}

	return VPConditionDto{
		Amount:     vp.Amount,
		Condition:  VPConditionType(vp.Condition),
		MaxTrigger: vp.MaxTrigger,
		Per:        per,
	}
}

// getCorporationCard fetches the corporation card for a player using the card registry
func getCorporationCard(p *player.Player, cardRegistry cards.CardRegistry) *CardDto {
	if p.CorporationID() == "" {
		return nil
	}

	card, err := cardRegistry.GetByID(p.CorporationID())
	if err != nil {
		log := logger.Get()
		log.Warn("Failed to fetch corporation card",
			zap.String("player_id", p.ID()),
			zap.String("corporation_id", p.CorporationID()),
			zap.Error(err))
		return nil
	}

	cardDto := ToCardDto(*card)
	return &cardDto
}

// getPlayedCards converts a slice of card IDs to CardDto objects using the card registry
func getPlayedCards(cardIDs []string, cardRegistry cards.CardRegistry) []CardDto {
	cardDtos := make([]CardDto, 0, len(cardIDs))
	log := logger.Get()

	for _, cardID := range cardIDs {
		card, err := cardRegistry.GetByID(cardID)
		if err != nil {
			log.Warn("Failed to fetch played card",
				zap.String("card_id", cardID),
				zap.Error(err))
			continue // Skip cards that can't be found
		}
		cardDtos = append(cardDtos, ToCardDto(*card))
	}

	return cardDtos
}

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

	// Get hand cards with full card details
	handCardIDs := p.Hand().Cards()
	handCards := getPlayedCards(handCardIDs, cardRegistry)

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
		Cards:            handCards,
		PlayedCards:      playedCards,
		Passed:           p.HasPassed(),
		AvailableActions: getAvailableActionsForPlayer(g, p.ID()),
		IsConnected:      p.IsConnected(),
		Effects:          convertPlayerEffects(p.Effects().List()),
		Actions:          convertPlayerActions(p.Actions().List()),

		SelectStartingCardsPhase: convertSelectStartingCardsPhase(g.GetSelectStartingCardsPhase(p.ID()), cardRegistry),
		ProductionPhase:          convertProductionPhase(g.GetProductionPhase(p.ID()), cardRegistry),
		StartingCards:            []CardDto{},
		PendingTileSelection:     convertPendingTileSelection(g.GetPendingTileSelection(p.ID())),
		PendingCardSelection:     convertPendingCardSelection(p.Selection().GetPendingCardSelection(), cardRegistry),
		PendingCardDrawSelection: convertPendingCardDrawSelection(p.Selection().GetPendingCardDrawSelection(), cardRegistry),
		ForcedFirstAction:        convertForcedFirstAction(g.GetForcedFirstAction(p.ID())),
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
		fmt.Println("⚠️  convertSelectStartingCardsPhase: phase is nil")
		return nil
	}

	fmt.Printf("🔧 convertSelectStartingCardsPhase: cards=%d, corps=%v\n",
		len(phase.AvailableCards), phase.AvailableCorporations)

	// Get full card details for available cards
	availableCards := getPlayedCards(phase.AvailableCards, cardRegistry)

	// Get full card details for available corporations
	availableCorporations := getPlayedCards(phase.AvailableCorporations, cardRegistry)

	result := &SelectStartingCardsPhaseDto{
		AvailableCards:        availableCards,
		AvailableCorporations: availableCorporations,
	}

	fmt.Printf("🔧 convertSelectStartingCardsPhase result: cards=%d, corps=%d\n",
		len(result.AvailableCards), len(result.AvailableCorporations))

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
func convertPlayerEffects(effects []cardtypes.CardEffect) []PlayerEffectDto {
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
func convertPlayerActions(actions []cardtypes.CardAction) []PlayerActionDto {
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
