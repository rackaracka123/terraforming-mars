import React, { useEffect, useState } from "react";
import SimpleGameCard from "../cards/SimpleGameCard.tsx";
import MegaCreditIcon from "../display/MegaCreditIcon.tsx";
import {
  CardDto,
  PendingCardSelectionDto,
} from "../../../types/generated/api-types.ts";

interface PendingCardSelectionOverlayProps {
  isOpen: boolean;
  selection: PendingCardSelectionDto;
  playerCredits: number;
  onSelectCards: (selectedCardIds: string[]) => void;
  onCancel?: () => void;
}

interface TitleInfo {
  title: string;
  description: string;
}

const PendingCardSelectionOverlay: React.FC<
  PendingCardSelectionOverlayProps
> = ({ isOpen, selection, playerCredits, onSelectCards, onCancel }) => {
  const [selectedCardIds, setSelectedCardIds] = useState<string[]>([]);
  const [totalCost, setTotalCost] = useState(0);
  const [totalReward, setTotalReward] = useState(0);
  const [showConfirmation, setShowConfirmation] = useState(false);

  // Initialize selection when overlay opens
  useEffect(() => {
    if (isOpen && selection.availableCards.length > 0) {
      setSelectedCardIds([]);
      setShowConfirmation(false);
      setTotalCost(0);
      setTotalReward(0);
    }
  }, [isOpen, selection.availableCards]);

  // Calculate total cost and reward whenever selection changes
  useEffect(() => {
    const cost = selectedCardIds.reduce((sum, cardId) => {
      return sum + (selection.cardCosts[cardId] || 0);
    }, 0);

    const reward = selectedCardIds.reduce((sum, cardId) => {
      return sum + (selection.cardRewards[cardId] || 0);
    }, 0);

    setTotalCost(cost);
    setTotalReward(reward);

    // Reset confirmation state when cards are selected
    if (selectedCardIds.length > 0 && showConfirmation) {
      setShowConfirmation(false);
    }
  }, [
    selectedCardIds,
    selection.cardCosts,
    selection.cardRewards,
    showConfirmation,
  ]);

  if (!isOpen || selection.availableCards.length === 0) return null;

  const getTitleAndDescription = (source: string): TitleInfo => {
    switch (source) {
      case "sell-patents":
        return {
          title: "Sell Cards for Credits",
          description: "Select cards to sell. Each card gives you 1 MC.",
        };
      case "research-phase":
        return {
          title: "Buy Research Cards",
          description: "Select cards to purchase. Each card costs 3 MC.",
        };
      default:
        return {
          title: "Select Cards",
          description: "Choose cards from the available options.",
        };
    }
  };

  const getCardBadge = (
    cardId: string,
  ): { type: "reward" | "cost" | "free"; value?: number } | null => {
    const cost = selection.cardCosts[cardId] || 0;
    const reward = selection.cardRewards[cardId] || 0;

    if (reward > 0) {
      return { type: "reward", value: reward };
    } else if (cost > 0) {
      return { type: "cost", value: cost };
    } else if (cost === 0 && reward === 0) {
      return { type: "free" };
    }
    return null;
  };

  const canAffordCard = (cardId: string): boolean => {
    const currentTotalCost = selectedCardIds.reduce((sum, id) => {
      return sum + (selection.cardCosts[id] || 0);
    }, 0);
    const cardCost = selection.cardCosts[cardId] || 0;
    return currentTotalCost + cardCost <= playerCredits;
  };

  const handleCardSelect = (cardId: string) => {
    setSelectedCardIds((prev) => {
      if (prev.includes(cardId)) {
        // Deselect card
        return prev.filter((id) => id !== cardId);
      } else {
        // Select card - check if player can afford it
        if (canAffordCard(cardId)) {
          return [...prev, cardId];
        } else {
          return prev;
        }
      }
    });
  };

  const handleConfirm = () => {
    const minCards = selection.minCards;
    const maxCards = selection.maxCards;
    const selectedCount = selectedCardIds.length;

    // Validate selection bounds
    if (selectedCount < minCards || selectedCount > maxCards) {
      return; // Invalid selection
    }

    if (selectedCount > 0) {
      // Player has selected cards - commit immediately
      onSelectCards(selectedCardIds);
    } else if (minCards === 0 && !showConfirmation) {
      // First click with no selection when 0 is allowed - show confirmation
      setShowConfirmation(true);
    } else if (minCards === 0 && showConfirmation) {
      // Second click with no selection - confirm with empty selection
      onSelectCards([]);
    }
  };

  const handleCancel = () => {
    if (onCancel) {
      onCancel();
    } else {
      // If no cancel handler, treat as selecting 0 cards (if allowed)
      if (selection.minCards === 0) {
        onSelectCards([]);
      }
    }
  };

  const titleInfo = getTitleAndDescription(selection.source);
  const netGain = totalReward - totalCost;
  const isValidSelection =
    selectedCardIds.length >= selection.minCards &&
    selectedCardIds.length <= selection.maxCards;

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
          <p className="mt-2 mb-0 text-base text-white/80 max-[768px]:text-sm">
            {titleInfo.description}
          </p>
        </div>

        {/* Cards display */}
        <div className="flex-1 overflow-x-auto overflow-y-hidden p-8 flex items-center bg-[radial-gradient(ellipse_at_center,rgba(139,69,19,0.1)_0%,transparent_70%)] [&::-webkit-scrollbar]:h-2 [&::-webkit-scrollbar-track]:bg-white/5 [&::-webkit-scrollbar-track]:rounded [&::-webkit-scrollbar-thumb]:bg-white/20 [&::-webkit-scrollbar-thumb]:rounded [&::-webkit-scrollbar-thumb:hover]:bg-white/30 max-[768px]:p-5">
          <div className="flex gap-6 mx-auto py-5 max-[768px]:gap-4">
            {selection.availableCards.map((card: CardDto, index: number) => {
              const isSelected = selectedCardIds.includes(card.id);
              const badge = getCardBadge(card.id);
              const canAfford = canAffordCard(card.id);

              return (
                <div key={card.id} className="relative">
                  <SimpleGameCard
                    card={card}
                    isSelected={isSelected}
                    onSelect={handleCardSelect}
                    animationDelay={index * 100}
                    showCheckbox={true}
                  />
                  {/* Cost/Reward Badge */}
                  {badge && (
                    <div
                      className={`absolute top-2 right-2 px-2 py-1 rounded-md font-bold text-sm shadow-lg ${
                        badge.type === "reward"
                          ? "bg-[#4caf50] text-white"
                          : badge.type === "cost"
                            ? "bg-[#f44336] text-white"
                            : "bg-[#4caf50] text-white"
                      }`}
                    >
                      {badge.type === "reward"
                        ? `+${badge.value} MC`
                        : badge.type === "cost"
                          ? `${badge.value} MC`
                          : "FREE"}
                    </div>
                  )}
                  {/* Unaffordable overlay */}
                  {!canAfford && !isSelected && (
                    <div className="absolute inset-0 bg-black/60 rounded-lg flex items-center justify-center">
                      <span className="text-white/80 font-bold">
                        Can't Afford
                      </span>
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
              <MegaCreditIcon value={playerCredits} size="large" />
            </div>
            {totalCost > 0 && (
              <div className="flex items-center gap-3">
                <span className="text-sm text-white/60 uppercase tracking-[0.5px]">
                  Total Cost:
                </span>
                <MegaCreditIcon value={totalCost} size="large" />
              </div>
            )}
            {totalReward > 0 && (
              <div className="flex items-center gap-3">
                <span className="text-sm text-white/60 uppercase tracking-[0.5px]">
                  Total Reward:
                </span>
                <div className="flex items-center gap-1">
                  <span className="text-[#4caf50] font-bold text-lg">+</span>
                  <MegaCreditIcon value={totalReward} size="large" />
                </div>
              </div>
            )}
            {netGain !== 0 && (
              <div className="flex items-center gap-3">
                <span className="text-sm text-white/60 uppercase tracking-[0.5px]">
                  Net Gain:
                </span>
                <div className="flex items-center gap-1">
                  <span
                    className={`font-bold text-lg ${
                      netGain > 0 ? "text-[#4caf50]" : "text-[#f44336]"
                    }`}
                  >
                    {netGain > 0 ? "+" : ""}
                    {netGain}
                  </span>
                  <span className="text-white/80 text-sm">MC</span>
                </div>
              </div>
            )}
          </div>

          <div className="flex items-center gap-6 max-[768px]:w-full max-[768px]:flex-col max-[768px]:gap-3">
            <div className="text-sm">
              {selectedCardIds.length === 0 ? (
                showConfirmation ? (
                  <span className="text-[#ff9800]">
                    Are you sure you don't want to select any cards?
                  </span>
                ) : (
                  <span className="text-white/70">No cards selected</span>
                )
              ) : (
                <span className="text-white/70">
                  {selectedCardIds.length} card
                  {selectedCardIds.length !== 1 ? "s" : ""} selected
                </span>
              )}
            </div>
            <div className="flex gap-3 items-center">
              {(onCancel || selection.minCards === 0) && (
                <button
                  className="py-3 px-6 bg-space-black-darker/60 border-2 border-space-blue-800/60 rounded-lg text-white font-medium cursor-pointer transition-all duration-200 whitespace-nowrap hover:-translate-y-px hover:bg-space-black-darker/80 hover:border-space-blue-600 active:translate-y-0"
                  onClick={handleCancel}
                >
                  {onCancel ? "Cancel" : "Skip"}
                </button>
              )}
              <button
                className="py-4 px-8 bg-space-black-darker/90 border-2 border-space-blue-800 rounded-xl text-xl font-bold text-white cursor-pointer transition-all duration-300 text-shadow-dark shadow-[0_4px_20px_rgba(30,60,150,0.3)] whitespace-nowrap hover:enabled:bg-space-black-darker/95 hover:enabled:border-space-blue-600 hover:enabled:-translate-y-0.5 hover:enabled:shadow-glow active:enabled:translate-y-0 disabled:bg-gray-700/50 disabled:border-gray-500/30 disabled:cursor-not-allowed disabled:transform-none disabled:shadow-none disabled:opacity-60 max-[768px]:w-full max-[768px]:py-3 max-[768px]:px-6 max-[768px]:text-lg"
                onClick={handleConfirm}
                disabled={!isValidSelection || totalCost > playerCredits}
              >
                {showConfirmation
                  ? "Confirm Skip"
                  : selection.source === "sell-patents"
                    ? "Sell Cards"
                    : "Confirm Selection"}
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default PendingCardSelectionOverlay;
