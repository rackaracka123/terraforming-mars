import React, { useEffect, useState } from "react";
import SimpleGameCard from "../cards/SimpleGameCard.tsx";
import MegaCreditIcon from "../display/MegaCreditIcon.tsx";
import { CardDto } from "../../../types/generated/api-types.ts";
import styles from "./StartingCardSelectionOverlay.module.css";

interface StartingCardSelectionOverlayProps {
  isOpen: boolean;
  cards: CardDto[];
  playerCredits: number;
  onConfirmSelection: (selectedCardIds: string[]) => void;
}

const StartingCardSelectionOverlay: React.FC<
  StartingCardSelectionOverlayProps
> = ({ isOpen, cards, playerCredits, onConfirmSelection }) => {
  const [selectedCardIds, setSelectedCardIds] = useState<string[]>([]);
  const [totalCost, setTotalCost] = useState(0);
  const [showConfirmation, setShowConfirmation] = useState(false);

  // Reset selection when overlay opens with new cards
  useEffect(() => {
    if (isOpen && cards.length > 0) {
      setSelectedCardIds([]);
      setTotalCost(0);
      setShowConfirmation(false);
    }
  }, [isOpen, cards]);

  // Calculate total cost whenever selection changes
  useEffect(() => {
    const cost = selectedCardIds.length * 3; // Each card costs 3 MC
    setTotalCost(cost);
    // Reset confirmation state when cards are selected
    if (selectedCardIds.length > 0 && showConfirmation) {
      setShowConfirmation(false);
    }
  }, [selectedCardIds, showConfirmation]);

  if (!isOpen || cards.length === 0) return null;

  const handleCardSelect = (cardId: string) => {
    setSelectedCardIds((prev) => {
      if (prev.includes(cardId)) {
        // Deselect card
        return prev.filter((id) => id !== cardId);
      } else {
        // Select card - check if player can afford it
        const newSelection = [...prev, cardId];
        const costForNewCard = newSelection.length > 1 ? 3 : 0;
        const newTotalCost = totalCost + costForNewCard;

        if (newTotalCost <= playerCredits) {
          return newSelection;
        } else {
          // Can't afford this card
          return prev;
        }
      }
    });
  };

  const handleConfirm = () => {
    if (selectedCardIds.length > 0) {
      // Player has selected cards, confirm normally
      onConfirmSelection(selectedCardIds);
    } else if (!showConfirmation) {
      // First click with no selection - show confirmation
      setShowConfirmation(true);
    } else {
      // Second click with no selection - confirm with empty selection
      onConfirmSelection([]);
    }
  };

  return (
    <div className={styles.overlay}>
      {/* Translucent background */}
      <div className={styles.backdrop} />

      {/* Content container */}
      <div className={styles.contentContainer}>
        {/* Header */}
        <div className={styles.header}>
          <h2 className={styles.title}>Select Starting Cards</h2>
          <p className={styles.subtitle}>
            Choose your starting cards. First card is FREE, additional cards
            cost 3 MC each.
          </p>
        </div>

        {/* Cards display */}
        <div className={styles.cardsContainer}>
          <div className={styles.cardsRow}>
            {cards.map((card, index) => {
              const cardIndex = selectedCardIds.indexOf(card.id);
              const isSelected = cardIndex !== -1;

              return (
                <SimpleGameCard
                  key={card.id}
                  card={card}
                  isSelected={isSelected}
                  onSelect={handleCardSelect}
                  animationDelay={index * 100}
                />
              );
            })}
          </div>
        </div>

        {/* Footer with cost and confirm button */}
        <div className={styles.footer}>
          <div className={styles.costInfo}>
            <div className={styles.creditsDisplay}>
              <span className={styles.label}>Your Credits:</span>
              <MegaCreditIcon value={playerCredits} size="large" />
            </div>
            <div className={styles.totalCost}>
              <span className={styles.label}>Total Cost:</span>
              {totalCost > 0 ? (
                <MegaCreditIcon value={totalCost} size="large" />
              ) : (
                <span className={styles.free}>FREE</span>
              )}
            </div>
          </div>

          <div className={styles.actions}>
            <div className={styles.selectionInfo}>
              {selectedCardIds.length === 0 ? (
                showConfirmation ? (
                  <span className={styles.warning}>
                    Are you sure you don't want to select any cards?
                  </span>
                ) : (
                  <span className={styles.info}>No cards selected</span>
                )
              ) : (
                <span className={styles.info}>
                  {selectedCardIds.length} card
                  {selectedCardIds.length !== 1 ? "s" : ""} selected
                </span>
              )}
            </div>
            <button
              className={styles.confirmButton}
              onClick={handleConfirm}
              disabled={totalCost > playerCredits}
            >
              {showConfirmation ? "Confirm Skip" : "Confirm Selection"}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default StartingCardSelectionOverlay;
