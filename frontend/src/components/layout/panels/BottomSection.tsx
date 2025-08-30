import React, { useState } from 'react';

interface BottomSectionProps {
  currentPlayer?: any;
  socket?: any;
  gameState?: any;
}

const BottomSection: React.FC<BottomSectionProps> = ({ 
  currentPlayer, 
  socket, 
  gameState 
}) => {
  const [activeTab, setActiveTab] = useState<'hand' | 'played' | 'resources'>('hand');
  
  // Mock data for demonstration
  const mockHand = [
    { id: 1, name: "Power Plant", cost: 11, type: "Building" },
    { id: 2, name: "Mining Rights", cost: 9, type: "Event" },
    { id: 3, name: "Asteroid", cost: 14, type: "Event" },
    { id: 4, name: "Research Outpost", cost: 18, type: "Building" },
    { id: 5, name: "Capital", cost: 26, type: "City" },
    { id: 6, name: "Greenery", cost: 23, type: "Greenery" },
    { id: 7, name: "Heat Trappers", cost: 6, type: "Automated" },
    { id: 8, name: "Solar Wind Power", cost: 11, type: "Power" },
    { id: 9, name: "Domed Crater", cost: 24, type: "City" },
    { id: 10, name: "Ice Cap Melting", cost: 5, type: "Event" }
  ];

  const mockPlayedCards = [
    { id: 101, name: "Inventors' Guild", type: "Automated" },
    { id: 102, name: "Business Network", type: "Effect" },
    { id: 103, name: "Investment Loan", type: "Event" }
  ];

  const resources = currentPlayer?.resources || {
    credits: 42,
    steel: 3,
    titanium: 1,
    plants: 8,
    energy: 2,
    heat: 15
  };

  const getCardTypeColor = (type: string) => {
    const colors = {
      'Building': '#8b4513',
      'Event': '#ff6b35',
      'City': '#f39c12',
      'Greenery': '#27ae60',
      'Automated': '#3498db',
      'Power': '#e74c3c',
      'Effect': '#9b59b6'
    };
    return colors[type as keyof typeof colors] || '#95a5a6';
  };

  const canAffordCard = (cost: number) => {
    return resources.credits >= cost;
  };

  return (
    <div className="bottom-section">
      <div className="section-header">
        <div className="tabs">
          <button 
            className={`tab ${activeTab === 'hand' ? 'active' : ''}`}
            onClick={() => setActiveTab('hand')}
          >
            Hand ({mockHand.length})
          </button>
          <button 
            className={`tab ${activeTab === 'played' ? 'active' : ''}`}
            onClick={() => setActiveTab('played')}
          >
            Played ({mockPlayedCards.length})
          </button>
          <button 
            className={`tab ${activeTab === 'resources' ? 'active' : ''}`}
            onClick={() => setActiveTab('resources')}
          >
            Resources
          </button>
        </div>
        
        <div className="quick-actions">
          <button className="action-btn pass">Pass Turn</button>
          <button className="action-btn draft">Draft Cards</button>
        </div>
      </div>
      
      <div className="section-content">
        {activeTab === 'hand' && (
          <div className="cards-container">
            <div className="cards-scroll">
              {mockHand.map((card) => (
                <div 
                  key={card.id} 
                  className={`card ${canAffordCard(card.cost) ? 'affordable' : 'expensive'}`}
                  onClick={() => console.log('Play card:', card.name)}
                >
                  <div className="card-cost">{card.cost}</div>
                  <div className="card-content">
                    <div 
                      className="card-type-indicator"
                      style={{ backgroundColor: getCardTypeColor(card.type) }}
                    />
                    <div className="card-info">
                      <div className="card-name">{card.name}</div>
                      <div className="card-type">{card.type}</div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {activeTab === 'played' && (
          <div className="cards-container">
            <div className="cards-scroll">
              {mockPlayedCards.map((card) => (
                <div key={card.id} className="card played">
                  <div className="card-content">
                    <div 
                      className="card-type-indicator"
                      style={{ backgroundColor: getCardTypeColor(card.type) }}
                    />
                    <div className="card-info">
                      <div className="card-name">{card.name}</div>
                      <div className="card-type">{card.type}</div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {activeTab === 'resources' && (
          <div className="resources-panel">
            <div className="resources-grid">
              <div className="resource-item">
                <div className="resource-icon">💰</div>
                <div className="resource-info">
                  <div className="resource-name">Credits</div>
                  <div className="resource-amount">{resources.credits} M€</div>
                </div>
                <div className="resource-production">+{currentPlayer?.production?.credits || 0}</div>
              </div>
              
              <div className="resource-item">
                <div className="resource-icon">🔩</div>
                <div className="resource-info">
                  <div className="resource-name">Steel</div>
                  <div className="resource-amount">{resources.steel}</div>
                </div>
                <div className="resource-production">+{currentPlayer?.production?.steel || 0}</div>
              </div>
              
              <div className="resource-item">
                <div className="resource-icon">⚗️</div>
                <div className="resource-info">
                  <div className="resource-name">Titanium</div>
                  <div className="resource-amount">{resources.titanium}</div>
                </div>
                <div className="resource-production">+{currentPlayer?.production?.titanium || 0}</div>
              </div>
              
              <div className="resource-item">
                <div className="resource-icon">🌿</div>
                <div className="resource-info">
                  <div className="resource-name">Plants</div>
                  <div className="resource-amount">{resources.plants}</div>
                </div>
                <div className="resource-production">+{currentPlayer?.production?.plants || 0}</div>
              </div>
              
              <div className="resource-item">
                <div className="resource-icon">⚡</div>
                <div className="resource-info">
                  <div className="resource-name">Energy</div>
                  <div className="resource-amount">{resources.energy}</div>
                </div>
                <div className="resource-production">+{currentPlayer?.production?.energy || 0}</div>
              </div>
              
              <div className="resource-item">
                <div className="resource-icon">🔥</div>
                <div className="resource-info">
                  <div className="resource-name">Heat</div>
                  <div className="resource-amount">{resources.heat}</div>
                </div>
                <div className="resource-production">+{currentPlayer?.production?.heat || 0}</div>
              </div>
            </div>
          </div>
        )}
      </div>
      
      <style jsx>{`
        .bottom-section {
          height: 200px;
          background: rgba(0, 0, 0, 0.95);
          border-top: 1px solid #333;
          display: flex;
          flex-direction: column;
        }
        
        .section-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: 10px 20px;
          border-bottom: 1px solid #333;
          background: rgba(0, 0, 0, 0.8);
        }
        
        .tabs {
          display: flex;
          gap: 5px;
        }
        
        .tab {
          background: none;
          border: none;
          color: #ccc;
          padding: 8px 16px;
          border-radius: 4px;
          cursor: pointer;
          font-size: 14px;
          transition: all 0.2s ease;
        }
        
        .tab:hover {
          background: rgba(255, 255, 255, 0.1);
          color: white;
        }
        
        .tab.active {
          background: #4a90e2;
          color: white;
        }
        
        .quick-actions {
          display: flex;
          gap: 10px;
        }
        
        .action-btn {
          padding: 8px 16px;
          border: none;
          border-radius: 4px;
          cursor: pointer;
          font-size: 14px;
          font-weight: bold;
          transition: all 0.2s ease;
        }
        
        .action-btn.pass {
          background: #e74c3c;
          color: white;
        }
        
        .action-btn.draft {
          background: #27ae60;
          color: white;
        }
        
        .action-btn:hover {
          opacity: 0.8;
        }
        
        .section-content {
          flex: 1;
          overflow: hidden;
        }
        
        .cards-container {
          height: 100%;
          padding: 10px;
        }
        
        .cards-scroll {
          display: flex;
          gap: 10px;
          overflow-x: auto;
          height: 100%;
          padding-bottom: 10px;
        }
        
        .card {
          min-width: 120px;
          width: 120px;
          background: rgba(255, 255, 255, 0.05);
          border-radius: 8px;
          cursor: pointer;
          transition: all 0.2s ease;
          display: flex;
          flex-direction: column;
          position: relative;
          border: 2px solid transparent;
        }
        
        .card:hover {
          background: rgba(255, 255, 255, 0.1);
          transform: translateY(-2px);
        }
        
        .card.affordable {
          border-color: #27ae60;
        }
        
        .card.expensive {
          opacity: 0.6;
        }
        
        .card.played {
          opacity: 0.8;
          border-color: #95a5a6;
        }
        
        .card-cost {
          position: absolute;
          top: 5px;
          right: 5px;
          background: #f39c12;
          color: white;
          border-radius: 50%;
          width: 24px;
          height: 24px;
          display: flex;
          align-items: center;
          justify-content: center;
          font-size: 12px;
          font-weight: bold;
        }
        
        .card-content {
          padding: 10px;
          flex: 1;
          display: flex;
          flex-direction: column;
        }
        
        .card-type-indicator {
          width: 100%;
          height: 4px;
          margin-bottom: 8px;
          border-radius: 2px;
        }
        
        .card-info {
          flex: 1;
        }
        
        .card-name {
          font-size: 12px;
          font-weight: bold;
          margin-bottom: 4px;
          line-height: 1.2;
        }
        
        .card-type {
          font-size: 10px;
          color: #888;
        }
        
        .resources-panel {
          padding: 20px;
          height: 100%;
          overflow-y: auto;
        }
        
        .resources-grid {
          display: grid;
          grid-template-columns: repeat(3, 1fr);
          gap: 15px;
        }
        
        .resource-item {
          display: flex;
          align-items: center;
          background: rgba(255, 255, 255, 0.05);
          padding: 12px;
          border-radius: 8px;
          gap: 12px;
        }
        
        .resource-icon {
          font-size: 24px;
        }
        
        .resource-info {
          flex: 1;
        }
        
        .resource-name {
          font-size: 14px;
          color: #ccc;
          margin-bottom: 2px;
        }
        
        .resource-amount {
          font-size: 18px;
          font-weight: bold;
          color: white;
        }
        
        .resource-production {
          font-size: 14px;
          color: #27ae60;
          font-weight: bold;
        }
      `}</style>
    </div>
  );
};

export default BottomSection;