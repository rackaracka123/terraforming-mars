package production

// GameStatus represents the status of a game
type GameStatus string

const (
	GameStatusActive GameStatus = "active"
)

// GamePhase represents the current phase of the game
type GamePhase string

const (
	GamePhaseProductionAndCardDraw GamePhase = "production_and_card_draw"
	GamePhaseAction                GamePhase = "action"
)

// Game represents the minimal game state needed for production phase management
type Game struct {
	ID           string
	Status       GameStatus
	PlayerIDs    []string
	Generation   int
	CurrentPhase GamePhase
}

// Player represents player state needed for production phase
type Player struct {
	ID              string
	Resources       Resources
	Production      Production
	TerraformRating int
	ProductionPhase ProductionPhase
}

// Resources represents a player's resource state
type Resources struct {
	Credits  int
	Steel    int
	Titanium int
	Plants   int
	Energy   int
	Heat     int
}

// Production represents a player's production rates
type Production struct {
	Credits  int
	Steel    int
	Titanium int
	Plants   int
	Energy   int
	Heat     int
}

// ProductionPhase represents the production phase state for a player
type ProductionPhase struct {
	AvailableCards    []string
	SelectionComplete bool
	BeforeResources   Resources
	AfterResources    Resources
	EnergyConverted   int
	CreditsIncome     int
}

// DeepCopy creates a deep copy of Resources
func (r Resources) DeepCopy() Resources {
	return Resources{
		Credits:  r.Credits,
		Steel:    r.Steel,
		Titanium: r.Titanium,
		Plants:   r.Plants,
		Energy:   r.Energy,
		Heat:     r.Heat,
	}
}
