import React, { useState } from 'react';
import ReactDOM from 'react-dom';

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
  // Player color system - 6 distinct colors for up to 6 players
  const playerColors = [
    '#ff4757', // Red
    '#3742fa', // Blue  
    '#2ed573', // Green
    '#ffa502', // Orange
    '#a55eea', // Purple
    '#26d0ce', // Cyan
  ];

  const getPlayerColor = (index: number) => {
    return playerColors[index % playerColors.length];
  };

  // Corporation information for tooltips
  const corporationInfo: { [key: string]: { name: string; description: string; ability: string } } = {
    'mars-direct': {
      name: 'Mars Direct',
      description: 'Direct approach to Mars terraforming with efficient resource management.',
      ability: 'Start with extra steel production and building discounts.'
    },
    'habitat-marte': {
      name: 'Habitat Marte',
      description: 'Specializes in creating sustainable living environments on Mars.',
      ability: 'Reduced costs for city and habitat construction projects.'
    },
    'aurorai': {
      name: 'Aurorai',
      description: 'Advanced atmospheric engineering and terraforming technology.',
      ability: 'Bonuses for oxygen and temperature raising projects.'
    },
    'bio-sol': {
      name: 'Bio-Sol',
      description: 'Biological solutions for Mars ecosystem development.',
      ability: 'Extra benefits from plant and microbe-based cards.'
    },
    'chimera': {
      name: 'Chimera',
      description: 'Hybrid technology corporation combining multiple approaches.',
      ability: 'Flexible bonuses that adapt to different card types.'
    },
    'odyssey': {
      name: 'Odyssey',
      description: 'Space exploration and logistics specialists.',
      ability: 'Reduced costs for space-based projects and colonies.'
    }
  };
  // Mock players with different corporations - reduced to 2 players
  const mockPlayers = [
    { 
      id: '1', 
      name: 'Alice Chen', 
      score: 76, 
      passed: true,
      corporation: 'mars-direct',
      terraformRating: 35,
      victoryPoints: 76
    },
    { 
      id: '3', 
      name: 'Carol Kim', 
      score: 28, 
      passed: false,
      corporation: 'aurorai',
      terraformRating: 28,
      victoryPoints: 28
    }
  ];

  const playersToShow = mockPlayers;
  
  // For demo purposes, simulate Carol Kim as the current player
  const mockCurrentPlayer = { id: '3', name: 'Carol Kim' };
  
  // State for tooltip management
  const [hoveredCorp, setHoveredCorp] = useState<string | null>(null);
  const [tooltipPosition, setTooltipPosition] = useState<{ top: number; left: number }>({ top: 0, left: 0 });

  const handleCorpHover = (playerId: string, corporation: string, event: React.MouseEvent) => {
    const rect = event.currentTarget.getBoundingClientRect();
    const tooltipWidth = 280; // Width of the tooltip as defined in CSS
    const tooltipHeight = 120; // Approximate height of tooltip
    const viewportWidth = window.innerWidth;
    const viewportHeight = window.innerHeight;
    const margin = 15;
    
    // Calculate preferred position (to the right of the logo)
    let left = rect.right + margin;
    let top = rect.top + window.scrollY - 10;
    
    // Check if tooltip would overflow right edge of viewport
    if (left + tooltipWidth > viewportWidth) {
      // Position to the left of the logo instead
      left = rect.left - tooltipWidth - margin;
    }
    
    // Ensure tooltip doesn't go off the left edge
    if (left < margin) {
      left = margin;
    }
    
    // Check vertical positioning and adjust if necessary
    if (top + tooltipHeight > window.scrollY + viewportHeight) {
      top = rect.bottom + window.scrollY + 10;
      
      // If still doesn't fit, position above the logo
      if (top + tooltipHeight > window.scrollY + viewportHeight) {
        top = rect.top + window.scrollY - tooltipHeight - 10;
      }
    }
    
    // Ensure tooltip doesn't go above viewport
    if (top < window.scrollY + margin) {
      top = window.scrollY + margin;
    }
    
    setTooltipPosition({ top, left });
    setHoveredCorp(`${playerId}-${corporation}`);
  };

  return (
    <div className="left-sidebar">
      <div className="players-list">
        {playersToShow.map((player, index) => {
          const score = player.score || player.victoryPoints || player.terraformRating || 20;
          const isPassed = player.passed;
          const isCurrentPlayer = player.id === mockCurrentPlayer.id;
          const corporation = player.corporation || 'polaris';
          const playerColor = getPlayerColor(index);
          
          return (
            <div 
              key={player.id || index} 
              className={`player-entry ${isCurrentPlayer ? 'current' : ''} ${isPassed ? 'passed' : ''}`}
              style={{ '--player-color': playerColor } as React.CSSProperties}
            >
              <div className="player-content">
                <div className="player-avatar">
                  <img 
                    src={`/assets/pathfinders/corp-logo-${corporation}.png`} 
                    alt={`${corporation} Corporation`}
                    className="corp-logo-img"
                    onMouseEnter={(e) => handleCorpHover(player.id, corporation, e)}
                    onMouseLeave={() => setHoveredCorp(null)}
                  />
                </div>
                <div className="player-info-section">
                  <div className="player-name">{player.name}</div>
                  <div className="player-score">{score}</div>
                  {isCurrentPlayer && <div className="you-indicator">YOU</div>}
                  {isPassed && <div className="passed-indicator">PASSED</div>}
                </div>
              </div>
            </div>
          );
        })}
      </div>
      
      {/* Global tooltip rendered as a portal to document body */}
      {hoveredCorp && ReactDOM.createPortal(
        <div 
          className="corp-tooltip"
          style={{
            top: tooltipPosition.top,
            left: tooltipPosition.left
          }}
        >
          {(() => {
            const [, corporation] = hoveredCorp.split('-');
            const corpInfo = corporationInfo[corporation];
            return corpInfo ? (
              <>
                <div className="corp-tooltip-header">
                  <strong>{corpInfo.name}</strong>
                </div>
                <div className="corp-tooltip-description">
                  {corpInfo.description}
                </div>
                <div className="corp-tooltip-ability">
                  <strong>Ability:</strong> {corpInfo.ability}
                </div>
              </>
            ) : null;
          })()}
        </div>,
        document.body
      )}
      
      <style jsx global>{`
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
          overflow: visible;
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
          gap: 12px;
          padding: 20px 15px;
          overflow: visible;
          height: 100%;
          justify-content: flex-start;
        }
        
        .player-entry {
          position: relative;
          background: var(--player-color, rgba(30, 60, 90, 0.8));
          border: 2px solid rgba(255, 255, 255, 0.2);
          padding: 16px;
          transition: all 0.3s ease;
          clip-path: polygon(0 0, calc(100% - 15px) 0, 100% 100%, 0 100%);
          min-height: 80px;
          flex: 0 0 auto;
          box-shadow: 
            0 4px 15px rgba(0, 0, 0, 0.4),
            0 0 15px var(--player-color, rgba(100, 150, 200, 0.3));
        }
        
        .player-entry::before {
          content: '';
          position: absolute;
          top: 0;
          left: 0;
          right: 0;
          bottom: 0;
          background: rgba(0, 0, 0, 0.4);
          transition: opacity 0.3s ease;
          clip-path: inherit;
        }
        
        .player-entry:hover::before {
          opacity: 1;
        }
        
        .player-entry:hover {
          border-color: var(--player-color, rgba(150, 200, 255, 0.5));
          transform: translateX(3px);
          box-shadow: 
            0 6px 25px rgba(0, 0, 0, 0.4),
            0 0 25px var(--player-color, rgba(100, 150, 255, 0.3)),
            inset 0 1px 0 rgba(255, 255, 255, 0.1);
        }
        
        .player-entry.current {
          border: 3px solid rgba(255, 255, 255, 0.9);
          box-shadow: 
            0 0 20px rgba(255, 255, 255, 0.4),
            0 0 40px rgba(255, 255, 255, 0.2),
            inset 0 1px 0 rgba(255, 255, 255, 0.3);
          transform: translateX(2px);
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
          justify-content: center;
          width: 48px;
          height: 48px;
          flex-shrink: 0;
          position: relative;
        }
        
        .corp-logo-img {
          width: 44px;
          height: 44px;
          border-radius: 8px;
          border: 2px solid rgba(255, 255, 255, 0.6);
          object-fit: contain;
          object-position: center;
          background: rgba(0, 0, 0, 0.2);
          transition: all 0.3s ease;
          max-width: 100%;
          max-height: 100%;
        }
        
        .corp-logo-img:hover {
          border-color: rgba(255, 255, 255, 0.9);
          transform: scale(1.05);
          cursor: pointer;
          box-shadow: 0 0 15px rgba(255, 255, 255, 0.4);
        }
        
        .corp-tooltip {
          position: fixed;
          background: linear-gradient(
            135deg,
            rgba(5, 15, 35, 0.98) 0%,
            rgba(15, 25, 45, 0.96) 50%,
            rgba(10, 20, 40, 0.98) 100%
          );
          border: 2px solid rgba(120, 170, 255, 0.7);
          border-radius: 10px;
          padding: 14px;
          width: 280px;
          max-width: calc(100vw - 30px);
          font-size: 12px;
          color: white;
          box-shadow: 
            0 12px 35px rgba(0, 0, 0, 0.8),
            0 0 25px rgba(120, 170, 255, 0.4),
            inset 0 1px 0 rgba(255, 255, 255, 0.1);
          z-index: 10000;
          backdrop-filter: blur(15px);
          animation: tooltipFadeIn 0.25s ease-out;
          pointer-events: none;
          word-wrap: break-word;
          overflow-wrap: break-word;
        }
        
        .corp-tooltip::before {
          content: '';
          position: absolute;
          top: 0;
          left: 0;
          right: 0;
          bottom: 0;
          background: linear-gradient(
            45deg,
            rgba(150, 200, 255, 0.08) 0%,
            transparent 50%,
            rgba(100, 150, 255, 0.04) 100%
          );
          border-radius: inherit;
          pointer-events: none;
        }
        
        @keyframes tooltipFadeIn {
          from {
            opacity: 0;
            transform: translateY(8px) scale(0.95);
          }
          to {
            opacity: 1;
            transform: translateY(0) scale(1);
          }
        }
        
        .corp-tooltip-header {
          font-size: 15px;
          font-weight: 600;
          color: rgba(170, 220, 255, 1);
          margin-bottom: 10px;
          text-shadow: 0 1px 3px rgba(0, 0, 0, 0.9);
          letter-spacing: 0.3px;
        }
        
        .corp-tooltip-description {
          line-height: 1.5;
          margin-bottom: 10px;
          color: rgba(255, 255, 255, 0.92);
          font-size: 12px;
        }
        
        .corp-tooltip-ability {
          font-size: 11px;
          color: rgba(255, 210, 120, 0.95);
          line-height: 1.4;
          padding-top: 8px;
          border-top: 1px solid rgba(120, 170, 255, 0.4);
        }
        
        .corp-tooltip-ability strong {
          color: rgba(255, 220, 140, 1);
          font-weight: 600;
        }
        
        .player-info-section {
          display: flex;
          flex-direction: column;
          align-items: flex-end;
          gap: 2px;
        }
        
        .player-info-section .player-name {
          font-size: 12px !important;
          font-weight: bold;
          color: #ffffff !important;
          text-shadow: 0 1px 2px rgba(0, 0, 0, 0.7);
          margin-bottom: 2px;
          text-align: right;
        }
        
        .player-score {
          font-size: 28px;
          font-weight: bold;
          color: #ffffff;
          text-shadow: 
            0 1px 2px rgba(0, 0, 0, 0.9),
            0 0 10px rgba(0, 0, 0, 0.8);
          font-family: 'Courier New', monospace;
        }
        
        .you-indicator {
          font-size: 10px;
          font-weight: bold;
          color: rgba(255, 200, 100, 1);
          background: rgba(150, 100, 50, 0.4);
          padding: 2px 6px;
          border-radius: 3px;
          border: 1px solid rgba(255, 200, 100, 0.6);
          text-shadow: 0 1px 2px rgba(0, 0, 0, 0.7);
          letter-spacing: 0.5px;
          margin-bottom: 2px;
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