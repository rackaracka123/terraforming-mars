import React from 'react';
import LeftSidebar from '../panels/LeftSidebar.tsx';
import TopMenuBar from '../panels/TopMenuBar.tsx';
import RightSidebar from '../panels/RightSidebar.tsx';
import BottomSection from '../panels/BottomSection.tsx';
import GameBoard from '../../game/board/GameBoard.tsx';

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