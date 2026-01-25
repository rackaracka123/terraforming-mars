import React, { useState } from "react";
import {
  PlayerActionDto,
  GameDto,
  GameStatusActive,
  GamePhaseAction,
} from "@/types/generated/api-types.ts";
import BehaviorSection from "../cards/BehaviorSection";
import { canPerformActions, hasActionsAvailable } from "@/utils/actionUtils.ts";
import GameIcon from "../display/GameIcon.tsx";
import { GameModal, GameModalHeader, GameModalContent, GameModalEmpty } from "../GameModal";

const isActionAvailable = (action: PlayerActionDto): boolean => {
  return action.available;
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

  const isGameActive = gameState?.status === GameStatusActive;
  const isActionPhase = gameState?.currentPhase === GamePhaseAction;
  const isCurrentPlayerTurn = gameState?.currentTurn === gameState?.viewingPlayerId;
  const hasActionsLeft = hasActionsAvailable(gameState?.currentPlayer?.availableActions);
  const hasPendingTileSelection = gameState?.currentPlayer?.pendingTileSelection;

  const showPlayButton = isGameActive && isActionPhase;
  const isPlayButtonEnabled =
    showPlayButton && canPerformActions(gameState) && !hasPendingTileSelection;

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

  const statsContent = (
    <div className="flex flex-col items-center gap-1">
      <span className="text-lg font-bold font-[Courier_New,monospace] text-white">
        {actions.length}
      </span>
      <span className="text-white/70 text-xs uppercase tracking-[0.5px]">Total Actions</span>
    </div>
  );

  const controlsContent = (
    <div className="flex gap-2 items-center text-white text-sm">
      <label>Sort:</label>
      <select
        value={sortType}
        onChange={(e) => setSortType(e.target.value as SortType)}
        className="bg-black/50 border border-[var(--modal-accent)]/40 rounded-md text-white py-1.5 px-3 text-sm"
      >
        <option value="cardName">Card Name</option>
      </select>
      <button
        className="bg-[var(--modal-accent)]/20 border border-[var(--modal-accent)]/40 rounded text-white py-1.5 px-2 cursor-pointer text-base transition-all duration-200 hover:bg-[var(--modal-accent)]/30 hover:scale-110"
        onClick={() => setSortOrder(sortOrder === "asc" ? "desc" : "asc")}
        title={`Sort ${sortOrder === "asc" ? "Descending" : "Ascending"}`}
      >
        {sortOrder === "asc" ? "↑" : "↓"}
      </button>
    </div>
  );

  return (
    <GameModal isVisible={isVisible} onClose={onClose} theme="actions">
      <GameModalHeader
        title="Card Actions"
        stats={statsContent}
        controls={controlsContent}
        onClose={onClose}
      />

      <GameModalContent>
        {sortedActions.length === 0 ? (
          <GameModalEmpty
            icon={<GameIcon iconType="card" size="large" />}
            title="No Card Actions Available"
            description="Play cards with manual triggers to gain actions"
          />
        ) : (
          <div className="grid grid-cols-[repeat(auto-fill,minmax(280px,1fr))] gap-5 justify-items-center max-[1200px]:grid-cols-[repeat(auto-fill,260px)] max-[1200px]:gap-[15px] max-md:grid-cols-[repeat(auto-fill,240px)] max-md:gap-[15px]">
            {sortedActions.map((action, index) => {
              const isAvailable = isActionAvailable(action);
              const isActionPlayable = isPlayButtonEnabled && isAvailable;

              return (
                <div
                  key={`${action.cardId}-${action.behaviorIndex}`}
                  className={`relative border-2 rounded-xl p-[15px] transition-all duration-200 [transition-timing-function:cubic-bezier(0.4,0,0.2,1)] backdrop-blur-[10px] animate-[actionSlideIn_0.6s_ease-out_both] w-full max-w-[320px] min-h-[200px] flex flex-col shadow-[0_4px_15px_rgba(0,0,0,0.3)] max-[1200px]:w-[260px] max-[1200px]:h-[180px] max-md:w-[240px] max-md:h-[160px] max-md:p-3 ${
                    action.available
                      ? "border-[rgba(255,100,100,0.4)] bg-[linear-gradient(135deg,rgba(30,60,90,0.4)_0%,rgba(20,40,70,0.3)_100%)]"
                      : "border-[rgba(255,100,100,0.2)] bg-[linear-gradient(135deg,rgba(30,60,90,0.2)_0%,rgba(20,40,70,0.15)_100%)] opacity-60"
                  }`}
                  style={{ animationDelay: `${index * 0.05}s` }}
                >
                  {!action.available && action.errors && action.errors.length > 0 && (
                    <div className="absolute top-2 right-2 z-[15] bg-[linear-gradient(135deg,#e74c3c,#c0392b)] text-white text-[9px] font-bold px-2 py-1 rounded border border-[rgba(231,76,60,0.8)] shadow-[0_2px_8px_rgba(231,76,60,0.4)] flex items-center gap-1">
                      <span>⚠</span>
                      <span className="max-w-[140px] truncate">
                        {action.errors[0].message}
                        {action.errors.length > 1 && ` (+${action.errors.length - 1})`}
                      </span>
                    </div>
                  )}
                  <div className="flex flex-col gap-2.5 flex-1 overflow-hidden">
                    <div className="text-white/90 text-sm font-semibold [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] leading-[1.3] text-center m-0 flex-shrink-0 flex items-center justify-center gap-2 flex-wrap max-md:text-xs">
                      {action.cardName}
                      {action.timesUsedThisGeneration > 0 && (
                        <span className="bg-[linear-gradient(135deg,rgba(120,120,120,0.8)_0%,rgba(80,80,80,0.9)_100%)] text-white/90 text-[10px] font-semibold uppercase tracking-[0.3px] py-[3px] px-2 rounded-[10px] border border-[rgba(120,120,120,0.6)] [text-shadow:none] opacity-100">
                          played
                        </span>
                      )}
                    </div>

                    <div className="relative w-full min-h-[40px] flex items-center justify-center [&>div]:!relative [&>div]:!bottom-auto [&>div]:!left-auto [&>div]:!right-auto [&>div]:w-full [&>div:hover]:!transform-none [&>div:hover]:!shadow-none [&>div:hover]:!filter-none">
                      <BehaviorSection
                        behaviors={[action.behavior]}
                        playerResources={gameState?.currentPlayer?.resources}
                      />
                    </div>
                  </div>

                  {showPlayButton && (
                    <button
                      className={`absolute bottom-2.5 right-2.5 bg-[linear-gradient(135deg,rgba(100,200,100,0.8)_0%,rgba(80,160,80,0.9)_100%)] border border-[rgba(100,200,100,0.6)] rounded-md text-black text-[11px] font-semibold py-1.5 px-3 cursor-pointer transition-all duration-200 [text-shadow:none] shadow-[0_2px_4px_rgba(0,0,0,0.3)] z-10 hover:bg-[linear-gradient(135deg,rgba(100,200,100,1)_0%,rgba(80,160,80,1)_100%)] hover:border-[rgba(100,200,100,0.8)] hover:-translate-y-px hover:shadow-[0_3px_8px_rgba(100,200,100,0.3)] disabled:!bg-[linear-gradient(135deg,rgba(120,120,120,0.5)_0%,rgba(80,80,80,0.6)_100%)] disabled:!border-[rgba(120,120,120,0.4)] disabled:!text-black/50 disabled:!cursor-not-allowed disabled:!transform-none disabled:!shadow-[0_1px_2px_rgba(0,0,0,0.2)] max-md:text-[10px] max-md:py-1 max-md:px-2 max-md:bottom-2 max-md:right-2 ${!isActionPlayable ? "!bg-[linear-gradient(135deg,rgba(120,120,120,0.5)_0%,rgba(80,80,80,0.6)_100%)] !border-[rgba(120,120,120,0.4)] !text-black/50 !cursor-not-allowed !transform-none !shadow-[0_1px_2px_rgba(0,0,0,0.2)]" : ""}`}
                      onClick={() => isActionPlayable && handleActionClick(action)}
                      disabled={!isActionPlayable}
                      title={
                        hasPendingTileSelection
                          ? "Complete tile placement first"
                          : !isCurrentPlayerTurn
                            ? "Wait for your turn"
                            : !hasActionsLeft
                              ? "No actions remaining"
                              : !isAvailable
                                ? action.errors && action.errors.length > 0
                                  ? action.errors[0].message
                                  : "Action not available"
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
      </GameModalContent>
    </GameModal>
  );
};

export default ActionsModal;
