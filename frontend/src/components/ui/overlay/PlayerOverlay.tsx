import React from 'react';

interface Player {
  id: string;
  name: string;
  terraformRating: number;
  victoryPoints: number;
  corporation?: string;
  passed?: boolean;
  resources?: {
    credits: number;
    steel: number;
    titanium: number;
    plants: number;
    energy: number;
    heat: number;
  };
}

interface PlayerOverlayProps {
  players: Player[];
  currentPlayer: Player | null;
}

const PlayerOverlay: React.FC<PlayerOverlayProps> = ({ players, currentPlayer }) => {
  // Player color system - 6 distinct colors for up to 6 players
  const playerColors = [
    '#ff4757', // Red
    '#3742fa', // Blue  
    '#2ed573', // Green
    '#ffa502', // Orange
    '#a55eea', // Purple
    '#26d0ce', // Cyan
  ];

  // Corporation logo mapping from available assets
  const corporationLogos: { [key: string]: string } = {
    'polaris': '/assets/pathfinders/corp-logo-polaris.png',
    'mars-direct': '/assets/pathfinders/corp-logo-mars-direct.png',
    'habitat-marte': '/assets/pathfinders/corp-logo-habitat-marte.png',
    'aurorai': '/assets/pathfinders/corp-logo-aurorai.png',
    'bio-sol': '/assets/pathfinders/corp-logo-bio-sol.png',
    'chimera': '/assets/pathfinders/corp-logo-chimera.png',
    'ambient': '/assets/pathfinders/corp-logo-ambient.png',
    'odyssey': '/assets/pathfinders/corp-logo-odyssey.png',
    'steelaris': '/assets/pathfinders/corp-logo-steelaris.png',
    'soylent': '/assets/pathfinders/corp-logo-soylent.png',
    'ringcom': '/assets/pathfinders/corp-logo-ringcom.png',
    'mind-set-mars': '/assets/pathfinders/corp-logo-mind-set-mars.png',
  };

  const getCorpLogo = (corporation?: string) => {
    if (!corporation) return '/assets/pathfinders/corp-logo-polaris.png'; // Default
    return corporationLogos[corporation] || '/assets/pathfinders/corp-logo-polaris.png';
  };

  const getPlayerColor = (index: number) => {
    return playerColors[index % playerColors.length];
  };

  // Use mock data if no real players - removed all 4 mock players
  const mockPlayers: Player[] = [];

  const playersToShow = players.length > 0 ? players : mockPlayers;

  return (
    <div className="player-overlay">
      <div className="player-tabs">
        {playersToShow.map((player, index) => {
          const isCurrentPlayer = player.id === currentPlayer?.id;
          const playerColor = getPlayerColor(index);
          const corpLogo = getCorpLogo(player.corporation);
          const isPassed = player.passed || false;
          
          return (
            <div
              key={player.id || index}
              className={`player-tab ${isCurrentPlayer ? 'current' : ''} ${isPassed ? 'passed' : ''}`}
              style={{ '--player-color': playerColor } as React.CSSProperties}
            >
              <div className="tab-content">
                <div className="corp-section">
                  <img 
                    src={corpLogo} 
                    alt={`${player.corporation || 'Unknown'} Corporation`}
                    className="corp-logo"
                  />
                </div>
                
                <div className="player-info">
                  <div className="player-name">{player.name}</div>
                  <div className="tr-section">
                    <span className="tr-label">TR</span>
                    <span className="tr-value">{player.terraformRating}</span>
                  </div>
                </div>
                
                {isPassed && (
                  <div className="pass-indicator">PASSED</div>
                )}
              </div>
            </div>
          );
        })}
      </div>
      
      <style jsx>{`
        .player-overlay {
          position: absolute;
          top: 70px;
          left: 50%;
          transform: translateX(-50%);
          z-index: 90;
          pointer-events: none;
        }
        
        .player-tabs {
          display: flex;
          gap: 8px;
          align-items: center;
          justify-content: center;
        }
        
        .player-tab {
          background: linear-gradient(
            135deg,
            rgba(30, 60, 90, 0.95) 0%,
            rgba(20, 40, 70, 0.9) 50%,
            rgba(10, 30, 60, 0.95) 100%
          );
          border: 2px solid var(--player-color);
          border-radius: 12px;
          padding: 8px 16px;
          min-width: 140px;
          backdrop-filter: blur(10px);
          transition: all 0.3s ease;
          pointer-events: auto;
          cursor: pointer;
          box-shadow: 
            0 4px 20px rgba(0, 0, 0, 0.4),
            0 0 20px var(--player-color, #4a90e2);
        }
        
        .player-tab::before {
          content: '';
          position: absolute;
          top: 0;
          left: 0;
          right: 0;
          bottom: 0;
          background: linear-gradient(
            45deg,
            var(--player-color, #4a90e2) 0%,
            transparent 50%,
            var(--player-color, #4a90e2) 100%
          );
          opacity: 0.1;
          border-radius: 10px;
          transition: opacity 0.3s ease;
        }
        
        .player-tab:hover::before {
          opacity: 0.2;
        }
        
        .player-tab:hover {
          transform: translateY(-2px);
          box-shadow: 
            0 6px 25px rgba(0, 0, 0, 0.5),
            0 0 30px var(--player-color, #4a90e2);
        }
        
        .player-tab.current {
          border-width: 3px;
          transform: translateY(-3px);
          box-shadow: 
            0 8px 30px rgba(0, 0, 0, 0.6),
            0 0 40px var(--player-color, #4a90e2);
        }
        
        .player-tab.current::before {
          opacity: 0.25;
        }
        
        .player-tab.passed {
          opacity: 0.6;
          border-color: rgba(150, 150, 150, 0.5);
          background: linear-gradient(
            135deg,
            rgba(50, 50, 50, 0.9) 0%,
            rgba(40, 40, 40, 0.8) 50%,
            rgba(30, 30, 30, 0.9) 100%
          );
          box-shadow: 
            0 4px 20px rgba(0, 0, 0, 0.4),
            0 0 10px rgba(150, 150, 150, 0.3);
        }
        
        .player-tab.passed:hover {
          opacity: 0.8;
          transform: translateY(-1px);
        }
        
        .player-tab.passed .corp-logo {
          filter: grayscale(100%) brightness(0.7);
        }
        
        .player-tab.passed .player-name,
        .player-tab.passed .tr-value {
          color: rgba(255, 255, 255, 0.7);
        }
        
        .tab-content {
          display: flex;
          align-items: center;
          gap: 12px;
          position: relative;
        }
        
        .corp-section {
          flex-shrink: 0;
        }
        
        .corp-logo {
          width: 36px;
          height: 36px;
          border-radius: 8px;
          border: 1px solid rgba(255, 255, 255, 0.2);
          object-fit: cover;
          transition: all 0.3s ease;
        }
        
        .player-info {
          display: flex;
          flex-direction: column;
          gap: 2px;
          flex-grow: 1;
          min-width: 0;
        }
        
        .player-name {
          font-size: 14px;
          font-weight: bold;
          color: #ffffff;
          text-shadow: 0 1px 2px rgba(0, 0, 0, 0.8);
          white-space: nowrap;
          overflow: hidden;
          text-overflow: ellipsis;
        }
        
        .tr-section {
          display: flex;
          align-items: center;
          gap: 4px;
        }
        
        .tr-label {
          font-size: 10px;
          font-weight: bold;
          color: rgba(255, 255, 255, 0.7);
          text-transform: uppercase;
          letter-spacing: 0.5px;
        }
        
        .tr-value {
          font-size: 18px;
          font-weight: bold;
          color: var(--player-color, #4a90e2);
          text-shadow: 0 1px 3px rgba(0, 0, 0, 0.8);
          font-family: 'Courier New', monospace;
        }
        
        .pass-indicator {
          position: absolute;
          top: -8px;
          right: -8px;
          font-size: 8px;
          font-weight: bold;
          color: rgba(100, 255, 100, 0.9);
          background: rgba(50, 150, 50, 0.8);
          padding: 2px 6px;
          border-radius: 4px;
          border: 1px solid rgba(100, 255, 100, 0.4);
          text-shadow: 0 1px 2px rgba(0, 0, 0, 0.7);
          letter-spacing: 0.3px;
          backdrop-filter: blur(5px);
        }
        
        @media (max-width: 1200px) {
          .player-tab {
            min-width: 120px;
            padding: 6px 12px;
          }
          
          .corp-logo {
            width: 32px;
            height: 32px;
          }
          
          .player-name {
            font-size: 12px;
          }
          
          .tr-value {
            font-size: 16px;
          }
        }
        
        @media (max-width: 800px) {
          .player-tabs {
            gap: 4px;
          }
          
          .player-tab {
            min-width: 100px;
            padding: 4px 8px;
          }
          
          .corp-logo {
            width: 28px;
            height: 28px;
          }
          
          .player-name {
            font-size: 11px;
          }
          
          .tr-value {
            font-size: 14px;
          }
        }
      `}</style>
    </div>
  );
};

export default PlayerOverlay;