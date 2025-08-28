import React from 'react';
import LeftSidebar from './LeftSidebar.tsx';
import TopMenuBar from './TopMenuBar.tsx';
import RightSidebar from './RightSidebar.tsx';
import BottomSection from './BottomSection.tsx';
import GameBoard from './GameBoard.tsx';

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
        />
        
        <GameBoard gameState={gameState} />
        
        <RightSidebar 
          globalParameters={gameState?.globalParameters}
          generation={gameState?.generation}
          currentPlayer={currentPlayer}
        />
      </div>
      
      <BottomSection 
        currentPlayer={currentPlayer}
        socket={socket}
        gameState={gameState}
      />
      
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