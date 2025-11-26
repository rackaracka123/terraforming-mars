package player

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session/game/card"
	"terraforming-mars-backend/internal/session/types"
)

var log = logger.Get()

// ProductionPhase contains both card selection and production phase state for a player
type ProductionPhase struct {
	AvailableCards    []string // card.Card IDs available for selection
	SelectionComplete bool     // Whether player completed card selection
	BeforeResources   types.Resources
	AfterResources    types.Resources
	EnergyConverted   int
	CreditsIncome     int
}

// DeepCopy creates a deep copy of the ProductionPhase
func (p *ProductionPhase) DeepCopy() *ProductionPhase {
	if p == nil {
		return nil
	}

	return &ProductionPhase{
		AvailableCards:    p.AvailableCards,
		SelectionComplete: p.SelectionComplete,
		BeforeResources:   p.BeforeResources.DeepCopy(),
		AfterResources:    p.AfterResources.DeepCopy(),
		EnergyConverted:   p.EnergyConverted,
		CreditsIncome:     p.CreditsIncome,
	}
}

type SelectStartingCardsPhase struct {
	AvailableCards        []string // card.Card IDs available for selection
	AvailableCorporations []string // Corporation IDs available for selection (2 corporations)
}

// PendingTileSelection represents a pending tile placement action
type PendingTileSelection struct {
	TileType       string   // "city", "greenery", "ocean"
	AvailableHexes []string // Backend-calculated valid hex coordinates
	Source         string   // What triggered this selection (card ID, standard project, etc.)
}

// PendingTileSelectionQueue represents a queue of tile placements to be made
type PendingTileSelectionQueue struct {
	Items  []string // Queue of tile types: ["city", "city", "ocean"]
	Source string   // card.Card ID that triggered all placements
}

// PendingCardSelection represents a pending card selection action (e.g., sell patents, card effects)
type PendingCardSelection struct {
	AvailableCards []string       // card.Card IDs player can select from
	CardCosts      map[string]int // card.Card ID -> cost to select (0 for sell patents, 3 for buying cards)
	CardRewards    map[string]int // card.Card ID -> reward for selecting (1 MC for sell patents)
	Source         string         // What triggered this selection ("sell-patents", card ID, etc.)
	MinCards       int            // Minimum cards to select (0 for sell patents)
	MaxCards       int            // Maximum cards to select (hand size for sell patents)
}

// PendingCardDrawSelection represents a pending card draw/peek/take/buy action from card effects
type PendingCardDrawSelection struct {
	AvailableCards []string // card.Card IDs shown to player (drawn or peeked)
	FreeTakeCount  int      // Number of cards to take for free (mandatory, 0 = optional)
	MaxBuyCount    int      // Maximum cards to buy (optional, 0 = no buying allowed)
	CardBuyCost    int      // Cost per card when buying (typically 3 MC, 0 if no buying)
	Source         string   // card.Card ID or action that triggered this
}

// ForcedFirstAction represents an action that must be completed as the player's first turn action
// Examples: Tharsis Republic must place a city as their first action
type ForcedFirstAction struct {
	ActionType    string // Type of action: "city_placement", "card_draw", etc.
	CorporationID string // Corporation that requires this action
	Source        string // Source to match for completion (corporation ID)
	Completed     bool   // Whether the forced action has been completed
	Description   string // Human-readable description for UI
}

// RequirementModifier represents a discount or lenience that modifies card/standard project requirements
// These are calculated from player effects and automatically updated when card hand or effects change
type RequirementModifier struct {
	Amount                int                    // Modifier amount (discount/lenience value)
	AffectedResources     []types.ResourceType   // types.Resources affected (e.g., ["credits"] for price discount, ["temperature"] for global param)
	CardTarget            *string                // Optional: specific card ID this applies to
	StandardProjectTarget *types.StandardProject // Optional: specific standard project this applies to
}

