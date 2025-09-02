package domain

// BoardSpace represents a space on the Mars board
type BoardSpace struct {
	Position     HexCoordinate   `json:"position" ts:"HexCoordinate"`
	Type         SpaceType       `json:"type" ts:"SpaceType"`
	Bonus        []ResourceType  `json:"bonus" ts:"ResourceType[]"`
	IsOceanSpace bool            `json:"isOceanSpace" ts:"boolean"`
	IsReserved   bool            `json:"isReserved" ts:"boolean"`
	ReservedFor  *string         `json:"reservedFor,omitempty" ts:"string | undefined"`
	Tile         *Tile           `json:"tile,omitempty" ts:"Tile | undefined"`
	AdjacentTo   []HexCoordinate `json:"adjacentTo" ts:"HexCoordinate[]"`
}

// SpaceType represents the type of board space
type SpaceType string

const (
	SpaceTypeLand    SpaceType = "land"
	SpaceTypeOcean   SpaceType = "ocean"
	SpaceTypeCity    SpaceType = "city"     // Tharsis tholus, Ascraeus Mons, Pavonis Mons
	SpaceTypeVolcano SpaceType = "volcano"  // Olympus Mons, Arsia Mons
	SpaceTypeNoctis  SpaceType = "noctis"   // Noctis City
)

// PlacementRule represents rules for tile placement
type PlacementRule struct {
	TileType      TileType           `json:"tileType" ts:"TileType"`
	Requirements  []PlacementReq     `json:"requirements" ts:"PlacementReq[]"`
	Restrictions  []PlacementReq     `json:"restrictions" ts:"PlacementReq[]"`
	BonusRules    []PlacementBonus   `json:"bonusRules" ts:"PlacementBonus[]"`
}

// PlacementReq represents a placement requirement or restriction
type PlacementReq struct {
	Type         PlacementReqType `json:"type" ts:"PlacementReqType"`
	Target       PlacementTarget  `json:"target" ts:"PlacementTarget"`
	Distance     *int             `json:"distance,omitempty" ts:"number | undefined"`
	Count        *int             `json:"count,omitempty" ts:"number | undefined"`
	TileType     *TileType        `json:"tileType,omitempty" ts:"TileType | undefined"`
	SpaceType    *SpaceType       `json:"spaceType,omitempty" ts:"SpaceType | undefined"`
	PlayerID     *string          `json:"playerId,omitempty" ts:"string | undefined"`
}

// PlacementReqType defines types of placement requirements
type PlacementReqType string

const (
	PlacementReqTypeAdjacent     PlacementReqType = "adjacent"      // Must be adjacent to target
	PlacementReqTypeNotAdjacent  PlacementReqType = "not_adjacent"  // Must not be adjacent to target
	PlacementReqTypeDistance     PlacementReqType = "distance"      // Must be within/outside distance
	PlacementReqTypeOceanSpace   PlacementReqType = "ocean_space"   // Must be on ocean-reserved space
	PlacementReqTypeLandSpace    PlacementReqType = "land_space"    // Must be on land space
	PlacementReqTypeEmptySpace   PlacementReqType = "empty_space"   // Must be on empty space
	PlacementReqTypeOwned        PlacementReqType = "owned"         // Must be owned by player
	PlacementReqTypeReserved     PlacementReqType = "reserved"      // Must be on reserved space
)

// PlacementTarget defines what the placement requirement targets
type PlacementTarget string

const (
	PlacementTargetOwnTile    PlacementTarget = "own_tile"
	PlacementTargetAnyTile    PlacementTarget = "any_tile"
	PlacementTargetCity       PlacementTarget = "city"
	PlacementTargetGreenery   PlacementTarget = "greenery"
	PlacementTargetOcean      PlacementTarget = "ocean"
	PlacementTargetVolcano    PlacementTarget = "volcano"
	PlacementTargetBoardEdge  PlacementTarget = "board_edge"
)

// PlacementBonus represents bonuses gained from tile placement
type PlacementBonus struct {
	Condition    PlacementCondition `json:"condition" ts:"PlacementCondition"`
	Effect       CardEffect         `json:"effect" ts:"CardEffect"`
	Description  string             `json:"description" ts:"string"`
}

// PlacementCondition defines when placement bonuses are triggered
type PlacementCondition struct {
	Type         PlacementCondType `json:"type" ts:"PlacementCondType"`
	Target       PlacementTarget   `json:"target" ts:"PlacementTarget"`
	Count        *int              `json:"count,omitempty" ts:"number | undefined"`
	TileType     *TileType         `json:"tileType,omitempty" ts:"TileType | undefined"`
	MinDistance  *int              `json:"minDistance,omitempty" ts:"number | undefined"`
	MaxDistance  *int              `json:"maxDistance,omitempty" ts:"number | undefined"`
}

// PlacementCondType defines types of placement conditions
type PlacementCondType string

