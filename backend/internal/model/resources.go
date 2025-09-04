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