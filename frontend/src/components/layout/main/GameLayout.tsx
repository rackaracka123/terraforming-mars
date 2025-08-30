import React from 'react';
import LeftSidebar from '../panels/LeftSidebar.tsx';
import TopMenuBar from '../panels/TopMenuBar.tsx';
import RightSidebar from '../panels/RightSidebar.tsx';
import Game3DView from '../../game/view/Game3DView.tsx';
import BottomResourceBar from '../../ui/overlay/BottomResourceBar.tsx';
import CardsHandOverlay from '../../ui/overlay/CardsHandOverlay.tsx';
import PlayerOverlay from '../../ui/overlay/PlayerOverlay.tsx';

interface GameLayoutProps {
  gameState: any;
  currentPlayer: any;
  socket: any;
}

const GameLayout: React.FC<GameLayoutProps> = ({ gameState, currentPlayer, socket }) => {
  return (
    <div className="game-layout">
      <TopMenuBar />
      
      <div className="game-content">
        <LeftSidebar 
          players={gameState?.players || []} 
          currentPlayer={currentPlayer}
          socket={socket}
        />
        
        <Game3DView gameState={gameState} />
        
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
      
      <BottomResourceBar />
      
      <CardsHandOverlay />
      
      <style jsx>{`
        .game-layout {
          display: flex;
          flex-direction: column;
          width: 100vw;
          height: 100vh;
          background: #000011;
          color: white;
          overflow: hidden;
        }
        
        .game-content {
          display: flex;
          flex: 1;
          min-height: 0;
        }
      `}</style>
    </div>
  );
};

export default GameLayout;