// Player represents a player in the game with encapsulated state
type Player struct {
	// Infrastructure (private)
	mu       sync.RWMutex
	eventBus *events.EventBusImpl

	// Identity (private)
	id     string
	name   string
	gameID string

	// Corporation (private)
	corporation   *card.Card
	corporationID string

	// Cards (private)
	cards       []string // Hand
	playedCards []string // Played cards

	// Resources (private)
	resources          types.Resources
	production         types.Production
	terraformRating    int
	victoryPoints      int
	resourceStorage    map[string]int
	paymentSubstitutes []card.PaymentSubstitute

	// Turn State (private)
	passed           bool
	availableActions int
	isConnected      bool

	// Effects (private)
	effects              []card.PlayerEffect
	actions              []PlayerAction
	requirementModifiers []RequirementModifier

	// Phase States (private)
	productionPhase          *ProductionPhase
	selectStartingCardsPhase *SelectStartingCardsPhase

	// Pending Selections (private)
	pendingTileSelection      *PendingTileSelection
	pendingTileSelectionQueue *PendingTileSelectionQueue
	pendingCardSelection      *PendingCardSelection
	pendingCardDrawSelection  *PendingCardDrawSelection
	forcedFirstAction         *ForcedFirstAction
}

// ================== Identity Getters (read-only) ==================

func (p *Player) ID() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.id
}

func (p *Player) Name() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.name
}

func (p *Player) GameID() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.gameID
}

// ================== Corporation ==================

func (p *Player) Corporation() *card.Card {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.corporation == nil {
		return nil
	}
	corpCopy := p.corporation.DeepCopy()
	return &corpCopy
}

func (p *Player) CorporationID() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.corporationID
}

func (p *Player) SetCorporationID(ctx context.Context, corporationID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.corporationID = corporationID
	return nil
}

func (p *Player) SetCorporation(ctx context.Context, corporation card.Card) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	corpCopy := corporation.DeepCopy()
	p.corporation = &corpCopy
	p.corporationID = corporation.ID
	return nil
}

// ================== Cards ==================

func (p *Player) Cards() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	cardsCopy := make([]string, len(p.cards))
	copy(cardsCopy, p.cards)
	return cardsCopy
}

func (p *Player) PlayedCards() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	playedCardsCopy := make([]string, len(p.playedCards))
	copy(playedCardsCopy, p.playedCards)
	return playedCardsCopy
}

func (p *Player) AddCardToHand(ctx context.Context, cardID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.cards = append(p.cards, cardID)
	return nil
}

func (p *Player) RemoveCardFromHand(ctx context.Context, cardID string) (bool, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for i, id := range p.cards {
		if id == cardID {
			p.cards = append(p.cards[:i], p.cards[i+1:]...)
			return true, nil
		}
	}
	return false, nil
}

func (p *Player) PlayCard(ctx context.Context, cardID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.playedCards = append(p.playedCards, cardID)

	// Remove from hand if present
	for i, id := range p.cards {
		if id == cardID {
			p.cards = append(p.cards[:i], p.cards[i+1:]...)
			break
		}
	}
	return nil
}

func (p *Player) SetCards(ctx context.Context, cards []string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	cardsCopy := make([]string, len(cards))
	copy(cardsCopy, cards)
	p.cards = cardsCopy
	return nil
}

func (p *Player) SetPlayedCards(ctx context.Context, playedCards []string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	playedCardsCopy := make([]string, len(playedCards))
	copy(playedCardsCopy, playedCards)
	p.playedCards = playedCardsCopy
	return nil
}

// ================== Resources ==================

func (p *Player) Resources() types.Resources {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.resources
}

func (p *Player) Production() types.Production {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.production
}

func (p *Player) TerraformRating() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.terraformRating
}

func (p *Player) VictoryPoints() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.victoryPoints
}

func (p *Player) ResourceStorage() map[string]int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	storageCopy := make(map[string]int)
	for k, v := range p.resourceStorage {
		storageCopy[k] = v
	}
	return storageCopy
}

func (p *Player) PaymentSubstitutes() []card.PaymentSubstitute {
	p.mu.RLock()
	defer p.mu.RUnlock()
	substitutesCopy := make([]card.PaymentSubstitute, len(p.paymentSubstitutes))
	copy(substitutesCopy, p.paymentSubstitutes)
	return substitutesCopy
}

func (p *Player) SetResources(ctx context.Context, resources types.Resources) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Calculate delta for event
	oldResources := p.resources
	changes := make(map[string]int)
	if oldResources.Credits != resources.Credits {
		changes["credits"] = resources.Credits - oldResources.Credits
	}
	if oldResources.Steel != resources.Steel {
		changes["steel"] = resources.Steel - oldResources.Steel
	}
	if oldResources.Titanium != resources.Titanium {
		changes["titanium"] = resources.Titanium - oldResources.Titanium
	}
	if oldResources.Plants != resources.Plants {
		changes["plants"] = resources.Plants - oldResources.Plants
	}
	if oldResources.Energy != resources.Energy {
		changes["energy"] = resources.Energy - oldResources.Energy
	}
	if oldResources.Heat != resources.Heat {
		changes["heat"] = resources.Heat - oldResources.Heat
	}

	// Update state
	p.resources = resources

	// Publish event if resources changed
	if p.eventBus != nil && len(changes) > 0 {
		events.Publish(p.eventBus, events.ResourcesChangedEvent{
			GameID:    p.gameID,
			PlayerID:  p.id,
			Changes:   changes,
			Timestamp: time.Now(),
		})
	}

	return nil
}

