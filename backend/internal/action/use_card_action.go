package action

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// UseCardActionAction handles the business logic for using a card's manual action
// Card actions are repeatable blue card abilities with inputs and outputs
type UseCardActionAction struct {
	BaseAction
}

// NewUseCardActionAction creates a new use card action action
func NewUseCardActionAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	logger *zap.Logger,
) *UseCardActionAction {
	return &UseCardActionAction{
		BaseAction: BaseAction{
			gameRepo:     gameRepo,
			cardRegistry: cardRegistry,
			logger:       logger,
		},
	}
}

// Execute performs the use card action
func (a *UseCardActionAction) Execute(
	ctx context.Context,
	gameID string,
	playerID string,
	cardID string,
	behaviorIndex int,
) error {
	log := a.InitLogger(gameID, playerID).With(
		zap.String("card_id", cardID),
		zap.Int("behavior_index", behaviorIndex),
		zap.String("action", "use_card_action"),
	)
	log.Info("🎯 Player attempting to use card action")

	// 1. Validate game exists and is active
	g, err := ValidateActiveGame(ctx, a.GameRepository(), gameID, log)
	if err != nil {
		return err
	}

	// 2. Validate game is in action phase
	if err := ValidateGamePhase(g, game.GamePhaseAction, log); err != nil {
		return err
	}

	// 3. Validate it's the player's turn
	if err := ValidateCurrentTurn(g, playerID, log); err != nil {
		return err
	}

	// 4. Get player from game
	p, err := a.GetPlayerFromGame(g, playerID, log)
	if err != nil {
		return err
	}

	// 5. BUSINESS LOGIC: Find the card action in player's available actions
	cardAction, err := a.findCardAction(p, cardID, behaviorIndex, log)
	if err != nil {
		return err
	}

	log.Info("✅ Found card action",
		zap.String("card_name", cardAction.CardName),
		zap.Int("play_count", cardAction.PlayCount))

	// 6. BUSINESS LOGIC: Validate and apply inputs (resource costs)
	if err := a.applyInputs(ctx, cardAction.Behavior.Inputs, p, log); err != nil {
		return err
	}

	// 7. BUSINESS LOGIC: Apply outputs (resource gains, etc.)
	if err := a.applyOutputs(ctx, cardAction.Behavior.Outputs, p, g, cardAction.CardName, log); err != nil {
		return err
	}

	// 8. BUSINESS LOGIC: Increment play count for the action
	a.incrementPlayCount(p, cardID, behaviorIndex, log)

	// 9. BUSINESS LOGIC: Consume a player action
	consumed := a.ConsumePlayerAction(g, log)
	if !consumed {
		log.Warn("⚠️ Action not consumed (unlimited actions or already at 0)")
	}

	// NO MANUAL BROADCAST - BroadcastEvent automatically triggered by:
	//    - p.Resources().Add/Deduct() publishes ResourcesChangedEvent
	//    - g.GlobalParameters() methods publish domain events
	//    - CurrentTurn().ConsumeAction() publishes BroadcastEvent
	//    Broadcaster subscribes to BroadcastEvent and handles WebSocket updates

	log.Info("🎉 Card action executed successfully")
	return nil
}

// findCardAction finds a card action in the player's available actions
func (a *UseCardActionAction) findCardAction(
	p *player.Player,
	cardID string,
	behaviorIndex int,
	log *zap.Logger,
) (*player.CardAction, error) {
	actions := p.Actions().List()

	for i := range actions {
		if actions[i].CardID == cardID && actions[i].BehaviorIndex == behaviorIndex {
			return &actions[i], nil
		}
	}

	log.Error("Card action not found in player's available actions",
		zap.String("card_id", cardID),
		zap.Int("behavior_index", behaviorIndex))
	return nil, fmt.Errorf("card action not found: %s[%d]", cardID, behaviorIndex)
}

