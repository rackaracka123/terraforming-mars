import React, { useEffect, useRef } from "react";
import {
  PlayerActionDto,
  GameDto,
} from "../../../types/generated/api-types.ts";
import BehaviorSection from "../cards/BehaviorSection.tsx";
import {
  canPerformActions,
  hasActionsAvailable,
} from "../../../utils/actionUtils.ts";

// Utility function to check if an action is affordable and available
const isActionAvailable = (
  action: PlayerActionDto,
  gameState?: GameDto,
): boolean => {
  // Check if action has been played this generation
  if (action.playCount > 0) {
    return false;
  }

  // Check if player can afford the action's input costs
  if (!gameState?.currentPlayer) {
    return false;
  }

  const playerResources = gameState.currentPlayer.resources;
  const actionInputs = action.behavior.inputs || [];

  for (const input of actionInputs) {
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
      // Add more resource types as needed
    }
  }

  return true;
};

interface ActionsPopoverProps {
  isVisible: boolean;
  onClose: () => void;
  actions: PlayerActionDto[];
  playerName?: string;
  onActionSelect?: (action: PlayerActionDto) => void;
  onOpenDetails?: () => void;
  anchorRef: React.RefObject<HTMLElement>;
  gameState?: GameDto;
}

const ActionsPopover: React.FC<ActionsPopoverProps> = ({
  isVisible,
  onClose,
  actions,
  playerName: _playerName = "Player",
  onActionSelect,
  onOpenDetails,
  anchorRef,
  gameState,
}) => {
  const popoverRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        onClose();
      }
    };

    const handleClickOutside = (event: MouseEvent) => {
      if (
        popoverRef.current &&
        !popoverRef.current.contains(event.target as Node) &&
        anchorRef.current &&
        !anchorRef.current.contains(event.target as Node)
      ) {
        onClose();
      }
    };

    if (isVisible) {
      document.addEventListener("keydown", handleEscape);
      document.addEventListener("mousedown", handleClickOutside);
    }

    return () => {
      document.removeEventListener("keydown", handleEscape);
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [isVisible, onClose, anchorRef]);

  if (!isVisible) return null;

  // Determine if actions can be played using utility function
  const isCurrentPlayerTurn =
    gameState?.currentTurn === gameState?.viewingPlayerId;
  const hasActionsLeft = hasActionsAvailable(
    gameState?.currentPlayer?.availableActions,
  );

  // Actions should be clickable only if all conditions are met
  const canPlayActions = canPerformActions(gameState);

  const handleActionClick = (action: PlayerActionDto) => {
    if (onActionSelect) {
      onActionSelect(action);
      onClose();
    }
  };

  return (
    <div
      className="fixed bottom-[85px] right-[30px] w-[320px] max-h-[400px] bg-space-black-darker/95 border-2 border-[#ff6464] rounded-xl shadow-[0_15px_40px_rgba(0,0,0,0.8),0_0_15px_#ff6464] backdrop-blur-space z-[10001] animate-[popoverSlideUp_0.3s_ease-out] flex flex-col overflow-hidden isolate pointer-events-auto max-[768px]:w-[280px] max-[768px]:right-[15px] max-[768px]:bottom-[70px]"
      ref={popoverRef}
    >
      <div className="absolute -bottom-2 right-[50px] w-0 h-0 border-l-[8px] border-l-transparent border-r-[8px] border-r-transparent border-t-[8px] border-t-[#ff6464] max-[768px]:right-[40px]" />

      <div className="flex items-center justify-between py-[15px] px-5 bg-black/40 border-b border-b-[#ff6464]/60">
        <div className="flex items-center gap-2.5">
          <h3 className="m-0 font-orbitron text-white text-base font-bold text-shadow-glow">
            Card Actions
          </h3>
        </div>
        <div className="flex items-center gap-2">
          <div className="text-white/80 text-xs bg-[#ff6464]/20 py-1 px-2 rounded-md border border-[#ff6464]/30">
            {actions.length} available
          </div>
          {onOpenDetails && (
            <button
              className="bg-space-black-darker/90 border-2 border-[#ff6464] rounded-lg text-white text-[11px] font-semibold py-1 px-2.5 cursor-pointer transition-all duration-200 text-shadow-dark pointer-events-auto relative z-[1] hover:bg-space-black-darker/95 hover:border-[#ff6464] hover:-translate-y-px hover:shadow-[0_2px_8px_#ff646460]"
              onClick={() => {
                onOpenDetails();
                onClose();
              }}
              title="Open detailed actions view"
            >
              Details
            </button>
          )}
        </div>
      </div>

      <div className="flex-1 overflow-y-auto [scrollbar-width:thin] [scrollbar-color:#ff6464_rgba(30,60,150,0.3)] [&::-webkit-scrollbar]:w-1.5 [&::-webkit-scrollbar-track]:bg-[rgba(30,60,150,0.3)] [&::-webkit-scrollbar-track]:rounded [&::-webkit-scrollbar-thumb]:bg-[#ff6464]/70 [&::-webkit-scrollbar-thumb]:rounded [&::-webkit-scrollbar-thumb:hover]:bg-[#ff6464]">
        {actions.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-10 px-5 text-center">
            <img
              src="/assets/misc/corpCard.png"
              alt="No actions"
              className="w-10 h-10 mb-[15px] opacity-60"
            />
            <div className="text-white text-sm font-medium mb-2">
              No card actions available
            </div>
            <div className="text-white/60 text-xs leading-[1.4]">
              Play cards with manual triggers to gain actions
            </div>
          </div>
        ) : (
          <div className="p-2 flex flex-col gap-2">
            {actions.map((action, index) => {
              const isAvailable = isActionAvailable(action, gameState);
              const isActionPlayable = canPlayActions && isAvailable;

              return (
                <div
                  key={`${action.cardId}-${action.behaviorIndex}`}
                  className={`flex items-center gap-3 py-2.5 px-[15px] bg-space-black-darker/60 border border-[#ff6464]/30 rounded-lg cursor-pointer transition-all duration-300 animate-[actionSlideIn_0.4s_ease-out_both] max-[768px]:py-2 max-[768px]:px-3 ${!isActionPlayable ? "opacity-50 !bg-space-black-darker/30 !border-[#ff6464]/15 !transform-none !shadow-none" : "hover:translate-x-1 hover:border-[#ff6464] hover:bg-space-black-darker/80 hover:shadow-[0_4px_15px_#ff646440]"}`}
                  onClick={() => isActionPlayable && handleActionClick(action)}
                  style={{
                    animationDelay: `${index * 0.05}s`,
                    cursor: isActionPlayable ? "pointer" : "default",
                  }}
                  title={
                    !canPlayActions
                      ? !isCurrentPlayerTurn
                        ? "Wait for your turn"
                        : !hasActionsLeft
                          ? "No actions remaining"
                          : "Actions not available in this phase"
                      : !isAvailable
                        ? action.playCount > 0
                          ? "Already played this generation"
                          : "Cannot afford this action"
                        : "Click to play this action"
                  }
                >
                  <div className="flex flex-col gap-2 flex-1">
                    <div className="text-white/70 text-[11px] font-medium uppercase tracking-[0.5px] [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] leading-[1.2] opacity-80 flex items-center gap-2 max-[768px]:text-[10px]">
                      {action.cardName}
                      {action.playCount > 0 && (
                        <span className="bg-[linear-gradient(135deg,rgba(120,120,120,0.8)_0%,rgba(80,80,80,0.9)_100%)] text-white/90 text-[8px] font-semibold uppercase tracking-[0.3px] py-0.5 px-1.5 rounded-lg border border-[rgba(120,120,120,0.6)] [text-shadow:none] opacity-100">
                          played
                        </span>
                      )}
                    </div>

                    <div className="relative w-full min-h-[32px] [&>div]:!relative [&>div]:!bottom-auto [&>div]:!left-auto [&>div]:!right-auto [&>div]:w-full [&>div:hover]:!transform-none [&>div:hover]:!shadow-none [&>div:hover]:!filter-none">
                      <BehaviorSection
                        behaviors={[action.behavior]}
                        playerResources={gameState?.currentPlayer?.resources}
                        greyOutAll={action.playCount > 0}
                      />
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
};

export default ActionsPopover;
