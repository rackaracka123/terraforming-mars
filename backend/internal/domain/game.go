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
	TurnPhase           *TurnPhase        `json:"turnPhase,omitempty" ts:"TurnPhase | undefined"`
	GlobalParameters    GlobalParameters  `json:"globalParameters" ts:"GlobalParameters"`
	EndGameConditions   []EndGameCondition `json:"endGameConditions" ts:"EndGameCondition[]"`
	Board               []BoardSpace      `json:"board" ts:"BoardSpace[]"`
	Milestones          []Milestone       `json:"milestones" ts:"Milestone[]"`
	Awards              []Award           `json:"awards" ts:"Award[]"`
	AvailableStandardProjects []StandardProject `json:"availableStandardProjects" ts:"StandardProject[]"`
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
	ActionHistory       []Action          `json:"actionHistory" ts:"Action[]"`
	Events              []GameEvent       `json:"events" ts:"GameEvent[]"`
	IsGameEnded         bool              `json:"isGameEnded" ts:"boolean"`
	WinnerID            *string           `json:"winnerId,omitempty" ts:"string | undefined"`
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
	GamePhaseGeneration          GamePhase = "generation"  // Main game phase
	GamePhaseGameEnd             GamePhase = "game_end"
	GamePhaseFinalScoring        GamePhase = "final_scoring"
)

// TurnPhase represents the phase within each generation
type TurnPhase string

const (
	TurnPhasePlayerOrder  TurnPhase = "player_order"   // Determine turn order
	TurnPhaseResearch     TurnPhase = "research"       // Buy cards phase  
	TurnPhaseAction       TurnPhase = "action"         // Main action phase
	TurnPhaseProduction   TurnPhase = "production"     // Resource production
	TurnPhaseDraft        TurnPhase = "draft"          // Card drafting (optional)
	TurnPhaseGeneration   TurnPhase = "generation_end" // End of generation cleanup
)

// TurnOrder represents player order for the current generation
type TurnOrder struct {
	Generation   int      `json:"generation" ts:"number"`
	PlayerOrder  []string `json:"playerOrder" ts:"string[]"`
	CurrentIndex int      `json:"currentIndex" ts:"number"`
}

// GenerationPhaseConfig represents configuration for each generation phase
type GenerationPhaseConfig struct {
	Phase               TurnPhase `json:"phase" ts:"TurnPhase"`
	Description         string    `json:"description" ts:"string"`
	IsSimultaneous      bool      `json:"isSimultaneous" ts:"boolean"`      // All players act simultaneously
	RequiresAllPassed   bool      `json:"requiresAllPassed" ts:"boolean"`   // Phase ends when all players pass
	HasActionLimit      bool      `json:"hasActionLimit" ts:"boolean"`      // Phase has action limits
	MaxActionsPerPlayer *int      `json:"maxActionsPerPlayer,omitempty" ts:"number | undefined"`
}

// GetGenerationPhases returns the phases that occur in each generation
func GetGenerationPhases() []GenerationPhaseConfig {
	return []GenerationPhaseConfig{
		{
			Phase:               TurnPhasePlayerOrder,
			Description:         "Determine player order for this generation",
			IsSimultaneous:      true,
			RequiresAllPassed:   false,
			HasActionLimit:      false,
		},
		{
			Phase:               TurnPhaseResearch,
			Description:         "Players buy cards from the market",
			IsSimultaneous:      true,
			RequiresAllPassed:   false,
			HasActionLimit:      false,
		},
		{
			Phase:               TurnPhaseAction,
			Description:         "Players take turns performing 1-2 actions",
			IsSimultaneous:      false,
			RequiresAllPassed:   true,
			HasActionLimit:      true,
			MaxActionsPerPlayer: intPtr(2),
		},
		{
			Phase:               TurnPhaseProduction,
			Description:         "All players receive production and resources",
			IsSimultaneous:      true,
			RequiresAllPassed:   false,
			HasActionLimit:      false,
		},
	}
}

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


// Action represents a player action that can be taken during the game
type Action struct {
	ID           string       `json:"id" ts:"string"`
	Type         ActionType   `json:"type" ts:"ActionType"`
	PlayerID     string       `json:"playerId" ts:"string"`
	CardID       *string      `json:"cardId,omitempty" ts:"string | undefined"`
	ProjectID    *string      `json:"projectId,omitempty" ts:"string | undefined"`
	Position     *HexCoordinate `json:"position,omitempty" ts:"HexCoordinate | undefined"`
	Resources    *ResourcesMap `json:"resources,omitempty" ts:"ResourcesMap | undefined"`
	Target       *string      `json:"target,omitempty" ts:"string | undefined"`
	Data         interface{}  `json:"data,omitempty" ts:"any"`
}

// ActionType represents the different types of actions a player can take
type ActionType string

const (
	ActionTypePlayCard          ActionType = "play_card"
	ActionTypeStandardProject   ActionType = "standard_project"
	ActionTypeClaimMilestone    ActionType = "claim_milestone"
	ActionTypeFundAward         ActionType = "fund_award"
	ActionTypeUseCardAction     ActionType = "use_card_action"
	ActionTypeConvertPlants     ActionType = "convert_plants"
	ActionTypeConvertHeat       ActionType = "convert_heat"
	ActionTypePlaceTile         ActionType = "place_tile"
	ActionTypeTradeWithColony   ActionType = "trade_with_colony"
	ActionTypePass              ActionType = "pass"
)

