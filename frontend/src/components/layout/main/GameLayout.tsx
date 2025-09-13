import React from "react";
import LeftSidebar from "../panels/LeftSidebar.tsx";
import TopMenuBar from "../panels/TopMenuBar.tsx";
import RightSidebar from "../panels/RightSidebar.tsx";
import MainContentDisplay from "../../ui/display/MainContentDisplay.tsx";
import BottomResourceBar from "../../ui/overlay/BottomResourceBar.tsx";
import CardsHandOverlay from "../../ui/overlay/CardsHandOverlay.tsx";
import PlayerOverlay from "../../ui/overlay/PlayerOverlay.tsx";
import VictoryPointIcon from "../../ui/display/VictoryPointIcon.tsx";
import { MainContentProvider } from "../../../contexts/MainContentContext.tsx";
import {
  GameDto,
  PlayerDto,
  OtherPlayerDto,
} from "../../../types/generated/api-types.ts";
import styles from "./GameLayout.module.css";

interface GameLayoutProps {
  gameState: GameDto;
  currentPlayer: PlayerDto | null;
  isAnyModalOpen?: boolean;
  isLobbyPhase?: boolean;
  onOpenCardEffectsModal?: () => void;
  onOpenActionsModal?: () => void;
  onOpenCardsPlayedModal?: () => void;
  onOpenTagsModal?: () => void;
  onOpenVictoryPointsModal?: () => void;
}

const GameLayout: React.FC<GameLayoutProps> = ({
  gameState,
  currentPlayer,
  isAnyModalOpen = false,
  isLobbyPhase = false,
  onOpenCardEffectsModal,
  onOpenActionsModal,
  onOpenCardsPlayedModal,
  onOpenTagsModal,
  onOpenVictoryPointsModal,
}) => {
  // Convert OtherPlayerDto to PlayerDto for LeftSidebar compatibility
  const convertOtherPlayerToPlayerDto = (
    otherPlayer: OtherPlayerDto,
  ): PlayerDto => ({
    ...otherPlayer,
    cards: [], // OtherPlayerDto doesn't expose actual cards, only handCardCount
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
        <TopMenuBar />

        <div className={styles.gameContent}>
          <LeftSidebar
            players={allPlayers}
            currentPlayer={currentPlayer}
            currentPlayerId={gameState?.currentTurn || ""}
            currentPhase={gameState?.currentPhase}
            gameState={gameState}
          />

          <MainContentDisplay gameState={gameState} />

          <RightSidebar
            globalParameters={gameState?.globalParameters}
            generation={gameState?.generation}
            currentPlayer={currentTurnPlayer}
          />
        </div>

        {/* Overlay Components */}
        <PlayerOverlay players={allPlayers} currentPlayer={currentPlayer} />

        {!isLobbyPhase && (
          <>
            <BottomResourceBar
              currentPlayer={currentPlayer}
              onOpenCardEffectsModal={onOpenCardEffectsModal}
              onOpenActionsModal={onOpenActionsModal}
              onOpenCardsPlayedModal={onOpenCardsPlayedModal}
              onOpenTagsModal={onOpenTagsModal}
              onOpenVictoryPointsModal={onOpenVictoryPointsModal}
            />

            <CardsHandOverlay hideWhenModalOpen={isAnyModalOpen} />

            {/* Victory Points display in bottom right */}
            <div className={styles.victoryPointsOverlay}>
              <VictoryPointIcon
                value={currentPlayer?.victoryPoints || 0}
                size="large"
              />
            </div>
          </>
        )}
      </div>
    </MainContentProvider>
  );
};

export default GameLayout;
