package model

// Corporation represents a corporation card with special abilities
type Corporation struct {
	ID                 string      `json:"id" ts:"string"`
	Name               string      `json:"name" ts:"string"`
	Description        string      `json:"description" ts:"string"`
	StartingCredits    int         `json:"startingCredits" ts:"number"`
	StartingResources  ResourceSet `json:"startingResources" ts:"ResourceSet"`
	StartingProduction ResourceSet `json:"startingProduction" ts:"ResourceSet"`
	Tags               []CardTag   `json:"tags" ts:"CardTag[]"`
	SpecialEffects     []string    `json:"specialEffects" ts:"string[]"` // Descriptions of special abilities
}

// ConvertCardToCorporation converts a corporation Card to a Corporation struct
func ConvertCardToCorporation(card Card) Corporation {
	corp := Corporation{
		ID:                 card.ID,
		Name:               card.Name,
		Description:        card.Description,
		StartingCredits:    42, // Default starting credits
		StartingResources:  ResourceSet{Credits: 42},
		StartingProduction: ResourceSet{},
		Tags:               card.Tags,
		SpecialEffects:     []string{card.Description},
	}

	// Parse corporation-specific starting conditions from description
	// This is a simplified version - full implementation would parse the actual JSON effects
	if corp.Name == "Credicor" {
		corp.StartingCredits = 57
		corp.StartingResources.Credits = 57
	}

	return corp
}
