import React, { useEffect, useState } from "react";
import SimpleGameCard from "../cards/SimpleGameCard.tsx";
import GameIcon from "../display/GameIcon.tsx";
import {
  CardDto,
  PendingCardDrawSelectionDto,
  ResourceTypeCredits,
} from "../../../types/generated/api-types.ts";

interface CardDrawSelectionOverlayProps {
  isOpen: boolean;
  selection: PendingCardDrawSelectionDto;
  playerCredits: number;
  onConfirm: (cardsToTake: string[], cardsToBuy: string[]) => void;
}

const CardDrawSelectionOverlay: React.FC<CardDrawSelectionOverlayProps> = ({
  isOpen,
  selection,
  playerCredits,
  onConfirm,
}) => {
  const [cardsToTake, setCardsToTake] = useState<string[]>([]);
  const [cardsToBuy, setCardsToBuy] = useState<string[]>([]);
  const [showConfirmation, setShowConfirmation] = useState(false);

  // Initialize selection when overlay opens
  useEffect(() => {
    if (isOpen && selection.availableCards.length > 0) {
      setCardsToTake([]);
      setCardsToBuy([]);
      setShowConfirmation(false);
    }
  }, [isOpen, selection.availableCards]);

  if (!isOpen || selection.availableCards.length === 0) return null;

  // Pure card-draw: All shown cards must be taken (no choice)
  // Peek+Draw/Take: Some cards must/can be taken (player has choice)
  const isCardDraw =
    selection.maxBuyCount === 0 &&
    selection.freeTakeCount === selection.availableCards.length;

  const getTitleAndDescription = (): {
    title: string;
    description: string;
  } => {
    if (isCardDraw) {
      return {
        title: "New cards",
        description: "",
      };
    }

    // For all peek/take/buy scenarios, use consistent "Select cards" title
    const maxCards = selection.freeTakeCount + selection.maxBuyCount;
    return {
      title: "Select cards",
      description: `Select up to ${maxCards} card${maxCards !== 1 ? "s" : ""}`,
    };
  };

  const getCardBadge = (
    cardId: string,
  ): { type: "free" | "buy"; value?: number } => {
    if (cardsToTake.includes(cardId)) {
      return { type: "free" };
    } else {
      return { type: "buy", value: selection.cardBuyCost };
    }
  };

  const canAffordBuy = (): boolean => {
    const currentBuyCost = cardsToBuy.length * selection.cardBuyCost;
    return currentBuyCost + selection.cardBuyCost <= playerCredits;
  };

  const handleCardSelect = (cardId: string) => {
    // For card-draw scenarios, auto-select all cards
    if (isCardDraw) {
      return;
    }

    // Reset confirmation when user selects cards
    if (showConfirmation) {
      setShowConfirmation(false);
    }

    // If card is already in take list, remove it
    if (cardsToTake.includes(cardId)) {
      setCardsToTake((prev) => prev.filter((id) => id !== cardId));
      return;
    }

    // If card is already in buy list, remove it
    if (cardsToBuy.includes(cardId)) {
      setCardsToBuy((prev) => prev.filter((id) => id !== cardId));
      return;
    }

    // Try to add to free take list first
    if (cardsToTake.length < selection.freeTakeCount) {
      setCardsToTake((prev) => [...prev, cardId]);
    } else if (cardsToBuy.length < selection.maxBuyCount && canAffordBuy()) {
      // Otherwise add to buy list if we can afford it
      setCardsToBuy((prev) => [...prev, cardId]);
    }
  };

  const handleConfirm = () => {
    // For card-draw, automatically select all cards
    if (isCardDraw) {
      const allCardIds = selection.availableCards.map((c) => c.id);
      onConfirm(allCardIds, []);
      return;
    }

    const totalSelected = cardsToTake.length + cardsToBuy.length;
    const maxAllowed = selection.freeTakeCount + selection.maxBuyCount;

    if (totalSelected > maxAllowed) {
      return; // Invalid selection
    }

    // Require confirmation in two scenarios:
    // 1. Discarding all cards (totalSelected === 0)
    // 2. Not taking all available free cards (cardsToTake.length < freeTakeCount)
    const needsConfirmation =
      totalSelected === 0 || cardsToTake.length < selection.freeTakeCount;

    if (needsConfirmation && !showConfirmation) {
      // First click - show confirmation
      setShowConfirmation(true);
      return;
    }

    // Second click or no confirmation needed - proceed with selection
    onConfirm(cardsToTake, cardsToBuy);
  };

  const getButtonText = (): string => {
    if (isCardDraw) {
      return "Return";
    }

    // For peek/take/buy scenarios
    const totalSelected = cardsToTake.length + cardsToBuy.length;

    // Check if confirmation is needed
    const needsConfirmation =
      totalSelected === 0 || cardsToTake.length < selection.freeTakeCount;

    if (needsConfirmation && showConfirmation) {
      return "Confirm Selection";
    }

    if (totalSelected === 0) {
      return "Discard all";
    }

    // Show buy count if any cards are being bought
    if (cardsToBuy.length > 0) {
      return cardsToBuy.length === 1
        ? "Buy 1 card"
        : `Buy ${cardsToBuy.length} cards`;
    }

    // Otherwise just confirm the free selection
    return "Confirm Selection";
  };

  const titleInfo = getTitleAndDescription();
  const totalBuyCost = cardsToBuy.length * selection.cardBuyCost;
  const totalSelected = cardsToTake.length + cardsToBuy.length;
  // For peek scenarios, allow any selection from 0 to max (including discarding all)
  const isValidSelection =
    isCardDraw ||
    totalSelected <= selection.freeTakeCount + selection.maxBuyCount;

  return (
    <div className="fixed inset-0 z-[1000] flex items-center justify-center animate-[fadeIn_0.3s_ease]">
      {/* Translucent background */}
      <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" />

      {/* Content container */}
      <div className="relative z-[1] w-[90%] max-w-[1400px] max-h-[90vh] flex flex-col bg-space-black-darker/95 border-2 border-space-blue-400 rounded-[20px] overflow-hidden backdrop-blur-space shadow-[0_20px_60px_rgba(0,0,0,0.6),0_0_60px_rgba(30,60,150,0.5)] max-[768px]:w-full max-[768px]:h-screen max-[768px]:max-h-screen max-[768px]:rounded-none">
        {/* Header */}
        <div className="py-6 px-8 bg-black/40 border-b border-space-blue-600 max-[768px]:p-5">
          <h2 className="m-0 font-orbitron text-[28px] font-bold text-white text-shadow-glow tracking-wider max-[768px]:text-2xl">
            {titleInfo.title}
          </h2>
          {titleInfo.description && (
            <p className="mt-2 mb-0 text-base text-white/80 max-[768px]:text-sm">
              {titleInfo.description}
            </p>
          )}
        </div>

        {/* Cards display */}
        <div className="flex-1 overflow-x-auto overflow-y-hidden p-8 flex items-center bg-[radial-gradient(ellipse_at_center,rgba(139,69,19,0.1)_0%,transparent_70%)] [&::-webkit-scrollbar]:h-2 [&::-webkit-scrollbar-track]:bg-white/5 [&::-webkit-scrollbar-track]:rounded [&::-webkit-scrollbar-thumb]:bg-white/20 [&::-webkit-scrollbar-thumb]:rounded [&::-webkit-scrollbar-thumb:hover]:bg-white/30 max-[768px]:p-5">
          <div className="flex gap-6 mx-auto py-5 max-[768px]:gap-4">
            {selection.availableCards.map((card: CardDto, index: number) => {
              const isSelected =
                cardsToTake.includes(card.id) || cardsToBuy.includes(card.id);
              const badge = getCardBadge(card.id);

              return (
                <div key={card.id} className="relative">
                  <SimpleGameCard
                    card={card}
                    isSelected={isSelected}
                    onSelect={handleCardSelect}
                    animationDelay={index * 100}
                    showCheckbox={!isCardDraw}
                  />
                  {/* FREE Badge - only show for cards in free take list */}
                  {!isCardDraw && badge.type === "free" && (
                    <div className="absolute top-2 right-2 px-2 py-1 rounded-md font-bold text-sm shadow-lg bg-[#4caf50] text-white">
                      FREE
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        </div>

        {/* Footer with cost and confirm button */}
        <div className="py-6 px-8 bg-black/40 border-t border-space-blue-600 flex justify-between items-center max-[768px]:p-5 max-[768px]:flex-col max-[768px]:gap-5">
          <div className="flex gap-8 items-center max-[768px]:w-full max-[768px]:justify-between max-[768px]:flex-wrap">
            <div className="flex items-center gap-3">
              <span className="text-sm text-white/60 uppercase tracking-[0.5px]">
                Your Credits:
              </span>
              <GameIcon
                iconType={ResourceTypeCredits}
                amount={playerCredits}
                size="large"
              />
            </div>
            {totalBuyCost > 0 && (
              <div className="flex items-center gap-3">
                <span className="text-sm text-white/60 uppercase tracking-[0.5px]">
                  Buy Cost:
                </span>
                <GameIcon
                  iconType={ResourceTypeCredits}
                  amount={totalBuyCost}
                  size="large"
                />
              </div>
            )}
          </div>

          <div className="flex items-center gap-6 max-[768px]:w-full max-[768px]:flex-col max-[768px]:gap-3">
            <div className="text-sm">
              {isCardDraw ? (
                <span className="text-white/70">
                  Drawing {selection.freeTakeCount} card
                  {selection.freeTakeCount !== 1 ? "s" : ""}
                </span>
              ) : showConfirmation ? (
                <span className="text-[#ff9800]">
                  {totalSelected === 0
                    ? "Are you sure you want to discard all?"
                    : (() => {
                        const remainingFreeTakes =
                          selection.freeTakeCount - cardsToTake.length;
                        return `You can take ${remainingFreeTakes} more card${remainingFreeTakes !== 1 ? "s" : ""} for free. Confirm?`;
                      })()}
                </span>
              ) : (
                (() => {
                  const discardCount =
                    selection.availableCards.length - totalSelected;
                  return discardCount > 0 ? (
                    <span className="text-white/70">
                      Discard {discardCount} card
                      {discardCount !== 1 ? "s" : ""}
                    </span>
                  ) : null;
                })()
              )}
            </div>
            <div className="flex gap-3 items-center">
              <button
                className="py-4 px-8 bg-space-black-darker/90 border-2 border-space-blue-800 rounded-xl text-xl font-bold text-white cursor-pointer transition-all duration-300 text-shadow-dark shadow-[0_4px_20px_rgba(30,60,150,0.3)] whitespace-nowrap hover:enabled:bg-space-black-darker/95 hover:enabled:border-space-blue-600 hover:enabled:-translate-y-0.5 hover:enabled:shadow-glow active:enabled:translate-y-0 disabled:bg-gray-700/50 disabled:border-gray-500/30 disabled:cursor-not-allowed disabled:transform-none disabled:shadow-none disabled:opacity-60 max-[768px]:w-full max-[768px]:py-3 max-[768px]:px-6 max-[768px]:text-lg"
                onClick={handleConfirm}
                disabled={!isValidSelection || totalBuyCost > playerCredits}
              >
                {getButtonText()}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default CardDrawSelectionOverlay;
