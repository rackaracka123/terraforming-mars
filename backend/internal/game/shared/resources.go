package shared

// Resources represents a player's resources
type Resources struct {
	Credits  int
	Steel    int
	Titanium int
	Plants   int
	Energy   int
	Heat     int
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
