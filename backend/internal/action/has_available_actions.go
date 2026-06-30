package action

import (
	"terraforming-mars-backend/internal/awards"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/internal/milestones"
	"terraforming-mars-backend/internal/standardprojects"
)

// conversionStandardProjects are the two resource-conversion "standard projects"
// that are not part of the standard-project registry (they are surfaced as
// resource buttons), but still constitute legal action-phase moves.
var conversionStandardProjects = []shared.StandardProject{
	shared.StandardProjectConvertHeatToTemperature,
	shared.StandardProjectConvertPlantsToGreenery,
}

// HasAvailableActions reports whether the player has at least one legal move on
// their turn. It is the single aggregate "can this player act at all?" signal,
// built entirely from the existing per-surface state calculators so it can never
// diverge from what the UI shows as available.
//
// A player who has a pending selection or an incomplete forced first action is
// NOT stuck: they have a mandatory thing to do, so the function returns true.
//
// The function short-circuits on the first available surface. The caller is
// responsible for only invoking it once every registry is present.
func HasAvailableActions(
	g *game.Game,
	p *player.Player,
	cardRegistry cards.CardRegistry,
	stdProjRegistry standardprojects.StandardProjectRegistry,
	milestoneRegistry milestones.MilestoneRegistry,
	awardRegistry awards.AwardRegistry,
) bool {
	if g.HasAnyPendingSelection(p.ID()) {
		return true
	}
	if forced := g.GetForcedFirstAction(p.ID()); forced != nil && !forced.Completed {
		return true
	}

	for _, cardID := range p.Hand().Cards() {
		card, err := cardRegistry.GetByID(cardID)
		if err != nil {
			continue
		}
		if CalculatePlayerCardState(card, p, g, cardRegistry).Available() {
			return true
		}
	}

	for _, act := range p.Actions().List() {
		state := CalculatePlayerCardActionState(
			act.CardID, act.Behavior, act.TimesUsedThisGeneration, p, g, cardRegistry,
		)
		if state.Available() {
			return true
		}
	}

	if hasAvailableStandardProject(g, p, cardRegistry, stdProjRegistry) {
		return true
	}

	for _, def := range milestoneRegistry.GetAll() {
		state := CalculateMilestoneState(
			shared.MilestoneType(def.ID), p, g, cardRegistry, milestoneRegistry,
		)
		if state.Available() {
			return true
		}
	}

	for _, def := range awardRegistry.GetAll() {
		if CalculateAwardState(shared.AwardType(def.ID), p, g, awardRegistry).Available() {
			return true
		}
	}

	if g.HasColonies() {
		if CalculateColonyTradeState(p, g, cardRegistry).Available() {
			return true
		}
	}

	return false
}

// hasAvailableStandardProject reports whether any standard project the player can
// reach this game is available, covering the registry-backed projects (filtered to
// the enabled packs, mirroring the DTO mapper) plus the two resource conversions
// and sell-patents, which the per-project calculator already validates.
func hasAvailableStandardProject(
	g *game.Game,
	p *player.Player,
	cardRegistry cards.CardRegistry,
	stdProjRegistry standardprojects.StandardProjectRegistry,
) bool {
	settings := g.Settings()
	enabledPacks := make(map[string]bool, len(settings.CardPacks))
	for _, pack := range settings.CardPacks {
		enabledPacks[pack] = true
	}
	if settings.VenusNextEnabled {
		enabledPacks[shared.PackVenus] = true
	}

	for _, def := range stdProjRegistry.GetAll() {
		if def.Pack != "" && !enabledPacks[def.Pack] {
			continue
		}
		if CalculatePlayerStandardProjectState(shared.StandardProject(def.ID), p, g, cardRegistry).Available() {
			return true
		}
	}

	for _, projectType := range conversionStandardProjects {
		if CalculatePlayerStandardProjectState(projectType, p, g, cardRegistry).Available() {
			return true
		}
	}

	return false
}
