package resources

// Resources represents a player's resource state
type Resources struct {
	Credits  int
	Steel    int
	Titanium int
	Plants   int
	Energy   int
	Heat     int
}

// Production represents a player's production rates
type Production struct {
	Credits  int
	Steel    int
	Titanium int
	Plants   int
	Energy   int
	Heat     int
}

// ResourceSet represents a set of resource amounts (for costs, gains, etc.)
type ResourceSet struct {
	Credits  int
	Steel    int
	Titanium int
	Plants   int
	Energy   int
	Heat     int
}
