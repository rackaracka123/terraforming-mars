import React from "react";
import { PlayerDto, OtherPlayerDto, GamePhase } from "@/types/generated/api-types.ts";
import PlayerList from "@/components/ui/list/PlayerList.tsx";

interface LeftSidebarProps {
  players: (PlayerDto | OtherPlayerDto)[];
  currentPlayer: PlayerDto | null;
  turnPlayerId: string;
  currentPhase?: GamePhase;
  hasPendingTilePlacement?: boolean;
}

const LeftSidebar: React.FC<LeftSidebarProps> = ({
  players,
  currentPlayer,
  turnPlayerId,
  currentPhase,
  hasPendingTilePlacement = false,
}) => {
  return (
    <div className="absolute top-[15%] left-0 z-10 w-[240px] h-[calc(85vh-120px)] bg-transparent py-[15px] flex flex-col overflow-visible pointer-events-none">
      <PlayerList
        players={players}
        currentPlayer={currentPlayer}
        turnPlayerId={turnPlayerId}
        currentPhase={currentPhase}
        hasPendingTilePlacement={hasPendingTilePlacement}
      />
    </div>
  );
};

export default LeftSidebar;
