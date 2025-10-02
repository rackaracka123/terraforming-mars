import React, { useEffect, useState } from "react";
import SimpleGameCard from "../cards/SimpleGameCard.tsx";
import MegaCreditIcon from "../display/MegaCreditIcon.tsx";
import { CardDto } from "../../../types/generated/api-types.ts";
import styles from "./ProductionCardSelectionOverlay.module.css";

interface ProductionCardSelectionOverlayProps {
  isOpen: boolean;
  cards: CardDto[];
  playerCredits: number;
  onSelectCards: (selectedCardIds: string[]) => void;
  onReturn: () => void;
}

const ProductionCardSelectionOverlay: React.FC<
  ProductionCardSelectionOverlayProps
> = ({ isOpen, cards, playerCredits, onSelectCards, onReturn }) => {
  const [selectedCardIds, setSelectedCardIds] = useState<string[]>([]);
  const [totalCost, setTotalCost] = useState(0);
  const [showConfirmation, setShowConfirmation] = useState(false);

  // Initialize selection when overlay opens
  useEffect(() => {
    if (isOpen && cards.length > 0) {
      setSelectedCardIds([]);
      setShowConfirmation(false);
      setTotalCost(0);
    }
  }, [isOpen, cards]);

  // Calculate total cost whenever selection changes - ALL cards cost 3 MC
  useEffect(() => {
    const cost = selectedCardIds.length * 3; // Each card costs 3 MC (no free cards)
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
        const newSelection = prev.filter((id) => id !== cardId);
        return newSelection;
      } else {
        // Select card - check if player can afford it (ALL cards cost 3 MC)
        const newSelection = [...prev, cardId];
        const newTotalCost = newSelection.length * 3;

        if (newTotalCost <= playerCredits) {
          return newSelection;
        } else {
          return prev;
        }
      }
    });
  };

  const handleConfirm = () => {
    if (selectedCardIds.length > 0) {
      // Player has selected cards - commit immediately
      onSelectCards(selectedCardIds);
    } else if (!showConfirmation) {
      // First click with no selection - show confirmation
      setShowConfirmation(true);
    } else {
      // Second click with no selection - confirm with empty selection
      onSelectCards([]);
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
          <h2 className={styles.title}>Select Cards to Buy</h2>
          <p className={styles.subtitle}>
            Choose cards to buy for your next turn. Each card costs 3 MC.
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
                  showCheckbox={true}
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
                    Are you sure you don't want to buy any cards?
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
            <div className={styles.buttonsContainer}>
              <button className={styles.returnButton} onClick={onReturn}>
                Hide
              </button>
              <button
                className={styles.confirmButton}
                onClick={handleConfirm}
                disabled={totalCost > playerCredits}
              >
                {showConfirmation ? "Confirm Skip" : "Buy Cards"}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ProductionCardSelectionOverlay;
