import React from "react";
import { PlayerDto, OtherPlayerDto, GameDto, GamePhase } from "@/types/generated/api-types.ts";
import PlayerList from "@/components/ui/list/PlayerList.tsx";

interface LeftSidebarProps {
  players: (PlayerDto | OtherPlayerDto)[];
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
    <div className="absolute top-[15%] left-0 z-10 w-[240px] h-[calc(85vh-120px)] bg-transparent py-[15px] flex flex-col overflow-visible pointer-events-none">
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