func (p *Player) SetProduction(ctx context.Context, production types.Production) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.production = production
	return nil
}

func (p *Player) SetTerraformRating(ctx context.Context, rating int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	oldRating := p.terraformRating
	p.terraformRating = rating

	// Publish event if rating changed
	if p.eventBus != nil && oldRating != rating {
		log.Debug("📡 Publishing TerraformRatingChangedEvent",
			zap.String("game_id", p.gameID),
			zap.String("player_id", p.id),
			zap.Int("old_rating", oldRating),
			zap.Int("new_rating", rating))

		events.Publish(p.eventBus, events.TerraformRatingChangedEvent{
			GameID:    p.gameID,
			PlayerID:  p.id,
			OldRating: oldRating,
			NewRating: rating,
			Timestamp: time.Now(),
		})
	}

	return nil
}

func (p *Player) SetVictoryPoints(ctx context.Context, victoryPoints int) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.victoryPoints = victoryPoints
	return nil
}

func (p *Player) SetResourceStorage(ctx context.Context, storage map[string]int) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	storageCopy := make(map[string]int)
	for k, v := range storage {
		storageCopy[k] = v
	}
	p.resourceStorage = storageCopy
	return nil
}

func (p *Player) SetPaymentSubstitutes(ctx context.Context, substitutes []card.PaymentSubstitute) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	substitutesCopy := make([]card.PaymentSubstitute, len(substitutes))
	copy(substitutesCopy, substitutes)
	p.paymentSubstitutes = substitutesCopy
	return nil
}

// AddResources adds or removes resources (supports multiple resources at once)
// changes is a map of resource type to delta (positive or negative)
func (p *Player) AddResources(changes map[types.ResourceType]int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for resourceType, delta := range changes {
		switch resourceType {
		case types.ResourceCredits:
			p.resources.Credits += delta
		case types.ResourceSteel:
			p.resources.Steel += delta
		case types.ResourceTitanium:
			p.resources.Titanium += delta
		case types.ResourcePlants:
			p.resources.Plants += delta
		case types.ResourceEnergy:
			p.resources.Energy += delta
		case types.ResourceHeat:
			p.resources.Heat += delta
		}
	}
}

// AddProduction adds or removes production (supports multiple resources at once)
// changes is a map of resource type to delta (positive or negative)
func (p *Player) AddProduction(changes map[types.ResourceType]int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for resourceType, delta := range changes {
		switch resourceType {
		case types.ResourceCredits:
			p.production.Credits += delta
		case types.ResourceSteel:
			p.production.Steel += delta
		case types.ResourceTitanium:
			p.production.Titanium += delta
		case types.ResourcePlants:
			p.production.Plants += delta
		case types.ResourceEnergy:
			p.production.Energy += delta
		case types.ResourceHeat:
			p.production.Heat += delta
		}
	}
}

// UpdateTerraformRating updates the player's terraform rating
func (p *Player) UpdateTerraformRating(delta int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.terraformRating += delta
}

// ================== Turn State ==================

func (p *Player) Passed() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.passed
}

func (p *Player) AvailableActions() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.availableActions
}

func (p *Player) IsConnected() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.isConnected
}

func (p *Player) SetPassed(ctx context.Context, passed bool) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.passed = passed
	return nil
}

func (p *Player) SetAvailableActions(ctx context.Context, actions int) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.availableActions = actions
	return nil
}

func (p *Player) SetConnectionStatus(ctx context.Context, isConnected bool) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.isConnected = isConnected
	return nil
}

func (p *Player) ConsumeAction() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.availableActions > 0 {
		p.availableActions--
		return true
	}
	return false
}

// ================== Effects ==================

func (p *Player) Effects() []card.PlayerEffect {
	p.mu.RLock()
	defer p.mu.RUnlock()
	effectsCopy := make([]card.PlayerEffect, len(p.effects))
	for i, effect := range p.effects {
		effectsCopy[i] = *effect.DeepCopy()
	}
	return effectsCopy
}

