package usecase

import (
	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"
	"terraforming-mars-backend/internal/usecase/common"
)

// UseCase contains all business logic for the application
type UseCase struct {
	actionValidator common.ActionValidator
	gameRepo        repository.GameRepository
	playerRepo      repository.PlayerRepository
	gameService     service.GameService
	sessionManager  session.SessionManager
}

// NewUseCase creates a new UseCase instance with all dependencies
func NewUseCase(
	actionValidator common.ActionValidator,
	gameRepo repository.GameRepository,
	playerRepo repository.PlayerRepository,
	gameService service.GameService,
	sessionManager session.SessionManager,
) *UseCase {
	return &UseCase{
		actionValidator: actionValidator,
		gameRepo:        gameRepo,
		playerRepo:      playerRepo,
		gameService:     gameService,
		sessionManager:  sessionManager,
	}
}
