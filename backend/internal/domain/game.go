package domain

import "time"

// GameState represents the complete state of a Terraforming Mars game
// @name GameState
type GameState struct {
	ID                  string            `json:"id" ts:"string"`
	Players             []Player          `json:"players" ts:"Player[]"`
	CurrentPlayer       string            `json:"currentPlayer" ts:"string"`
	Generation          int               `json:"generation" ts:"number"`
	Phase               GamePhase         `json:"phase" ts:"GamePhase"`
	GlobalParameters    GlobalParameters  `json:"globalParameters" ts:"GlobalParameters"`
	Milestones          []Milestone       `json:"milestones" ts:"Milestone[]"`
	Awards              []Award           `json:"awards" ts:"Award[]"`
	FirstPlayer         string            `json:"firstPlayer" ts:"string"`
	CorporationDraft    *bool             `json:"corporationDraft,omitempty" ts:"boolean | undefined"`
	Deck                []string          `json:"deck" ts:"string[]"`
	DiscardPile         []string          `json:"discardPile" ts:"string[]"`
	SoloMode            bool              `json:"soloMode" ts:"boolean"`
	Turn                int               `json:"turn" ts:"number"`
	DraftDirection      *DraftDirection   `json:"draftDirection,omitempty" ts:"DraftDirection | undefined"`
	GameSettings        GameSettings      `json:"gameSettings" ts:"GameSettings"`
	CurrentActionCount  *int              `json:"currentActionCount,omitempty" ts:"number | undefined"`
	MaxActionsPerTurn   *int              `json:"maxActionsPerTurn,omitempty" ts:"number | undefined"`
	CreatedAt           time.Time         `json:"createdAt" ts:"string"`
	UpdatedAt           time.Time         `json:"updatedAt" ts:"string"`
}

// GamePhase represents the current phase of the game
type GamePhase string

const (
	GamePhaseSetup               GamePhase = "setup"
	GamePhaseCorporationSelection GamePhase = "corporation_selection"
	GamePhaseInitialResearch     GamePhase = "initial_research"
	GamePhasePrelude             GamePhase = "prelude"
	GamePhaseResearch            GamePhase = "research"
	GamePhaseAction              GamePhase = "action"
	GamePhaseProduction          GamePhase = "production"
	GamePhaseDraft               GamePhase = "draft"
	GamePhaseGameEnd             GamePhase = "game_end"
)

// DraftDirection represents the direction cards are passed during draft
type DraftDirection string

const (
	DraftDirectionClockwise        DraftDirection = "clockwise"
	DraftDirectionCounterClockwise DraftDirection = "counter_clockwise"
)

// GameSettings contains all configurable game options
type GameSettings struct {
	Expansions                    []GameExpansion `json:"expansions" ts:"GameExpansion[]"`
	CorporateEra                  bool            `json:"corporateEra" ts:"boolean"`
	DraftVariant                  bool            `json:"draftVariant" ts:"boolean"`
	InitialDraft                  bool            `json:"initialDraft" ts:"boolean"`
	PreludeExtension              bool            `json:"preludeExtension" ts:"boolean"`
	VenusNextExtension            bool            `json:"venusNextExtension" ts:"boolean"`
	ColoniesExtension             bool            `json:"coloniesExtension" ts:"boolean"`
	TurmoilExtension              bool            `json:"turmoilExtension" ts:"boolean"`
	RemoveNegativeAttackCards     bool            `json:"removeNegativeAttackCards" ts:"boolean"`
	IncludeVenusMA                bool            `json:"includeVenusMA" ts:"boolean"`
	MoonExpansion                 bool            `json:"moonExpansion" ts:"boolean"`
	PathfindersExpansion          bool            `json:"pathfindersExpansion" ts:"boolean"`
	UnderworldExpansion           bool            `json:"underworldExpansion" ts:"boolean"`
	EscapeVelocityExpansion       bool            `json:"escapeVelocityExpansion" ts:"boolean"`
	Fast                          bool            `json:"fast" ts:"boolean"`
	ShowOtherPlayersVP            bool            `json:"showOtherPlayersVP" ts:"boolean"`
	CustomCorporationsList       *[]string       `json:"customCorporationsList,omitempty" ts:"string[] | undefined"`
	BannedCards                   *[]string       `json:"bannedCards,omitempty" ts:"string[] | undefined"`
	IncludedCards                 *[]string       `json:"includedCards,omitempty" ts:"string[] | undefined"`
	SoloTR                        bool            `json:"soloTR" ts:"boolean"`
	RandomFirstPlayer             bool            `json:"randomFirstPlayer" ts:"boolean"`
	RequiresVenusTrackCompletion  bool            `json:"requiresVenusTrackCompletion" ts:"boolean"`
	RequiresMoonTrackCompletion   bool            `json:"requiresMoonTrackCompletion" ts:"boolean"`
	MoonStandardProjectVariant    bool            `json:"moonStandardProjectVariant" ts:"boolean"`
	AltVenusBoard                 bool            `json:"altVenusBoard" ts:"boolean"`
	EscapeVelocityMode            bool            `json:"escapeVelocityMode" ts:"boolean"`
	EscapeVelocityThreshold       int             `json:"escapeVelocityThreshold" ts:"number"`
	EscapeVelocityPeriod          int             `json:"escapeVelocityPeriod" ts:"number"`
	EscapeVelocityPenalty         int             `json:"escapeVelocityPenalty" ts:"number"`
	TwoTempTerraformingThreshold  bool            `json:"twoTempTerraformingThreshold" ts:"boolean"`
	HeatFor                       bool            `json:"heatFor" ts:"boolean"`
	Breakthrough                  bool            `json:"breakthrough" ts:"boolean"`
}

