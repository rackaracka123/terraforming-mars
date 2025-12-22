import React from "react";
import SimpleGameCard from "../cards/SimpleGameCard.tsx";
import GameIcon from "../display/GameIcon.tsx";
import {
  PendingCardSelectionDto,
  ResourceTypeCredit,
} from "../../../types/generated/api-types.ts";
import { useCardSelection } from "../../../hooks/useCardSelection.ts";
import {
  OVERLAY_BACKGROUND_CLASS,
  OVERLAY_CONTAINER_CLASS,
  OVERLAY_HEADER_CLASS,
  OVERLAY_TITLE_CLASS,
  OVERLAY_DESCRIPTION_CLASS,
  OVERLAY_CARDS_CONTAINER_CLASS,
  OVERLAY_CARDS_INNER_CLASS,
  OVERLAY_FOOTER_CLASS,
  PRIMARY_BUTTON_CLASS,
  SECONDARY_BUTTON_CLASS,
  RESOURCE_LABEL_CLASS,
} from "./overlayStyles.ts";

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
  const {
    selectedCardIds,
    totalCost,
    totalReward,
    showConfirmation,
    isValidSelection,
    handleCardSelect,
    handleConfirm: handleCardConfirm,
    canAffordCard,
  } = useCardSelection({
    cards: selection.availableCards,
    isOpen,
    playerCredits,
    getCardCost: (cardId) => selection.cardCosts[cardId] || 0,
    getCardReward: (cardId) => selection.cardRewards[cardId] || 0,
    minCards: selection.minCards,
    maxCards: selection.maxCards,
  });

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

  const handleConfirm = () => {
    handleCardConfirm(onSelectCards);
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

  return (
    <div className="fixed inset-0 z-[1000] flex items-center justify-center animate-[fadeIn_0.3s_ease]">
      {/* Translucent background */}
      <div className={OVERLAY_BACKGROUND_CLASS} />

      {/* Content container */}
      <div className={OVERLAY_CONTAINER_CLASS}>
        {/* Header */}
        <div className={OVERLAY_HEADER_CLASS}>
          <h2 className={OVERLAY_TITLE_CLASS}>{titleInfo.title}</h2>
          <p className={OVERLAY_DESCRIPTION_CLASS}>{titleInfo.description}</p>
        </div>

        {/* Cards display */}
        <div className={OVERLAY_CARDS_CONTAINER_CLASS}>
          <div className={OVERLAY_CARDS_INNER_CLASS}>
            {selection.availableCards.map((card, index) => {
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
        <div className={OVERLAY_FOOTER_CLASS}>
          <div className="flex gap-8 items-center max-[768px]:w-full max-[768px]:justify-between max-[768px]:flex-wrap">
            <div className="flex items-center gap-3">
              <span className={RESOURCE_LABEL_CLASS}>Your Credits:</span>
              <GameIcon
                iconType={ResourceTypeCredit}
                amount={playerCredits}
                size="large"
              />
            </div>
            {totalCost > 0 && (
              <div className="flex items-center gap-3">
                <span className={RESOURCE_LABEL_CLASS}>Total Cost:</span>
                <GameIcon
                  iconType={ResourceTypeCredit}
                  amount={totalCost}
                  size="large"
                />
              </div>
            )}
            {totalReward > 0 && (
              <div className="flex items-center gap-3">
                <span className={RESOURCE_LABEL_CLASS}>Total Reward:</span>
                <div className="flex items-center gap-1">
                  <span className="text-[#4caf50] font-bold text-lg">+</span>
                  <GameIcon
                    iconType={ResourceTypeCredit}
                    amount={totalReward}
                    size="large"
                  />
                </div>
              </div>
            )}
            {netGain !== 0 && (
              <div className="flex items-center gap-3">
                <span className={RESOURCE_LABEL_CLASS}>Net Gain:</span>
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
                  className={SECONDARY_BUTTON_CLASS}
                  onClick={handleCancel}
                >
                  {onCancel ? "Cancel" : "Skip"}
                </button>
              )}
              <button
                className={PRIMARY_BUTTON_CLASS}
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
