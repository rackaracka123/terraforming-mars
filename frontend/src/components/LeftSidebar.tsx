import React from 'react';

interface Player {
  id: string;
  name: string;
  terraformRating: number;
  victoryPoints: number;
  passed?: boolean;
  resources: {
    credits: number;
    steel: number;
    titanium: number;
    plants: number;
    energy: number;
    heat: number;
  };
}

interface LeftSidebarProps {
  players: Player[];
  currentPlayer: Player | null;
}

const LeftSidebar: React.FC<LeftSidebarProps> = ({ players, currentPlayer }) => {
  // Mock data to match reference image scores
  const mockPlayers = [
    { id: '1', name: 'Player 1', score: 76, passed: true },
    { id: '2', name: 'Player 2', score: 76, passed: true },
    { id: '3', name: 'Player 3', score: 28, passed: false },
    { id: '4', name: 'Player 4', score: 24, passed: false },
    { id: '5', name: 'Player 5', score: 27, passed: false },
    { id: '6', name: 'Player 6', score: 19, passed: true },
  ];

  const playersToShow = players.length > 0 ? players : mockPlayers;

  return (
    <div className="left-sidebar">
      <div className="players-list">
        {playersToShow.map((player, index) => {
          const score = player.score || player.victoryPoints || player.terraformRating || mockPlayers[index]?.score || 20;
          const isPassed = player.passed || mockPlayers[index]?.passed;
          const isCurrentPlayer = player.id === currentPlayer?.id;
          
          return (
            <div 
              key={player.id || index} 
              className={`player-entry ${isCurrentPlayer ? 'current' : ''} ${isPassed ? 'passed' : ''}`}
            >
              <div className="player-content">
                <div className="player-avatar">
                  <img 
                    src="/assets/pathfinders/corp-logo-polaris.png" 
                    alt="Corp Logo"
                    className="corp-logo-img"
                  />
                </div>
                <div className="player-score-section">
                  <div className="player-score">{score}</div>
                  {isPassed && <div className="passed-indicator">PASSED</div>}
                </div>
              </div>
            </div>
          );
        })}
      </div>
      
      <style jsx>{`
        .left-sidebar {
          width: 200px;
          background: linear-gradient(180deg, 
            rgba(0, 20, 40, 0.95) 0%, 
            rgba(0, 10, 30, 0.95) 50%, 
            rgba(0, 5, 20, 0.95) 100%
          );
          backdrop-filter: blur(10px);
          border-right: 1px solid rgba(100, 150, 200, 0.2);
          padding: 20px 0;
          display: flex;
          flex-direction: column;
          position: relative;
        }
        
        .left-sidebar::before {
          content: '';
          position: absolute;
          top: 0;
          left: 0;
          right: 0;
          bottom: 0;
          background: linear-gradient(
            135deg,
            rgba(100, 150, 255, 0.05) 0%,
            transparent 50%,
            rgba(0, 100, 200, 0.03) 100%
          );
          pointer-events: none;
        }
        
        .players-list {
          display: flex;
          flex-direction: column;
          gap: 8px;
          padding: 0 15px;
        }
        
        .player-entry {
          position: relative;
          background: linear-gradient(
            135deg,
            rgba(30, 60, 90, 0.8) 0%,
            rgba(20, 40, 70, 0.6) 50%,
            rgba(10, 30, 60, 0.8) 100%
          );
          border: 1px solid rgba(100, 150, 200, 0.3);
          padding: 12px;
          transition: all 0.3s ease;
          clip-path: polygon(0 0, calc(100% - 15px) 0, 100% 100%, 0 100%);
        }
        
        .player-entry::before {
          content: '';
          position: absolute;
          top: 0;
          left: 0;
          right: 0;
          bottom: 0;
          background: linear-gradient(
            45deg,
            rgba(150, 200, 255, 0.1) 0%,
            transparent 50%,
            rgba(100, 150, 255, 0.05) 100%
          );
          opacity: 0;
          transition: opacity 0.3s ease;
          clip-path: inherit;
        }
        
        .player-entry:hover::before {
          opacity: 1;
        }
        
        .player-entry:hover {
          border-color: rgba(150, 200, 255, 0.5);
          transform: translateX(3px);
          box-shadow: 
            0 4px 20px rgba(100, 150, 255, 0.15),
            inset 0 1px 0 rgba(255, 255, 255, 0.1);
        }
        
        .player-entry.current {
          border-color: rgba(255, 200, 100, 0.8);
          background: linear-gradient(
            135deg,
            rgba(60, 50, 30, 0.9) 0%,
            rgba(40, 35, 20, 0.7) 50%,
            rgba(30, 25, 15, 0.9) 100%
          );
          box-shadow: 
            0 0 20px rgba(255, 200, 100, 0.3),
            inset 0 1px 0 rgba(255, 200, 100, 0.2);
        }
        
        .player-entry.passed {
          opacity: 0.5;
          background: linear-gradient(
            135deg,
            rgba(50, 50, 50, 0.8) 0%,
            rgba(40, 40, 40, 0.6) 50%,
            rgba(30, 30, 30, 0.8) 100%
          );
          border-color: rgba(100, 100, 100, 0.3);
        }
        
        .player-entry.passed:hover {
          opacity: 0.7;
          border-color: rgba(120, 120, 120, 0.4);
          transform: translateX(1px);
        }
        
        .player-entry.passed .player-score {
          color: rgba(255, 255, 255, 0.6);
        }
        
        .player-entry.passed .corp-logo-img {
          filter: grayscale(100%) brightness(0.7);
        }
        
        .player-content {
          display: flex;
          align-items: center;
          justify-content: space-between;
          position: relative;
          z-index: 1;
        }
        
        .player-avatar {
          display: flex;
          align-items: center;
        }
        
        .corp-logo-img {
          width: 32px;
          height: 32px;
          border-radius: 6px;
          border: 2px solid rgba(150, 200, 255, 0.5);
          object-fit: cover;
          transition: all 0.3s ease;
        }
        
        .corp-logo-img:hover {
          border-color: rgba(200, 250, 255, 0.8);
          transform: scale(1.05);
        }
        
        .player-score-section {
          display: flex;
          flex-direction: column;
          align-items: flex-end;
          gap: 2px;
        }
        
        .player-score {
          font-size: 28px;
          font-weight: bold;
          color: #ffffff;
          text-shadow: 
            0 1px 2px rgba(0, 0, 0, 0.8),
            0 0 10px rgba(150, 200, 255, 0.3);
          font-family: 'Courier New', monospace;
        }
        
        .passed-indicator {
          font-size: 10px;
          font-weight: bold;
          color: rgba(100, 255, 100, 0.9);
          background: rgba(50, 150, 50, 0.3);
          padding: 2px 6px;
          border-radius: 3px;
          border: 1px solid rgba(100, 255, 100, 0.4);
          text-shadow: 0 1px 2px rgba(0, 0, 0, 0.7);
          letter-spacing: 0.5px;
        }
      `}</style>
    </div>
  );
};

export default LeftSidebar;