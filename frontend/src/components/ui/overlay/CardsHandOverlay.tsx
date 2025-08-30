import React, { useState } from 'react';

interface MockCard {
  id: string;
  name: string;
  cost: number;
  type: 'event' | 'automated' | 'active';
  tags: string[];
  description: string;
  playable: boolean;
}

interface CardsHandOverlayProps {}

const CardsHandOverlay: React.FC<CardsHandOverlayProps> = () => {
  const [isExpanded, setIsExpanded] = useState(false);
  const [selectedCard, setSelectedCard] = useState<string | null>(null);

  // Mock cards data
  const mockCards: MockCard[] = [
    {
      id: '1',
      name: 'Power Plant',
      cost: 11,
      type: 'automated',
      tags: ['energy', 'building'],
      description: 'Increase your energy production by 1 step.',
      playable: true
    },
    {
      id: '2',
      name: 'Asteroid',
      cost: 14,
      type: 'event',
      tags: ['space'],
      description: 'Raise temperature 1 step.',
      playable: true
    },
    {
      id: '3',
      name: 'Greenery',
      cost: 23,
      type: 'automated',
      tags: ['plant', 'building'],
      description: 'Place a greenery tile. Increase oxygen 1 step.',
      playable: false
    },
    {
      id: '4',
      name: 'Mining Rights',
      cost: 9,
      type: 'automated',
      tags: ['building'],
      description: 'Place this tile on an area with a steel or titanium bonus.',
      playable: true
    },
    {
      id: '5',
      name: 'Research',
      cost: 11,
      type: 'event',
      tags: ['science'],
      description: 'Draw 2 cards.',
      playable: true
    },
    {
      id: '6',
      name: 'Solar Power',
      cost: 10,
      type: 'automated',
      tags: ['energy', 'building'],
      description: 'Increase your energy production by 1 step.',
      playable: true
    },
    {
      id: '7',
      name: 'Colony',
      cost: 8,
      type: 'automated',
      tags: ['building'],
      description: 'Place a city tile.',
      playable: false
    },
    {
      id: '8',
      name: 'Investment Loan',
      cost: 3,
      type: 'event',
      tags: ['earth'],
      description: 'Decrease your M€ production 1 step and gain 10 M€.',
      playable: true
    },
    {
      id: '9',
      name: 'Carbonate Processing',
      cost: 6,
      type: 'automated',
      tags: ['building'],
      description: 'Increase your energy production 1 step and your heat production 1 step.',
      playable: false
    },
    {
      id: '10',
      name: 'Breathing Filters',
      cost: 11,
      type: 'automated',
      tags: ['science'],
      description: 'Increase your oxygen requirement by 2.',
      playable: true
    },
    {
      id: '11',
      name: 'Underground City',
      cost: 18,
      type: 'automated',
      tags: ['city', 'building'],
      description: 'Decrease your energy production 2 steps and place a city tile.',
      playable: false
    },
    {
      id: '12',
      name: 'Lightning Harvest',
      cost: 8,
      type: 'automated',
      tags: ['energy'],
      description: 'Increase your energy production 1 step and your M€ production 1 step.',
      playable: true
    }
  ];

  const getCardTypeColor = (type: string) => {
    switch (type) {
      case 'event': return '#e74c3c';
      case 'automated': return '#27ae60';
      case 'active': return '#3498db';
      default: return '#95a5a6';
    }
  };

  const playableCards = mockCards.filter(card => card.playable);
  const unplayableCards = mockCards.filter(card => !card.playable);

  return (
    <div className="cards-hand-overlay">
      {/* Hand Toggle Button */}
      <div 
        className={`hand-toggle ${isExpanded ? 'expanded' : ''}`}
        onClick={() => setIsExpanded(!isExpanded)}
      >
        <div className="toggle-icon">🃏</div>
        <div className="cards-count">{mockCards.length}</div>
        <div className="playable-indicator">
          {playableCards.length} playable
        </div>
      </div>

      {/* Expanded Cards View */}
      {isExpanded && (
        <div className="cards-hand-expanded">
          <div className="cards-header">
            <div className="section-title">
              Playable Cards ({playableCards.length})
            </div>
          </div>
          
          <div className="cards-grid playable-cards">
            {playableCards.map((card) => (
              <div 
                key={card.id}
                className={`card-item playable ${selectedCard === card.id ? 'selected' : ''}`}
                onClick={() => setSelectedCard(selectedCard === card.id ? null : card.id)}
                style={{ '--card-color': getCardTypeColor(card.type) } as React.CSSProperties}
              >
                <div className="card-cost">{card.cost}</div>
                <div className="card-content">
                  <div className="card-name">{card.name}</div>
                  <div className="card-tags">
                    {card.tags.slice(0, 2).map((tag, i) => (
                      <span key={i} className="card-tag">{tag}</span>
                    ))}
                  </div>
                </div>
                <div className="card-type-indicator" />
              </div>
            ))}
          </div>

          {unplayableCards.length > 0 && (
            <>
              <div className="cards-header">
                <div className="section-title disabled">
                  Unplayable Cards ({unplayableCards.length})
                </div>
              </div>
              
              <div className="cards-grid unplayable-cards">
                {unplayableCards.map((card) => (
                  <div 
                    key={card.id}
                    className="card-item unplayable"
                    style={{ '--card-color': getCardTypeColor(card.type) } as React.CSSProperties}
                  >
                    <div className="card-cost">{card.cost}</div>
                    <div className="card-content">
                      <div className="card-name">{card.name}</div>
                      <div className="card-tags">
                        {card.tags.slice(0, 2).map((tag, i) => (
                          <span key={i} className="card-tag">{tag}</span>
                        ))}
                      </div>
                    </div>
                    <div className="card-type-indicator" />
                  </div>
                ))}
              </div>
            </>
          )}
        </div>
      )}

      {/* Selected Card Details */}
      {selectedCard && (
        <div className="card-detail-popup">
          {(() => {
            const card = mockCards.find(c => c.id === selectedCard);
            return card ? (
              <div className="card-detail">
                <div className="detail-header">
                  <div className="detail-name">{card.name}</div>
                  <div className="detail-cost">{card.cost} M€</div>
                </div>
                <div className="detail-tags">
                  {card.tags.map((tag, i) => (
                    <span key={i} className="detail-tag">{tag}</span>
                  ))}
                </div>
                <div className="detail-description">{card.description}</div>
                <div className="detail-actions">
                  <button 
                    className="play-card-btn"
                    disabled={!card.playable}
                  >
                    Play Card
                  </button>
                  <button 
                    className="close-detail-btn"
                    onClick={() => setSelectedCard(null)}
                  >
                    Close
                  </button>
                </div>
              </div>
            ) : null;
          })()}
        </div>
      )}

      <style jsx>{`
        .cards-hand-overlay {
          position: fixed;
          bottom: 140px;
          left: 50%;
          transform: translateX(-50%);
          z-index: 150;
          pointer-events: auto;
        }

        .hand-toggle {
          background: linear-gradient(
            135deg,
            rgba(60, 40, 90, 0.9) 0%,
            rgba(40, 20, 70, 0.8) 100%
          );
          border: 2px solid rgba(150, 100, 255, 0.6);
          border-radius: 20px;
          padding: 15px 25px;
          cursor: pointer;
          transition: all 0.3s ease;
          backdrop-filter: blur(10px);
          box-shadow: 
            0 4px 20px rgba(0, 0, 0, 0.4),
            0 0 20px rgba(150, 100, 255, 0.3);
          display: flex;
          align-items: center;
          gap: 12px;
        }

        .hand-toggle:hover {
          transform: translateY(-3px);
          box-shadow: 
            0 8px 30px rgba(0, 0, 0, 0.5),
            0 0 30px rgba(150, 100, 255, 0.4);
        }

        .hand-toggle.expanded {
          border-color: rgba(150, 100, 255, 0.9);
          background: linear-gradient(
            135deg,
            rgba(60, 40, 90, 1) 0%,
            rgba(40, 20, 70, 0.9) 100%
          );
        }

        .toggle-icon {
          font-size: 20px;
          filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.5));
        }

        .cards-count {
          font-size: 18px;
          font-weight: bold;
          color: #ffffff;
          text-shadow: 0 1px 3px rgba(0, 0, 0, 0.8);
        }

        .playable-indicator {
          font-size: 12px;
          color: rgba(150, 255, 150, 0.9);
          font-weight: 500;
        }

        .cards-hand-expanded {
          position: absolute;
          bottom: 100%;
          left: 50%;
          transform: translateX(-50%);
          width: 800px;
          max-width: 90vw;
          max-height: 400px;
          background: linear-gradient(
            135deg,
            rgba(10, 20, 40, 0.98) 0%,
            rgba(20, 30, 50, 0.96) 100%
          );
          border: 2px solid rgba(150, 100, 255, 0.5);
          border-radius: 15px;
          padding: 20px;
          margin-bottom: 15px;
          backdrop-filter: blur(15px);
          box-shadow: 
            0 12px 40px rgba(0, 0, 0, 0.7),
            0 0 30px rgba(150, 100, 255, 0.3);
          overflow-y: auto;
        }

        .cards-header {
          margin-bottom: 15px;
        }

        .section-title {
          font-size: 16px;
          font-weight: bold;
          color: rgba(150, 255, 150, 1);
          text-transform: uppercase;
          letter-spacing: 0.5px;
          border-bottom: 1px solid rgba(150, 255, 150, 0.3);
          padding-bottom: 5px;
        }

        .section-title.disabled {
          color: rgba(255, 150, 150, 0.8);
          border-bottom-color: rgba(255, 150, 150, 0.3);
        }

        .cards-grid {
          display: grid;
          grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
          gap: 12px;
          margin-bottom: 20px;
        }

        .card-item {
          background: linear-gradient(
            135deg,
            rgba(30, 50, 80, 0.6) 0%,
            rgba(20, 40, 70, 0.5) 100%
          );
          border: 2px solid rgba(255, 255, 255, 0.2);
          border-radius: 10px;
          padding: 12px;
          cursor: pointer;
          transition: all 0.3s ease;
          position: relative;
          overflow: hidden;
        }

        .card-item.playable {
          border-color: rgba(150, 255, 150, 0.4);
        }

        .card-item.playable:hover {
          transform: translateY(-2px);
          border-color: rgba(150, 255, 150, 0.8);
          box-shadow: 
            0 6px 20px rgba(0, 0, 0, 0.4),
            0 0 15px rgba(150, 255, 150, 0.4);
        }

        .card-item.selected {
          border-color: rgba(255, 200, 100, 0.9);
          background: linear-gradient(
            135deg,
            rgba(60, 50, 30, 0.8) 0%,
            rgba(40, 35, 20, 0.7) 100%
          );
          box-shadow: 
            0 4px 15px rgba(0, 0, 0, 0.5),
            0 0 20px rgba(255, 200, 100, 0.5);
        }

        .card-item.unplayable {
          opacity: 0.5;
          border-color: rgba(255, 150, 150, 0.3);
          cursor: not-allowed;
        }

        .card-cost {
          position: absolute;
          top: 8px;
          right: 8px;
          background: rgba(241, 196, 15, 0.9);
          color: #000;
          font-size: 12px;
          font-weight: bold;
          padding: 4px 8px;
          border-radius: 6px;
          box-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
        }

        .card-content {
          margin-right: 40px;
        }

        .card-name {
          font-size: 14px;
          font-weight: bold;
          color: #ffffff;
          text-shadow: 0 1px 2px rgba(0, 0, 0, 0.8);
          margin-bottom: 8px;
          line-height: 1.2;
        }

        .card-tags {
          display: flex;
          flex-wrap: wrap;
          gap: 4px;
        }

        .card-tag {
          font-size: 10px;
          background: rgba(100, 150, 200, 0.3);
          color: rgba(100, 150, 200, 1);
          padding: 2px 6px;
          border-radius: 4px;
          text-transform: uppercase;
          font-weight: 500;
        }

        .card-type-indicator {
          position: absolute;
          left: 0;
          top: 0;
          bottom: 0;
          width: 4px;
          background: var(--card-color);
        }

        .card-detail-popup {
          position: fixed;
          top: 50%;
          left: 50%;
          transform: translate(-50%, -50%);
          background: linear-gradient(
            135deg,
            rgba(15, 25, 45, 0.98) 0%,
            rgba(25, 35, 55, 0.96) 100%
          );
          border: 2px solid rgba(255, 200, 100, 0.7);
          border-radius: 15px;
          padding: 25px;
          width: 400px;
          max-width: 90vw;
          backdrop-filter: blur(20px);
          box-shadow: 
            0 20px 60px rgba(0, 0, 0, 0.8),
            0 0 40px rgba(255, 200, 100, 0.3);
          z-index: 300;
        }

        .detail-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 15px;
        }

        .detail-name {
          font-size: 20px;
          font-weight: bold;
          color: #ffffff;
          text-shadow: 0 1px 3px rgba(0, 0, 0, 0.8);
        }

        .detail-cost {
          font-size: 16px;
          font-weight: bold;
          color: #f1c40f;
          background: rgba(241, 196, 15, 0.2);
          padding: 4px 8px;
          border-radius: 6px;
        }

        .detail-tags {
          display: flex;
          flex-wrap: wrap;
          gap: 6px;
          margin-bottom: 15px;
        }

        .detail-tag {
          font-size: 12px;
          background: rgba(100, 150, 200, 0.4);
          color: rgba(150, 200, 255, 1);
          padding: 4px 8px;
          border-radius: 6px;
          text-transform: uppercase;
          font-weight: 500;
        }

        .detail-description {
          font-size: 14px;
          color: rgba(255, 255, 255, 0.9);
          line-height: 1.5;
          margin-bottom: 20px;
          background: rgba(0, 0, 0, 0.2);
          padding: 12px;
          border-radius: 8px;
          border-left: 3px solid rgba(100, 150, 200, 0.5);
        }

        .detail-actions {
          display: flex;
          gap: 10px;
          justify-content: flex-end;
        }

        .play-card-btn,
        .close-detail-btn {
          padding: 8px 16px;
          border: none;
          border-radius: 6px;
          font-weight: bold;
          cursor: pointer;
          transition: all 0.2s ease;
        }

        .play-card-btn {
          background: linear-gradient(135deg, #27ae60 0%, #2ecc71 100%);
          color: white;
        }

        .play-card-btn:disabled {
          background: rgba(100, 100, 100, 0.5);
          color: rgba(255, 255, 255, 0.5);
          cursor: not-allowed;
        }

        .play-card-btn:hover:not(:disabled) {
          background: linear-gradient(135deg, #229954 0%, #27ae60 100%);
          transform: translateY(-1px);
        }

        .close-detail-btn {
          background: rgba(100, 100, 100, 0.6);
          color: white;
        }

        .close-detail-btn:hover {
          background: rgba(120, 120, 120, 0.8);
        }

        @media (max-width: 900px) {
          .cards-hand-expanded {
            width: 95vw;
          }

          .cards-grid {
            grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
            gap: 8px;
          }
        }
      `}</style>
    </div>
  );
};

export default CardsHandOverlay;