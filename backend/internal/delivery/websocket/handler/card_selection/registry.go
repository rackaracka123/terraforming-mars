package card_selection

import (
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/handler/card_selection/select_cards"
	"terraforming-mars-backend/internal/delivery/websocket/handler/card_selection/select_starting_card"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/service"
)

// SetupCardSelectionRegistry registers all card selection handlers with the registry
func SetupCardSelectionRegistry(
	cardService service.CardService,
	gameService service.GameService,
) *core.ActionRegistry {
	registry := core.NewActionRegistry()
	parser := utils.NewMessageParser()

	// Register card selection actions
	registry.Register(dto.ActionTypeSelectStartingCard, select_starting_card.NewHandler(cardService, gameService, parser))
	registry.Register(dto.ActionTypeSelectCards, select_cards.NewHandler(cardService, gameService, parser))

	return registry
}
