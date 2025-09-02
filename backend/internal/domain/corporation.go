package domain

// Corporation represents a corporation that a player can choose
type Corporation struct {
	ID                 string       `json:"id" ts:"string"`
	Name               string       `json:"name" ts:"string"`
	Description        string       `json:"description" ts:"string"`
	StartingMegaCredits int          `json:"startingMegaCredits" ts:"number"`
	StartingProduction *ResourcesMap `json:"startingProduction,omitempty" ts:"ResourcesMap | undefined"`
	StartingResources  *ResourcesMap `json:"startingResources,omitempty" ts:"ResourcesMap | undefined"`
	StartingTR         *int         `json:"startingTR,omitempty" ts:"number | undefined"`
	StartingCards      []string     `json:"startingCards" ts:"string[]"`
	Tags               []Tag        `json:"tags" ts:"Tag[]"`
	LogoPath           string       `json:"logoPath" ts:"string"`
	Color              string       `json:"color" ts:"string"`
	Ability            string       `json:"ability" ts:"string"`
}

// CorporationType represents the type identifier for corporations
type CorporationType string

const (
	CorporationTypeCredicor        CorporationType = "CREDICOR"
	CorporationTypeEcoline         CorporationType = "ECOLINE"
	CorporationTypeHelion          CorporationType = "HELION"
	CorporationTypeMiningGuild     CorporationType = "MINING_GUILD"
	CorporationTypeInventrix       CorporationType = "INVENTRIX"
	CorporationTypeThorgate        CorporationType = "THORGATE"
	CorporationTypeTharsisRepublic CorporationType = "THARSIS_REPUBLIC"
	CorporationTypePhobolog        CorporationType = "PHOBOLOG"
	CorporationTypeUnmi            CorporationType = "UNMI"
	CorporationTypeTeractor        CorporationType = "TERACTOR"
	CorporationTypeSaturnSystems   CorporationType = "SATURN_SYSTEMS"
	CorporationTypeInterplanetaryCinematics CorporationType = "INTERPLANETARY_CINEMATICS"
)