const (
	PlacementCondTypeAdjacentCount PlacementCondType = "adjacent_count"  // Number of adjacent tiles of type
	PlacementCondTypeSpaceBonus    PlacementCondType = "space_bonus"     // Bonus from the space itself
	PlacementCondTypeFirstTile     PlacementCondType = "first_tile"      // First tile of this type placed
)

// GetPlacementRules returns standard placement rules
func GetPlacementRules() []PlacementRule {
	return []PlacementRule{
		{
			TileType: TileTypeOcean,
			Requirements: []PlacementReq{
				{
					Type:      PlacementReqTypeOceanSpace,
					Target:    PlacementTargetOcean,
				},
			},
			Restrictions: []PlacementReq{},
			BonusRules: []PlacementBonus{
				{
					Condition: PlacementCondition{
						Type:   PlacementCondTypeSpaceBonus,
						Target: PlacementTargetOcean,
					},
					Effect: CardEffect{
						Type:   EffectTypeGainResource,
						Target: EffectTargetSelf,
						Amount: intPtr(2),
					},
					Description: "Gain bonus resources from ocean space",
				},
			},
		},
		{
			TileType: TileTypeGreenery,
			Requirements: []PlacementReq{
				{
					Type:   PlacementReqTypeLandSpace,
					Target: PlacementTargetGreenery,
				},
			},
			Restrictions: []PlacementReq{},
			BonusRules: []PlacementBonus{
				{
					Condition: PlacementCondition{
						Type:   PlacementCondTypeAdjacentCount,
						Target: PlacementTargetOwnTile,
						Count:  intPtr(1),
					},
					Effect: CardEffect{
						Type:   EffectTypeGainResource,
						Target: EffectTargetSelf,
						Amount: intPtr(1),
					},
					Description: "Gain 1 M€ for each adjacent tile you own",
				},
			},
		},
		{
			TileType: TileTypeCity,
			Requirements: []PlacementReq{
				{
					Type:   PlacementReqTypeLandSpace,
					Target: PlacementTargetCity,
				},
			},
			Restrictions: []PlacementReq{
				{
					Type:     PlacementReqTypeNotAdjacent,
					Target:   PlacementTargetCity,
					TileType: tileTypePtr(TileTypeCity),
				},
			},
			BonusRules: []PlacementBonus{
				{
					Condition: PlacementCondition{
						Type:     PlacementCondTypeAdjacentCount,
						Target:   PlacementTargetGreenery,
						TileType: tileTypePtr(TileTypeGreenery),
					},
					Effect: CardEffect{
						Type:   EffectTypeGainResource,
						Target: EffectTargetSelf,
						Amount: intPtr(1),
					},
					Description: "Gain 1 M€ for each adjacent greenery",
				},
			},
		},
	}
}

// GetBoardSpaces returns the standard Mars board layout
func GetBoardSpaces() []BoardSpace {
	spaces := []BoardSpace{}
	
	// This would contain the actual 61 hexagonal spaces of the Mars board
	// For now, just return a few example spaces
	spaces = append(spaces, BoardSpace{
		Position:     HexCoordinate{Q: 0, R: 0, S: 0},
		Type:         SpaceTypeLand,
		Bonus:        []ResourceType{ResourceTypeCredits, ResourceTypeCredits},
		IsOceanSpace: false,
		IsReserved:   false,
		AdjacentTo:   []HexCoordinate{},
	})
	
	spaces = append(spaces, BoardSpace{
		Position:     HexCoordinate{Q: 1, R: 0, S: -1},
		Type:         SpaceTypeOcean,
		Bonus:        []ResourceType{ResourceTypeSteel},
		IsOceanSpace: true,
		IsReserved:   false,
		AdjacentTo:   []HexCoordinate{{Q: 0, R: 0, S: 0}},
	})
	
	return spaces
}

// AdjacencyBonus represents bonuses from tile adjacency
type AdjacencyBonus struct {
	SourceTile  TileType     `json:"sourceTile" ts:"TileType"`
	TargetTile  TileType     `json:"targetTile" ts:"TileType"`
	Bonus       CardEffect   `json:"bonus" ts:"CardEffect"`
	Description string       `json:"description" ts:"string"`
}

// GetAdjacencyBonuses returns standard adjacency bonuses
func GetAdjacencyBonuses() []AdjacencyBonus {
	return []AdjacencyBonus{
		{
			SourceTile: TileTypeCity,
			TargetTile: TileTypeGreenery,
			Bonus: CardEffect{
				Type:         EffectTypeGainResource,
				Target:       EffectTargetSelf,
				Amount:       intPtr(1),
				ResourceType: resourceTypePtr(ResourceTypeCredits),
			},
			Description: "Cities gain 1 M€ for each adjacent greenery at game end",
		},
		{
			SourceTile: TileTypeGreenery,
			TargetTile: TileTypeCity,
			Bonus: CardEffect{
				Type:   EffectTypeGainVictoryPoints,
				Target: EffectTargetSelf,
				Amount: intPtr(1),
			},
			Description: "Greeneries are worth 1 VP each",
		},
	}
}