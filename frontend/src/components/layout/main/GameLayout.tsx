import React from "react";
import LeftSidebar from "../panels/LeftSidebar.tsx";
import TopMenuBar from "../panels/TopMenuBar.tsx";
import RightSidebar from "../panels/RightSidebar.tsx";
import MainContentDisplay from "../../ui/display/MainContentDisplay.tsx";
import BottomResourceBar from "../../ui/overlay/BottomResourceBar.tsx";
import CardsHandOverlay from "../../ui/overlay/CardsHandOverlay.tsx";
import PlayerOverlay from "../../ui/overlay/PlayerOverlay.tsx";
import { MainContentProvider } from "../../../contexts/MainContentContext.tsx";

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
  onOpenCardEffectsModal,
  onOpenActionsModal,
  onOpenCardsPlayedModal,
  onOpenTagsModal,
  onOpenVictoryPointsModal,
}) => {
  return (
    <MainContentProvider>
      <div className="game-layout">
        <TopMenuBar />

        <div className="game-content">
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

        <BottomResourceBar
          currentPlayer={currentPlayer}
          onOpenCardEffectsModal={onOpenCardEffectsModal}
          onOpenActionsModal={onOpenActionsModal}
          onOpenCardsPlayedModal={onOpenCardsPlayedModal}
          onOpenTagsModal={onOpenTagsModal}
          onOpenVictoryPointsModal={onOpenVictoryPointsModal}
        />

        <CardsHandOverlay hideWhenModalOpen={isAnyModalOpen} />

        <style jsx>{`
          .game-layout {
            display: grid;
            grid-template-rows: auto 1fr;
            width: 100vw;
            height: 100vh;
            background: #000011 url("/assets/background-noise.png");
            background-attachment: fixed;
            background-repeat: repeat;
            color: white;
            overflow: hidden;
          }

          .game-content {
            display: grid;
            grid-template-columns: minmax(280px, 320px) 1fr minmax(150px, 250px);
            min-height: 0;
            gap: 0;
          }

          @media (max-width: 1200px) {
            .game-content {
              grid-template-columns: minmax(250px, 280px) 1fr minmax(
                  120px,
                  180px
                );
            }
          }

          @media (max-width: 900px) {
            .game-content {
              grid-template-columns: minmax(200px, 240px) 1fr minmax(
                  100px,
                  150px
                );
            }
          }

          @media (max-width: 768px) {
            .game-content {
              grid-template-columns: 1fr;
              grid-template-rows: auto 1fr auto;
            }
          }
        `}</style>
      </div>
    </MainContentProvider>
  );
};

export default GameLayout;
