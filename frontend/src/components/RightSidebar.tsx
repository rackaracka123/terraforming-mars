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
  const getTemperatureProgress = () => {
    if (!globalParameters) return 0;
    return ((globalParameters.temperature + 30) / 38) * 100;
  };

  const getOxygenProgress = () => {
    if (!globalParameters) return 0;
    return (globalParameters.oxygen / 14) * 100;
  };

  const getOceansProgress = () => {
    if (!globalParameters) return 0;
    return (globalParameters.oceans / 9) * 100;
  };

  return (
    <div className="right-sidebar">
      <div className="sidebar-header">
        <h3>Terra Status</h3>
        <div className="generation">Gen {generation || 1}</div>
      </div>
      
      <div className="parameters-section">
        <div className="parameter-item">
          <div className="parameter-header">
            <span className="parameter-icon">🌡️</span>
            <span className="parameter-name">Temperature</span>
            <span className="parameter-value">{globalParameters?.temperature || -30}°C</span>
          </div>
          <div className="progress-bar">
            <div 
              className="progress-fill temperature"
              style={{ width: `${getTemperatureProgress()}%` }}
            />
          </div>
          <div className="parameter-range">-30°C → +8°C</div>
        </div>

        <div className="parameter-item">
          <div className="parameter-header">
            <span className="parameter-icon">💨</span>
            <span className="parameter-name">Oxygen</span>
            <span className="parameter-value">{globalParameters?.oxygen || 0}%</span>
          </div>
          <div className="progress-bar">
            <div 
              className="progress-fill oxygen"
              style={{ width: `${getOxygenProgress()}%` }}
            />
          </div>
          <div className="parameter-range">0% → 14%</div>
        </div>

        <div className="parameter-item">
          <div className="parameter-header">
            <span className="parameter-icon">🌊</span>
            <span className="parameter-name">Oceans</span>
            <span className="parameter-value">{globalParameters?.oceans || 0}/9</span>
          </div>
          <div className="progress-bar">
            <div 
              className="progress-fill oceans"
              style={{ width: `${getOceansProgress()}%` }}
            />
          </div>
          <div className="parameter-range">0 → 9 tiles</div>
        </div>
      </div>

      <div className="current-player-section">
        {currentPlayer && (
          <>
            <div className="section-header">
              <h4>{currentPlayer.name}</h4>
            </div>
            <div className="player-stats">
              <div className="stat-row">
                <span>Terraform Rating</span>
                <span className="stat-value">{currentPlayer.terraformRating || 20}</span>
              </div>
              <div className="stat-row">
                <span>Victory Points</span>
                <span className="stat-value">{currentPlayer.victoryPoints || 0}</span>
              </div>
            </div>
            
            <div className="production-track">
              <h5>Production</h5>
              <div className="production-grid">
                <div className="production-item">
                  <span className="prod-icon">💰</span>
                  <span>{currentPlayer.production?.credits || 0}</span>
                </div>
                <div className="production-item">
                  <span className="prod-icon">🔩</span>
                  <span>{currentPlayer.production?.steel || 0}</span>
                </div>
                <div className="production-item">
                  <span className="prod-icon">⚡</span>
                  <span>{currentPlayer.production?.energy || 0}</span>
                </div>
                <div className="production-item">
                  <span className="prod-icon">🌿</span>
                  <span>{currentPlayer.production?.plants || 0}</span>
                </div>
                <div className="production-item">
                  <span className="prod-icon">🔥</span>
                  <span>{currentPlayer.production?.heat || 0}</span>
                </div>
              </div>
            </div>
          </>
        )}
      </div>
      
      <style jsx>{`
        .right-sidebar {
          width: 300px;
          background: rgba(0, 0, 0, 0.9);
          border-left: 1px solid #333;
          padding: 20px;
          overflow-y: auto;
          display: flex;
          flex-direction: column;
          gap: 25px;
        }
        
        .sidebar-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          border-bottom: 1px solid #333;
          padding-bottom: 10px;
        }
        
        .sidebar-header h3 {
          margin: 0;
          color: #fff;
          font-size: 18px;
        }
        
        .generation {
          background: #4a90e2;
          color: white;
          padding: 4px 12px;
          border-radius: 12px;
          font-size: 14px;
          font-weight: bold;
        }
        
        .parameters-section {
          display: flex;
          flex-direction: column;
          gap: 20px;
        }
        
        .parameter-item {
          background: rgba(255, 255, 255, 0.05);
          border-radius: 8px;
          padding: 15px;
        }
        
        .parameter-header {
          display: flex;
          align-items: center;
          margin-bottom: 10px;
        }
        
        .parameter-icon {
          font-size: 20px;
          margin-right: 10px;
        }
        
        .parameter-name {
          flex: 1;
          font-weight: bold;
        }
        
        .parameter-value {
          font-weight: bold;
          color: #4a90e2;
        }
        
        .progress-bar {
          width: 100%;
          height: 8px;
          background: #333;
          border-radius: 4px;
          overflow: hidden;
          margin-bottom: 5px;
        }
        
        .progress-fill {
          height: 100%;
          transition: width 0.3s ease;
        }
        
        .progress-fill.temperature {
          background: linear-gradient(90deg, #3498db, #e74c3c);
        }
        
        .progress-fill.oxygen {
          background: linear-gradient(90deg, #95a5a6, #2ecc71);
        }
        
        .progress-fill.oceans {
          background: linear-gradient(90deg, #34495e, #3498db);
        }
        
        .parameter-range {
          font-size: 12px;
          color: #888;
        }
        
        .current-player-section {
          background: rgba(74, 144, 226, 0.1);
          border: 1px solid rgba(74, 144, 226, 0.3);
          border-radius: 8px;
          padding: 15px;
        }
        
        .section-header h4 {
          margin: 0 0 15px 0;
          color: #4a90e2;
          font-size: 16px;
        }
        
        .player-stats {
          display: flex;
          flex-direction: column;
          gap: 8px;
          margin-bottom: 15px;
        }
        
        .stat-row {
          display: flex;
          justify-content: space-between;
          font-size: 14px;
        }
        
        .stat-value {
          font-weight: bold;
          color: #4a90e2;
        }
        
        .production-track h5 {
          margin: 0 0 10px 0;
          font-size: 14px;
          color: #ccc;
        }
        
        .production-grid {
          display: grid;
          grid-template-columns: repeat(3, 1fr);
          gap: 8px;
        }
        
        .production-item {
          display: flex;
          align-items: center;
          gap: 5px;
          font-size: 12px;
          background: rgba(255, 255, 255, 0.05);
          padding: 4px 8px;
          border-radius: 4px;
        }
        
        .prod-icon {
          font-size: 14px;
        }
      `}</style>
    </div>
  );
};

export default RightSidebar;