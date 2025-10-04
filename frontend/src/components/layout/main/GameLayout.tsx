import React from "react";
import LeftSidebar from "../panels/LeftSidebar.tsx";
import TopMenuBar from "../panels/TopMenuBar.tsx";
import RightSidebar from "../panels/RightSidebar.tsx";
import MainContentDisplay from "../../ui/display/MainContentDisplay.tsx";
import BottomResourceBar from "../../ui/overlay/BottomResourceBar.tsx";
import PlayerOverlay from "../../ui/overlay/PlayerOverlay.tsx";
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
  isAnyModalOpen?: boolean;
  isLobbyPhase?: boolean;
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
  isAnyModalOpen: _isAnyModalOpen = false,
  isLobbyPhase = false,
  onOpenCardEffectsModal,
  onOpenCardsPlayedModal,
  onOpenVictoryPointsModal,
  onOpenActionsModal,
  onActionSelect,
  showStandardProjectsPopover = false,
  onToggleStandardProjectsPopover,
  standardProjectsButtonRef,
}) => {
  const convertOtherPlayerToPlayerDto = (
    otherPlayer: OtherPlayerDto,
  ): PlayerDto => ({
    ...otherPlayer,
    cards: [],
    selectStartingCardsPhase: undefined,
    productionPhase: undefined,
    startingCards: [],
  });

  // Create a map of all players (current + others) for easy lookup
  const playerMap = new Map<string, PlayerDto>();
  if (gameState?.currentPlayer) {
    playerMap.set(gameState.currentPlayer.id, gameState.currentPlayer);
  }
  gameState?.otherPlayers?.forEach((otherPlayer) => {
    playerMap.set(otherPlayer.id, convertOtherPlayerToPlayerDto(otherPlayer));
  });

  // Construct allPlayers using the turn order from the backend
  const allPlayers: PlayerDto[] =
    (gameState?.turnOrder
      ?.map((playerId) => playerMap.get(playerId))
      .filter((player) => player !== undefined) as PlayerDto[]) || [];

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
              onOpenCardEffectsModal={onOpenCardEffectsModal}
              onOpenCardsPlayedModal={onOpenCardsPlayedModal}
              onOpenVictoryPointsModal={onOpenVictoryPointsModal}
              onOpenActionsModal={onOpenActionsModal}
              onActionSelect={onActionSelect}
            />
          </>
        )}
      </div>
    </MainContentProvider>
  );
};

export default GameLayout;
