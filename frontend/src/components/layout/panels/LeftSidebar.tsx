import React from "react";
import { PlayerDto, GameDto, GamePhase } from "@/types/generated/api-types.ts";
import PlayerList from "@/components/ui/list/PlayerList.tsx";
import styles from "./LeftSidebar.module.css";
import "./LeftSidebar.global.css";

interface LeftSidebarProps {
  players: PlayerDto[];
  currentPlayer: PlayerDto | null;
  currentPlayerId: string;
  currentPhase?: GamePhase;
  gameState?: GameDto;
  onPass?: () => void;
}

const LeftSidebar: React.FC<LeftSidebarProps> = ({
  players,
  currentPlayer,
  currentPlayerId,
  currentPhase,
  gameState: _gameState,
}) => {
  return (
    <div className={styles.leftSidebar}>
      <PlayerList
        players={players}
        currentPlayer={currentPlayer}
        currentPlayerId={currentPlayerId}
        currentPhase={currentPhase}
        remainingActions={currentPlayer?.availableActions}
      />
    </div>
  );
};

export default LeftSidebar;
