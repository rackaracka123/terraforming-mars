package corporation_test

import (
	"testing"

	"terraforming-mars-backend/internal/session/game/card"
	"terraforming-mars-backend/internal/session/game/player/corporation"
)

import (
	"testing"

	"terraforming-mars-backend/internal/session/game/card"
)

func TestNewCorporation(t *testing.T) {
	corp := NewCorporation()

	if corp.Card() != nil {
		t.Error("Expected new corporation to have no card")
	}
	if corp.ID() != "" {
		t.Error("Expected new corporation to have empty ID")
	}
	if corp.HasCorporation() {
		t.Error("Expected HasCorporation() to return false")
	}
}

func TestCorporation_SetID(t *testing.T) {
	corp := NewCorporation()

	corp.SetID("credicor")
	if corp.ID() != "credicor" {
		t.Errorf("Expected ID to be 'credicor', got '%s'", corp.ID())
	}
}

func TestCorporation_SetCard(t *testing.T) {
	corp := NewCorporation()

	testCard := card.Card{
		ID:   "credicor",
		Name: "Credicor",
		Cost: 0,
		Type: "corporation",
	}

	corp.SetCard(testCard)

	if !corp.HasCorporation() {
		t.Error("Expected HasCorporation() to return true after setting card")
	}
	if corp.ID() != "credicor" {
		t.Errorf("Expected ID to be 'credicor', got '%s'", corp.ID())
	}

	retrievedCard := corp.Card()
	if retrievedCard == nil {
		t.Fatal("Expected card to not be nil")
	}
	if retrievedCard.ID != "credicor" {
		t.Errorf("Expected card ID to be 'credicor', got '%s'", retrievedCard.ID)
	}
	if retrievedCard.Name != "Credicor" {
		t.Errorf("Expected card name to be 'Credicor', got '%s'", retrievedCard.Name)
	}
}

func TestCorporation_HasCorporation(t *testing.T) {
	corp := NewCorporation()

	if corp.HasCorporation() {
		t.Error("Expected HasCorporation() to return false initially")
	}

	corp.SetID("credicor")
	if corp.HasCorporation() {
		t.Error("Expected HasCorporation() to return false after setting only ID")
	}

	testCard := card.Card{
		ID:   "credicor",
		Name: "Credicor",
	}
	corp.SetCard(testCard)
	if !corp.HasCorporation() {
		t.Error("Expected HasCorporation() to return true after setting card")
	}
}

func TestCorporation_DeepCopy(t *testing.T) {
	original := NewCorporation()
	testCard := card.Card{
		ID:   "mining_guild",
		Name: "Mining Guild",
		Cost: 0,
		Type: "corporation",
	}
	original.SetCard(testCard)

	copy := original.DeepCopy()

	// Verify copy has same values
	if copy.ID() != original.ID() {
		t.Error("Expected copy to have same ID")
	}
	if !copy.HasCorporation() {
		t.Error("Expected copy to have corporation")
	}

	copyCard := copy.Card()
	originalCard := original.Card()
	if copyCard == nil || originalCard == nil {
		t.Fatal("Expected both cards to not be nil")
	}
	if copyCard.ID != originalCard.ID {
		t.Error("Expected copy card to have same ID")
	}
	if copyCard.Name != originalCard.Name {
		t.Error("Expected copy card to have same name")
	}

	// Verify modifying copy doesn't affect original
	copy.SetID("different_corp")
	if original.ID() != "mining_guild" {
		t.Error("Expected original ID to remain unchanged")
	}

	// Modify the copy's card
	newCard := card.Card{
		ID:   "different_corp",
		Name: "Different Corp",
	}
	copy.SetCard(newCard)
	if original.Card().ID != "mining_guild" {
		t.Error("Expected original card ID to remain unchanged")
	}
}

func TestCorporation_DeepCopy_Nil(t *testing.T) {
	var corp *Corporation = nil
	copy := corp.DeepCopy()

	if copy != nil {
		t.Error("Expected DeepCopy of nil to return nil")
	}
}

func TestCorporation_DeepCopy_NoCorporation(t *testing.T) {
	original := NewCorporation()
	original.SetID("some_id")

	copy := original.DeepCopy()

	if copy.Card() != nil {
		t.Error("Expected copy to have no card")
	}
	if copy.ID() != "some_id" {
		t.Errorf("Expected copy to have ID 'some_id', got '%s'", copy.ID())
	}
}
