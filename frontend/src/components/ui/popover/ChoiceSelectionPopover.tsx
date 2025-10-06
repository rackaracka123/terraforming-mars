import React, { useEffect, useRef, useState } from "react";
import {
  ChoiceDto,
  CardBehaviorDto,
  ResourcesDto,
} from "../../../types/generated/api-types.ts";
import BehaviorSection from "../cards/BehaviorSection.tsx";

interface ChoiceItem {
  index: number;
  choice: ChoiceDto;
}

interface ChoiceSelectionPopoverProps {
  cardId: string;
  cardName: string;
  behaviors: CardBehaviorDto[];
  behaviorIndex: number;
  onChoiceSelect: (choiceIndex: number) => void;
  onCancel: () => void;
  isVisible: boolean;
  isAction?: boolean; // True if this is for an action, false/undefined if for card play
  playerResources?: ResourcesDto;
  resourceStorage?: { [key: string]: number };
}

const ChoiceSelectionPopover: React.FC<ChoiceSelectionPopoverProps> = ({
  cardId,
  cardName,
  behaviors,
  behaviorIndex,
  onChoiceSelect,
  onCancel,
  isVisible,
  isAction = false,
  playerResources,
  resourceStorage,
}) => {
  const popoverRef = useRef<HTMLDivElement>(null);
  const [isClosing, setIsClosing] = useState(false);

  // Get the behavior with choices
  const behavior = behaviors?.[behaviorIndex];
  const choices: ChoiceItem[] =
    behavior?.choices?.map((choice, index) => ({
      index,
      choice,
    })) || [];

  // Helper to check if a choice is affordable
  const isChoiceAffordable = (choice: ChoiceDto): boolean => {
    if (!playerResources) return true; // If no resources provided, show as affordable

    const storage = resourceStorage || {};

    // Check all inputs for this choice
    const inputs = [
      ...(behavior?.inputs || []), // Base behavior inputs
      ...(choice.inputs || []), // Choice-specific inputs
    ];

    for (const input of inputs) {
      switch (input.type) {
        case "credits":
          if (playerResources.credits < input.amount) return false;
          break;
        case "steel":
          if (playerResources.steel < input.amount) return false;
          break;
        case "titanium":
          if (playerResources.titanium < input.amount) return false;
          break;
        case "plants":
          if (playerResources.plants < input.amount) return false;
          break;
        case "energy":
          if (playerResources.energy < input.amount) return false;
          break;
        case "heat":
          if (playerResources.heat < input.amount) return false;
          break;

        // Card storage resources
        case "animals":
        case "microbes":
        case "floaters":
        case "science":
        case "asteroid":
          if (input.target === "self-card") {
            const cardStorage = storage[cardId] || 0;
            if (cardStorage < input.amount) return false;
          }
          break;
      }
    }

    return true;
  };

  const handleCancelClick = () => {
    setIsClosing(true);
    setTimeout(() => {
      setIsClosing(false);
      onCancel();
    }, 200); // Match fadeOut animation duration
  };

  const handleChoiceClick = (choiceIndex: number) => {
    // Call immediately without delay for choice selection
    onChoiceSelect(choiceIndex);
  };

  useEffect(() => {
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        handleCancelClick();
      }
    };

    const handleClickOutside = (event: MouseEvent) => {
      // Only close on left click (button 0), ignore right click (button 2) and middle click (button 1)
      if (
        event.button === 0 &&
        popoverRef.current &&
        !popoverRef.current.contains(event.target as Node)
      ) {
        handleCancelClick();
      }
    };

    const preventScroll = (event: WheelEvent | TouchEvent) => {
      event.preventDefault();
      event.stopPropagation();
    };

    if (isVisible) {
      // Prevent body scroll
      document.body.style.overflow = "hidden";

      // Add event listeners
      document.addEventListener("keydown", handleEscape);
      document.addEventListener("mousedown", handleClickOutside);
      document.addEventListener("wheel", preventScroll, { passive: false });
      document.addEventListener("touchmove", preventScroll, { passive: false });
    }

    return () => {
      // Restore body scroll
      document.body.style.overflow = "";

      // Remove event listeners
      document.removeEventListener("keydown", handleEscape);
      document.removeEventListener("mousedown", handleClickOutside);
      document.removeEventListener("wheel", preventScroll);
      document.removeEventListener("touchmove", preventScroll);
    };
  }, [isVisible, onCancel]);

  if (!isVisible || choices.length === 0) return null;

  return (
    <div className="fixed top-0 left-0 right-0 bottom-0 z-[10002] flex items-center justify-center pointer-events-auto overflow-hidden">
      <div
        className={`
          min-w-[240px] w-fit max-w-[90vw] max-h-[500px]
          bg-space-black-darker/95
          border-2 border-space-blue-500
          rounded-xl
          shadow-[0_15px_40px_rgba(0,0,0,0.8),0_0_15px_rgba(30,60,150,1)]
          backdrop-blur-space
          flex flex-col overflow-hidden isolate
          pointer-events-auto
          ${isClosing ? "animate-fadeOut" : "animate-popIn"}
        `}
        ref={popoverRef}
      >
        {/* Header */}
        <div className="py-[15px] px-5 bg-black/40 border-b border-b-space-blue-500/60">
          <h3 className="m-0 font-orbitron text-white text-base font-bold text-shadow-glow">
            {isAction ? "Choose Action Effect" : "Choose One Effect"}
          </h3>
          <div className="text-white/60 text-xs text-shadow-glow mt-1">
            {cardName}
          </div>
        </div>

        {/* Choices Container */}
        <div className="flex-1 overflow-y-auto p-2.5 scrollbar-thin scrollbar-thumb-space-blue-500/50 scrollbar-track-space-blue-900/30">
          {choices.map(({ index, choice }) => {
            // Convert the choice into a CardBehaviorDto that BehaviorSection can render
            const behaviorForChoice: CardBehaviorDto = {
              triggers: behavior?.triggers || [],
              inputs: choice.inputs,
              outputs: choice.outputs,
              choices: undefined, // Don't show nested choices
            };

            const delay = index * 0.05;
            const isAffordable = isChoiceAffordable(choice);

            return (
              <div
                key={index}
                className={`
                  bg-black/30
                  border-2 border-space-blue-500/40
                  rounded-[10px] px-3.5 py-3
                  mb-2
                  transition-all duration-[250ms] ease-out
                  animate-choiceSlideIn
                  ${
                    isAffordable
                      ? "cursor-pointer hover:translate-y-[-2px] hover:scale-[1.01] hover:border-space-blue-500/80 hover:bg-black/50 hover:shadow-[0_4px_16px_rgba(30,60,150,0.5)]"
                      : "cursor-default"
                  }
                `}
                style={{ animationDelay: `${delay}s` }}
                onClick={() => isAffordable && handleChoiceClick(index)}
                title={
                  isAffordable
                    ? "Click to select this choice"
                    : "Cannot afford this choice"
                }
              >
                <div className="text-white/60 text-[11px] font-semibold uppercase tracking-wider mb-3 text-shadow-glow">
                  Choice {index + 1}
                </div>
                <div className="flex items-center justify-center w-full [&>div]:!relative [&>div]:!bottom-auto [&>div]:!left-auto [&>div]:!right-auto [&>div]:w-auto [&>div]:max-w-full [&>div:hover]:!transform-none [&>div:hover]:!shadow-none [&>div:hover]:!filter-none">
                  <BehaviorSection
                    behaviors={[behaviorForChoice]}
                    playerResources={playerResources}
                    resourceStorage={resourceStorage}
                    cardId={cardId}
                    greyOutAll={!isAffordable}
                  />
                </div>
              </div>
            );
          })}
        </div>

        {/* Footer */}
        <div className="px-4 py-3 bg-black/40 border-t border-space-blue-500/60 flex justify-center">
          <button
            className="
              bg-space-blue-600/50
              border-2 border-space-blue-500/60
              rounded-md text-white text-xs font-semibold
              px-6 py-2 cursor-pointer
              transition-all duration-200
              text-shadow-glow font-orbitron
              shadow-[0_0_8px_rgba(30,60,150,0.4)]
              hover:bg-space-blue-500/60
              hover:border-space-blue-500/80
              hover:translate-y-[-2px]
              hover:shadow-[0_0_12px_rgba(30,60,150,0.6)]
            "
            onClick={handleCancelClick}
          >
            Cancel
          </button>
        </div>
      </div>

      <style>{`
        @keyframes popIn {
          from {
            opacity: 0;
            transform: scale(0.9) translateY(-20px);
          }
          to {
            opacity: 1;
            transform: scale(1) translateY(0);
          }
        }

        @keyframes fadeOut {
          from {
            opacity: 1;
          }
          to {
            opacity: 0;
          }
        }

        @keyframes choiceSlideIn {
          from {
            opacity: 0;
            transform: translateX(-20px);
          }
          to {
            opacity: 1;
            transform: translateX(0);
          }
        }

        .animate-popIn {
          animation: popIn 0.25s ease-out;
        }

        .animate-fadeOut {
          animation: fadeOut 0.2s ease-out forwards;
        }

        .animate-choiceSlideIn {
          animation: choiceSlideIn 0.3s ease-out both;
        }

        /* Media queries */
        @media (max-width: 768px) {
          .min-w-\\[240px\\] {
            min-width: 180px;
          }
          .max-w-\\[90vw\\] {
            max-width: 95vw;
          }
        }
      `}</style>
    </div>
  );
};

export default ChoiceSelectionPopover;
