package shared

// Production represents a player's production values
type Production struct {
	Credits  int
	Steel    int
	Titanium int
	Plants   int
	Energy   int
	Heat     int
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
