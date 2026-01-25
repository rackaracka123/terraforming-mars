import React from "react";
import { PlayerActionDto, GameDto } from "../../../types/generated/api-types.ts";
import BehaviorSection from "../cards/BehaviorSection";
import { canPerformActions } from "../../../utils/actionUtils.ts";
import GameIcon from "../display/GameIcon.tsx";
import { GamePopover, GamePopoverEmpty, GamePopoverItem } from "../GamePopover";

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
  onActionSelect,
  onOpenDetails,
  anchorRef,
  gameState,
}) => {
  const hasPendingTileSelection = gameState?.currentPlayer?.pendingTileSelection;
  const canPlayActions = canPerformActions(gameState) && !hasPendingTileSelection;

  const handleActionClick = (action: PlayerActionDto) => {
    if (onActionSelect) {
      onActionSelect(action);
      onClose();
    }
  };

  return (
    <GamePopover
      isVisible={isVisible}
      onClose={onClose}
      position={{ type: "anchor", anchorRef, placement: "above" }}
      theme="actions"
      header={{
        title: "Card Actions",
        badge: `${actions.length} available`,
        rightContent: onOpenDetails ? (
          <button
            className="bg-space-black-darker/90 border-2 border-[var(--popover-accent)] rounded-lg text-white text-[11px] font-semibold py-1 px-2.5 cursor-pointer transition-all duration-200 text-shadow-dark pointer-events-auto relative z-[1] hover:bg-space-black-darker/95 hover:-translate-y-px hover:shadow-[0_2px_8px_rgba(var(--popover-accent-rgb),0.4)]"
            onClick={() => {
              onOpenDetails();
              onClose();
            }}
            title="Open detailed actions view"
          >
            Details
          </button>
        ) : undefined,
      }}
      arrow={{ enabled: true, position: "right", offset: 30 }}
      width={320}
      maxHeight={400}
    >
      {actions.length === 0 ? (
        <GamePopoverEmpty
          icon={<GameIcon iconType="card" size="medium" />}
          title="No card actions available"
          description="Play cards with manual triggers to gain actions"
        />
      ) : (
        <div className="p-2 flex flex-col gap-2">
          {actions.map((action, index) => {
            const isAvailable = action.available;
            const isActionPlayable = canPlayActions && isAvailable;

            return (
              <GamePopoverItem
                key={`${action.cardId}-${action.behaviorIndex}`}
                state={isAvailable ? "available" : "disabled"}
                onClick={isActionPlayable ? () => handleActionClick(action) : undefined}
                error={
                  !isAvailable && action.errors && action.errors.length > 0
                    ? { message: action.errors[0].message, count: action.errors.length }
                    : undefined
                }
                hoverEffect="translate-x"
                animationDelay={index * 0.05}
                className={!isActionPlayable && isAvailable ? "cursor-default" : ""}
              >
                <div className="flex flex-col gap-2 flex-1">
                  <div className="text-white/70 text-[11px] font-medium uppercase tracking-[0.5px] [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] leading-[1.2] opacity-80 flex items-center gap-2 max-[768px]:text-[10px]">
                    {action.cardName}
                    {action.timesUsedThisGeneration > 0 && (
                      <span className="bg-[linear-gradient(135deg,rgba(120,120,120,0.8)_0%,rgba(80,80,80,0.9)_100%)] text-white/90 text-[8px] font-semibold uppercase tracking-[0.3px] py-0.5 px-1.5 rounded-lg border border-[rgba(120,120,120,0.6)] [text-shadow:none] opacity-100">
                        played
                      </span>
                    )}
                  </div>

                  <div className="relative w-full min-h-[32px] [&>div]:!relative [&>div]:!bottom-auto [&>div]:!left-auto [&>div]:!right-auto [&>div]:w-full [&>div:hover]:!transform-none [&>div:hover]:!shadow-none [&>div:hover]:!filter-none">
                    <BehaviorSection
                      behaviors={[action.behavior]}
                      playerResources={gameState?.currentPlayer?.resources}
                      resourceStorage={gameState?.currentPlayer?.resourceStorage}
                      cardId={action.cardId}
                    />
                  </div>
                </div>
              </GamePopoverItem>
            );
          })}
        </div>
      )}
    </GamePopover>
  );
};

export default ActionsPopover;
