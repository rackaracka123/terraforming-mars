import React, { useEffect } from 'react';
import { CardType } from '../../../types/cards.ts';

interface Card {
  id: string;
  name: string;
  type: CardType;
  cost: number;
  description: string;
}

interface PlayedCardsViewModalProps {
  isVisible: boolean;
  onClose: () => void;
  cards: Card[];
  playerName?: string;
}

const PlayedCardsViewModal: React.FC<PlayedCardsViewModalProps> = ({ 
  isVisible, 
  onClose, 
  cards, 
  playerName = "Player" 
}) => {
  // Handle escape key
  useEffect(() => {
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        onClose();
      }
    };

    if (isVisible) {
      document.addEventListener('keydown', handleEscape);
      document.body.style.overflow = 'hidden';
    }

    return () => {
      document.removeEventListener('keydown', handleEscape);
      document.body.style.overflow = 'unset';
    };
  }, [isVisible, onClose]);

  if (!isVisible) return null;

  const getCardTypeStyle = (type: CardType) => {
    const styles = {
      [CardType.CORPORATION]: {
        background: 'linear-gradient(145deg, rgba(0, 200, 100, 0.15) 0%, rgba(0, 150, 80, 0.25) 100%)',
        borderColor: 'rgba(0, 255, 120, 0.6)',
        glowColor: 'rgba(0, 255, 120, 0.3)'
      },
      [CardType.AUTOMATED]: {
        background: 'linear-gradient(145deg, rgba(0, 150, 255, 0.15) 0%, rgba(0, 100, 200, 0.25) 100%)',
        borderColor: 'rgba(0, 180, 255, 0.6)',
        glowColor: 'rgba(0, 180, 255, 0.3)'
      },
      [CardType.ACTIVE]: {
        background: 'linear-gradient(145deg, rgba(255, 150, 0, 0.15) 0%, rgba(200, 100, 0, 0.25) 100%)',
        borderColor: 'rgba(255, 180, 0, 0.6)',
        glowColor: 'rgba(255, 180, 0, 0.3)'
      },
      [CardType.EVENT]: {
        background: 'linear-gradient(145deg, rgba(255, 80, 80, 0.15) 0%, rgba(200, 50, 50, 0.25) 100%)',
        borderColor: 'rgba(255, 120, 120, 0.6)',
        glowColor: 'rgba(255, 120, 120, 0.3)'
      },
      [CardType.PRELUDE]: {
        background: 'linear-gradient(145deg, rgba(200, 100, 255, 0.15) 0%, rgba(150, 50, 200, 0.25) 100%)',
        borderColor: 'rgba(220, 120, 255, 0.6)',
        glowColor: 'rgba(220, 120, 255, 0.3)'
      }
    };
    return styles[type] || styles[CardType.AUTOMATED];
  };

  const getCardTypeName = (type: CardType) => {
    const names = {
      [CardType.CORPORATION]: 'Corporation',
      [CardType.AUTOMATED]: 'Automated',
      [CardType.ACTIVE]: 'Active', 
      [CardType.EVENT]: 'Event',
      [CardType.PRELUDE]: 'Prelude'
    };
    return names[type] || 'Card';
  };

  return (
    <div className="played-cards-modal-overlay">
      {/* Backdrop */}
      <div className="backdrop" onClick={onClose} />
      
      {/* Modal Container */}
      <div className="modal-container">
        {/* Header */}
        <div className="modal-header">
          <h1 className="modal-title">
            {playerName}'s Played Cards
          </h1>
          <div className="cards-count">
            {cards.length} {cards.length === 1 ? 'Card' : 'Cards'}
          </div>
          <button className="close-button" onClick={onClose} aria-label="Close modal">
            ×
          </button>
        </div>

        {/* Cards Grid */}
        <div className="cards-container">
          {cards.length === 0 ? (
            <div className="empty-state">
              <div className="empty-icon">🃏</div>
              <h3>No Cards Played Yet</h3>
              <p>Cards played during the game will appear here</p>
            </div>
          ) : (
            <div className="cards-grid">
              {cards.map((card, index) => {
                const cardStyle = getCardTypeStyle(card.type);
                return (
                  <div 
                    key={card.id} 
                    className="card"
                    style={{
                      background: cardStyle.background,
                      borderColor: cardStyle.borderColor,
                      boxShadow: `0 4px 20px rgba(0, 0, 0, 0.3), 0 0 30px ${cardStyle.glowColor}`,
                      animationDelay: `${index * 0.1}s`
                    }}
                  >
                    {/* Card Type Badge */}
                    <div className="card-type-badge">
                      {getCardTypeName(card.type)}
                    </div>

                    {/* Card Cost */}
                    <div className="card-cost">
                      {card.cost}
                    </div>

                    {/* Card Content */}
                    <div className="card-content">
                      <h3 className="card-name">{card.name}</h3>
                      <p className="card-description">{card.description}</p>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      </div>

      <style jsx>{`
        .played-cards-modal-overlay {
          position: fixed;
          top: 0;
          left: 0;
          right: 0;
          bottom: 0;
          z-index: 3000;
          display: flex;
          align-items: center;
          justify-content: center;
          padding: 20px;
          animation: modalFadeIn 0.3s ease-out;
        }

        .backdrop {
          position: absolute;
          top: 0;
          left: 0;
          right: 0;
          bottom: 0;
          background: rgba(0, 0, 0, 0.85);
          backdrop-filter: blur(10px);
          cursor: pointer;
        }

        .modal-container {
          position: relative;
          width: 100%;
          max-width: 1200px;
          max-height: 90vh;
          background: linear-gradient(145deg, rgba(15, 25, 45, 0.98) 0%, rgba(25, 35, 55, 0.95) 100%);
          border: 3px solid rgba(100, 150, 255, 0.4);
          border-radius: 20px;
          overflow: hidden;
          box-shadow: 0 25px 80px rgba(0, 0, 0, 0.8), 0 0 60px rgba(50, 100, 200, 0.4);
          backdrop-filter: blur(20px);
          animation: modalSlideIn 0.4s ease-out;
        }

        .modal-header {
          display: flex;
          align-items: center;
          justify-content: space-between;
          padding: 25px 30px;
          background: linear-gradient(90deg, rgba(20, 30, 50, 0.9) 0%, rgba(30, 40, 60, 0.7) 100%);
          border-bottom: 2px solid rgba(100, 150, 255, 0.3);
        }

        .modal-title {
          margin: 0;
          color: #ffffff;
          font-size: 28px;
          font-weight: bold;
          text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.8);
        }

        .cards-count {
          color: rgba(255, 255, 255, 0.8);
          font-size: 16px;
          font-weight: 500;
          background: rgba(100, 150, 255, 0.2);
          padding: 8px 16px;
          border-radius: 20px;
          border: 1px solid rgba(100, 150, 255, 0.3);
        }

        .close-button {
          background: linear-gradient(135deg, rgba(255, 80, 80, 0.8) 0%, rgba(200, 40, 40, 0.9) 100%);
          border: 2px solid rgba(255, 120, 120, 0.6);
          border-radius: 50%;
          width: 45px;
          height: 45px;
          color: #ffffff;
          font-size: 24px;
          font-weight: bold;
          cursor: pointer;
          display: flex;
          align-items: center;
          justify-content: center;
          transition: all 0.3s ease;
          box-shadow: 0 4px 15px rgba(0, 0, 0, 0.4);
        }

        .close-button:hover {
          transform: scale(1.1);
          box-shadow: 0 6px 25px rgba(255, 80, 80, 0.5);
        }

        .close-button:active {
          transform: scale(0.95);
        }

        .cards-container {
          padding: 30px;
          max-height: calc(90vh - 120px);
          overflow-y: auto;
          scrollbar-width: thin;
          scrollbar-color: rgba(100, 150, 255, 0.5) rgba(50, 75, 125, 0.3);
        }

        .cards-container::-webkit-scrollbar {
          width: 8px;
        }

        .cards-container::-webkit-scrollbar-track {
          background: rgba(50, 75, 125, 0.3);
          border-radius: 4px;
        }

        .cards-container::-webkit-scrollbar-thumb {
          background: rgba(100, 150, 255, 0.5);
          border-radius: 4px;
        }

        .cards-container::-webkit-scrollbar-thumb:hover {
          background: rgba(100, 150, 255, 0.7);
        }

        .empty-state {
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          padding: 60px 20px;
          text-align: center;
        }

        .empty-icon {
          font-size: 64px;
          margin-bottom: 20px;
          opacity: 0.6;
        }

        .empty-state h3 {
          color: #ffffff;
          font-size: 24px;
          margin-bottom: 10px;
        }

        .empty-state p {
          color: rgba(255, 255, 255, 0.7);
          font-size: 16px;
          margin: 0;
        }

        .cards-grid {
          display: grid;
          grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
          gap: 25px;
          justify-items: center;
        }

        .card {
          width: 100%;
          max-width: 320px;
          min-height: 260px;
          border: 2px solid;
          border-radius: 15px;
          padding: 20px;
          position: relative;
          cursor: pointer;
          transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
          backdrop-filter: blur(10px);
          animation: cardSlideIn 0.6s ease-out both;
        }

        .card:hover {
          transform: translateY(-8px) scale(1.02);
          box-shadow: 0 12px 40px rgba(0, 0, 0, 0.4), 0 0 50px var(--glow-color, rgba(100, 150, 255, 0.4)) !important;
        }

        .card-type-badge {
          position: absolute;
          top: 15px;
          right: 15px;
          background: rgba(0, 0, 0, 0.8);
          color: #ffffff;
          padding: 6px 12px;
          border-radius: 12px;
          font-size: 11px;
          font-weight: bold;
          text-transform: uppercase;
          letter-spacing: 0.5px;
          border: 1px solid rgba(255, 255, 255, 0.2);
        }

        .card-cost {
          position: absolute;
          top: 15px;
          left: 15px;
          width: 40px;
          height: 40px;
          background: linear-gradient(135deg, rgba(255, 215, 0, 0.9) 0%, rgba(255, 165, 0, 1) 100%);
          border: 2px solid rgba(255, 255, 255, 0.8);
          border-radius: 50%;
          display: flex;
          align-items: center;
          justify-content: center;
          color: #000000;
          font-weight: bold;
          font-size: 14px;
          font-family: 'Courier New', monospace;
          box-shadow: 0 4px 12px rgba(255, 165, 0, 0.4);
        }

        .card-content {
          margin-top: 35px;
        }

        .card-name {
          color: #ffffff;
          font-size: 20px;
          font-weight: bold;
          margin: 0 0 15px 0;
          text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.8);
          line-height: 1.3;
        }

        .card-description {
          color: rgba(255, 255, 255, 0.9);
          font-size: 14px;
          line-height: 1.5;
          margin: 0;
          background: rgba(0, 0, 0, 0.3);
          padding: 15px;
          border-radius: 10px;
          border: 1px solid rgba(255, 255, 255, 0.1);
        }

        @keyframes modalFadeIn {
          from {
            opacity: 0;
          }
          to {
            opacity: 1;
          }
        }

        @keyframes modalSlideIn {
          from {
            opacity: 0;
            transform: translateY(-50px) scale(0.9);
          }
          to {
            opacity: 1;
            transform: translateY(0) scale(1);
          }
        }

        @keyframes cardSlideIn {
          from {
            opacity: 0;
            transform: translateY(30px) scale(0.9);
          }
          to {
            opacity: 1;
            transform: translateY(0) scale(1);
          }
        }

        /* Responsive Design */
        @media (max-width: 768px) {
          .modal-container {
            margin: 10px;
            max-width: calc(100vw - 20px);
          }

          .modal-header {
            padding: 20px;
            flex-wrap: wrap;
            gap: 15px;
          }

          .modal-title {
            font-size: 22px;
            flex: 1;
            min-width: 200px;
          }

          .cards-container {
            padding: 20px;
          }

          .cards-grid {
            grid-template-columns: 1fr;
            gap: 20px;
          }

          .card {
            max-width: 100%;
          }
        }

        @media (max-width: 480px) {
          .modal-header {
            padding: 15px;
          }

          .modal-title {
            font-size: 20px;
          }

          .cards-container {
            padding: 15px;
          }

          .card {
            min-height: 220px;
            padding: 15px;
          }

          .card-name {
            font-size: 18px;
          }

          .card-description {
            font-size: 13px;
            padding: 12px;
          }
        }
      `}</style>
    </div>
  );
};

export default PlayedCardsViewModal;