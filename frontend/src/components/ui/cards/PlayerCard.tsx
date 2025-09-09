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
}) => {
  const isPassed = player.passed;
  const isDisconnected = player.connectionStatus === "disconnected";
  const actionsRemaining = totalActions - actionsUsed;

  // Determine button text based on actions used
  const buttonText = actionsUsed > 0 ? "SKIP" : "PASS";

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
                {actionsRemaining}{" "}
                {actionsRemaining === 1 ? "action" : "actions"} left
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
