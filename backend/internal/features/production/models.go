package production

// ProductionPhaseState tracks card selection state during production phase
type ProductionPhaseState struct {
	AvailableCards    []string // Card IDs available for selection
	SelectionComplete bool     // Whether player has completed their selection
}
