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
  gameState,
}) => {
  return (
    <div className={styles.leftSidebar}>
      <div className={styles.sidebarTitle}>PLAYERS</div>
      <PlayerList
        players={players}
        currentPlayer={currentPlayer}
        currentPlayerId={currentPlayerId}
        currentPhase={currentPhase}
        remainingActions={gameState?.remainingActions}
      />
    </div>
  );
};

export default LeftSidebar;
