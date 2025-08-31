import React from 'react';
import { Z_INDEX } from '../../../constants/zIndex.ts';

interface GameBoardProps {
  gameState?: any;
}

const GameBoard: React.FC<GameBoardProps> = ({ gameState }) => {
  // Mock hex grid data - representing the 3D Mars surface
  const generateHexGrid = () => {
    const hexes = [];
    const rows = 6;
    const maxCols = 7;
    
    for (let row = 0; row < rows; row++) {
      const cols = row < 3 ? maxCols - row : row - 2;
      const offset = row < 3 ? row * 0.5 : (5 - row) * 0.5;
      
      for (let col = 0; col < cols; col++) {
        const hexId = `${row}-${col}`;
        const type = Math.random() > 0.7 ? 'special' : 'normal';
        const hasOcean = Math.random() > 0.85;
        const hasForest = Math.random() > 0.8 && !hasOcean;
        const hasCity = Math.random() > 0.9 && !hasOcean && !hasForest;
        
        hexes.push({
          id: hexId,
          row,
          col,
          x: (col + offset) * 60,
          y: row * 52,
          type,
          hasOcean,
          hasForest,
          hasCity,
          temperature: gameState?.globalParameters?.temperature || -30
        });
      }
    }
    
    return hexes;
  };

  const hexGrid = generateHexGrid();

  const getHexColor = (hex: any) => {
    if (hex.hasOcean) return '#3498db';
    if (hex.hasForest) return '#27ae60';
    if (hex.hasCity) return '#f39c12';
    if (hex.type === 'special') return '#e74c3c';
    
    // Temperature-based coloring for Mars surface
    const temp = hex.temperature;
    if (temp < -20) return '#8b4513';  // Cold brown
    if (temp < 0) return '#cd853f';    // Warmer brown
    return '#daa520';                  // Desert gold
  };

  const getHexIcon = (hex: any) => {
    if (hex.hasOcean) return '🌊';
    if (hex.hasForest) return '🌲';
    if (hex.hasCity) return '🏙️';
    if (hex.type === 'special') return '⭐';
    return '';
  };

  return (
    <div className="game-board">
      <div className="board-background">
        <div className="mars-atmosphere"></div>
        <div className="space-bg">
          <div className="star star-1"></div>
          <div className="star star-2"></div>
          <div className="star star-3"></div>
          <div className="star star-4"></div>
          <div className="star star-5"></div>
        </div>
      </div>
      
      <div className="hex-grid-container">
        <div className="hex-grid">
          {hexGrid.map((hex) => (
            <div
              key={hex.id}
              className="hex"
              style={{
                left: `${hex.x}px`,
                top: `${hex.y}px`,
                backgroundColor: getHexColor(hex)
              }}
              onClick={() => console.log('Clicked hex:', hex.id)}
            >
              <div className="hex-inner">
                <div className="hex-content">
                  {getHexIcon(hex) && (
                    <div className="hex-icon">{getHexIcon(hex)}</div>
                  )}
                  <div className="hex-coords">{hex.id}</div>
                </div>
              </div>
            </div>
          ))}
        </div>
        
        <div className="board-overlay">
          <div className="planet-info">
            <h3>🪐 Mars</h3>
            <div className="planet-status">
              <div>Temperature: {gameState?.globalParameters?.temperature || -30}°C</div>
              <div>Oxygen: {gameState?.globalParameters?.oxygen || 0}%</div>
              <div>Oceans: {gameState?.globalParameters?.oceans || 0}/9</div>
            </div>
          </div>
        </div>
      </div>
      
      <style jsx>{`
        .game-board {
          flex: 1;
          position: relative;
          overflow: hidden;
          display: flex;
          align-items: center;
          justify-content: center;
          background: radial-gradient(circle at center, #1a1a2e, #16213e, #0f0f23);
        }
        
        .board-background {
          position: absolute;
          inset: 0;
          z-index: 0;
        }
        
        .mars-atmosphere {
          position: absolute;
          inset: 0;
          background: radial-gradient(
            ellipse at center,
            rgba(205, 92, 92, 0.1) 0%,
            rgba(139, 69, 19, 0.05) 40%,
            transparent 70%
          );
          animation: atmosphereGlow 4s ease-in-out infinite alternate;
        }
        
        .space-bg {
          position: absolute;
          inset: 0;
        }
        
        .star {
          position: absolute;
          width: 2px;
          height: 2px;
          background: white;
          border-radius: 50%;
          animation: twinkle 3s infinite;
        }
        
        .star-1 { top: 20%; left: 10%; animation-delay: 0s; }
        .star-2 { top: 30%; right: 15%; animation-delay: 1s; }
        .star-3 { top: 60%; left: 20%; animation-delay: 2s; }
        .star-4 { bottom: 30%; right: 25%; animation-delay: 0.5s; }
        .star-5 { bottom: 20%; left: 80%; animation-delay: 1.5s; }
        
        .hex-grid-container {
          position: relative;
          z-index: 1;
          padding: 50px;
        }
        
        .hex-grid {
          position: relative;
          width: 500px;
          height: 400px;
        }
        
        .hex {
          position: absolute;
          width: 50px;
          height: 43.3px;
          cursor: pointer;
          transition: all 0.2s ease;
        }
        
        .hex:before {
          content: '';
          position: absolute;
          top: -12.5px;
          left: 0;
          width: 0;
          height: 0;
          border-left: 25px solid transparent;
          border-right: 25px solid transparent;
          border-bottom: 12.5px solid;
          border-bottom-color: inherit;
        }
        
        .hex:after {
          content: '';
          position: absolute;
          bottom: -12.5px;
          left: 0;
          width: 0;
          height: 0;
          border-left: 25px solid transparent;
          border-right: 25px solid transparent;
          border-top: 12.5px solid;
          border-top-color: inherit;
        }
        
        .hex:hover {
          transform: scale(1.1);
          z-index: 10;
          filter: brightness(1.2);
        }
        
        .hex-inner {
          width: 100%;
          height: 100%;
          background-color: inherit;
          display: flex;
          align-items: center;
          justify-content: center;
          position: relative;
          z-index: 1;
        }
        
        .hex-content {
          text-align: center;
          color: white;
          font-size: 10px;
        }
        
        .hex-icon {
          font-size: 16px;
          margin-bottom: 2px;
        }
        
        .hex-coords {
          font-size: 8px;
          opacity: 0.7;
        }
        
        .board-overlay {
          position: absolute;
          top: 20px;
          left: 20px;
          z-index: 5;
        }
        
        .planet-info {
          background: rgba(0, 0, 0, 0.8);
          border-radius: 8px;
          padding: 15px;
          color: white;
          border: 1px solid #333;
        }
        
        .planet-info h3 {
          margin: 0 0 10px 0;
          color: #cd853f;
        }
        
        .planet-status {
          font-size: 12px;
          display: flex;
          flex-direction: column;
          gap: 4px;
        }
        
        @keyframes atmosphereGlow {
          from {
            opacity: 0.3;
          }
          to {
            opacity: 0.6;
          }
        }
        
        @keyframes twinkle {
          0%, 100% {
            opacity: 0.3;
          }
          50% {
            opacity: 1;
          }
        }
      `}</style>
    </div>
  );
};

export default GameBoard;