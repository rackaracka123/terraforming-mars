import React from "react";
import LeftSidebar from "../panels/LeftSidebar.tsx";
import TopMenuBar from "../panels/TopMenuBar.tsx";
import RightSidebar from "../panels/RightSidebar.tsx";
import MainContentDisplay from "../../ui/display/MainContentDisplay.tsx";
import BottomResourceBar from "../../ui/overlay/BottomResourceBar.tsx";
import PlayerOverlay from "../../ui/overlay/PlayerOverlay.tsx";
import CorporationDisplay from "../../ui/display/CorporationDisplay.tsx";
import { MainContentProvider } from "../../../contexts/MainContentContext.tsx";
import {
  GameDto,
  PlayerDto,
  OtherPlayerDto,
  PlayerActionDto,
  CardDto,
} from "../../../types/generated/api-types.ts";

interface GameLayoutProps {
  gameState: GameDto;
  currentPlayer: PlayerDto | null;
  playedCards?: CardDto[];
  corporationCard?: CardDto | null;
  isAnyModalOpen?: boolean;
  isLobbyPhase?: boolean;
  changedPaths?: Set<string>;
  onOpenCardEffectsModal?: () => void;
  onOpenCardsPlayedModal?: () => void;
  onOpenVictoryPointsModal?: () => void;
  onOpenActionsModal?: () => void;
  onActionSelect?: (action: PlayerActionDto) => void;
  showStandardProjectsPopover?: boolean;
  onToggleStandardProjectsPopover?: () => void;
  standardProjectsButtonRef?: React.RefObject<HTMLButtonElement | null>;
}

const GameLayout: React.FC<GameLayoutProps> = ({
  gameState,
  currentPlayer,
  playedCards = [],
  corporationCard = null,
  isAnyModalOpen: _isAnyModalOpen = false,
  isLobbyPhase = false,
  changedPaths = new Set(),
  onOpenCardEffectsModal,
  onOpenCardsPlayedModal,
  onOpenVictoryPointsModal,
  onOpenActionsModal,
  onActionSelect,
  showStandardProjectsPopover = false,
  onToggleStandardProjectsPopover,
  standardProjectsButtonRef,
}) => {
  // Create a map of all players (current + others) for easy lookup
  const playerMap = new Map<string, PlayerDto | OtherPlayerDto>();
  if (gameState?.currentPlayer) {
    playerMap.set(gameState.currentPlayer.id, gameState.currentPlayer);
  }
  gameState?.otherPlayers?.forEach((otherPlayer) => {
    playerMap.set(otherPlayer.id, otherPlayer);
  });

  // Construct allPlayers using the turn order from the backend
  const allPlayers: (PlayerDto | OtherPlayerDto)[] =
    (gameState?.turnOrder
      ?.map((playerId) => playerMap.get(playerId))
      .filter((player) => player !== undefined) as (
      | PlayerDto
      | OtherPlayerDto
    )[]) || [];

  // Find the current turn player for the right sidebar
  const currentTurnPlayer =
    allPlayers.find((player) => player.id === gameState?.currentTurn) || null;

  return (
    <MainContentProvider>
      <div className="grid grid-rows-[auto_1fr] w-screen h-screen bg-[#000011] bg-[url('/assets/background-noise.png')] [background-attachment:fixed] bg-repeat text-white overflow-hidden">
        <TopMenuBar
          gameState={gameState}
          showStandardProjectsPopover={showStandardProjectsPopover}
          onToggleStandardProjectsPopover={onToggleStandardProjectsPopover}
          standardProjectsButtonRef={standardProjectsButtonRef}
        />

        <div className="grid grid-cols-1 min-h-0 gap-0 relative">
          <MainContentDisplay gameState={gameState} />
        </div>

        {/* Overlay Components */}
        <LeftSidebar
          players={allPlayers}
          currentPlayer={currentPlayer}
          currentPlayerId={gameState?.currentTurn || ""}
          currentPhase={gameState?.currentPhase}
          gameState={gameState}
        />

        <RightSidebar
          globalParameters={gameState?.globalParameters}
          generation={gameState?.generation}
          currentPlayer={currentTurnPlayer}
        />

        <PlayerOverlay players={allPlayers} currentPlayer={currentPlayer} />

        {!isLobbyPhase && (
          <>
            <BottomResourceBar
              currentPlayer={currentPlayer}
              gameState={gameState}
              playedCards={playedCards}
              changedPaths={changedPaths}
              onOpenCardEffectsModal={onOpenCardEffectsModal}
              onOpenCardsPlayedModal={onOpenCardsPlayedModal}
              onOpenVictoryPointsModal={onOpenVictoryPointsModal}
              onOpenActionsModal={onOpenActionsModal}
              onActionSelect={onActionSelect}
            />

            {corporationCard && (
              <CorporationDisplay corporation={corporationCard} />
            )}
          </>
        )}
      </div>
    </MainContentProvider>
  );
};

export default GameLayout;