func (p *Player) Actions() []PlayerAction {
	p.mu.RLock()
	defer p.mu.RUnlock()
	actionsCopy := make([]PlayerAction, len(p.actions))
	for i, action := range p.actions {
		actionsCopy[i] = *action.DeepCopy()
	}
	return actionsCopy
}

func (p *Player) RequirementModifiers() []RequirementModifier {
	p.mu.RLock()
	defer p.mu.RUnlock()

	modifiersCopy := make([]RequirementModifier, len(p.requirementModifiers))
	for i, modifier := range p.requirementModifiers {
		affectedResourcesCopy := make([]types.ResourceType, len(modifier.AffectedResources))
		copy(affectedResourcesCopy, modifier.AffectedResources)

		var cardTargetCopy *string
		if modifier.CardTarget != nil {
			val := *modifier.CardTarget
			cardTargetCopy = &val
		}

		var standardProjectTargetCopy *types.StandardProject
		if modifier.StandardProjectTarget != nil {
			val := *modifier.StandardProjectTarget
			standardProjectTargetCopy = &val
		}

		modifiersCopy[i] = RequirementModifier{
			Amount:                modifier.Amount,
			AffectedResources:     affectedResourcesCopy,
			CardTarget:            cardTargetCopy,
			StandardProjectTarget: standardProjectTargetCopy,
		}
	}
	return modifiersCopy
}

func (p *Player) ForcedFirstAction() *ForcedFirstAction {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.forcedFirstAction == nil {
		return nil
	}

	return &ForcedFirstAction{
		ActionType:    p.forcedFirstAction.ActionType,
		CorporationID: p.forcedFirstAction.CorporationID,
		Source:        p.forcedFirstAction.Source,
		Completed:     p.forcedFirstAction.Completed,
		Description:   p.forcedFirstAction.Description,
	}
}

func (p *Player) SetEffects(ctx context.Context, effects []card.PlayerEffect) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	effectsCopy := make([]card.PlayerEffect, len(effects))
	for i, effect := range effects {
		effectsCopy[i] = *effect.DeepCopy()
	}
	p.effects = effectsCopy
	return nil
}

func (p *Player) SetActions(ctx context.Context, actions []PlayerAction) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	actionsCopy := make([]PlayerAction, len(actions))
	for i, action := range actions {
		actionsCopy[i] = *action.DeepCopy()
	}
	p.actions = actionsCopy
	return nil
}

func (p *Player) SetRequirementModifiers(ctx context.Context, modifiers []RequirementModifier) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	modifiersCopy := make([]RequirementModifier, len(modifiers))
	for i, modifier := range modifiers {
		affectedResourcesCopy := make([]types.ResourceType, len(modifier.AffectedResources))
		copy(affectedResourcesCopy, modifier.AffectedResources)

		var cardTargetCopy *string
		if modifier.CardTarget != nil {
			val := *modifier.CardTarget
			cardTargetCopy = &val
		}

		var standardProjectTargetCopy *types.StandardProject
		if modifier.StandardProjectTarget != nil {
			val := *modifier.StandardProjectTarget
			standardProjectTargetCopy = &val
		}

		modifiersCopy[i] = RequirementModifier{
			Amount:                modifier.Amount,
			AffectedResources:     affectedResourcesCopy,
			CardTarget:            cardTargetCopy,
			StandardProjectTarget: standardProjectTargetCopy,
		}
	}
	p.requirementModifiers = modifiersCopy
	return nil
}

func (p *Player) SetForcedFirstAction(ctx context.Context, action *ForcedFirstAction) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if action == nil {
		p.forcedFirstAction = nil
		return nil
	}

	p.forcedFirstAction = &ForcedFirstAction{
		ActionType:    action.ActionType,
		CorporationID: action.CorporationID,
		Source:        action.Source,
		Completed:     action.Completed,
		Description:   action.Description,
	}
	return nil
}

// ================== Phase States ==================

func (p *Player) SelectStartingCardsPhase() *SelectStartingCardsPhase {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.selectStartingCardsPhase == nil {
		return nil
	}

	availableCardsCopy := make([]string, len(p.selectStartingCardsPhase.AvailableCards))
	copy(availableCardsCopy, p.selectStartingCardsPhase.AvailableCards)

	availableCorporationsCopy := make([]string, len(p.selectStartingCardsPhase.AvailableCorporations))
	copy(availableCorporationsCopy, p.selectStartingCardsPhase.AvailableCorporations)

	return &SelectStartingCardsPhase{
		AvailableCards:        availableCardsCopy,
		AvailableCorporations: availableCorporationsCopy,
	}
}

