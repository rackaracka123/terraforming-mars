import React, { useState, useEffect } from "react";
import SimpleGameCard from "../cards/SimpleGameCard.tsx";
import styles from "./StartingCardSelectionOverlay.module.css";

interface Card {
  id: string;
  name: string;
  cost: number;
  tags?: string[];
  description?: string;
  requirements?: string;
}

interface StartingCardSelectionOverlayProps {
  isOpen: boolean;
  cards: Card[];
  playerCredits: number;
  onConfirmSelection: (selectedCardIds: string[]) => void;
}

const StartingCardSelectionOverlay: React.FC<StartingCardSelectionOverlayProps> = ({
  isOpen,
  cards,
  playerCredits,
  onConfirmSelection,
}) => {
  const [selectedCardIds, setSelectedCardIds] = useState<string[]>([]);
  const [totalCost, setTotalCost] = useState(0);

  // Reset selection when overlay opens with new cards
  useEffect(() => {
    if (isOpen && cards.length > 0) {
      setSelectedCardIds([]);
      setTotalCost(0);
    }
  }, [isOpen, cards]);

  // Calculate total cost whenever selection changes
  useEffect(() => {
    let cost = 0;
    selectedCardIds.forEach((cardId, index) => {
      // First selected card is free, others cost 3 MC each
      if (index > 0) {
        cost += 3;
      }
    });
    setTotalCost(cost);
  }, [selectedCardIds]);

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
      onConfirmSelection(selectedCardIds);
    }
  };

  const canAffordMore = totalCost + 3 <= playerCredits || selectedCardIds.length === 0;

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
            Choose your starting cards. First card is FREE, additional cards cost 3 MC each.
          </p>
        </div>

        {/* Cards display */}
        <div className={styles.cardsContainer}>
          <div className={styles.cardsRow}>
            {cards.map((card, index) => {
              const cardIndex = selectedCardIds.indexOf(card.id);
              const isSelected = cardIndex !== -1;
              const isFree = isSelected && cardIndex === 0;
              
              return (
                <SimpleGameCard
                  key={card.id}
                  card={card}
                  isSelected={isSelected}
                  isFree={isFree}
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
              <div className={styles.creditsAmount}>
                <img
                  src="/assets/resources/megacredit.png"
                  alt="MC"
                  className={styles.creditsIcon}
                />
                <span>{playerCredits}</span>
              </div>
            </div>
            <div className={styles.totalCost}>
              <span className={styles.label}>Total Cost:</span>
              <div className={styles.costAmount}>
                {totalCost > 0 ? (
                  <>
                    <img
                      src="/assets/resources/megacredit.png"
                      alt="MC"
                      className={styles.creditsIcon}
                    />
                    <span className={totalCost > playerCredits ? styles.insufficientFunds : ""}>
                      {totalCost}
                    </span>
                  </>
                ) : (
                  <span className={styles.free}>FREE</span>
                )}
              </div>
            </div>
          </div>

          <div className={styles.actions}>
            <div className={styles.selectionInfo}>
              {selectedCardIds.length === 0 ? (
                <span className={styles.warning}>Select at least one card to continue</span>
              ) : (
                <span className={styles.info}>
                  {selectedCardIds.length} card{selectedCardIds.length !== 1 ? 's' : ''} selected
                </span>
              )}
            </div>
            <button
              className={styles.confirmButton}
              onClick={handleConfirm}
              disabled={selectedCardIds.length === 0 || totalCost > playerCredits}
            >
              Confirm Selection
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default StartingCardSelectionOverlay;