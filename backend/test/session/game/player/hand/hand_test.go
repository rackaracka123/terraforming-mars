package hand_test

import (
	"testing"

	"terraforming-mars-backend/internal/session/game/player/hand"
)

import (
	"testing"
)

func TestNewHand(t *testing.T) {
	h := NewHand()

	if h.CardCount() != 0 {
		t.Errorf("Expected 0 cards, got %d", h.CardCount())
	}
	if len(h.Cards()) != 0 {
		t.Error("Expected empty cards slice")
	}
	if len(h.PlayedCards()) != 0 {
		t.Error("Expected empty played cards slice")
	}
	if h.HasCard("any-card") {
		t.Error("Expected HasCard to return false for any card")
	}
}

func TestHand_AddCard(t *testing.T) {
	h := NewHand()

	h.AddCard("card-1")
	if h.CardCount() != 1 {
		t.Errorf("Expected 1 card, got %d", h.CardCount())
	}
	if !h.HasCard("card-1") {
		t.Error("Expected HasCard('card-1') to return true")
	}

	h.AddCard("card-2")
	h.AddCard("card-3")
	if h.CardCount() != 3 {
		t.Errorf("Expected 3 cards, got %d", h.CardCount())
	}
}

func TestHand_RemoveCard(t *testing.T) {
	h := NewHand()
	h.AddCard("card-1")
	h.AddCard("card-2")
	h.AddCard("card-3")

	// Remove existing card
	removed := h.RemoveCard("card-2")
	if !removed {
		t.Error("Expected RemoveCard to return true")
	}
	if h.CardCount() != 2 {
		t.Errorf("Expected 2 cards, got %d", h.CardCount())
	}
	if h.HasCard("card-2") {
		t.Error("Expected HasCard('card-2') to return false")
	}

	// Try to remove non-existent card
	removed = h.RemoveCard("non-existent")
	if removed {
		t.Error("Expected RemoveCard to return false for non-existent card")
	}
	if h.CardCount() != 2 {
		t.Errorf("Expected 2 cards, got %d", h.CardCount())
	}
}

func TestHand_PlayCard(t *testing.T) {
	h := NewHand()
	h.AddCard("card-1")
	h.AddCard("card-2")

	// Play existing card
	played := h.PlayCard("card-1")
	if !played {
		t.Error("Expected PlayCard to return true")
	}
	if h.CardCount() != 1 {
		t.Errorf("Expected 1 card in hand, got %d", h.CardCount())
	}
	if h.HasCard("card-1") {
		t.Error("Expected card-1 to be removed from hand")
	}

	playedCards := h.PlayedCards()
	if len(playedCards) != 1 {
		t.Errorf("Expected 1 played card, got %d", len(playedCards))
	}
	if playedCards[0] != "card-1" {
		t.Errorf("Expected played card to be 'card-1', got '%s'", playedCards[0])
	}

	// Try to play non-existent card
	played = h.PlayCard("non-existent")
	if played {
		t.Error("Expected PlayCard to return false for non-existent card")
	}
	if len(h.PlayedCards()) != 1 {
		t.Error("Expected played cards count to remain unchanged")
	}
}

func TestHand_AddPlayedCard(t *testing.T) {
	h := NewHand()

	// Add card directly to played cards (e.g., corporation)
	h.AddPlayedCard("corporation-1")
	if len(h.PlayedCards()) != 1 {
		t.Errorf("Expected 1 played card, got %d", len(h.PlayedCards()))
	}
	if h.HasCard("corporation-1") {
		t.Error("Expected card to not be in hand")
	}
}

func TestHand_SetCards(t *testing.T) {
	h := NewHand()

	cards := []string{"card-1", "card-2", "card-3"}
	h.SetCards(cards)

	if h.CardCount() != 3 {
		t.Errorf("Expected 3 cards, got %d", h.CardCount())
	}
	for _, cardID := range cards {
		if !h.HasCard(cardID) {
			t.Errorf("Expected HasCard('%s') to return true", cardID)
		}
	}

	// Verify defensive copy
	cards[0] = "modified"
	if h.HasCard("modified") {
		t.Error("Expected hand to not be affected by external slice modification")
	}
}

func TestHand_SetCards_Nil(t *testing.T) {
	h := NewHand()
	h.AddCard("card-1")

	h.SetCards(nil)
	if h.CardCount() != 0 {
		t.Errorf("Expected 0 cards after SetCards(nil), got %d", h.CardCount())
	}
}

func TestHand_SetPlayedCards(t *testing.T) {
	h := NewHand()

	playedCards := []string{"card-1", "card-2"}
	h.SetPlayedCards(playedCards)

	if len(h.PlayedCards()) != 2 {
		t.Errorf("Expected 2 played cards, got %d", len(h.PlayedCards()))
	}

	// Verify defensive copy
	playedCards[0] = "modified"
	result := h.PlayedCards()
	if result[0] == "modified" {
		t.Error("Expected played cards to not be affected by external slice modification")
	}
}

func TestHand_SetPlayedCards_Nil(t *testing.T) {
	h := NewHand()
	h.AddPlayedCard("card-1")

	h.SetPlayedCards(nil)
	if len(h.PlayedCards()) != 0 {
		t.Errorf("Expected 0 played cards after SetPlayedCards(nil), got %d", len(h.PlayedCards()))
	}
}

func TestHand_Cards_DefensiveCopy(t *testing.T) {
	h := NewHand()
	h.AddCard("card-1")
	h.AddCard("card-2")

	cards := h.Cards()
	cards[0] = "modified"

	// Verify original hand is not affected
	if !h.HasCard("card-1") {
		t.Error("Expected hand to still have card-1")
	}
	if h.HasCard("modified") {
		t.Error("Expected hand to not have 'modified'")
	}
}

func TestHand_PlayedCards_DefensiveCopy(t *testing.T) {
	h := NewHand()
	h.AddPlayedCard("card-1")
	h.AddPlayedCard("card-2")

	played := h.PlayedCards()
	played[0] = "modified"

	// Verify original played cards are not affected
	playedAgain := h.PlayedCards()
	if playedAgain[0] != "card-1" {
		t.Error("Expected played cards to not be affected by external modification")
	}
}

func TestHand_DeepCopy(t *testing.T) {
	original := NewHand()
	original.AddCard("card-1")
	original.AddCard("card-2")
	original.AddPlayedCard("played-1")

	copy := original.DeepCopy()

	// Verify copy has same values
	if copy.CardCount() != original.CardCount() {
		t.Error("Expected copy to have same card count")
	}
	if len(copy.PlayedCards()) != len(original.PlayedCards()) {
		t.Error("Expected copy to have same played card count")
	}
	if !copy.HasCard("card-1") || !copy.HasCard("card-2") {
		t.Error("Expected copy to have same cards")
	}

	// Verify modifying copy doesn't affect original
	copy.AddCard("card-3")
	copy.AddPlayedCard("played-2")

	if original.CardCount() != 2 {
		t.Error("Expected original to still have 2 cards")
	}
	if len(original.PlayedCards()) != 1 {
		t.Error("Expected original to still have 1 played card")
	}

	// Verify modifying returned slices doesn't affect copies
	copiedCards := copy.Cards()
	copiedCards[0] = "modified"
	if !copy.HasCard("card-1") {
		t.Error("Expected copy to still have card-1")
	}
}

func TestHand_DeepCopy_Nil(t *testing.T) {
	var h *Hand = nil
	copy := h.DeepCopy()

	if copy != nil {
		t.Error("Expected DeepCopy of nil to return nil")
	}
}