// StandardProject represents standard projects available to all players
type StandardProject struct {
	ID           string              `json:"id" ts:"string"`
	Name         string              `json:"name" ts:"string"`
	Type         StandardProjectType `json:"type" ts:"StandardProjectType"`
	Cost         int                 `json:"cost" ts:"number"`
	Requirements []Requirement       `json:"requirements" ts:"Requirement[]"`
	Effects      []CardEffect        `json:"effects" ts:"CardEffect[]"`
	Description  string              `json:"description" ts:"string"`
	IsRepeatable bool                `json:"isRepeatable" ts:"boolean"`
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

// GetStandardProjects returns all available standard projects
func GetStandardProjects() []StandardProject {
	return []StandardProject{
		{
			ID:   string(StandardProjectTypeSellPatents),
			Name: "Sell Patents",
			Type: StandardProjectTypeSellPatents,
			Cost: 0,
			Requirements: []Requirement{},
			Effects: []CardEffect{
				{
					Type:         EffectTypeGainResource,
					Target:       EffectTargetSelf,
					Amount:       intPtr(1),
					ResourceType: resourceTypePtr(ResourceTypeCredits),
				},
			},
			Description:  "Gain 1 M€ for each card in your hand",
			IsRepeatable: true,
		},
		{
			ID:   string(StandardProjectTypePowerPlant),
			Name: "Power Plant",
			Type: StandardProjectTypePowerPlant,
			Cost: 11,
			Requirements: []Requirement{},
			Effects: []CardEffect{
				{
					Type:         EffectTypeGainProduction,
					Target:       EffectTargetSelf,
					Amount:       intPtr(1),
					ResourceType: resourceTypePtr(ResourceTypeEnergy),
				},
			},
			Description:  "Increase your energy production 1 step",
			IsRepeatable: true,
		},
		{
			ID:   string(StandardProjectTypeAsteroid),
			Name: "Asteroid",
			Type: StandardProjectTypeAsteroid,
			Cost: 14,
			Requirements: []Requirement{},
			Effects: []CardEffect{
				{
					Type:      EffectTypeIncreaseParameter,
					Target:    EffectTargetGlobal,
					Amount:    intPtr(1),
				},
			},
			Description:  "Raise temperature 1 step",
			IsRepeatable: true,
		},
		{
			ID:   string(StandardProjectTypeAquifer),
			Name: "Aquifer",
			Type: StandardProjectTypeAquifer,
			Cost: 18,
			Requirements: []Requirement{},
			Effects: []CardEffect{
				{
					Type:     EffectTypePlaceTile,
					Target:   EffectTargetGlobal,
					TileType: tileTypePtr(TileTypeOcean),
				},
			},
			Description:  "Place 1 ocean tile",
			IsRepeatable: true,
		},
		{
			ID:   string(StandardProjectTypeGreenery),
			Name: "Greenery",
			Type: StandardProjectTypeGreenery,
			Cost: 23,
			Requirements: []Requirement{},
			Effects: []CardEffect{
				{
					Type:     EffectTypePlaceTile,
					Target:   EffectTargetSelf,
					TileType: tileTypePtr(TileTypeGreenery),
				},
				{
					Type:      EffectTypeIncreaseParameter,
					Target:    EffectTargetGlobal,
					Amount:    intPtr(1),
				},
			},
			Description:  "Place a greenery tile and raise oxygen 1 step",
			IsRepeatable: true,
		},
		{
			ID:   string(StandardProjectTypeCity),
			Name: "City",
			Type: StandardProjectTypeCity,
			Cost: 25,
			Requirements: []Requirement{},
			Effects: []CardEffect{
				{
					Type:     EffectTypePlaceTile,
					Target:   EffectTargetSelf,
					TileType: tileTypePtr(TileTypeCity),
				},
				{
					Type:         EffectTypeGainProduction,
					Target:       EffectTargetSelf,
					Amount:       intPtr(1),
					ResourceType: resourceTypePtr(ResourceTypeCredits),
				},
			},
			Description:  "Place a city tile and increase M€ production 1 step",
			IsRepeatable: true,
		},
		{
			ID:   string(StandardProjectTypeAirScrapping),
			Name: "Air Scrapping",
			Type: StandardProjectTypeAirScrapping,
			Cost: 15,
			Requirements: []Requirement{
				{
					Type:      RequirementTypeOxygen,
					Parameter: globalParamPtr(GlobalParamOxygen),
					MinValue:  intPtr(6),
				},
			},
			Effects: []CardEffect{
				{
					Type:      EffectTypeIncreaseParameter,
					Target:    EffectTargetGlobal,
					Amount:    intPtr(1),
				},
			},
			Description:  "Raise Venus scale 1 step. Requires 6% oxygen",
			IsRepeatable: true,
		},
	}
}

// Utility functions for pointer creation
func intPtr(i int) *int {
	return &i
}

func resourceTypePtr(rt ResourceType) *ResourceType {
	return &rt
}

func tileTypePtr(tt TileType) *TileType {
	return &tt
}

func globalParamPtr(gp GlobalParam) *GlobalParam {
	return &gp
}