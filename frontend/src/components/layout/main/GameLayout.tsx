import React from "react";
import LeftSidebar from "../panels/LeftSidebar.tsx";
import TopMenuBar from "../panels/TopMenuBar.tsx";
import RightSidebar from "../panels/RightSidebar.tsx";
import MainContentDisplay from "../../ui/display/MainContentDisplay.tsx";
import BottomResourceBar from "../../ui/overlay/BottomResourceBar.tsx";
import CardsHandOverlay from "../../ui/overlay/CardsHandOverlay.tsx";
import PlayerOverlay from "../../ui/overlay/PlayerOverlay.tsx";
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

  const allPlayers: PlayerDto[] = [
    ...(gameState?.currentPlayer ? [gameState.currentPlayer] : []),
    ...(gameState?.otherPlayers?.map(convertOtherPlayerToPlayerDto) || []),
  ];
  return (
    <MainContentProvider>
      <div className={styles.gameLayout}>
        <TopMenuBar />

        <div className={styles.gameContent}>
          <LeftSidebar
            players={allPlayers}
            currentPlayer={currentPlayer}
            currentPlayerId={gameState?.currentPlayer.id || ""}
            currentPhase={gameState?.currentPhase}
            gameState={gameState}
          />

          <MainContentDisplay gameState={gameState} />

          <RightSidebar
            globalParameters={gameState?.globalParameters}
            generation={gameState?.generation}
            currentPlayer={currentPlayer}
          />
        </div>

        {/* Overlay Components */}
        <PlayerOverlay
          players={[
            ...(gameState?.currentPlayer ? [gameState.currentPlayer] : []),
            ...(gameState?.otherPlayers || []),
          ]}
          currentPlayer={currentPlayer}
        />

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
          </>
        )}
      </div>
    </MainContentProvider>
  );
};

export default GameLayout;
