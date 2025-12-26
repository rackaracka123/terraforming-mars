import React, { useEffect, useState } from "react";
import SimpleGameCard from "../cards/SimpleGameCard.tsx";
import CorporationCard from "../cards/CorporationCard.tsx";
import GameIcon from "../display/GameIcon.tsx";
import { CardDto, ResourceTypeCredit } from "../../../types/generated/api-types.ts";
import { getCorporationBorderColor } from "@/utils/corporationColors.ts";
import { useCardSelection } from "../../../hooks/useCardSelection.ts";
import {
  OVERLAY_CONTAINER_CLASS,
  OVERLAY_HEADER_CLASS,
  OVERLAY_TITLE_CLASS,
  OVERLAY_DESCRIPTION_CLASS,
  OVERLAY_FOOTER_CLASS,
  PRIMARY_BUTTON_CLASS,
  SECONDARY_BUTTON_CLASS,
  RESOURCE_LABEL_CLASS,
  RESOURCE_DISPLAY_CLASS,
} from "./overlayStyles.ts";

interface StartingCardSelectionOverlayProps {
  isOpen: boolean;
  cards: CardDto[];
  availableCorporations: CardDto[];
  playerCredits: number;
  onSelectCards: (selectedCardIds: string[], corporationId: string) => void;
}

