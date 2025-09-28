import React from "react";
import { PlayerDto } from "@/types/generated/api-types.ts";
import styles from "./PlayerCard.module.css";

interface PlayerCardProps {
  player: PlayerDto;
  playerColor: string;
  isCurrentPlayer: boolean;
  isActivePlayer: boolean;
  isCurrentTurn: boolean;
  isActionPhase: boolean;
  onSkipAction?: () => void;
  actionsUsed?: number;
  totalActions?: number;
  totalPlayers?: number; // Added to determine solo vs last player
}

const PlayerCard: React.FC<PlayerCardProps> = ({
  player,
  playerColor,
  isCurrentPlayer,
  isActivePlayer,
  isCurrentTurn,
  isActionPhase,
  onSkipAction,
  actionsUsed = 0,
  totalActions = 2,
  totalPlayers = 1,
}) => {
  const isPassed = player.passed;
  const isDisconnected = !player.isConnected;
  const hasUnlimitedActions = player.availableActions === -1;

  // For unlimited actions, calculate actionsRemaining and button text differently
  const actionsRemaining = hasUnlimitedActions
    ? -1
    : totalActions - actionsUsed;

  // Determine button text - always PASS for unlimited actions, otherwise SKIP if actions used
  const buttonText = hasUnlimitedActions
    ? "PASS"
    : actionsUsed > 0
      ? "SKIP"
      : "PASS";

  return (
    <div
      className={`${styles.playerCardContainer} ${isCurrentTurn ? styles.active : ""}`}
    >
      {/* Main player card with angled edge */}
      <div
        className={`${styles.playerCard} ${isDisconnected ? styles.disconnected : ""} ${!isCurrentTurn ? styles.notInTurn : ""} ${isCurrentTurn ? styles.active : ""}`}
        style={{ "--player-color": playerColor } as React.CSSProperties}
      >
        <div className={styles.playerCardContent}>
          <div className={styles.playerChips}>
            {isCurrentPlayer && (
              <span className={`${styles.chip} ${styles.chipYou}`}>YOU</span>
            )}
            {isPassed && (
              <span className={`${styles.chip} ${styles.chipPassed}`}>
                PASSED
              </span>
            )}
            {isDisconnected && (
              <span className={`${styles.chip} ${styles.chipDisconnected}`}>
                DISCONNECTED
              </span>
            )}
            {isCurrentTurn && isActionPhase && (
              <span className={`${styles.chip} ${styles.chipActions}`}>
                {hasUnlimitedActions
                  ? totalPlayers === 1
                    ? "Solo"
                    : "Last player"
                  : `${actionsRemaining} ${actionsRemaining === 1 ? "action" : "actions"} left`}
              </span>
            )}
          </div>
          <span className={styles.playerName}>{player.name}</span>
        </div>
        {isActivePlayer && isCurrentTurn && isActionPhase && (
          <button className={styles.actionButton} onClick={onSkipAction}>
            {buttonText}
          </button>
        )}
      </div>
    </div>
  );
};

export default PlayerCard;
