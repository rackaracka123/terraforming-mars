package dto

import (
	"go.uber.org/zap"

	"terraforming-mars-backend/internal/cards"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/internal/logger"
)

// ToCardDto converts a Card to CardDto
func ToCardDto(card gamecards.Card) CardDto {
	return CardDto{
		ID:                 card.ID,
		Name:               card.Name,
		Type:               CardType(card.Type),
		Cost:               card.Cost,
		Description:        card.Description,
		Pack:               card.Pack,
		Tags:               mapSlice(card.Tags, func(t shared.CardTag) CardTag { return CardTag(t) }),
		Requirements:       mapSlice(card.Requirements, toRequirementDto),
		Behaviors:          mapSlice(card.Behaviors, toCardBehaviorDto),
		ResourceStorage:    ptrCast(card.ResourceStorage, toResourceStorageDto),
		VPConditions:       mapSlice(card.VPConditions, toVPConditionDto),
		StartingResources:  ptrCast(card.StartingResources, toResourceSetDto),
		StartingProduction: ptrCast(card.StartingProduction, toResourceSetDto),
	}
}

// toResourceSetDto converts shared.ResourceSet to ResourceSet DTO.
func toResourceSetDto(rs shared.ResourceSet) ResourceSet {
	return ResourceSet{
		Credits:  rs.Credits,
		Steel:    rs.Steel,
		Titanium: rs.Titanium,
		Plants:   rs.Plants,
		Energy:   rs.Energy,
		Heat:     rs.Heat,
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

func toRequirementDto(req gamecards.Requirement) RequirementDto {
	return RequirementDto{
		Type:     RequirementType(req.Type),
		Min:      req.Min,
		Max:      req.Max,
		Location: ptrCast(req.Location, func(l gamecards.CardApplyLocation) CardApplyLocation { return CardApplyLocation(l) }),
		Tag:      ptrCast(req.Tag, func(t shared.CardTag) CardTag { return CardTag(t) }),
		Resource: ptrCast(req.Resource, func(r shared.ResourceType) ResourceType { return ResourceType(r) }),
	}
}

func toCardBehaviorDto(behavior shared.CardBehavior) CardBehaviorDto {
	return CardBehaviorDto{
		Triggers: mapSlice(behavior.Triggers, toTriggerDto),
		Inputs:   mapSlice(behavior.Inputs, toResourceConditionDto),
		Outputs:  mapSlice(behavior.Outputs, toResourceConditionDto),
		Choices:  mapSlice(behavior.Choices, toChoiceDto),
	}
}

func toTriggerDto(trigger shared.Trigger) TriggerDto {
	return TriggerDto{
		Type:      ResourceTriggerType(trigger.Type),
		Condition: ptrCast(trigger.Condition, toResourceTriggerConditionDto),
	}
}

func toResourceTriggerConditionDto(cond shared.ResourceTriggerCondition) ResourceTriggerConditionDto {
	return ResourceTriggerConditionDto{
		Type:              TriggerType(cond.Type),
		Location:          ptrCast(cond.Location, func(l string) CardApplyLocation { return CardApplyLocation(l) }),
		AffectedTags:      mapSlice(cond.AffectedTags, func(t shared.CardTag) CardTag { return CardTag(t) }),
		AffectedResources: cond.AffectedResources,
		Target:            ptrCast(cond.Target, func(t string) TargetType { return TargetType(t) }),
	}
}

func toResourceConditionDto(rc shared.ResourceCondition) ResourceConditionDto {
	return ResourceConditionDto{
		Type:                     ResourceType(rc.ResourceType),
		Amount:                   rc.Amount,
		Target:                   TargetType(rc.Target),
		AffectedResources:        rc.AffectedResources,
		AffectedTags:             mapSlice(rc.AffectedTags, func(t shared.CardTag) CardTag { return CardTag(t) }),
		AffectedCardTypes:        mapSlice(rc.AffectedCardTypes, func(ct string) CardType { return CardType(ct) }),
		AffectedStandardProjects: mapSlice(rc.AffectedStandardProjects, func(sp shared.StandardProject) StandardProject { return StandardProject(sp) }),
		MaxTrigger:               rc.MaxTrigger,
		Per:                      ptrCast(rc.Per, toPerConditionDto),
	}
}

func toChoiceDto(choice shared.Choice) ChoiceDto {
	return ChoiceDto{
		Inputs:  mapSlice(choice.Inputs, toResourceConditionDto),
		Outputs: mapSlice(choice.Outputs, toResourceConditionDto),
	}
}

func toPerConditionDto(pc shared.PerCondition) PerConditionDto {
	return PerConditionDto{
		Type:     ResourceType(pc.ResourceType),
		Amount:   pc.Amount,
		Location: ptrCast(pc.Location, func(l string) CardApplyLocation { return CardApplyLocation(l) }),
		Target:   ptrCast(pc.Target, func(t string) TargetType { return TargetType(t) }),
		Tag:      ptrCast(pc.Tag, func(t shared.CardTag) CardTag { return CardTag(t) }),
	}
}

func toResourceStorageDto(storage gamecards.ResourceStorage) ResourceStorageDto {
	return ResourceStorageDto{
		Type:     ResourceType(storage.Type),
		Capacity: storage.Capacity,
		Starting: storage.Starting,
	}
}

func toVPConditionDto(vp gamecards.VictoryPointCondition) VPConditionDto {
	return VPConditionDto{
		Amount:     vp.Amount,
		Condition:  VPConditionType(vp.Condition),
		MaxTrigger: vp.MaxTrigger,
		Per:        ptrCast(vp.Per, toVPPerConditionDto),
	}
}

// toVPPerConditionDto converts gamecards.PerCondition (used in VP conditions) to PerConditionDto.
// This is separate from toPerConditionDto because gamecards.PerCondition uses Type instead of ResourceType field.
func toVPPerConditionDto(pc gamecards.PerCondition) PerConditionDto {
	return PerConditionDto{
		Type:     ResourceType(pc.Type),
		Amount:   pc.Amount,
		Location: ptrCast(pc.Location, func(l gamecards.CardApplyLocation) CardApplyLocation { return CardApplyLocation(l) }),
		Target:   ptrCast(pc.Target, func(t gamecards.TargetType) TargetType { return TargetType(t) }),
		Tag:      ptrCast(pc.Tag, func(t shared.CardTag) CardTag { return CardTag(t) }),
	}
}