func (p *Player) ProductionPhase() *ProductionPhase {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.productionPhase == nil {
		return nil
	}

	return p.productionPhase.DeepCopy()
}

func (p *Player) PendingCardSelection() *PendingCardSelection {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.pendingCardSelection == nil {
		return nil
	}

	availableCardsCopy := make([]string, len(p.pendingCardSelection.AvailableCards))
	copy(availableCardsCopy, p.pendingCardSelection.AvailableCards)

	cardCostsCopy := make(map[string]int)
	for k, v := range p.pendingCardSelection.CardCosts {
		cardCostsCopy[k] = v
	}

	cardRewardsCopy := make(map[string]int)
	for k, v := range p.pendingCardSelection.CardRewards {
		cardRewardsCopy[k] = v
	}

	return &PendingCardSelection{
		AvailableCards: availableCardsCopy,
		CardCosts:      cardCostsCopy,
		CardRewards:    cardRewardsCopy,
		Source:         p.pendingCardSelection.Source,
		MinCards:       p.pendingCardSelection.MinCards,
		MaxCards:       p.pendingCardSelection.MaxCards,
	}
}

func (p *Player) PendingCardDrawSelection() *PendingCardDrawSelection {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.pendingCardDrawSelection == nil {
		return nil
	}

	availableCardsCopy := make([]string, len(p.pendingCardDrawSelection.AvailableCards))
	copy(availableCardsCopy, p.pendingCardDrawSelection.AvailableCards)

	return &PendingCardDrawSelection{
		AvailableCards: availableCardsCopy,
		FreeTakeCount:  p.pendingCardDrawSelection.FreeTakeCount,
		MaxBuyCount:    p.pendingCardDrawSelection.MaxBuyCount,
		CardBuyCost:    p.pendingCardDrawSelection.CardBuyCost,
		Source:         p.pendingCardDrawSelection.Source,
	}
}

func (p *Player) SetStartingCardsSelection(ctx context.Context, cardIDs, corpIDs []string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.selectStartingCardsPhase = &SelectStartingCardsPhase{
		AvailableCards:        cardIDs,
		AvailableCorporations: corpIDs,
	}
	return nil
}

func (p *Player) CompleteStartingSelection(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.selectStartingCardsPhase = nil
	return nil
}

func (p *Player) CompleteProductionSelection(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.productionPhase != nil {
		p.productionPhase.SelectionComplete = true
	}
	return nil
}

func (p *Player) SetStartingCardsPhase(ctx context.Context, phase *SelectStartingCardsPhase) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.selectStartingCardsPhase = phase
	return nil
}

func (p *Player) SetProductionPhase(ctx context.Context, phase *ProductionPhase) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.productionPhase = phase
	return nil
}

func (p *Player) SetPendingCardSelection(ctx context.Context, selection *PendingCardSelection) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.pendingCardSelection = selection

	log.Debug("✅ Player pending card selection updated",
		zap.String("game_id", p.gameID),
		zap.String("player_id", p.id),
		zap.String("source", selection.Source),
		zap.Int("available_cards", len(selection.AvailableCards)))

	return nil
}

func (p *Player) ClearPendingCardSelection(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.pendingCardSelection = nil

	log.Debug("✅ Player pending card selection cleared",
		zap.String("game_id", p.gameID),
		zap.String("player_id", p.id))

	return nil
}

func (p *Player) SetPendingCardDrawSelection(ctx context.Context, selection *PendingCardDrawSelection) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pendingCardDrawSelection = selection
	return nil
}

// ================== Tile Queue ==================

func (p *Player) PendingTileSelection() *PendingTileSelection {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.pendingTileSelection == nil {
		return nil
	}

	availableHexesCopy := make([]string, len(p.pendingTileSelection.AvailableHexes))
	copy(availableHexesCopy, p.pendingTileSelection.AvailableHexes)

	return &PendingTileSelection{
		TileType:       p.pendingTileSelection.TileType,
		AvailableHexes: availableHexesCopy,
		Source:         p.pendingTileSelection.Source,
	}
}

