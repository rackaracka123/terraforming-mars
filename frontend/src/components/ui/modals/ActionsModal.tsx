import React, { useEffect, useState } from "react";
import {
  PlayerActionDto,
  GameDto,
  GameStatusActive,
  GamePhaseAction,
} from "@/types/generated/api-types.ts";
import BehaviorSection from "../cards/BehaviorSection.tsx";
import { canPerformActions, hasActionsAvailable } from "@/utils/actionUtils.ts";

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

interface ActionsModalProps {
  isVisible: boolean;
  onClose: () => void;
  actions: PlayerActionDto[];
  onActionSelect?: (action: PlayerActionDto) => void;
  gameState?: GameDto;
}

type SortType = "cardName";

const ActionsModal: React.FC<ActionsModalProps> = ({
  isVisible,
  onClose,
  actions,
  onActionSelect,
  gameState,
}) => {
  const [sortType, setSortType] = useState<SortType>("cardName");
  const [sortOrder, setSortOrder] = useState<"asc" | "desc">("asc");

  useEffect(() => {
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        onClose();
      }
    };

    if (isVisible) {
      document.addEventListener("keydown", handleEscape);
      document.body.style.overflow = "hidden";
    }

    return () => {
      document.removeEventListener("keydown", handleEscape);
      document.body.style.overflow = "unset";
    };
  }, [isVisible, onClose]);

  if (!isVisible) return null;

  // Determine if actions can be played using utility function
  const isGameActive = gameState?.status === GameStatusActive;
  const isActionPhase = gameState?.currentPhase === GamePhaseAction;
  const isCurrentPlayerTurn =
    gameState?.currentTurn === gameState?.viewingPlayerId;
  const hasActionsLeft = hasActionsAvailable(
    gameState?.currentPlayer?.availableActions,
  );

  // Button should be visible only if game is active and in action phase
  const showPlayButton = isGameActive && isActionPhase;

  // Button should be enabled only if player can perform actions (handles unlimited actions)
  const isPlayButtonEnabled = showPlayButton && canPerformActions(gameState);

  // Sort actions
  const sortedActions = [...actions].sort((a, b) => {
    const aValue = a.cardName.toLowerCase();
    const bValue = b.cardName.toLowerCase();

    if (sortOrder === "asc") {
      return aValue < bValue ? -1 : aValue > bValue ? 1 : 0;
    } else {
      return aValue > bValue ? -1 : aValue < bValue ? 1 : 0;
    }
  });

  const handleActionClick = (action: PlayerActionDto) => {
    if (onActionSelect) {
      onActionSelect(action);
      onClose();
    }
  };

  return (
    <div className="fixed top-0 left-0 right-0 bottom-0 z-[3000] flex items-center justify-center p-5 animate-[modalFadeIn_0.3s_ease-out]">
      <div
        className="absolute top-0 left-0 right-0 bottom-0 bg-black/60 backdrop-blur-sm cursor-pointer"
        onClick={onClose}
      />

      <div className="relative w-full max-w-[1200px] max-h-[90vh] bg-space-black-darker/95 border-2 border-space-blue-400 rounded-[20px] overflow-hidden shadow-[0_20px_60px_rgba(0,0,0,0.6),0_0_40px_rgba(30,60,150,0.3)] backdrop-blur-space animate-[modalSlideIn_0.4s_ease-out] flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between py-[25px] px-[30px] bg-black/40 border-b border-space-blue-600 flex-shrink-0 max-md:p-5 max-md:flex-col max-md:gap-[15px] max-md:items-start">
          <div className="flex flex-col gap-[15px]">
            <h1 className="m-0 font-orbitron text-white text-[28px] font-bold text-shadow-glow tracking-wider">
              Card Actions
            </h1>
            <div className="flex gap-5 items-center">
              <div className="flex flex-col items-center gap-1">
                <span className="text-lg font-bold font-[Courier_New,monospace] text-white">
                  {actions.length}
                </span>
                <span className="text-white/70 text-xs uppercase tracking-[0.5px]">
                  Total Actions
                </span>
              </div>
            </div>
          </div>

          <div className="flex gap-5 items-center max-md:flex-col max-md:gap-2.5 max-md:w-full">
            <div className="flex gap-2 items-center text-white text-sm">
              <label>Sort by:</label>
              <select
                value={sortType}
                onChange={(e) => setSortType(e.target.value as SortType)}
                className="bg-black/50 border border-[rgba(0,255,120,0.4)] rounded-md text-white py-1.5 px-3 text-sm"
              >
                <option value="cardName">Card Name</option>
              </select>
              <button
                className="bg-[rgba(0,255,120,0.2)] border border-[rgba(0,255,120,0.4)] rounded text-white py-1.5 px-2 cursor-pointer text-base transition-all duration-200 hover:bg-[rgba(0,255,120,0.3)] hover:scale-110"
                onClick={() =>
                  setSortOrder(sortOrder === "asc" ? "desc" : "asc")
                }
                title={`Sort ${sortOrder === "asc" ? "Descending" : "Ascending"}`}
              >
                {sortOrder === "asc" ? "↑" : "↓"}
              </button>
            </div>
          </div>

          <button
            className="bg-[linear-gradient(135deg,rgba(255,80,80,0.8)_0%,rgba(200,40,40,0.9)_100%)] border-2 border-[rgba(255,120,120,0.6)] rounded-full w-[45px] h-[45px] text-white text-2xl font-bold cursor-pointer flex items-center justify-center transition-all duration-300 shadow-[0_4px_15px_rgba(0,0,0,0.4)] hover:scale-110 hover:shadow-[0_6px_25px_rgba(255,80,80,0.5)]"
            onClick={onClose}
          >
            ×
          </button>
        </div>

        {/* Actions Content */}
        <div className="flex-1 py-[25px] px-[30px] overflow-y-auto [scrollbar-width:thin] [scrollbar-color:rgba(0,255,120,0.5)_rgba(50,75,125,0.3)] [&::-webkit-scrollbar]:w-2 [&::-webkit-scrollbar-track]:bg-[rgba(50,75,125,0.3)] [&::-webkit-scrollbar-track]:rounded [&::-webkit-scrollbar-thumb]:bg-[rgba(0,255,120,0.5)] [&::-webkit-scrollbar-thumb]:rounded [&::-webkit-scrollbar-thumb:hover]:bg-[rgba(0,255,120,0.7)] max-md:p-5">
          {sortedActions.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-[60px] px-5 text-center min-h-[300px]">
              <img
                src="/assets/misc/corpCard.png"
                alt="No actions"
                className="w-16 h-16 mb-5 opacity-60"
              />
              <h3 className="text-white text-2xl m-0 mb-2.5">
                No Card Actions Available
              </h3>
              <p className="text-white/70 text-base m-0">
                Play cards with manual triggers to gain actions
              </p>
            </div>
          ) : (
            <div className="grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-5 justify-items-center max-[1200px]:grid-cols-[repeat(auto-fill,260px)] max-[1200px]:gap-[15px] max-md:grid-cols-[repeat(auto-fill,240px)] max-md:gap-[15px]">
              {sortedActions.map((action, index) => {
                const isAvailable = isActionAvailable(action, gameState);
                const isActionPlayable = isPlayButtonEnabled && isAvailable;

                return (
                  <div
                    key={`${action.cardId}-${action.behaviorIndex}`}
                    className={`border-2 border-[rgba(255,100,100,0.4)] rounded-xl p-[15px] transition-all duration-300 [transition-timing-function:cubic-bezier(0.4,0,0.2,1)] backdrop-blur-[10px] animate-[actionSlideIn_0.6s_ease-out_both] w-full max-w-[320px] min-h-[200px] flex flex-col bg-[linear-gradient(135deg,rgba(30,60,90,0.4)_0%,rgba(20,40,70,0.3)_100%)] shadow-[0_4px_15px_rgba(0,0,0,0.3)] relative ${!isAvailable ? "opacity-60 !border-[rgba(255,100,100,0.2)] !bg-[linear-gradient(135deg,rgba(30,60,90,0.2)_0%,rgba(20,40,70,0.15)_100%)]" : ""} max-[1200px]:w-[260px] max-[1200px]:h-[180px] max-md:w-[240px] max-md:h-[160px] max-md:p-3`}
                    style={{ animationDelay: `${index * 0.05}s` }}
                  >
                    <div className="flex flex-col gap-2.5 flex-1 overflow-hidden">
                      <div className="text-white/90 text-sm font-semibold [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] leading-[1.3] text-center m-0 flex-shrink-0 flex items-center justify-center gap-2 flex-wrap max-md:text-xs">
                        {action.cardName}
                        {action.playCount > 0 && (
                          <span className="bg-[linear-gradient(135deg,rgba(120,120,120,0.8)_0%,rgba(80,80,80,0.9)_100%)] text-white/90 text-[10px] font-semibold uppercase tracking-[0.3px] py-[3px] px-2 rounded-[10px] border border-[rgba(120,120,120,0.6)] [text-shadow:none] opacity-100">
                            played
                          </span>
                        )}
                      </div>

                      <div className="relative w-full min-h-[40px] flex items-center justify-center [&>div]:!relative [&>div]:!bottom-auto [&>div]:!left-auto [&>div]:!right-auto [&>div]:w-full [&>div:hover]:!transform-none [&>div:hover]:!shadow-none [&>div:hover]:!filter-none">
                        <BehaviorSection
                          behaviors={[action.behavior]}
                          playerResources={gameState?.currentPlayer?.resources}
                          greyOutAll={action.playCount > 0}
                        />
                      </div>
                    </div>

                    {showPlayButton && (
                      <button
                        className={`absolute bottom-2.5 right-2.5 bg-[linear-gradient(135deg,rgba(100,200,100,0.8)_0%,rgba(80,160,80,0.9)_100%)] border border-[rgba(100,200,100,0.6)] rounded-md text-black text-[11px] font-semibold py-1.5 px-3 cursor-pointer transition-all duration-200 [text-shadow:none] shadow-[0_2px_4px_rgba(0,0,0,0.3)] z-10 hover:bg-[linear-gradient(135deg,rgba(100,200,100,1)_0%,rgba(80,160,80,1)_100%)] hover:border-[rgba(100,200,100,0.8)] hover:-translate-y-px hover:shadow-[0_3px_8px_rgba(100,200,100,0.3)] disabled:!bg-[linear-gradient(135deg,rgba(120,120,120,0.5)_0%,rgba(80,80,80,0.6)_100%)] disabled:!border-[rgba(120,120,120,0.4)] disabled:!text-black/50 disabled:!cursor-not-allowed disabled:!transform-none disabled:!shadow-[0_1px_2px_rgba(0,0,0,0.2)] max-md:text-[10px] max-md:py-1 max-md:px-2 max-md:bottom-2 max-md:right-2 ${!isActionPlayable ? "!bg-[linear-gradient(135deg,rgba(120,120,120,0.5)_0%,rgba(80,80,80,0.6)_100%)] !border-[rgba(120,120,120,0.4)] !text-black/50 !cursor-not-allowed !transform-none !shadow-[0_1px_2px_rgba(0,0,0,0.2)]" : ""}`}
                        onClick={() =>
                          isActionPlayable && handleActionClick(action)
                        }
                        disabled={!isActionPlayable}
                        title={
                          !isCurrentPlayerTurn
                            ? "Wait for your turn"
                            : !hasActionsLeft
                              ? "No actions remaining"
                              : !isAvailable
                                ? action.playCount > 0
                                  ? "Already played this generation"
                                  : "Cannot afford this action"
                                : "Play this action"
                        }
                      >
                        Play
                      </button>
                    )}
                  </div>
                );
              })}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default ActionsModal;
