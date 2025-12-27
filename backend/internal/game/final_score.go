package game

// FinalScore represents a player's final score with VP breakdown
type FinalScore struct {
	PlayerID   string
	PlayerName string
	Breakdown  VPBreakdown
	Credits    int // For tiebreaker
	Placement  int // 1st, 2nd, 3rd, etc.
	IsWinner   bool
}

// CardVPConditionDetail represents the detailed calculation of a single VP condition
type CardVPConditionDetail struct {
	ConditionType  string
	Amount         int
	Count          int
	MaxTrigger     *int
	ActualTriggers int
	TotalVP        int
	Explanation    string
}

// CardVPDetail represents VP calculation for a single card
type CardVPDetail struct {
	CardID     string
	CardName   string
	Conditions []CardVPConditionDetail
	TotalVP    int
}

// GreeneryVPDetail represents VP from a single greenery tile
type GreeneryVPDetail struct {
	Coordinate string // Format: "q,r,s"
	VP         int    // Always 1 per greenery
}

// CityVPDetail represents VP from a single city tile and its adjacent greeneries
type CityVPDetail struct {
	CityCoordinate     string   // Format: "q,r,s"
	AdjacentGreeneries []string // Coordinates of adjacent greenery tiles
	VP                 int      // Number of adjacent greeneries
}

// VPBreakdown contains the detailed breakdown of a player's victory points
type VPBreakdown struct {
	TerraformRating   int
	CardVP            int
	CardVPDetails     []CardVPDetail
	MilestoneVP       int
	AwardVP           int
	GreeneryVP        int
	GreeneryVPDetails []GreeneryVPDetail
	CityVP            int
	CityVPDetails     []CityVPDetail
	TotalVP           int
}
