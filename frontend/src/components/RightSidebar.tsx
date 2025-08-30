import React from 'react';

interface GlobalParameters {
  temperature: number;
  oxygen: number;
  oceans: number;
}

interface RightSidebarProps {
  globalParameters?: GlobalParameters;
  generation?: number;
  currentPlayer?: any;
}

const RightSidebar: React.FC<RightSidebarProps> = ({ 
  globalParameters, 
  generation, 
  currentPlayer 
}) => {
  // Temperature ranges from -30°C to +8°C in steps of 2°C (19 total steps)
  const getTemperatureSteps = () => {
    const temp = globalParameters?.temperature || -30;
    const steps = [];
    for (let i = 0; i < 19; i++) {
      const stepTemp = 8 - (i * 2); // 8, 6, 4, 2, 0, -2, -4, ..., -30
      steps.push({
        value: stepTemp,
        isActive: temp >= stepTemp
      });
    }
    return steps;
  };

  // Oxygen from 0% to 14% (14 steps)
  const getOxygenSteps = () => {
    const oxygen = globalParameters?.oxygen || 0;
    const steps = [];
    for (let i = 1; i <= 14; i++) {
      steps.push({
        value: i,
        isActive: oxygen >= i
      });
    }
    return steps;
  };

  // Oceans from 0 to 9 (9 steps)
  const getOceanSteps = () => {
    const oceans = globalParameters?.oceans || 0;
    const steps = [];
    for (let i = 1; i <= 9; i++) {
      steps.push({
        value: i,
        isActive: oceans >= i
      });
    }
    return steps;
  };

  return (
    <div className="right-sidebar">
      {/* Generation Counter - matching reference design */}
      <div className="generation-counter">
        <div className="generation-hex">
          <div className="gen-text">GEN</div>
          <div className="gen-number">{generation || 1}</div>
        </div>
      </div>
      
      {/* Global Parameters Vertical Tracks - matching reference */}
      <div className="global-parameters">
        {/* Temperature Track */}
        <div className="parameter-track temperature-track">
          <div className="track-header">
            <div className="track-icon">🌡️</div>
            <div className="track-bar temperature-bar">
              {getTemperatureSteps().map((step, i) => (
                <div 
                  key={i}
                  className={`track-segment temperature ${step.isActive ? 'active' : ''}`}
                  title={`${step.value}°C`}
                />
              ))}
            </div>
            <div className="track-values">
              <div className="track-value top">8</div>
              <div className="track-value bottom">-30</div>
            </div>
          </div>
          <div className="current-value">{globalParameters?.temperature || -30}°C</div>
        </div>

        {/* Oxygen Track */}
        <div className="parameter-track oxygen-track">
          <div className="track-header">
            <div className="track-icon">💨</div>
            <div className="track-bar oxygen-bar">
              {getOxygenSteps().map((step, i) => (
                <div 
                  key={i}
                  className={`track-segment oxygen ${step.isActive ? 'active' : ''}`}
                  title={`${step.value}%`}
                />
              ))}
            </div>
            <div className="track-values">
              <div className="track-value top">14</div>
              <div className="track-value bottom">0</div>
            </div>
          </div>
          <div className="current-value">{globalParameters?.oxygen || 0}%</div>
        </div>

        {/* Oceans Track */}
        <div className="parameter-track ocean-track">
          <div className="track-header">
            <div className="track-icon">🌊</div>
            <div className="track-bar ocean-bar">
              {getOceanSteps().map((step, i) => (
                <div 
                  key={i}
                  className={`track-segment ocean ${step.isActive ? 'active' : ''}`}
                  title={`${step.value} oceans`}
                />
              ))}
            </div>
            <div className="track-values">
              <div className="track-value top">9</div>
              <div className="track-value bottom">0</div>
            </div>
          </div>
          <div className="current-value">{globalParameters?.oceans || 0}/9</div>
        </div>
      </div>

      {/* Player Score Section */}
      <div className="player-score-section">
        {currentPlayer && (
          <div className="score-hex-container">
            <div className="score-hex">
              <div className="score-value">{currentPlayer.terraformRating || 20}</div>
            </div>
            <div className="player-name">{currentPlayer.name}</div>
          </div>
        )}
      </div>
      
      <style>{`
        .right-sidebar {
          width: 60px;
          height: 100vh;
          background: linear-gradient(180deg, rgba(0, 0, 0, 0.9) 0%, rgba(0, 0, 20, 0.9) 100%);
          border-left: 1px solid rgba(60, 60, 60, 0.5);
          padding: 5px 2px;
          overflow: hidden;
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: flex-start;
          position: relative;
        }
        
        .generation-counter {
          margin-bottom: 15px;
          flex-shrink: 0;
        }
        
        .generation-hex {
          width: 40px;
          height: 40px;
          background: linear-gradient(135deg, #d4af37 0%, #b8860b 100%);
          clip-path: polygon(30% 0%, 70% 0%, 100% 50%, 70% 100%, 30% 100%, 0% 50%);
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          color: #000;
          font-weight: bold;
          border: 1px solid #8b7355;
          box-shadow: 0 0 8px rgba(212, 175, 55, 0.4);
        }
        
        .gen-text {
          font-size: 8px;
          line-height: 1;
        }
        
        .gen-number {
          font-size: 16px;
          line-height: 1;
        }
        
        .global-parameters {
          flex: 1;
          display: flex;
          flex-direction: column;
          align-items: center;
          gap: 10px;
          width: 100%;
          height: calc(100vh - 100px);
        }
        
        .parameter-track {
          display: flex;
          flex-direction: column;
          align-items: center;
          gap: 3px;
          width: 100%;
          flex: 1;
        }
        
        .track-header {
          display: flex;
          align-items: flex-start;
          justify-content: center;
          gap: 4px;
          flex: 1;
          width: 100%;
        }
        
        .track-icon {
          font-size: 10px;
          width: 12px;
          text-align: center;
          margin-top: 5px;
        }
        
        .track-bar {
          width: 12px;
          flex: 1;
          min-height: 200px;
          background: rgba(40, 40, 40, 0.8);
          border: 1px solid rgba(100, 100, 100, 0.4);
          border-radius: 6px;
          overflow: hidden;
          position: relative;
          display: flex;
          flex-direction: column-reverse;
          padding: 1px;
        }
        
        .track-segment {
          flex: 1;
          background: rgba(60, 60, 60, 0.5);
          margin: 1px 0;
          transition: all 0.2s ease;
        }
        
        .track-segment.temperature.active {
          background: linear-gradient(to top, #ff4500 0%, #ff8c00 100%);
          box-shadow: 0 0 3px rgba(255, 69, 0, 0.8);
        }
        
        .track-segment.oxygen.active {
          background: linear-gradient(to top, #00bfff 0%, #87ceeb 100%);
          box-shadow: 0 0 3px rgba(0, 191, 255, 0.8);
        }
        
        .track-segment.ocean.active {
          background: linear-gradient(to top, #0066cc 0%, #4da6ff 100%);
          box-shadow: 0 0 3px rgba(0, 102, 204, 0.8);
        }
        
        .track-values {
          display: flex;
          flex-direction: column;
          justify-content: space-between;
          flex: 1;
          min-height: 200px;
          margin-left: 2px;
        }
        
        .track-value {
          font-size: 6px;
          color: #ccc;
          font-weight: bold;
          width: 8px;
          text-align: center;
        }
        
        .current-value {
          font-size: 7px;
          font-weight: bold;
          color: #fff;
          text-align: center;
          margin-top: 2px;
        }
        
        .player-score-section {
          flex-shrink: 0;
          margin-top: 20px;
        }
        
        .score-hex-container {
          display: flex;
          flex-direction: column;
          align-items: center;
          gap: 8px;
        }
        
        .score-hex {
          width: 45px;
          height: 45px;
          background: linear-gradient(135deg, #4a90e2 0%, #357abd 100%);
          clip-path: polygon(30% 0%, 70% 0%, 100% 50%, 70% 100%, 30% 100%, 0% 50%);
          display: flex;
          align-items: center;
          justify-content: center;
          color: #fff;
          font-weight: bold;
          border: 2px solid #2c5aa0;
          box-shadow: 0 0 8px rgba(74, 144, 226, 0.4);
        }
        
        .score-value {
          font-size: 14px;
          line-height: 1;
        }
        
        .player-name {
          font-size: 8px;
          color: #4a90e2;
          text-align: center;
          max-width: 80px;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
        }
      `}</style>
    </div>
  );
};

export default RightSidebar;