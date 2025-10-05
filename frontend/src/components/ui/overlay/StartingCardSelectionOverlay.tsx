import React, { useEffect, useState } from "react";
import SimpleGameCard from "../cards/SimpleGameCard.tsx";
import CorporationCard from "../cards/CorporationCard.tsx";
import MegaCreditIcon from "../display/MegaCreditIcon.tsx";
import { CardDto } from "../../../types/generated/api-types.ts";
import { fetchCorporations } from "../../../utils/cardPlayabilityUtils.ts";

interface StartingCardSelectionOverlayProps {
  isOpen: boolean;
  cards: CardDto[];
  availableCorporations: string[];
  playerCredits: number;
  onSelectCards: (selectedCardIds: string[], corporationId: string) => void;
}

const StartingCardSelectionOverlay: React.FC<
  StartingCardSelectionOverlayProps
> = ({
  isOpen,
  cards,
  availableCorporations,
  playerCredits,
  onSelectCards,
}) => {
  const [selectedCardIds, setSelectedCardIds] = useState<string[]>([]);
  const [selectedCorporationId, setSelectedCorporationId] = useState<
    string | null
  >(null);
  const [corporationCards, setCorporationCards] = useState<CardDto[]>([]);
  const [totalCost, setTotalCost] = useState(0);
  const [showConfirmation, setShowConfirmation] = useState(false);
  const [currentStep, setCurrentStep] = useState<"corporation" | "cards">(
    "corporation",
  );

  // Fetch corporation cards when corporations are available
  useEffect(() => {
    const loadCorporations = async () => {
      if (availableCorporations.length === 0) {
        setCorporationCards([]);
        return;
      }

      try {
        const allCorporations = await fetchCorporations();
        const corps = allCorporations.filter((corp) =>
          availableCorporations.includes(corp.id),
        );

        setCorporationCards(corps);
      } catch (error) {
        console.error("Failed to load corporation cards:", error);
        setCorporationCards([]);
      }
    };

    void loadCorporations();
  }, [availableCorporations]);

  // Initialize selection when overlay opens
  useEffect(() => {
    if (isOpen && cards.length > 0) {
      setSelectedCardIds([]);
      setSelectedCorporationId(null);
      setShowConfirmation(false);
      setTotalCost(0);
      setCurrentStep("corporation");
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
        const newSelection = prev.filter((id) => id !== cardId);
        return newSelection;
      } else {
        // Select card - check if player can afford it
        const newSelection = [...prev, cardId];
        const costForNewCard = newSelection.length > 1 ? 3 : 0;
        const newTotalCost = totalCost + costForNewCard;

        if (newTotalCost <= playerCredits) {
          return newSelection;
        } else {
          return prev;
        }
      }
    });
  };

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

    if (selectedCardIds.length > 0) {
      // Player has selected cards - commit immediately
      onSelectCards(selectedCardIds, selectedCorporationId);
    } else if (!showConfirmation) {
      // First click with no selection - show confirmation
      setShowConfirmation(true);
    } else {
      // Second click with no selection - confirm with empty selection
      onSelectCards([], selectedCorporationId);
    }
  };

  return (
    <div className="fixed inset-0 z-[1000] flex items-center justify-center animate-[fadeIn_0.3s_ease]">
      {/* Translucent background */}
      <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" />

      {/* Content container */}
      <div className="relative z-[1] w-[90%] max-w-[1400px] max-h-[90vh] flex flex-col bg-space-black-darker/95 border-2 border-space-blue-400 rounded-[20px] overflow-hidden backdrop-blur-space shadow-[0_20px_60px_rgba(0,0,0,0.6),0_0_60px_rgba(30,60,150,0.5)] max-[768px]:w-full max-[768px]:h-screen max-[768px]:max-h-screen max-[768px]:rounded-none">
        {/* Header */}
        <div className="py-6 px-8 bg-black/40 border-b border-space-blue-600 max-[768px]:p-5">
          <h2 className="m-0 font-orbitron text-[28px] font-bold text-white text-shadow-glow tracking-wider max-[768px]:text-2xl">
            {currentStep === "corporation"
              ? "Select Your Corporation"
              : "Select Starting Cards"}
          </h2>
          <p className="mt-2 mb-0 text-base text-white/80 max-[768px]:text-sm">
            {currentStep === "corporation"
              ? "Choose your corporation to begin the game"
              : "Choose your starting cards. First card is FREE, additional cards cost 3 MC each."}
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
        {currentStep === "corporation" && corporationCards.length > 0 && (
          <div className="flex-1 overflow-y-auto p-8 bg-black/20 max-[768px]:p-5">
            <div className="flex gap-4 justify-center flex-wrap">
              {corporationCards.map((corp) => (
                <div key={corp.id} className="w-[400px] max-[768px]:w-full">
                  <CorporationCard
                    corporation={{
                      id: corp.id,
                      name: corp.name,
                      description: corp.description,
                      startingMegaCredits: corp.startingCredits ?? 0,
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
        <div className="py-6 px-8 bg-black/40 border-t border-space-blue-600 flex justify-between items-center max-[768px]:p-5 max-[768px]:flex-col max-[768px]:gap-5">
          {/* Step 1: Corporation Footer */}
          {currentStep === "corporation" && (
            <>
              <div className="text-sm text-white/70">
                {selectedCorporationId
                  ? "Corporation selected"
                  : "Please select a corporation"}
              </div>
              <button
                className="py-4 px-8 bg-space-black-darker/90 border-2 border-space-blue-800 rounded-xl text-xl font-bold text-white cursor-pointer transition-all duration-300 text-shadow-dark shadow-[0_4px_20px_rgba(30,60,150,0.3)] hover:enabled:bg-space-black-darker/95 hover:enabled:border-space-blue-600 hover:enabled:-translate-y-0.5 hover:enabled:shadow-glow active:enabled:translate-y-0 disabled:bg-gray-700/50 disabled:border-gray-500/30 disabled:cursor-not-allowed disabled:transform-none disabled:shadow-none disabled:opacity-60 max-[768px]:w-full max-[768px]:py-3 max-[768px]:px-6 max-[768px]:text-lg"
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
                <div className="flex items-center gap-3">
                  <span className="text-sm text-white/60 uppercase tracking-[0.5px]">
                    Your Credits:
                  </span>
                  <MegaCreditIcon value={playerCredits} size="large" />
                </div>
                <div className="flex items-center gap-3">
                  <span className="text-sm text-white/60 uppercase tracking-[0.5px]">
                    Total Cost:
                  </span>
                  {totalCost > 0 ? (
                    <MegaCreditIcon value={totalCost} size="large" />
                  ) : (
                    <span className="!text-[#4caf50] font-bold tracking-[1px]">
                      FREE
                    </span>
                  )}
                </div>
              </div>

              <div className="flex items-center gap-4 max-[768px]:w-full max-[768px]:flex-col max-[768px]:gap-3">
                <button
                  className="py-3 px-6 bg-transparent border-2 border-space-blue-800 rounded-xl text-base font-semibold text-white cursor-pointer transition-all duration-300 hover:bg-space-black-darker/50 hover:border-space-blue-600 max-[768px]:w-full"
                  onClick={handleNextToCorporation}
                >
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
                  className="py-4 px-8 bg-space-black-darker/90 border-2 border-space-blue-800 rounded-xl text-xl font-bold text-white cursor-pointer transition-all duration-300 text-shadow-dark shadow-[0_4px_20px_rgba(30,60,150,0.3)] hover:enabled:bg-space-black-darker/95 hover:enabled:border-space-blue-600 hover:enabled:-translate-y-0.5 hover:enabled:shadow-glow active:enabled:translate-y-0 disabled:bg-gray-700/50 disabled:border-gray-500/30 disabled:cursor-not-allowed disabled:transform-none disabled:shadow-none disabled:opacity-60 max-[768px]:w-full max-[768px]:py-3 max-[768px]:px-6 max-[768px]:text-lg"
                  onClick={handleConfirm}
                  disabled={totalCost > playerCredits}
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