const StartingCardSelectionOverlay: React.FC<StartingCardSelectionOverlayProps> = ({
  isOpen,
  cards,
  availableCorporations,
  playerCredits,
  onSelectCards,
}) => {
  const [selectedCorporationId, setSelectedCorporationId] = useState<string | null>(null);
  const [currentStep, setCurrentStep] = useState<"corporation" | "cards">("corporation");

  const {
    selectedCardIds,
    totalCost,
    showConfirmation,
    isValidSelection,
    handleCardSelect,
    handleConfirm: handleCardConfirm,
  } = useCardSelection({
    cards,
    isOpen: isOpen && currentStep === "cards",
    playerCredits,
    costPerCard: 3,
    minCards: 0,
  });

  // Initialize selection when overlay opens
  useEffect(() => {
    if (isOpen && cards.length > 0) {
      setSelectedCorporationId(null);
      setCurrentStep("corporation");
    }
  }, [isOpen, cards]);

  if (!isOpen || cards.length === 0) return null;

  const handleCorporationSelect = (corporationId: string) => {
    setSelectedCorporationId(corporationId);
  };

  const handleNextToCorporation = () => {
    setCurrentStep("corporation");
  };

  const handleNextToCards = () => {
    if (!selectedCorporationId) {
      return;
    }
    setCurrentStep("cards");
  };

  const handleConfirm = () => {
    // Must select a corporation before confirming
    if (!selectedCorporationId) {
      return;
    }

    handleCardConfirm((cardIds) => {
      onSelectCards(cardIds, selectedCorporationId);
    });
  };

  return (
    <div className="fixed inset-0 z-[1000] flex items-center justify-center animate-[fadeIn_0.3s_ease]">
      {/* Content container */}
      <div className={OVERLAY_CONTAINER_CLASS}>
        {/* Header */}
        <div className={OVERLAY_HEADER_CLASS}>
          <h2 className={OVERLAY_TITLE_CLASS}>
            {currentStep === "corporation" ? "Select Your Corporation" : "Select Starting Cards"}
          </h2>
          <p className={OVERLAY_DESCRIPTION_CLASS}>
            {currentStep === "corporation"
              ? "Choose your corporation to begin the game"
              : "Choose your starting cards. Each card costs 3 MC."}
          </p>
          <div className="mt-3 flex gap-2 items-center">
            <div
              className={`h-2 w-2 rounded-full ${currentStep === "corporation" ? "bg-space-blue-400" : "bg-white/30"}`}
            />
            <div
              className={`h-2 w-2 rounded-full ${currentStep === "cards" ? "bg-space-blue-400" : "bg-white/30"}`}
            />
          </div>
        </div>

        {/* Step 1: Corporation Selection */}
        {currentStep === "corporation" && availableCorporations.length > 0 && (
          <div className="flex-1 overflow-y-auto p-8 bg-black/20 max-[768px]:p-5">
            <div className="flex gap-4 justify-center flex-wrap">
              {availableCorporations.map((corp) => (
                <div key={corp.id} className="w-[400px] max-[768px]:w-full">
                  <CorporationCard
                    corporation={{
                      id: corp.id,
                      name: corp.name,
                      description: corp.description,
                      startingMegaCredits: corp.startingResources?.credits ?? 0,
                      startingProduction: corp.startingProduction
                        ? {
                            credits: corp.startingProduction.credits,
                            steel: corp.startingProduction.steel,
                            titanium: corp.startingProduction.titanium,
                            plants: corp.startingProduction.plants,
                            energy: corp.startingProduction.energy,
                            heat: corp.startingProduction.heat,
                          }
                        : undefined,
                      startingResources: corp.startingResources
                        ? {
                            credits: corp.startingResources.credits,
                            steel: corp.startingResources.steel,
                            titanium: corp.startingResources.titanium,
                            plants: corp.startingResources.plants,
                            energy: corp.startingResources.energy,
                            heat: corp.startingResources.heat,
                          }
                        : undefined,
                      behaviors: corp.behaviors,
                      logoPath: undefined,
                    }}
                    isSelected={selectedCorporationId === corp.id}
                    onSelect={handleCorporationSelect}
                    borderColor={getCorporationBorderColor(corp.name)}
                  />
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Step 2: Cards Selection */}
        {currentStep === "cards" && (
          <div className="flex-1 overflow-x-auto overflow-y-hidden p-8 flex items-center bg-[radial-gradient(ellipse_at_center,rgba(139,69,19,0.1)_0%,transparent_70%)] [&::-webkit-scrollbar]:h-2 [&::-webkit-scrollbar-track]:bg-white/5 [&::-webkit-scrollbar-track]:rounded [&::-webkit-scrollbar-thumb]:bg-white/20 [&::-webkit-scrollbar-thumb]:rounded [&::-webkit-scrollbar-thumb:hover]:bg-white/30 max-[768px]:p-5">
            <div className="flex gap-6 mx-auto py-5 max-[768px]:gap-4">
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
        )}

        {/* Footer with navigation buttons */}
        <div className={OVERLAY_FOOTER_CLASS}>
          {/* Step 1: Corporation Footer */}
          {currentStep === "corporation" && (
            <>
              <div className="text-sm text-white/70">
                {selectedCorporationId ? "Corporation selected" : "Please select a corporation"}
              </div>
              <button
                className={PRIMARY_BUTTON_CLASS}
                onClick={handleNextToCards}
                disabled={!selectedCorporationId}
              >
                Next: Select Cards →
              </button>
            </>
          )}

          {/* Step 2: Cards Footer */}
          {currentStep === "cards" && (
            <>
              <div className="flex gap-8 items-center max-[768px]:w-full max-[768px]:justify-between">
                <div className={RESOURCE_DISPLAY_CLASS}>
                  <span className={RESOURCE_LABEL_CLASS}>Your Credits:</span>
                  <GameIcon iconType={ResourceTypeCredit} amount={playerCredits} size="large" />
                </div>
                <div className={RESOURCE_DISPLAY_CLASS}>
                  <span className={RESOURCE_LABEL_CLASS}>Total Cost:</span>
                  {totalCost > 0 ? (
                    <GameIcon iconType={ResourceTypeCredit} amount={totalCost} size="large" />
                  ) : (
                    <span className="!text-[#4caf50] font-bold tracking-[1px]">FREE</span>
                  )}
                </div>
              </div>

              <div className="flex items-center gap-4 max-[768px]:w-full max-[768px]:flex-col max-[768px]:gap-3">
                <button className={SECONDARY_BUTTON_CLASS} onClick={handleNextToCorporation}>
                  ← Back
                </button>

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

                <button
                  className={PRIMARY_BUTTON_CLASS}
                  onClick={handleConfirm}
                  disabled={!isValidSelection}
                >
                  {showConfirmation ? "Confirm Skip" : "Confirm Selection"}
                </button>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
};

export default StartingCardSelectionOverlay;
