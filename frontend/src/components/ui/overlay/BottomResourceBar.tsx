import React, { useState } from 'react';

interface ResourceData {
  id: string;
  name: string;
  current: number;
  production: number;
  icon: string;
  color: string;
}

interface BottomResourceBarProps {}

const BottomResourceBar: React.FC<BottomResourceBarProps> = () => {
  const [cardsExpanded, setCardsExpanded] = useState(false);

  // Mock resource data
  const mockResources: ResourceData[] = [
    { id: 'credits', name: 'Credits', current: 45, production: 12, icon: '💰', color: '#f1c40f' },
    { id: 'steel', name: 'Steel', current: 8, production: 3, icon: '🔩', color: '#95a5a6' },
    { id: 'titanium', name: 'Titanium', current: 4, production: 1, icon: '⚙️', color: '#e74c3c' },
    { id: 'plants', name: 'Plants', current: 12, production: 5, icon: '🌱', color: '#27ae60' },
    { id: 'energy', name: 'Energy', current: 6, production: 2, icon: '⚡', color: '#3498db' },
    { id: 'heat', name: 'Heat', current: 9, production: 1, icon: '🔥', color: '#e67e22' }
  ];

  const mockCardCount = 12;

  return (
    <div className="bottom-resource-bar">
      {/* Resource Grid */}
      <div className="resources-section">
        <div className="resources-grid">
          {mockResources.map((resource) => (
            <div 
              key={resource.id}
              className="resource-item"
              style={{ '--resource-color': resource.color } as React.CSSProperties}
            >
              <div className="resource-icon">{resource.icon}</div>
              <div className="resource-info">
                <div className="resource-current">{resource.current}</div>
                <div className="resource-production">+{resource.production}</div>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Cards Section */}
      <div className="cards-section">
        <div 
          className={`cards-indicator ${cardsExpanded ? 'expanded' : ''}`}
          onClick={() => setCardsExpanded(!cardsExpanded)}
        >
          <div className="cards-icon">🃏</div>
          <div className="cards-count">{mockCardCount}</div>
        </div>
        
        {cardsExpanded && (
          <div className="cards-preview">
            <div className="cards-grid">
              {/* Mock cards preview */}
              {Array.from({ length: 6 }, (_, i) => (
                <div key={i} className="card-thumbnail">
                  <div className="card-cost">{Math.floor(Math.random() * 20) + 5}</div>
                  <div className="card-name">Card {i + 1}</div>
                </div>
              ))}
              <div className="more-cards">+{mockCardCount - 6} more</div>
            </div>
          </div>
        )}
      </div>

      {/* Game Info Section */}
      <div className="game-info-section">
        <div className="turn-phase">
          <div className="phase-label">ACTION PHASE</div>
          <div className="actions-left">2 actions left</div>
        </div>
      </div>

      <style jsx>{`
        .bottom-resource-bar {
          position: fixed;
          bottom: 0;
          left: 0;
          right: 0;
          height: 120px;
          background: linear-gradient(
            180deg,
            rgba(5, 15, 35, 0.95) 0%,
            rgba(10, 25, 45, 0.98) 50%,
            rgba(5, 20, 40, 0.99) 100%
          );
          backdrop-filter: blur(15px);
          border-top: 2px solid rgba(100, 150, 255, 0.3);
          display: flex;
          align-items: center;
          justify-content: space-between;
          padding: 15px 30px;
          z-index: 200;
          box-shadow: 
            0 -8px 32px rgba(0, 0, 0, 0.6),
            0 0 20px rgba(100, 150, 255, 0.2);
        }

        .bottom-resource-bar::before {
          content: '';
          position: absolute;
          top: 0;
          left: 0;
          right: 0;
          bottom: 0;
          background: linear-gradient(
            45deg,
            rgba(150, 200, 255, 0.05) 0%,
            transparent 50%,
            rgba(100, 150, 255, 0.03) 100%
          );
          pointer-events: none;
        }

        .resources-section {
          flex: 2;
        }

        .resources-grid {
          display: grid;
          grid-template-columns: repeat(6, 1fr);
          gap: 15px;
          max-width: 500px;
        }

        .resource-item {
          display: flex;
          flex-direction: column;
          align-items: center;
          gap: 8px;
          background: linear-gradient(
            135deg,
            rgba(30, 60, 90, 0.4) 0%,
            rgba(20, 40, 70, 0.3) 100%
          );
          border: 2px solid var(--resource-color);
          border-radius: 12px;
          padding: 12px 8px;
          transition: all 0.3s ease;
          cursor: pointer;
          position: relative;
          overflow: hidden;
        }

        .resource-item::before {
          content: '';
          position: absolute;
          top: 0;
          left: 0;
          right: 0;
          bottom: 0;
          background: var(--resource-color);
          opacity: 0.1;
          transition: opacity 0.3s ease;
        }

        .resource-item:hover::before {
          opacity: 0.2;
        }

        .resource-item:hover {
          transform: translateY(-2px);
          box-shadow: 
            0 6px 20px rgba(0, 0, 0, 0.4),
            0 0 15px var(--resource-color);
        }

        .resource-icon {
          font-size: 18px;
          filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.5));
        }

        .resource-info {
          display: flex;
          flex-direction: column;
          align-items: center;
          gap: 2px;
        }

        .resource-current {
          font-size: 16px;
          font-weight: bold;
          color: #ffffff;
          text-shadow: 0 1px 3px rgba(0, 0, 0, 0.8);
        }

        .resource-production {
          font-size: 12px;
          font-weight: bold;
          color: var(--resource-color);
          text-shadow: 0 1px 2px rgba(0, 0, 0, 0.6);
        }

        .cards-section {
          flex: 1;
          display: flex;
          flex-direction: column;
          align-items: center;
          gap: 15px;
          position: relative;
        }

        .cards-indicator {
          display: flex;
          flex-direction: column;
          align-items: center;
          gap: 8px;
          background: linear-gradient(
            135deg,
            rgba(60, 40, 90, 0.6) 0%,
            rgba(40, 20, 70, 0.5) 100%
          );
          border: 2px solid rgba(150, 100, 255, 0.6);
          border-radius: 15px;
          padding: 15px 20px;
          cursor: pointer;
          transition: all 0.3s ease;
          box-shadow: 
            0 4px 15px rgba(0, 0, 0, 0.3),
            0 0 15px rgba(150, 100, 255, 0.3);
        }

        .cards-indicator:hover {
          transform: translateY(-3px);
          box-shadow: 
            0 8px 25px rgba(0, 0, 0, 0.4),
            0 0 25px rgba(150, 100, 255, 0.4);
        }

        .cards-indicator.expanded {
          border-color: rgba(150, 100, 255, 0.9);
          background: linear-gradient(
            135deg,
            rgba(60, 40, 90, 0.8) 0%,
            rgba(40, 20, 70, 0.7) 100%
          );
        }

        .cards-icon {
          font-size: 24px;
          filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.5));
        }

        .cards-count {
          font-size: 20px;
          font-weight: bold;
          color: #ffffff;
          text-shadow: 0 1px 3px rgba(0, 0, 0, 0.8);
        }

        .cards-preview {
          position: absolute;
          bottom: 100%;
          left: 50%;
          transform: translateX(-50%);
          background: linear-gradient(
            135deg,
            rgba(10, 20, 40, 0.95) 0%,
            rgba(20, 30, 50, 0.95) 100%
          );
          border: 2px solid rgba(150, 100, 255, 0.5);
          border-radius: 12px;
          padding: 15px;
          margin-bottom: 10px;
          backdrop-filter: blur(10px);
          box-shadow: 
            0 8px 25px rgba(0, 0, 0, 0.6),
            0 0 20px rgba(150, 100, 255, 0.3);
          z-index: 300;
        }

        .cards-grid {
          display: grid;
          grid-template-columns: repeat(3, 1fr);
          gap: 10px;
          width: 300px;
        }

        .card-thumbnail {
          background: rgba(255, 255, 255, 0.1);
          border-radius: 6px;
          padding: 8px;
          text-align: center;
          border: 1px solid rgba(255, 255, 255, 0.2);
          cursor: pointer;
          transition: all 0.2s ease;
        }

        .card-thumbnail:hover {
          background: rgba(255, 255, 255, 0.2);
          transform: scale(1.05);
        }

        .card-cost {
          font-size: 12px;
          color: #f1c40f;
          font-weight: bold;
        }

        .card-name {
          font-size: 10px;
          color: #ffffff;
          margin-top: 4px;
        }

        .more-cards {
          grid-column: 1 / -1;
          text-align: center;
          font-size: 12px;
          color: rgba(255, 255, 255, 0.7);
          padding: 8px;
          border-top: 1px solid rgba(255, 255, 255, 0.2);
          margin-top: 5px;
        }

        .game-info-section {
          flex: 1;
          display: flex;
          flex-direction: column;
          align-items: flex-end;
          gap: 8px;
        }

        .turn-phase {
          background: linear-gradient(
            135deg,
            rgba(80, 60, 20, 0.6) 0%,
            rgba(60, 40, 10, 0.5) 100%
          );
          border: 2px solid rgba(255, 200, 100, 0.6);
          border-radius: 10px;
          padding: 10px 15px;
          text-align: center;
          box-shadow: 
            0 4px 15px rgba(0, 0, 0, 0.3),
            0 0 15px rgba(255, 200, 100, 0.2);
        }

        .phase-label {
          font-size: 12px;
          font-weight: bold;
          color: rgba(255, 200, 100, 1);
          text-transform: uppercase;
          letter-spacing: 0.5px;
        }

        .actions-left {
          font-size: 14px;
          color: #ffffff;
          margin-top: 4px;
        }

        @media (max-width: 1200px) {
          .bottom-resource-bar {
            height: 100px;
            padding: 10px 20px;
          }

          .resources-grid {
            gap: 10px;
            max-width: 400px;
          }

          .resource-item {
            padding: 8px 6px;
          }

          .resource-icon {
            font-size: 16px;
          }

          .resource-current {
            font-size: 14px;
          }
        }

        @media (max-width: 800px) {
          .bottom-resource-bar {
            flex-direction: column;
            height: auto;
            padding: 15px;
            gap: 15px;
          }

          .resources-grid {
            grid-template-columns: repeat(3, 1fr);
            max-width: none;
            width: 100%;
          }

          .cards-section,
          .game-info-section {
            width: 100%;
            align-items: center;
          }
        }
      `}</style>
    </div>
  );
};

export default BottomResourceBar;