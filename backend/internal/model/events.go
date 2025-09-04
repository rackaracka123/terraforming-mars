package model

// Event type constants
const (
	EventTypeGameCreated                = "game.created"
	EventTypePlayerJoined               = "player.joined"
	EventTypeGameStarted                = "game.started"
	EventTypePlayerStartingCardOptions  = "player.starting_card_options"
	EventTypeStartingCardSelected       = "starting_card.selected"
	EventTypePhaseChanged               = "game.phase_changed"
	EventTypeGameUpdated                = "game.updated"
	EventTypeCardPlayed                 = "card.played"
	EventTypePlayerResourcesChanged     = "player.resources_changed"
	EventTypePlayerProductionChanged    = "player.production_changed"
)