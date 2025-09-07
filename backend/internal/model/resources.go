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

// DeepCopy creates a deep copy of the Resources
func (r *Resources) DeepCopy() *Resources {
	if r == nil {
		return nil
	}

	return &Resources{
		Credits:  r.Credits,
		Steel:    r.Steel,
		Titanium: r.Titanium,
		Plants:   r.Plants,
		Energy:   r.Energy,
		Heat:     r.Heat,
	}
}

// DeepCopy creates a deep copy of the Production
func (p *Production) DeepCopy() *Production {
	if p == nil {
		return nil
	}

	return &Production{
		Credits:  p.Credits,
		Steel:    p.Steel,
		Titanium: p.Titanium,
		Plants:   p.Plants,
		Energy:   p.Energy,
		Heat:     p.Heat,
	}
}
