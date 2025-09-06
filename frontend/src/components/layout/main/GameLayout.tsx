import React from "react";
import LeftSidebar from "../panels/LeftSidebar.tsx";
import TopMenuBar from "../panels/TopMenuBar.tsx";
import RightSidebar from "../panels/RightSidebar.tsx";
import MainContentDisplay from "../../ui/display/MainContentDisplay.tsx";
import BottomResourceBar from "../../ui/overlay/BottomResourceBar.tsx";
import CardsHandOverlay from "../../ui/overlay/CardsHandOverlay.tsx";
import PlayerOverlay from "../../ui/overlay/PlayerOverlay.tsx";
import { MainContentProvider } from "../../../contexts/MainContentContext.tsx";
import styles from "./GameLayout.module.css";

// Mock interfaces for compatibility
interface MockGameState {
  id: string;
  players: MockPlayer[];
  currentPlayer: string;
  generation: number;
  phase: string;
  globalParameters: {
    temperature: number;
    oxygen: number;
    oceans: number;
  };
}

interface MockPlayer {
  id: string;
  name: string;
  resources: {
    credits: number;
    steel: number;
    titanium: number;
    plants: number;
    energy: number;
    heat: number;
  };
  production: {
    credits: number;
    steel: number;
    titanium: number;
    plants: number;
    energy: number;
    heat: number;
  };
  terraformRating: number;
  victoryPoints: number;
  corporation?: string;
  passed?: boolean;
  availableActions?: number;
}

interface GameLayoutProps {
  gameState: MockGameState;
  currentPlayer: MockPlayer | null;
  socket: WebSocket | null;
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
  socket,
  isAnyModalOpen = false,
  isLobbyPhase = false,
  onOpenCardEffectsModal,
  onOpenActionsModal,
  onOpenCardsPlayedModal,
  onOpenTagsModal,
  onOpenVictoryPointsModal,
}) => {
  return (
    <MainContentProvider>
      <div className={styles.gameLayout}>
        <TopMenuBar />

        <div className={styles.gameContent}>
          <LeftSidebar
            players={gameState?.players || []}
            currentPlayer={currentPlayer}
            socket={socket}
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
          players={gameState?.players || []}
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
