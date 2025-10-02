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
import styles from "./ActionsPopover.module.css";

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
    <div className={styles.actionsPopover} ref={popoverRef}>
      <div className={styles.popoverArrow} />

      <div className={styles.popoverHeader}>
        <div className={styles.headerTitle}>
          <h3>Card Actions</h3>
        </div>
        <div className={styles.headerControls}>
          <div className={styles.actionsCount}>{actions.length} available</div>
          {onOpenDetails && (
            <button
              className={styles.detailsButton}
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

      <div className={styles.popoverContent}>
        {actions.length === 0 ? (
          <div className={styles.emptyState}>
            <img
              src="/assets/misc/corpCard.png"
              alt="No actions"
              className={styles.emptyIcon}
            />
            <div className={styles.emptyText}>No card actions available</div>
            <div className={styles.emptySubtitle}>
              Play cards with manual triggers to gain actions
            </div>
          </div>
        ) : (
          <div className={styles.actionsList}>
            {actions.map((action, index) => {
              const isAvailable = isActionAvailable(action, gameState);
              const isActionPlayable = canPlayActions && isAvailable;

              return (
                <div
                  key={`${action.cardId}-${action.behaviorIndex}`}
                  className={`${styles.actionItem} ${!isActionPlayable ? styles.actionItemDisabled : ""}`}
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
                  <div className={styles.actionContent}>
                    <div className={styles.actionTitle}>
                      {action.cardName}
                      {action.playCount > 0 && (
                        <span className={styles.playedChip}>played</span>
                      )}
                    </div>

                    <div className={styles.behaviorContainer}>
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
