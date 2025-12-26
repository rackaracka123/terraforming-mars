import React from "react";
import { PlayerDto, OtherPlayerDto, GamePhase } from "@/types/generated/api-types.ts";
import { globalWebSocketManager } from "@/services/globalWebSocketManager.ts";
import PlayerCard from "../cards/PlayerCard.tsx";

interface PlayerListProps {
  players: (PlayerDto | OtherPlayerDto)[];
  currentPlayer: PlayerDto | null;
  turnPlayerId: string;
  currentPhase?: GamePhase;
}

const PlayerList: React.FC<PlayerListProps> = ({
  players,
  currentPlayer,
  turnPlayerId,
  currentPhase,
}) => {
  const isActionPhase = currentPhase === "action";

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

  const handleSkipAction = async () => {
    try {
      await globalWebSocketManager.skipAction();
    } catch (error) {
      console.error("Failed to skip action:", error);
    }
  };

  return (
    <div className="flex flex-col w-full gap-0 overflow-y-auto overflow-x-visible max-h-[calc(100vh-200px)] [scrollbar-width:none] [-ms-overflow-style:none] [&::-webkit-scrollbar]:hidden">
      {players.map((player, index) => (
        <PlayerCard
          key={player.id}
          player={player}
          playerColor={getPlayerColor(index)}
          isCurrentPlayer={player.id === currentPlayer?.id}
          isCurrentTurn={player.id === turnPlayerId}
          isActionPhase={isActionPhase}
          onSkipAction={handleSkipAction}
          totalPlayers={players.length}
        />
      ))}
    </div>
  );
};

export default PlayerList;