func (p *Player) PendingTileSelectionQueue() *PendingTileSelectionQueue {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.pendingTileSelectionQueue == nil {
		return nil
	}

	itemsCopy := make([]string, len(p.pendingTileSelectionQueue.Items))
	copy(itemsCopy, p.pendingTileSelectionQueue.Items)

	return &PendingTileSelectionQueue{
		Items:  itemsCopy,
		Source: p.pendingTileSelectionQueue.Source,
	}
}

func (p *Player) CreateTileQueue(ctx context.Context, cardID string, tileTypes []string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(tileTypes) > 0 {
		p.pendingTileSelectionQueue = &PendingTileSelectionQueue{
			Items:  tileTypes,
			Source: cardID,
		}

		// Publish TileQueueCreatedEvent
		if p.eventBus != nil {
			log.Debug("📡 Publishing TileQueueCreatedEvent",
				zap.String("game_id", p.gameID),
				zap.String("player_id", p.id),
				zap.Int("queue_size", len(tileTypes)),
				zap.String("source", cardID))

			events.Publish(p.eventBus, TileQueueCreatedEvent{
				GameID:    p.gameID,
				PlayerID:  p.id,
				QueueSize: len(tileTypes),
				Source:    cardID,
				Timestamp: time.Now(),
			})
		}
	}

	return nil
}

func (p *Player) GetTileQueue(ctx context.Context) (*PendingTileSelectionQueue, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.pendingTileSelectionQueue == nil {
		return nil, nil
	}

	itemsCopy := make([]string, len(p.pendingTileSelectionQueue.Items))
	copy(itemsCopy, p.pendingTileSelectionQueue.Items)

	return &PendingTileSelectionQueue{
		Items:  itemsCopy,
		Source: p.pendingTileSelectionQueue.Source,
	}, nil
}

func (p *Player) ProcessNextTile(ctx context.Context) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// If no queue exists or queue is empty, nothing to process
	if p.pendingTileSelectionQueue == nil || len(p.pendingTileSelectionQueue.Items) == 0 {
		log.Debug("No tile placements in queue",
			zap.String("game_id", p.gameID),
			zap.String("player_id", p.id))
		return "", nil
	}

	// Pop the first item from the queue
	nextTileType := p.pendingTileSelectionQueue.Items[0]
	remainingItems := p.pendingTileSelectionQueue.Items[1:]

	log.Info("🎯 Popping next tile from queue",
		zap.String("game_id", p.gameID),
		zap.String("player_id", p.id),
		zap.String("tile_type", nextTileType),
		zap.String("source", p.pendingTileSelectionQueue.Source),
		zap.Int("remaining_in_queue", len(remainingItems)))

	source := p.pendingTileSelectionQueue.Source

	// Update queue with remaining items or clear if empty
	if len(remainingItems) > 0 {
		p.pendingTileSelectionQueue = &PendingTileSelectionQueue{
			Items:  remainingItems,
			Source: p.pendingTileSelectionQueue.Source,
		}
	} else {
		// This is the last item, clear the queue
		p.pendingTileSelectionQueue = nil
	}

	log.Debug("✅ Tile popped from queue",
		zap.String("game_id", p.gameID),
		zap.String("player_id", p.id),
		zap.String("tile_type", nextTileType))

	// If there are more tiles in queue after popping, publish event to trigger processing
	if p.eventBus != nil && len(remainingItems) > 0 {
		log.Debug("📡 Publishing TileQueueCreatedEvent for remaining tiles",
			zap.String("game_id", p.gameID),
			zap.String("player_id", p.id),
			zap.Int("remaining_count", len(remainingItems)),
			zap.String("source", source))

		events.Publish(p.eventBus, TileQueueCreatedEvent{
			GameID:    p.gameID,
			PlayerID:  p.id,
			QueueSize: len(remainingItems),
			Source:    source,
			Timestamp: time.Now(),
		})
	}

	return nextTileType, nil
}

func (p *Player) SetPendingTileSelection(ctx context.Context, selection *PendingTileSelection) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pendingTileSelection = selection
	return nil
}

func (p *Player) ClearPendingTileSelection(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pendingTileSelection = nil
	return nil
}

func (p *Player) SetTileQueue(ctx context.Context, queue *PendingTileSelectionQueue) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pendingTileSelectionQueue = queue
	return nil
}

func (p *Player) ClearTileQueue(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.pendingTileSelectionQueue = nil
	return nil
}

func (p *Player) QueueTilePlacement(source string, tileTypes []string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.pendingTileSelectionQueue = &PendingTileSelectionQueue{
		Items:  tileTypes,
		Source: source,
	}
}

