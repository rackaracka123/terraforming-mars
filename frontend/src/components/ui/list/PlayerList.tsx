import React, { useEffect, useState } from "react";
import { PlayerDto, GamePhase } from "@/types/generated/api-types.ts";
import { globalWebSocketManager } from "@/services/globalWebSocketManager.ts";
import PlayerCard from "../cards/PlayerCard.tsx";
import styles from "./PlayerList.module.css";

interface PlayerListProps {
  players: PlayerDto[];
  currentPlayer: PlayerDto | null;
  currentPlayerId: string;
  currentPhase?: GamePhase;
  remainingActions?: number;
}

const PlayerList: React.FC<PlayerListProps> = ({
  players,
  currentPlayer,
  currentPlayerId,
  currentPhase,
  remainingActions,
}) => {
  const [previousTurnPlayer, setPreviousTurnPlayer] = useState<string | null>(
    null,
  );
  const [actionsUsed, setActionsUsed] = useState(0);

  // Calculate if we're in the action phase where action UI should be visible
  const isActionPhase = currentPhase === "action";

  // Player color system - 6 distinct colors for up to 6 players
  const playerColors = [
    "#b91c2b", // Red
    "#232dc7", // Blue
    "#3abe3a", // Green
    "#ffa502", // Orange
    "#a55eea", // Purple
    "#26d0ce", // Cyan
  ];

  const getPlayerColor = (index: number) => {
    return playerColors[index % playerColors.length];
  };

  // Track turn changes for animation
  useEffect(() => {
    if (previousTurnPlayer !== currentPlayerId && previousTurnPlayer !== null) {
      // Turn has changed, reset actions used
      setActionsUsed(0);
    }
    setPreviousTurnPlayer(currentPlayerId);
  }, [currentPlayerId, previousTurnPlayer]);

  // Calculate actions used based on remaining actions
  useEffect(() => {
    const totalActions = 2; // Default total actions per turn
    const used = totalActions - (remainingActions || 0);
    setActionsUsed(Math.max(0, used));
  }, [remainingActions]);

  const handleSkipAction = async () => {
    try {
      await globalWebSocketManager.skipAction();
    } catch (error) {
      console.error("Failed to skip action:", error);
    }
  };

  return (
    <div className={styles.playerList}>
      {players.map((player, index) => {
        const isCurrentPlayer = player.id === currentPlayer?.id;
        const isCurrentTurn = player.id === currentPlayerId;
        const playerColor = getPlayerColor(index);

        return (
          <PlayerCard
            key={player.id}
            player={player}
            playerColor={playerColor}
            isCurrentPlayer={isCurrentPlayer}
            isActivePlayer={isCurrentPlayer}
            isCurrentTurn={isCurrentTurn}
            isActionPhase={isActionPhase}
            onSkipAction={handleSkipAction}
            actionsUsed={actionsUsed}
            totalActions={2}
          />
        );
      })}
    </div>
  );
};

export default PlayerList;
