package shared

// AwardType represents the type of award
type AwardType string

// Award types available in the base game
const (
	AwardLandlord   AwardType = "landlord"   // Most tiles on Mars
	AwardBanker     AwardType = "banker"     // Most MC production
	AwardScientist  AwardType = "scientist"  // Most science tags
	AwardThermalist AwardType = "thermalist" // Most heat resources
	AwardMiner      AwardType = "miner"      // Most steel + titanium resources
)

// ValidAwardType returns true if the string is a valid award type
func ValidAwardType(s string) bool {
	switch AwardType(s) {
	case AwardLandlord, AwardBanker, AwardScientist, AwardThermalist, AwardMiner:
		return true
	default:
		return false
	}
}

// MilestoneType represents the type of milestone
type MilestoneType string

// Milestone types available in the base game
const (
	MilestoneTerraformer MilestoneType = "terraformer" // 35+ TR
	MilestoneMayor       MilestoneType = "mayor"       // 3+ cities
	MilestoneGardener    MilestoneType = "gardener"    // 3+ greenery tiles
	MilestoneBuilder     MilestoneType = "builder"     // 8+ building tags
	MilestonePlanner     MilestoneType = "planner"     // 16+ cards in hand
)

// ValidMilestoneType returns true if the string is a valid milestone type
func ValidMilestoneType(s string) bool {
	switch MilestoneType(s) {
	case MilestoneTerraformer, MilestoneMayor, MilestoneGardener, MilestoneBuilder, MilestonePlanner:
		return true
	default:
		return false
	}
}