// GetBaseCorporations returns the base game corporations
func GetBaseCorporations() []Corporation {
	return []Corporation{
		{
			ID:                  string(CorporationTypeCredicor),
			Name:                "Credicor",
			Description:         "Having the largest credit rating in the solar system, Credicor can build the most ambitious projects.",
			StartingMegaCredits: 57,
			StartingProduction:  nil,
			StartingResources:   nil,
			StartingTR:          nil,
			StartingCards:       []string{},
			Tags:                []Tag{},
			LogoPath:           "/assets/corporations/credicor.png",
			Color:              "#4CAF50",
			Ability:            "Start with 57 M€",
		},
		{
			ID:                  string(CorporationTypeEcoline),
			Name:                "Ecoline",
			Description:         "Specialists in bioengineering of plants, Ecoline can terraform Mars using plant life.",
			StartingMegaCredits: 36,
			StartingProduction:  &ResourcesMap{Plants: 2},
			StartingResources:   &ResourcesMap{Plants: 3},
			StartingTR:          nil,
			StartingCards:       []string{},
			Tags:                []Tag{TagPlant},
			LogoPath:           "/assets/corporations/ecoline.png",
			Color:              "#8BC34A",
			Ability:            "Start with 2 plant production and 3 plants. You may always pay 7 plants to place a greenery tile.",
		},
		{
			ID:                  string(CorporationTypeHelion),
			Name:                "Helion",
			Description:         "Helion specializes in fusion power, making energy production more efficient.",
			StartingMegaCredits: 42,
			StartingProduction:  &ResourcesMap{Heat: 3},
			StartingResources:   nil,
			StartingTR:          nil,
			StartingCards:       []string{},
			Tags:                []Tag{TagPower},
			LogoPath:           "/assets/corporations/helion.png",
			Color:              "#FF5722",
			Ability:            "Start with 3 heat production. You may use heat as M€ with a 1:1 conversion rate.",
		},
		{
			ID:                  string(CorporationTypeMiningGuild),
			Name:                "Mining Guild",
			Description:         "Experienced in mining operations, this guild gains bonuses from steel and titanium production.",
			StartingMegaCredits: 30,
			StartingProduction:  &ResourcesMap{Steel: 1},
			StartingResources:   &ResourcesMap{Steel: 5},
			StartingTR:          nil,
			StartingCards:       []string{},
			Tags:                []Tag{TagBuilding},
			LogoPath:           "/assets/corporations/mining_guild.png",
			Color:              "#795548",
			Ability:            "Start with 1 steel production and 5 steel. When you increase production of steel or titanium, gain 1 M€.",
		},
		{
			ID:                  string(CorporationTypeInventrix),
			Name:                "Inventrix",
			Description:         "Inventrix is a technology company that starts with additional cards and card draw.",
			StartingMegaCredits: 45,
			StartingProduction:  nil,
			StartingResources:   nil,
			StartingTR:          nil,
			StartingCards:       []string{}, // Would be 3 additional cards in implementation
			Tags:                []Tag{TagScience},
			LogoPath:           "/assets/corporations/inventrix.png",
			Color:              "#2196F3",
			Ability:            "Start with 3 additional cards. Your hand limit is increased by 2.",
		},
		{
			ID:                  string(CorporationTypeThorgate),
			Name:                "Thorgate",
			Description:         "Thorgate specializes in power infrastructure and gets discounts on power-related cards.",
			StartingMegaCredits: 48,
			StartingProduction:  &ResourcesMap{Energy: 1},
			StartingResources:   nil,
			StartingTR:          nil,
			StartingCards:       []string{},
			Tags:                []Tag{TagPower},
			LogoPath:           "/assets/corporations/thorgate.png",
			Color:              "#FF9800",
			Ability:            "Start with 1 energy production. When you play a power tag, gain 3 M€.",
		},
		{
			ID:                  string(CorporationTypeTharsisRepublic),
			Name:                "Tharsis Republic",
			Description:         "Focused on city development, Tharsis Republic gains bonuses from city placement.",
			StartingMegaCredits: 40,
			StartingProduction:  &ResourcesMap{Credits: 1},
			StartingResources:   nil,
			StartingTR:          nil,
			StartingCards:       []string{},
			Tags:                []Tag{TagBuilding},
			LogoPath:           "/assets/corporations/tharsis_republic.png",
			Color:              "#9C27B0",
			Ability:            "Start with 1 M€ production. When you place a city tile, gain 3 M€.",
		},
		{
			ID:                  string(CorporationTypePhobolog),
			Name:                "Phobolog",
			Description:         "Phobolog specializes in space operations and titanium usage.",
			StartingMegaCredits: 23,
			StartingProduction:  nil,
			StartingResources:   &ResourcesMap{Titanium: 10},
			StartingTR:          nil,
			StartingCards:       []string{},
			Tags:                []Tag{TagSpace},
			LogoPath:           "/assets/corporations/phobolog.png",
			Color:              "#607D8B",
			Ability:            "Start with 10 titanium. Your titanium is worth 1 M€ extra when paying for space cards.",
		},
	}
}

// Milestone represents claimable milestones in the game
type Milestone struct {
	ID              string           `json:"id" ts:"string"`
	Name            string           `json:"name" ts:"string"`
	Description     string           `json:"description" ts:"string"`
	Cost            int              `json:"cost" ts:"number"`
	ClaimedBy       *string          `json:"claimedBy,omitempty" ts:"string | undefined"`
	AchievementType AchievementType  `json:"achievementType" ts:"AchievementType"`
	RequiredValue   int              `json:"requiredValue" ts:"number"`
}

// Award represents fundable awards in the game
type Award struct {
	ID              string          `json:"id" ts:"string"`
	Name            string          `json:"name" ts:"string"`
	Description     string          `json:"description" ts:"string"`
	Cost            int             `json:"cost" ts:"number"`
	FundedBy        *string         `json:"fundedBy,omitempty" ts:"string | undefined"`
	CompetitionType CompetitionType `json:"competitionType" ts:"CompetitionType"`
	Ranking         []AwardRanking  `json:"ranking" ts:"AwardRanking[]"`
}