// ================== Utility ==================

// GetStartingSelectionCards returns the player's starting card selection, nil if not in that phase
func (p *Player) GetStartingSelectionCards() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.selectStartingCardsPhase == nil {
		return nil
	}

	return p.selectStartingCardsPhase.AvailableCards
}

// GetProductionPhaseCards returns the player's production phase card selection, nil if not in that phase
func (p *Player) GetProductionPhaseCards() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.productionPhase == nil {
		return nil
	}

	return p.productionPhase.AvailableCards
}

// DeepCopy creates a deep copy of the Player
func (p *Player) DeepCopy() *Player {
	if p == nil {
		return nil
	}

	p.mu.RLock()
	defer p.mu.RUnlock()

	// Copy cards slice
	cardsCopy := make([]string, len(p.cards))
	copy(cardsCopy, p.cards)

	// Copy played cards slice
	playedCardsCopy := make([]string, len(p.playedCards))
	copy(playedCardsCopy, p.playedCards)

	// Deep copy production selection if it exists
	var productionSelectionCopy *ProductionPhase
	if p.productionPhase != nil {
		productionSelectionCopy = p.productionPhase.DeepCopy()
	}

	// Copy starting selection slice
	var startingSelectionCopy *SelectStartingCardsPhase
	if p.selectStartingCardsPhase != nil {
		availableCardsCopy := make([]string, len(p.selectStartingCardsPhase.AvailableCards))
		copy(availableCardsCopy, p.selectStartingCardsPhase.AvailableCards)

		availableCorporationsCopy := make([]string, len(p.selectStartingCardsPhase.AvailableCorporations))
		copy(availableCorporationsCopy, p.selectStartingCardsPhase.AvailableCorporations)

		startingSelectionCopy = &SelectStartingCardsPhase{
			AvailableCards:        availableCardsCopy,
			AvailableCorporations: availableCorporationsCopy,
		}
	}

	// Deep copy pending tile selection if it exists
	var pendingTileSelectionCopy *PendingTileSelection
	if p.pendingTileSelection != nil {
		availableHexesCopy := make([]string, len(p.pendingTileSelection.AvailableHexes))
		copy(availableHexesCopy, p.pendingTileSelection.AvailableHexes)

		pendingTileSelectionCopy = &PendingTileSelection{
			TileType:       p.pendingTileSelection.TileType,
			AvailableHexes: availableHexesCopy,
			Source:         p.pendingTileSelection.Source,
		}
	}

	// Deep copy pending tile selection queue if it exists
	var pendingTileSelectionQueueCopy *PendingTileSelectionQueue
	if p.pendingTileSelectionQueue != nil {
		itemsCopy := make([]string, len(p.pendingTileSelectionQueue.Items))
		copy(itemsCopy, p.pendingTileSelectionQueue.Items)

		pendingTileSelectionQueueCopy = &PendingTileSelectionQueue{
			Items:  itemsCopy,
			Source: p.pendingTileSelectionQueue.Source,
		}
	}

	// Deep copy effects slice
	effectsCopy := make([]card.PlayerEffect, len(p.effects))
	for i, effect := range p.effects {
		effectsCopy[i] = *effect.DeepCopy()
	}

	// Deep copy actions slice
	actionsCopy := make([]PlayerAction, len(p.actions))
	for i, action := range p.actions {
		actionsCopy[i] = *action.DeepCopy()
	}

	// Deep copy resource storage map
	resourceStorageCopy := make(map[string]int)
	for cardID, count := range p.resourceStorage {
		resourceStorageCopy[cardID] = count
	}

	// Deep copy payment substitutes slice
	var paymentSubstitutesCopy []card.PaymentSubstitute
	if p.paymentSubstitutes != nil {
		paymentSubstitutesCopy = make([]card.PaymentSubstitute, len(p.paymentSubstitutes))
		copy(paymentSubstitutesCopy, p.paymentSubstitutes)
	}

	// Deep copy requirement modifiers slice
	var requirementModifiersCopy []RequirementModifier
	if p.requirementModifiers != nil {
		requirementModifiersCopy = make([]RequirementModifier, len(p.requirementModifiers))
		for i, modifier := range p.requirementModifiers {
			affectedResourcesCopy := make([]types.ResourceType, len(modifier.AffectedResources))
			copy(affectedResourcesCopy, modifier.AffectedResources)

			var cardTargetCopy *string
			if modifier.CardTarget != nil {
				val := *modifier.CardTarget
				cardTargetCopy = &val
			}

			var standardProjectTargetCopy *types.StandardProject
			if modifier.StandardProjectTarget != nil {
				val := *modifier.StandardProjectTarget
				standardProjectTargetCopy = &val
			}

			requirementModifiersCopy[i] = RequirementModifier{
				Amount:                modifier.Amount,
				AffectedResources:     affectedResourcesCopy,
				CardTarget:            cardTargetCopy,
				StandardProjectTarget: standardProjectTargetCopy,
			}
		}
	}

	// Deep copy pending card selection if it exists
	var pendingCardSelectionCopy *PendingCardSelection
	if p.pendingCardSelection != nil {
		availableCardsCopy := make([]string, len(p.pendingCardSelection.AvailableCards))
		copy(availableCardsCopy, p.pendingCardSelection.AvailableCards)

		cardCostsCopy := make(map[string]int)
		for cardID, cost := range p.pendingCardSelection.CardCosts {
			cardCostsCopy[cardID] = cost
		}

		cardRewardsCopy := make(map[string]int)
		for cardID, reward := range p.pendingCardSelection.CardRewards {
			cardRewardsCopy[cardID] = reward
		}

		pendingCardSelectionCopy = &PendingCardSelection{
			AvailableCards: availableCardsCopy,
			CardCosts:      cardCostsCopy,
			CardRewards:    cardRewardsCopy,
			Source:         p.pendingCardSelection.Source,
			MinCards:       p.pendingCardSelection.MinCards,
			MaxCards:       p.pendingCardSelection.MaxCards,
		}
	}

	// Deep copy pending card draw selection if it exists
	var pendingCardDrawSelectionCopy *PendingCardDrawSelection
	if p.pendingCardDrawSelection != nil {
		availableCardsCopy := make([]string, len(p.pendingCardDrawSelection.AvailableCards))
		copy(availableCardsCopy, p.pendingCardDrawSelection.AvailableCards)

		pendingCardDrawSelectionCopy = &PendingCardDrawSelection{
			AvailableCards: availableCardsCopy,
			FreeTakeCount:  p.pendingCardDrawSelection.FreeTakeCount,
			MaxBuyCount:    p.pendingCardDrawSelection.MaxBuyCount,
			CardBuyCost:    p.pendingCardDrawSelection.CardBuyCost,
			Source:         p.pendingCardDrawSelection.Source,
		}
	}

	// Deep copy forced first action if it exists
	var forcedFirstActionCopy *ForcedFirstAction
	if p.forcedFirstAction != nil {
		forcedFirstActionCopy = &ForcedFirstAction{
			ActionType:    p.forcedFirstAction.ActionType,
			CorporationID: p.forcedFirstAction.CorporationID,
			Source:        p.forcedFirstAction.Source,
			Completed:     p.forcedFirstAction.Completed,
			Description:   p.forcedFirstAction.Description,
		}
	}

	// Deep copy corporation if it exists
	var corporationCopy *card.Card
	if p.corporation != nil {
		corpCopy := p.corporation.DeepCopy()
		corporationCopy = &corpCopy
	}

	return &Player{
		mu:                        sync.RWMutex{},
		eventBus:                  p.eventBus,
		id:                        p.id,
		name:                      p.name,
		gameID:                    p.gameID,
		corporation:               corporationCopy,
		corporationID:             p.corporationID,
		cards:                     cardsCopy,
		resources:                 p.resources, // Resources is a struct, so this is copied by value
		production:                p.production, // Production is a struct, so this is copied by value
		terraformRating:           p.terraformRating,
		playedCards:               playedCardsCopy,
		passed:                    p.passed,
		availableActions:          p.availableActions,
		victoryPoints:             p.victoryPoints,
		isConnected:               p.isConnected,
		effects:                   effectsCopy,
		actions:                   actionsCopy,
		productionPhase:           productionSelectionCopy,
		selectStartingCardsPhase:  startingSelectionCopy,
		pendingTileSelection:      pendingTileSelectionCopy,
		pendingTileSelectionQueue: pendingTileSelectionQueueCopy,
		pendingCardSelection:      pendingCardSelectionCopy,
		pendingCardDrawSelection:  pendingCardDrawSelectionCopy,
		forcedFirstAction:         forcedFirstActionCopy,
		resourceStorage:           resourceStorageCopy,
		paymentSubstitutes:        paymentSubstitutesCopy,
		requirementModifiers:      requirementModifiersCopy,
	}
}
