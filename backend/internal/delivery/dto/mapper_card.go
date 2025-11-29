package dto

import (
	"go.uber.org/zap"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game/cardtypes"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/logger"
)

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

// Card-related helper functions for nested DTO conversions

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
