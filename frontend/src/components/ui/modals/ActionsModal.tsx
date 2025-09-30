import React, { useEffect, useState } from "react";
import {
  PlayerActionDto,
  GameDto,
  GameStatusActive,
  GamePhaseAction,
} from "@/types/generated/api-types.ts";
import BehaviorSection from "../cards/BehaviorSection.tsx";
import { canPerformActions, hasActionsAvailable } from "@/utils/actionUtils.ts";
import styles from "./ActionsModal.module.css";

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
    <div className={styles.actionsModal}>
      <div className={styles.backdrop} onClick={onClose} />

      <div className={styles.modalContainer}>
        {/* Header */}
        <div className={styles.modalHeader}>
          <div className={styles.headerLeft}>
            <h1 className={styles.modalTitle}>Card Actions</h1>
            <div className={styles.actionSummary}>
              <div className={styles.summaryItem}>
                <span className={styles.summaryValue}>{actions.length}</span>
                <span className={styles.summaryLabel}>Total Actions</span>
              </div>
            </div>
          </div>

          <div className={styles.headerControls}>
            <div className={styles.sortControls}>
              <label>Sort by:</label>
              <select
                value={sortType}
                onChange={(e) => setSortType(e.target.value as SortType)}
              >
                <option value="cardName">Card Name</option>
              </select>
              <button
                className={styles.sortOrderBtn}
                onClick={() =>
                  setSortOrder(sortOrder === "asc" ? "desc" : "asc")
                }
                title={`Sort ${sortOrder === "asc" ? "Descending" : "Ascending"}`}
              >
                {sortOrder === "asc" ? "↑" : "↓"}
              </button>
            </div>
          </div>

          <button className={styles.closeButton} onClick={onClose}>
            ×
          </button>
        </div>

        {/* Actions Content */}
        <div className={styles.actionsContent}>
          {sortedActions.length === 0 ? (
            <div className={styles.emptyState}>
              <img
                src="/assets/misc/corpCard.png"
                alt="No actions"
                className={styles.emptyIcon}
              />
              <h3>No Card Actions Available</h3>
              <p>Play cards with manual triggers to gain actions</p>
            </div>
          ) : (
            <div className={styles.actionsGrid}>
              {sortedActions.map((action, index) => {
                const isAvailable = isActionAvailable(action, gameState);
                const isActionPlayable = isPlayButtonEnabled && isAvailable;

                return (
                  <div
                    key={`${action.cardId}-${action.behaviorIndex}`}
                    className={`${styles.actionBox} ${!isAvailable ? styles.actionBoxUnavailable : ""}`}
                    style={{ animationDelay: `${index * 0.05}s` }}
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

                    {showPlayButton && (
                      <button
                        className={`${styles.actionButton} ${!isActionPlayable ? styles.actionButtonDisabled : ""}`}
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
