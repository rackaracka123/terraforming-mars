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
} from "../../../types/generated/api-types.ts";
import styles from "./GameLayout.module.css";

interface GameLayoutProps {
  gameState: GameDto;
  currentPlayer: PlayerDto | null;
  isAnyModalOpen?: boolean;
  isLobbyPhase?: boolean;
  onOpenCardEffectsModal?: () => void;
  onOpenCardsPlayedModal?: () => void;
  onOpenTagsModal?: () => void;
  onOpenVictoryPointsModal?: () => void;
  onOpenActionsModal?: () => void;
  onActionSelect?: (action: PlayerActionDto) => void;
}

const GameLayout: React.FC<GameLayoutProps> = ({
  gameState,
  currentPlayer,
  isAnyModalOpen: _isAnyModalOpen = false,
  isLobbyPhase = false,
  onOpenCardEffectsModal,
  onOpenCardsPlayedModal,
  onOpenTagsModal,
  onOpenVictoryPointsModal,
  onOpenActionsModal,
  onActionSelect,
}) => {
  // Convert OtherPlayerDto to PlayerDto for LeftSidebar compatibility
  const convertOtherPlayerToPlayerDto = (
    otherPlayer: OtherPlayerDto,
  ): PlayerDto => ({
    ...otherPlayer,
    cards: [], // OtherPlayerDto doesn't expose actual cards, only handCardCount
    startingSelection: [], // Add required startingSelection property
    hasSelectedStartingCards: true, // Default to true for other players
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
      <div className={styles.gameLayout}>
        <TopMenuBar gameState={gameState} />

        <div className={styles.gameContent}>
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
              onOpenCardEffectsModal={onOpenCardEffectsModal}
              onOpenCardsPlayedModal={onOpenCardsPlayedModal}
              onOpenTagsModal={onOpenTagsModal}
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
