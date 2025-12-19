import { useState, useEffect, useCallback } from "react";

export interface CardSelectionConfig {
  /** Cards or selection object with availableCards */
  cards: { id: string }[] | { availableCards: { id: string }[] };
  /** Whether the overlay is open */
  isOpen: boolean;
  /** Player's available credits */
  playerCredits: number;
  /** Cost per card (flat rate) - optional */
  costPerCard?: number;
  /** Custom cost function for individual cards - optional */
  getCardCost?: (cardId: string) => number;
  /** Custom reward function for individual cards - optional */
  getCardReward?: (cardId: string) => number;
  /** Minimum cards that must be selected */
  minCards?: number;
  /** Maximum cards that can be selected */
  maxCards?: number;
}

export interface CardSelectionState {
  /** Currently selected card IDs */
  selectedCardIds: string[];
  /** Total cost of selected cards */
  totalCost: number;
  /** Total reward from selected cards */
  totalReward: number;
  /** Whether confirmation dialog is shown */
  showConfirmation: boolean;
  /** Whether the current selection is valid */
  isValidSelection: boolean;
  /** Handle card selection/deselection */
  handleCardSelect: (cardId: string) => void;
  /** Handle confirmation */
  handleConfirm: (onConfirm: (selectedIds: string[]) => void) => void;
  /** Reset confirmation state */
  resetConfirmation: () => void;
  /** Check if a specific card can be afforded */
  canAffordCard: (cardId: string) => boolean;
}

/**
 * Custom hook for managing card selection logic across overlays
 * Handles selection state, cost tracking, validation, and confirmation flow
 */
export function useCardSelection(config: CardSelectionConfig): CardSelectionState {
  const {
    cards,
    isOpen,
    playerCredits,
    costPerCard = 0,
    getCardCost,
    getCardReward,
    minCards = 0,
    maxCards = Infinity,
  } = config;

  // Extract cards array from object if needed
  const cardsArray = Array.isArray(cards) ? cards : cards.availableCards;

  const [selectedCardIds, setSelectedCardIds] = useState<string[]>([]);
  const [totalCost, setTotalCost] = useState(0);
  const [totalReward, setTotalReward] = useState(0);
  const [showConfirmation, setShowConfirmation] = useState(false);

  // Reset when overlay opens
  useEffect(() => {
    if (isOpen && cardsArray.length > 0) {
      setSelectedCardIds([]);
      setShowConfirmation(false);
      setTotalCost(0);
      setTotalReward(0);
    }
  }, [isOpen, cardsArray.length]);

  // Calculate total cost and reward
  useEffect(() => {
    let cost = 0;
    let reward = 0;

    selectedCardIds.forEach((cardId) => {
      if (getCardCost) {
        cost += getCardCost(cardId);
      } else if (costPerCard) {
        cost += costPerCard;
      }

      if (getCardReward) {
        reward += getCardReward(cardId);
      }
    });

    setTotalCost(cost);
    setTotalReward(reward);

    // Reset confirmation when selection changes
    if (selectedCardIds.length > 0 && showConfirmation) {
      setShowConfirmation(false);
    }
  }, [selectedCardIds, costPerCard, getCardCost, getCardReward, showConfirmation]);

  // Check if a specific card can be afforded
  const canAffordCard = useCallback(
    (cardId: string): boolean => {
      const currentCost = selectedCardIds.reduce((sum, id) => {
        if (getCardCost) return sum + getCardCost(id);
        if (costPerCard) return sum + costPerCard;
        return sum;
      }, 0);

      const cardCost = getCardCost ? getCardCost(cardId) : costPerCard;
      return currentCost + cardCost <= playerCredits;
    },
    [selectedCardIds, playerCredits, costPerCard, getCardCost],
  );

  // Handle card selection/deselection
  const handleCardSelect = useCallback(
    (cardId: string) => {
      setSelectedCardIds((prev) => {
        if (prev.includes(cardId)) {
          // Deselect card
          return prev.filter((id) => id !== cardId);
        } else {
          // Check if we can afford it and haven't exceeded max
          if (prev.length >= maxCards) {
            return prev;
          }

          if (!canAffordCard(cardId)) {
            return prev;
          }

          // Select card
          return [...prev, cardId];
        }
      });
    },
    [maxCards, canAffordCard],
  );

  // Handle confirmation
  const handleConfirm = useCallback(
    (onConfirm: (selectedIds: string[]) => void) => {
      const selectedCount = selectedCardIds.length;

      // Check if selection is within bounds
      if (selectedCount < minCards || selectedCount > maxCards) {
        return; // Invalid selection
      }

      if (selectedCount > 0) {
        // Has selection - confirm immediately
        onConfirm(selectedCardIds);
      } else if (minCards === 0 && !showConfirmation) {
        // No selection but 0 is allowed - show confirmation first
        setShowConfirmation(true);
      } else if (minCards === 0 && showConfirmation) {
        // Second click - confirm with empty selection
        onConfirm([]);
      }
    },
    [selectedCardIds, minCards, maxCards, showConfirmation],
  );

  // Reset confirmation state
  const resetConfirmation = useCallback(() => {
    setShowConfirmation(false);
  }, []);

  // Check if current selection is valid
  const isValidSelection =
    selectedCardIds.length >= minCards &&
    selectedCardIds.length <= maxCards &&
    totalCost <= playerCredits;

  return {
    selectedCardIds,
    totalCost,
    totalReward,
    showConfirmation,
    isValidSelection,
    handleCardSelect,
    handleConfirm,
    resetConfirmation,
    canAffordCard,
  };
}
