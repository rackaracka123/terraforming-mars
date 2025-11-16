package player

import (
	"terraforming-mars-backend/internal/domain"
)

// ============================================================================
// RESOURCE CONVERSION UTILITIES
// ============================================================================

// ToResourceSet converts player Resources to domain.ResourceSet
func (r Resources) ToResourceSet() domain.ResourceSet {
	return domain.ResourceSet{
		Credits:  r.Credits,
		Steel:    r.Steel,
		Titanium: r.Titanium,
		Plants:   r.Plants,
		Energy:   r.Energy,
		Heat:     r.Heat,
	}
}

// FromResourceSet converts domain.ResourceSet to player Resources
func FromResourceSet(rs domain.ResourceSet) Resources {
	return Resources{
		Credits:  rs.Credits,
		Steel:    rs.Steel,
		Titanium: rs.Titanium,
		Plants:   rs.Plants,
		Energy:   rs.Energy,
		Heat:     rs.Heat,
	}
}

// ToResourceSet converts player Production to domain.ResourceSet
func (p Production) ToResourceSet() domain.ResourceSet {
	return domain.ResourceSet{
		Credits:  p.Credits,
		Steel:    p.Steel,
		Titanium: p.Titanium,
		Plants:   p.Plants,
		Energy:   p.Energy,
		Heat:     p.Heat,
	}
}

// FromResourceSetToProduction converts domain.ResourceSet to player Production
func FromResourceSetToProduction(rs domain.ResourceSet) Production {
	return Production{
		Credits:  rs.Credits,
		Steel:    rs.Steel,
		Titanium: rs.Titanium,
		Plants:   rs.Plants,
		Energy:   rs.Energy,
		Heat:     rs.Heat,
	}
}

// CanAfford checks if player resources can afford a cost (with optional discounts)
func (r Resources) CanAfford(cost domain.ResourceSet) bool {
	return r.Credits >= cost.Credits &&
		r.Steel >= cost.Steel &&
		r.Titanium >= cost.Titanium &&
		r.Plants >= cost.Plants &&
		r.Energy >= cost.Energy &&
		r.Heat >= cost.Heat
}

// Add adds resources from a ResourceSet to player Resources
func (r *Resources) Add(rs domain.ResourceSet) {
	r.Credits += rs.Credits
	r.Steel += rs.Steel
	r.Titanium += rs.Titanium
	r.Plants += rs.Plants
	r.Energy += rs.Energy
	r.Heat += rs.Heat
}

// Subtract subtracts resources from a ResourceSet from player Resources
func (r *Resources) Subtract(rs domain.ResourceSet) {
	r.Credits -= rs.Credits
	r.Steel -= rs.Steel
	r.Titanium -= rs.Titanium
	r.Plants -= rs.Plants
	r.Energy -= rs.Energy
	r.Heat -= rs.Heat
}

// Add adds production from a ResourceSet to player Production
func (p *Production) Add(rs domain.ResourceSet) {
	p.Credits += rs.Credits
	p.Steel += rs.Steel
	p.Titanium += rs.Titanium
	p.Plants += rs.Plants
	p.Energy += rs.Energy
	p.Heat += rs.Heat
}

// Subtract subtracts production from a ResourceSet from player Production
func (p *Production) Subtract(rs domain.ResourceSet) {
	p.Credits -= rs.Credits
	p.Steel -= rs.Steel
	p.Titanium -= rs.Titanium
	p.Plants -= rs.Plants
	p.Energy -= rs.Energy
	p.Heat -= rs.Heat
}
