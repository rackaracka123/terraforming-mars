package fixtures

import (
	"github.com/google/uuid"
	"terraforming-mars-backend/internal/player"
)

// PlayerOption is a function that modifies a test player
type PlayerOption func(*player.Player)

// NewTestPlayer creates a new player for testing with optional modifications
func NewTestPlayer(options ...PlayerOption) *player.Player {
	p := &player.Player{
		ID:              uuid.New().String(),
		Name:            "Test Player",
		TerraformRating: 20,
		VictoryPoints:   0,
		Resources:       player.Resources{},
		Production:      player.Production{},
		Cards:           []player.Card{},
		PlayedCards:     []player.Card{},
		Corporation:     nil,
		Effects:         []player.PlayerEffect{},
		Actions:         []player.PlayerAction{},
		ResourceStorage: map[string]int{},
	}

	for _, option := range options {
		option(p)
	}

	return p
}

// WithID sets a specific player ID
func WithID(id string) PlayerOption {
	return func(p *player.Player) {
		p.ID = id
	}
}

// WithName sets the player name
func WithName(name string) PlayerOption {
	return func(p *player.Player) {
		p.Name = name
	}
}

// WithTR sets the terraform rating
func WithTR(tr int) PlayerOption {
	return func(p *player.Player) {
		p.TerraformRating = tr
	}
}

// WithVP sets the victory points
func WithVP(vp int) PlayerOption {
	return func(p *player.Player) {
		p.VictoryPoints = vp
	}
}

// WithResources sets specific resource amounts
func WithResources(credits, steel, titanium, plants, energy, heat int) PlayerOption {
	return func(p *player.Player) {
		p.Resources = player.Resources{
			Credits:  credits,
			Steel:    steel,
			Titanium: titanium,
			Plants:   plants,
			Energy:   energy,
			Heat:     heat,
		}
	}
}

// WithCredits sets only credits
func WithCredits(amount int) PlayerOption {
	return func(p *player.Player) {
		p.Resources.Credits = amount
	}
}

// WithSteel sets only steel
func WithSteel(amount int) PlayerOption {
	return func(p *player.Player) {
		p.Resources.Steel = amount
	}
}

// WithTitanium sets only titanium
func WithTitanium(amount int) PlayerOption {
	return func(p *player.Player) {
		p.Resources.Titanium = amount
	}
}

// WithPlants sets only plants
func WithPlants(amount int) PlayerOption {
	return func(p *player.Player) {
		p.Resources.Plants = amount
	}
}

// WithEnergy sets only energy
func WithEnergy(amount int) PlayerOption {
	return func(p *player.Player) {
		p.Resources.Energy = amount
	}
}

// WithHeat sets only heat
func WithHeat(amount int) PlayerOption {
	return func(p *player.Player) {
		p.Resources.Heat = amount
	}
}

// WithProduction sets specific production amounts
func WithProduction(credits, steel, titanium, plants, energy, heat int) PlayerOption {
	return func(p *player.Player) {
		p.Production = player.Production{
			Credits:  credits,
			Steel:    steel,
			Titanium: titanium,
			Plants:   plants,
			Energy:   energy,
			Heat:     heat,
		}
	}
}

// WithCreditsProduction sets only credits production
func WithCreditsProduction(amount int) PlayerOption {
	return func(p *player.Player) {
		p.Production.Credits = amount
	}
}

// WithEnergyProduction sets only energy production
func WithEnergyProduction(amount int) PlayerOption {
	return func(p *player.Player) {
		p.Production.Energy = amount
	}
}

// WithCards adds cards to hand
func WithCards(cardIDs ...string) PlayerOption {
	return func(p *player.Player) {
		for _, id := range cardIDs {
			p.Cards = append(p.Cards, player.Card{
				ID:   id,
				Name: "Test Card " + id,
			})
		}
	}
}

// WithPlayedCards adds played cards
func WithPlayedCards(cardIDs ...string) PlayerOption {
	return func(p *player.Player) {
		for _, id := range cardIDs {
			p.PlayedCards = append(p.PlayedCards, player.Card{
				ID:   id,
				Name: "Test Card " + id,
			})
		}
	}
}

// WithCorporation sets the corporation
func WithCorporation(corpID, corpName string) PlayerOption {
	return func(p *player.Player) {
		p.Corporation = &player.Card{
			ID:   corpID,
			Name: corpName,
		}
	}
}

// WithResourceStorage adds resource storage
func WithResourceStorage(resourceType string, amount int) PlayerOption {
	return func(p *player.Player) {
		if p.ResourceStorage == nil {
			p.ResourceStorage = make(map[string]int)
		}
		p.ResourceStorage[resourceType] = amount
	}
}