// applyInputs validates and deducts input resources
func (a *UseCardActionAction) applyInputs(
	ctx context.Context,
	inputs []shared.ResourceCondition,
	p *player.Player,
	log *zap.Logger,
) error {
	if len(inputs) == 0 {
		log.Debug("No inputs to apply")
		return nil
	}

	log.Info("💰 Processing card action inputs", zap.Int("input_count", len(inputs)))

	// First validate player has all required resources
	resources := p.Resources().Get()

	for _, input := range inputs {
		switch input.ResourceType {
		case shared.ResourceCredits:
			if resources.Credits < input.Amount {
				log.Error("Insufficient credits",
					zap.Int("required", input.Amount),
					zap.Int("available", resources.Credits))
				return fmt.Errorf("insufficient credits: need %d, have %d", input.Amount, resources.Credits)
			}
		case shared.ResourceSteel:
			if resources.Steel < input.Amount {
				log.Error("Insufficient steel",
					zap.Int("required", input.Amount),
					zap.Int("available", resources.Steel))
				return fmt.Errorf("insufficient steel: need %d, have %d", input.Amount, resources.Steel)
			}
		case shared.ResourceTitanium:
			if resources.Titanium < input.Amount {
				log.Error("Insufficient titanium",
					zap.Int("required", input.Amount),
					zap.Int("available", resources.Titanium))
				return fmt.Errorf("insufficient titanium: need %d, have %d", input.Amount, resources.Titanium)
			}
		case shared.ResourcePlants:
			if resources.Plants < input.Amount {
				log.Error("Insufficient plants",
					zap.Int("required", input.Amount),
					zap.Int("available", resources.Plants))
				return fmt.Errorf("insufficient plants: need %d, have %d", input.Amount, resources.Plants)
			}
		case shared.ResourceEnergy:
			if resources.Energy < input.Amount {
				log.Error("Insufficient energy",
					zap.Int("required", input.Amount),
					zap.Int("available", resources.Energy))
				return fmt.Errorf("insufficient energy: need %d, have %d", input.Amount, resources.Energy)
			}
		case shared.ResourceHeat:
			if resources.Heat < input.Amount {
				log.Error("Insufficient heat",
					zap.Int("required", input.Amount),
					zap.Int("available", resources.Heat))
				return fmt.Errorf("insufficient heat: need %d, have %d", input.Amount, resources.Heat)
			}
		default:
			log.Warn("⚠️ Unhandled input type", zap.String("type", string(input.ResourceType)))
		}
	}

	// All validations passed, now deduct resources
	for _, input := range inputs {
		switch input.ResourceType {
		case shared.ResourceCredits:
			p.Resources().Add(map[shared.ResourceType]int{
				shared.ResourceCredits: -input.Amount,
			})
			log.Info("💸 Deducted credits", zap.Int("amount", input.Amount))

		case shared.ResourceSteel:
			p.Resources().Add(map[shared.ResourceType]int{
				shared.ResourceSteel: -input.Amount,
			})
			log.Info("🔩 Deducted steel", zap.Int("amount", input.Amount))

		case shared.ResourceTitanium:
			p.Resources().Add(map[shared.ResourceType]int{
				shared.ResourceTitanium: -input.Amount,
			})
			log.Info("⚙️ Deducted titanium", zap.Int("amount", input.Amount))

		case shared.ResourcePlants:
			p.Resources().Add(map[shared.ResourceType]int{
				shared.ResourcePlants: -input.Amount,
			})
			log.Info("🌱 Deducted plants", zap.Int("amount", input.Amount))

		case shared.ResourceEnergy:
			p.Resources().Add(map[shared.ResourceType]int{
				shared.ResourceEnergy: -input.Amount,
			})
			log.Info("⚡ Deducted energy", zap.Int("amount", input.Amount))

		case shared.ResourceHeat:
			p.Resources().Add(map[shared.ResourceType]int{
				shared.ResourceHeat: -input.Amount,
			})
			log.Info("🔥 Deducted heat", zap.Int("amount", input.Amount))
		}
	}

	return nil
}