// AchievementType defines what type of achievement is required for milestones
type AchievementType string

const (
	AchievementTypeTerraformRating AchievementType = "terraform_rating" // Terraformer: TR 35
	AchievementTypeCities          AchievementType = "cities"           // Mayor: 3 cities
	AchievementTypeGreeneries      AchievementType = "greeneries"      // Gardener: 3 greeneries
	AchievementTypeBuildings       AchievementType = "buildings"       // Builder: 8 building tags
	AchievementTypeCards           AchievementType = "cards"           // Planner: 16 cards in hand
)

// CompetitionType defines what type of competition awards are based on
type CompetitionType string

const (
	CompetitionTypeTileCount     CompetitionType = "tile_count"     // Landlord: most tiles
	CompetitionTypeCredits       CompetitionType = "credits"       // Banker: most M€
	CompetitionTypeScienceTags   CompetitionType = "science_tags"  // Scientist: most science tags
	CompetitionTypeHeatResource  CompetitionType = "heat_resource" // Thermalist: most heat resource
	CompetitionTypeSteelTitanium CompetitionType = "steel_titanium" // Miner: most steel and titanium
)

// AwardRanking represents a player's ranking in an award competition
type AwardRanking struct {
	PlayerID string `json:"playerId" ts:"string"`
	Value    int    `json:"value" ts:"number"`
	Rank     int    `json:"rank" ts:"number"`
}

// GetBaseMilestones returns the 5 base game milestones
func GetBaseMilestones() []Milestone {
	return []Milestone{
		{
			ID:              "terraformer",
			Name:            "Terraformer",
			Description:     "Have a terraform rating of at least 35",
			Cost:            8,
			AchievementType: AchievementTypeTerraformRating,
			RequiredValue:   35,
		},
		{
			ID:              "mayor",
			Name:            "Mayor", 
			Description:     "Own at least 3 city tiles",
			Cost:            8,
			AchievementType: AchievementTypeCities,
			RequiredValue:   3,
		},
		{
			ID:              "gardener",
			Name:            "Gardener",
			Description:     "Own at least 3 greenery tiles",
			Cost:            8,
			AchievementType: AchievementTypeGreeneries,
			RequiredValue:   3,
		},
		{
			ID:              "builder",
			Name:            "Builder",
			Description:     "Have at least 8 building tags in play",
			Cost:            8,
			AchievementType: AchievementTypeBuildings,
			RequiredValue:   8,
		},
		{
			ID:              "planner",
			Name:            "Planner",
			Description:     "Have at least 16 cards in hand",
			Cost:            8,
			AchievementType: AchievementTypeCards,
			RequiredValue:   16,
		},
	}
}

// GetBaseAwards returns the 5 base game awards
func GetBaseAwards() []Award {
	return []Award{
		{
			ID:              "landlord",
			Name:            "Landlord",
			Description:     "Own the most tiles on Mars",
			Cost:            8,
			CompetitionType: CompetitionTypeTileCount,
			Ranking:         []AwardRanking{},
		},
		{
			ID:              "banker",
			Name:            "Banker",
			Description:     "Have the most M€",
			Cost:            8,
			CompetitionType: CompetitionTypeCredits,
			Ranking:         []AwardRanking{},
		},
		{
			ID:              "scientist",
			Name:            "Scientist",
			Description:     "Have the most science tags in play",
			Cost:            8,
			CompetitionType: CompetitionTypeScienceTags,
			Ranking:         []AwardRanking{},
		},
		{
			ID:              "thermalist",
			Name:            "Thermalist",
			Description:     "Have the most heat resource",
			Cost:            8,
			CompetitionType: CompetitionTypeHeatResource,
			Ranking:         []AwardRanking{},
		},
		{
			ID:              "miner",
			Name:            "Miner",
			Description:     "Have the most steel and titanium resource",
			Cost:            8,
			CompetitionType: CompetitionTypeSteelTitanium,
			Ranking:         []AwardRanking{},
		},
	}
}