// GameExpansion represents available game expansions
type GameExpansion string

const (
	GameExpansionPrelude        GameExpansion = "prelude"
	GameExpansionVenus          GameExpansion = "venus"
	GameExpansionColonies       GameExpansion = "colonies"
	GameExpansionTurmoil        GameExpansion = "turmoil"
	GameExpansionBigBox         GameExpansion = "big_box"
	GameExpansionAres           GameExpansion = "ares"
	GameExpansionMoon           GameExpansion = "moon"
	GameExpansionPathfinders    GameExpansion = "pathfinders"
	GameExpansionPrelude2       GameExpansion = "prelude2"
	GameExpansionCEO            GameExpansion = "ceo"
	GameExpansionPromo          GameExpansion = "promo"
	GameExpansionCommunity      GameExpansion = "community"
	GameExpansionUnderworld     GameExpansion = "underworld"
	GameExpansionEscapeVelocity GameExpansion = "escape_velocity"
	GameExpansionStarWars       GameExpansion = "star_wars"
)

// GameEvent represents events that occur during the game
type GameEvent struct {
	ID          string        `json:"id" ts:"string"`
	Type        GameEventType `json:"type" ts:"GameEventType"`
	TriggeredBy *string       `json:"triggeredBy,omitempty" ts:"string | undefined"`
	Data        interface{}   `json:"data,omitempty" ts:"any"`
	Timestamp   int64         `json:"timestamp" ts:"number"`
}

// GameEventType represents the type of game event
type GameEventType string

const (
	GameEventTypeGameStarted        GameEventType = "game_started"
	GameEventTypePlayerJoined       GameEventType = "player_joined"
	GameEventTypePlayerLeft         GameEventType = "player_left"
	GameEventTypeCardPlayed         GameEventType = "card_played"
	GameEventTypeTilePlaced         GameEventType = "tile_placed"
	GameEventTypeParameterIncreased GameEventType = "parameter_increased"
	GameEventTypeMilestoneClaimed   GameEventType = "milestone_claimed"
	GameEventTypeAwardFunded        GameEventType = "award_funded"
	GameEventTypeGenerationEnd      GameEventType = "generation_end"
	GameEventTypeGameEnd            GameEventType = "game_end"
	GameEventTypeProductionPhase    GameEventType = "production_phase"
	GameEventTypeResearchPhase      GameEventType = "research_phase"
)

// StandardProject represents standard projects available to all players
type StandardProject struct {
	ID          string `json:"id" ts:"string"`
	Name        string `json:"name" ts:"string"`
	Cost        int    `json:"cost" ts:"number"`
	Description string `json:"description" ts:"string"`
}

// StandardProjectType represents the type of standard project
type StandardProjectType string

const (
	StandardProjectTypeSellPatents  StandardProjectType = "sell_patents"
	StandardProjectTypePowerPlant   StandardProjectType = "power_plant"
	StandardProjectTypeAsteroid     StandardProjectType = "asteroid"
	StandardProjectTypeAquifer      StandardProjectType = "aquifer"
	StandardProjectTypeGreenery     StandardProjectType = "greenery"
	StandardProjectTypeCity         StandardProjectType = "city"
	StandardProjectTypeAirScrapping StandardProjectType = "air_scrapping"
	StandardProjectTypeBufferGas    StandardProjectType = "buffer_gas"
)