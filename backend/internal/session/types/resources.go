package types

// Resources represents a player's resources
type Resources struct {
	Credits  int `json:"credits" ts:"number"`
	Steel    int `json:"steel" ts:"number"`
	Titanium int `json:"titanium" ts:"number"`
	Plants   int `json:"plants" ts:"number"`
	Energy   int `json:"energy" ts:"number"`
	Heat     int `json:"heat" ts:"number"`
}

// IsZero returns true if all resource values are zero
func (r Resources) IsZero() bool {
	return r.Credits == 0 && r.Steel == 0 && r.Titanium == 0 &&
		r.Plants == 0 && r.Energy == 0 && r.Heat == 0
}

// DeepCopy creates a deep copy of the Resources struct
func (r Resources) DeepCopy() Resources {
	return Resources{
		Credits:  r.Credits,
		Steel:    r.Steel,
		Titanium: r.Titanium,
		Plants:   r.Plants,
		Energy:   r.Energy,
		Heat:     r.Heat,
	}
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

// DeepCopy creates a deep copy of the Production struct
func (p Production) DeepCopy() Production {
	return Production{
		Credits:  p.Credits,
		Steel:    p.Steel,
		Titanium: p.Titanium,
		Plants:   p.Plants,
		Energy:   p.Energy,
		Heat:     p.Heat,
	}
}
