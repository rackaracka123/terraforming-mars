package model

// ConnectionStatus represents the connection status of a player
type ConnectionStatus string

const (
	ConnectionStatusConnected    ConnectionStatus = "connected"
	ConnectionStatusDisconnected ConnectionStatus = "disconnected"
)

// Player represents a player in the game
type Player struct {
	ID               string           `json:"id" ts:"string"`
	Name             string           `json:"name" ts:"string"`
	Corporation      string           `json:"corporation" ts:"string"`
	Cards            []string         `json:"cards" ts:"string[]"`
	Resources        Resources        `json:"resources" ts:"Resources"`
	Production       Production       `json:"production" ts:"Production"`
	TerraformRating  int              `json:"terraformRating" ts:"number"`
	IsActive         bool             `json:"isActive" ts:"boolean"`
	PlayedCards      []string         `json:"playedCards" ts:"string[]"`
	Passed           bool             `json:"passed" ts:"boolean"`
	AvailableActions int              `json:"availableActions" ts:"number"`
	VictoryPoints    int              `json:"victoryPoints" ts:"number"`
	MilestoneIcon    string           `json:"milestoneIcon" ts:"string"`
	ConnectionStatus ConnectionStatus `json:"connectionStatus" ts:"ConnectionStatus"`
}

// DeepCopy creates a deep copy of the Player
func (p *Player) DeepCopy() *Player {
	if p == nil {
		return nil
	}

	// Copy cards slice
	cardsCopy := make([]string, len(p.Cards))
	copy(cardsCopy, p.Cards)

	// Copy played cards slice
	playedCardsCopy := make([]string, len(p.PlayedCards))
	copy(playedCardsCopy, p.PlayedCards)

	return &Player{
		ID:               p.ID,
		Name:             p.Name,
		Corporation:      p.Corporation,
		Cards:            cardsCopy,
		Resources:        *p.Resources.DeepCopy(),
		Production:       *p.Production.DeepCopy(),
		TerraformRating:  p.TerraformRating,
		IsActive:         p.IsActive,
		PlayedCards:      playedCardsCopy,
		Passed:           p.Passed,
		AvailableActions: p.AvailableActions,
		VictoryPoints:    p.VictoryPoints,
		MilestoneIcon:    p.MilestoneIcon,
		ConnectionStatus: p.ConnectionStatus,
	}
}