// applyOutputs applies output resources and effects
func (a *UseCardActionAction) applyOutputs(
	ctx context.Context,
	outputs []shared.ResourceCondition,
	p *player.Player,
	g *game.Game,
	cardName string,
	log *zap.Logger,
) error {
	if len(outputs) == 0 {
		log.Debug("No outputs to apply")
		return nil
	}

	log.Info("✨ Processing card action outputs", zap.Int("output_count", len(outputs)))

	for _, output := range outputs {
		switch output.ResourceType {
		// Basic resources
		case shared.ResourceCredits:
			p.Resources().Add(map[shared.ResourceType]int{
				shared.ResourceCredits: output.Amount,
			})
			log.Info("💰 Added credits", zap.Int("amount", output.Amount))

		case shared.ResourceSteel:
			p.Resources().Add(map[shared.ResourceType]int{
				shared.ResourceSteel: output.Amount,
			})
			log.Info("🔩 Added steel", zap.Int("amount", output.Amount))

		case shared.ResourceTitanium:
			p.Resources().Add(map[shared.ResourceType]int{
				shared.ResourceTitanium: output.Amount,
			})
			log.Info("⚙️ Added titanium", zap.Int("amount", output.Amount))

		case shared.ResourcePlants:
			p.Resources().Add(map[shared.ResourceType]int{
				shared.ResourcePlants: output.Amount,
			})
			log.Info("🌱 Added plants", zap.Int("amount", output.Amount))

		case shared.ResourceEnergy:
			p.Resources().Add(map[shared.ResourceType]int{
				shared.ResourceEnergy: output.Amount,
			})
			log.Info("⚡ Added energy", zap.Int("amount", output.Amount))

		case shared.ResourceHeat:
			p.Resources().Add(map[shared.ResourceType]int{
				shared.ResourceHeat: output.Amount,
			})
			log.Info("🔥 Added heat", zap.Int("amount", output.Amount))

		// Production resources
		case shared.ResourceCreditsProduction:
			p.Resources().AddProduction(map[shared.ResourceType]int{
				shared.ResourceCreditsProduction: output.Amount,
			})
			log.Info("💰 Added credits production", zap.Int("amount", output.Amount))

		case shared.ResourceSteelProduction:
			p.Resources().AddProduction(map[shared.ResourceType]int{
				shared.ResourceSteelProduction: output.Amount,
			})
			log.Info("🔩 Added steel production", zap.Int("amount", output.Amount))

		case shared.ResourceTitaniumProduction:
			p.Resources().AddProduction(map[shared.ResourceType]int{
				shared.ResourceTitaniumProduction: output.Amount,
			})
			log.Info("⚙️ Added titanium production", zap.Int("amount", output.Amount))

		case shared.ResourcePlantsProduction:
			p.Resources().AddProduction(map[shared.ResourceType]int{
				shared.ResourcePlantsProduction: output.Amount,
			})
			log.Info("🌱 Added plants production", zap.Int("amount", output.Amount))

		case shared.ResourceEnergyProduction:
			p.Resources().AddProduction(map[shared.ResourceType]int{
				shared.ResourceEnergyProduction: output.Amount,
			})
			log.Info("⚡ Added energy production", zap.Int("amount", output.Amount))

		case shared.ResourceHeatProduction:
			p.Resources().AddProduction(map[shared.ResourceType]int{
				shared.ResourceHeatProduction: output.Amount,
			})
			log.Info("🔥 Added heat production", zap.Int("amount", output.Amount))

		// Global parameters
		case shared.ResourceOxygen:
			actualSteps, err := g.GlobalParameters().IncreaseOxygen(ctx, output.Amount)
			if err != nil {
				log.Error("Failed to increase oxygen", zap.Error(err))
				return fmt.Errorf("failed to increase oxygen: %w", err)
			}
			log.Info("🌊 Increased oxygen", zap.Int("steps", actualSteps))

		case shared.ResourceTemperature:
			actualSteps, err := g.GlobalParameters().IncreaseTemperature(ctx, output.Amount)
			if err != nil {
				log.Error("Failed to increase temperature", zap.Error(err))
				return fmt.Errorf("failed to increase temperature: %w", err)
			}
			log.Info("🌡️ Increased temperature", zap.Int("steps", actualSteps))

		// Tile placements - append to queue for user selection
		case shared.ResourceCityPlacement, shared.ResourceGreeneryPlacement, shared.ResourceOceanPlacement:
			// Map resource type to tile type string
			var tileType string
			switch output.ResourceType {
			case shared.ResourceCityPlacement:
				tileType = "city"
			case shared.ResourceGreeneryPlacement:
				tileType = "greenery"
			case shared.ResourceOceanPlacement:
				tileType = "ocean"
			}

			// Build array of tile types to append (for multiple placements)
			tileTypes := make([]string, output.Amount)
			for i := 0; i < output.Amount; i++ {
				tileTypes[i] = tileType
			}

			// Atomically append to queue (thread-safe)
			if err := g.AppendToPendingTileSelectionQueue(ctx, p.ID(), tileTypes, cardName); err != nil {
				return fmt.Errorf("failed to append to pending tile selection queue: %w", err)
			}

			log.Info("🏗️ Added tile placements to queue",
				zap.String("tile_type", tileType),
				zap.Int("count", output.Amount),
				zap.String("source", cardName))

		default:
			log.Warn("⚠️ Unhandled output type in card action",
				zap.String("type", string(output.ResourceType)))
		}
	}

	return nil
}

// incrementPlayCount increments the play count for a card action
func (a *UseCardActionAction) incrementPlayCount(
	p *player.Player,
	cardID string,
	behaviorIndex int,
	log *zap.Logger,
) {
	actions := p.Actions().List()

	// Find and increment play count
	for i := range actions {
		if actions[i].CardID == cardID && actions[i].BehaviorIndex == behaviorIndex {
			actions[i].PlayCount++
			log.Debug("📊 Incremented action play count",
				zap.Int("new_count", actions[i].PlayCount))
			break
		}
	}

	// Update player actions
	p.Actions().SetActions(actions)
}
