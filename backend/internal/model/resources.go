package model

// Resources represents a player's resources
type Resources struct {
	Credits  int `json:"credits" ts:"number"`
	Steel    int `json:"steel" ts:"number"`
	Titanium int `json:"titanium" ts:"number"`
	Plants   int `json:"plants" ts:"number"`
	Energy   int `json:"energy" ts:"number"`
	Heat     int `json:"heat" ts:"number"`
}

// Production represents a player's production values
type Production struct {
	Credits  int `json:"credits" ts:"number"`
	Steel    int `json:"steel" ts:"number"`
	Titanium int `json:"titanium" ts:"number"`
	Plants   int `json:"plants" ts:"number"`
	Energy   int `json:"energy" ts:"number"`
	Heat     int `json:"heat" ts:"number"`
}

// Add adds two resource sets together
func (r Resources) Add(other Resources) Resources {
	return Resources{
		Credits:  r.Credits + other.Credits,
		Steel:    r.Steel + other.Steel,
		Titanium: r.Titanium + other.Titanium,
		Plants:   r.Plants + other.Plants,
		Energy:   r.Energy + other.Energy,
		Heat:     r.Heat + other.Heat,
	}
}

// Subtract subtracts one resource set from another
func (r Resources) Subtract(other Resources) Resources {
	return Resources{
		Credits:  r.Credits - other.Credits,
		Steel:    r.Steel - other.Steel,
		Titanium: r.Titanium - other.Titanium,
		Plants:   r.Plants - other.Plants,
		Energy:   r.Energy - other.Energy,
		Heat:     r.Heat - other.Heat,
	}
}

// HasNegative checks if any resource values are negative
func (r Resources) HasNegative() bool {
	return r.Credits < 0 || r.Steel < 0 || r.Titanium < 0 ||
		r.Plants < 0 || r.Energy < 0 || r.Heat < 0
}

// CanAfford checks if current resources can afford the given cost
func (r Resources) CanAfford(cost Resources) bool {
	return r.Credits >= cost.Credits &&
		r.Steel >= cost.Steel &&
		r.Titanium >= cost.Titanium &&
		r.Plants >= cost.Plants &&
		r.Energy >= cost.Energy &&
		r.Heat >= cost.Heat
}

// Add adds two production sets together
func (p Production) Add(other Production) Production {
	return Production{
		Credits:  p.Credits + other.Credits,
		Steel:    p.Steel + other.Steel,
		Titanium: p.Titanium + other.Titanium,
		Plants:   p.Plants + other.Plants,
		Energy:   p.Energy + other.Energy,
		Heat:     p.Heat + other.Heat,
	}
}

// Subtract subtracts one production set from another
func (p Production) Subtract(other Production) Production {
	return Production{
		Credits:  p.Credits - other.Credits,
		Steel:    p.Steel - other.Steel,
		Titanium: p.Titanium - other.Titanium,
		Plants:   p.Plants - other.Plants,
		Energy:   p.Energy - other.Energy,
		Heat:     p.Heat - other.Heat,
	}
}
