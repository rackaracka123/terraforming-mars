import React, { useEffect } from 'react';
import { CardType } from '../../../types/cards.ts';

interface GameAction {
  id: string;
  name: string;
  type: 'standard' | 'card' | 'milestone' | 'award';
  cost?: number;
  description: string;
  requirement?: string;
  available: boolean;
  source?: string; // Card name or 'Standard Project' etc.
}

interface AvailableActionsModalProps {
  isVisible: boolean;
  onClose: () => void;
  actions: GameAction[];
  playerName?: string;
  onActionSelect?: (action: GameAction) => void;
}

const AvailableActionsModal: React.FC<AvailableActionsModalProps> = ({ 
  isVisible, 
  onClose, 
  actions, 
  playerName = "Player",
  onActionSelect
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

  const getActionTypeStyle = (type: string, available: boolean) => {
    const baseOpacity = available ? 1 : 0.4;
    const styles = {
      standard: {
        background: `linear-gradient(145deg, rgba(0, 150, 255, ${0.15 * baseOpacity}) 0%, rgba(0, 100, 200, ${0.25 * baseOpacity}) 100%)`,
        borderColor: `rgba(0, 180, 255, ${0.6 * baseOpacity})`,
        glowColor: `rgba(0, 180, 255, ${0.3 * baseOpacity})`
      },
      card: {
        background: `linear-gradient(145deg, rgba(255, 150, 0, ${0.15 * baseOpacity}) 0%, rgba(200, 100, 0, ${0.25 * baseOpacity}) 100%)`,
        borderColor: `rgba(255, 180, 0, ${0.6 * baseOpacity})`,
        glowColor: `rgba(255, 180, 0, ${0.3 * baseOpacity})`
      },
      milestone: {
        background: `linear-gradient(145deg, rgba(0, 200, 100, ${0.15 * baseOpacity}) 0%, rgba(0, 150, 80, ${0.25 * baseOpacity}) 100%)`,
        borderColor: `rgba(0, 255, 120, ${0.6 * baseOpacity})`,
        glowColor: `rgba(0, 255, 120, ${0.3 * baseOpacity})`
      },
      award: {
        background: `linear-gradient(145deg, rgba(200, 100, 255, ${0.15 * baseOpacity}) 0%, rgba(150, 50, 200, ${0.25 * baseOpacity}) 100%)`,
        borderColor: `rgba(220, 120, 255, ${0.6 * baseOpacity})`,
        glowColor: `rgba(220, 120, 255, ${0.3 * baseOpacity})`
      }
    };
    return styles[type as keyof typeof styles] || styles.standard;
  };

  const getActionTypeName = (type: string) => {
    const names = {
      standard: 'Standard Project',
      card: 'Card Action',
      milestone: 'Milestone',
      award: 'Award'
    };
    return names[type as keyof typeof names] || 'Action';
  };

  const getActionIcon = (type: string) => {
    const icons = {
      standard: '⚙️',
      card: '🃏',
      milestone: '🎯',
      award: '🏆'
    };
    return icons[type as keyof typeof icons] || '⚡';
  };

  const availableActions = actions.filter(action => action.available);
  const unavailableActions = actions.filter(action => !action.available);

  const handleActionClick = (action: GameAction) => {
    if (action.available && onActionSelect) {
      onActionSelect(action);
      onClose();
    }
  };

  return (
    <div className="actions-modal-overlay">
      {/* Backdrop */}
      <div className="backdrop" onClick={onClose} />
      
      {/* Modal Container */}
      <div className="modal-container">
        {/* Header */}
        <div className="modal-header">
          <h1 className="modal-title">
            Available Actions
          </h1>
          <div className="actions-count">
            {availableActions.length} Available
          </div>
          <button className="close-button" onClick={onClose} aria-label="Close modal">
            ×
          </button>
        </div>

        {/* Actions Content */}
        <div className="actions-container">
          {actions.length === 0 ? (
            <div className="empty-state">
              <div className="empty-icon">⚡</div>
              <h3>No Actions Available</h3>
              <p>Actions will appear here when available</p>
            </div>
          ) : (
            <>
              {/* Available Actions */}
              {availableActions.length > 0 && (
                <div className="actions-section">
                  <h2 className="section-title">Available Actions ({availableActions.length})</h2>
                  <div className="actions-grid">
                    {availableActions.map((action, index) => {
                      const actionStyle = getActionTypeStyle(action.type, true);
                      return (
                        <div 
                          key={action.id} 
                          className="action-card available"
                          style={{
                            background: actionStyle.background,
                            borderColor: actionStyle.borderColor,
                            boxShadow: `0 4px 20px rgba(0, 0, 0, 0.3), 0 0 30px ${actionStyle.glowColor}`,
                            animationDelay: `${index * 0.1}s`
                          }}
                          onClick={() => handleActionClick(action)}
                        >
                          {/* Action Type Badge */}
                          <div className="action-type-badge">
                            <span className="action-icon">{getActionIcon(action.type)}</span>
                            {getActionTypeName(action.type)}
                          </div>

                          {/* Action Cost */}
                          {action.cost !== undefined && (
                            <div className="action-cost">
                              {action.cost}
                            </div>
                          )}

                          {/* Action Content */}
                          <div className="action-content">
                            <h3 className="action-name">{action.name}</h3>
                            {action.source && (
                              <div className="action-source">Source: {action.source}</div>
                            )}
                            <p className="action-description">{action.description}</p>
                            {action.requirement && (
                              <div className="action-requirement">
                                <strong>Requirement:</strong> {action.requirement}
                              </div>
                            )}
                          </div>
                        </div>
                      );
                    })}
                  </div>
                </div>
              )}

              {/* Unavailable Actions */}
              {unavailableActions.length > 0 && (
                <div className="actions-section">
                  <h2 className="section-title">Unavailable Actions ({unavailableActions.length})</h2>
                  <div className="actions-grid">
                    {unavailableActions.map((action, index) => {
                      const actionStyle = getActionTypeStyle(action.type, false);
                      return (
                        <div 
                          key={action.id} 
                          className="action-card unavailable"
                          style={{
                            background: actionStyle.background,
                            borderColor: actionStyle.borderColor,
                            boxShadow: `0 2px 10px rgba(0, 0, 0, 0.2)`,
                            animationDelay: `${(availableActions.length + index) * 0.1}s`
                          }}
                        >
                          {/* Action Type Badge */}
                          <div className="action-type-badge">
                            <span className="action-icon">{getActionIcon(action.type)}</span>
                            {getActionTypeName(action.type)}
                          </div>

                          {/* Action Cost */}
                          {action.cost !== undefined && (
                            <div className="action-cost unavailable-cost">
                              {action.cost}
                            </div>
                          )}

                          {/* Action Content */}
                          <div className="action-content">
                            <h3 className="action-name">{action.name}</h3>
                            {action.source && (
                              <div className="action-source">Source: {action.source}</div>
                            )}
                            <p className="action-description">{action.description}</p>
                            {action.requirement && (
                              <div className="action-requirement">
                                <strong>Requirement:</strong> {action.requirement}
                              </div>
                            )}
                          </div>
                        </div>
                      );
                    })}
                  </div>
                </div>
              )}
            </>
          )}
        </div>
      </div>

      <style jsx>{`
        .actions-modal-overlay {
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
          max-width: 1000px;
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

        .actions-count {
          color: rgba(255, 255, 255, 0.8);
          font-size: 16px;
          font-weight: 500;
          background: rgba(0, 255, 120, 0.2);
          padding: 8px 16px;
          border-radius: 20px;
          border: 1px solid rgba(0, 255, 120, 0.3);
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

        .actions-container {
          padding: 30px;
          max-height: calc(90vh - 120px);
          overflow-y: auto;
          scrollbar-width: thin;
          scrollbar-color: rgba(100, 150, 255, 0.5) rgba(50, 75, 125, 0.3);
        }

        .actions-container::-webkit-scrollbar {
          width: 8px;
        }

        .actions-container::-webkit-scrollbar-track {
          background: rgba(50, 75, 125, 0.3);
          border-radius: 4px;
        }

        .actions-container::-webkit-scrollbar-thumb {
          background: rgba(100, 150, 255, 0.5);
          border-radius: 4px;
        }

        .actions-container::-webkit-scrollbar-thumb:hover {
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

        .actions-section {
          margin-bottom: 40px;
        }

        .actions-section:last-child {
          margin-bottom: 0;
        }

        .section-title {
          color: #ffffff;
          font-size: 20px;
          font-weight: bold;
          margin: 0 0 20px 0;
          padding-bottom: 10px;
          border-bottom: 2px solid rgba(100, 150, 255, 0.3);
          text-shadow: 1px 1px 2px rgba(0, 0, 0, 0.8);
        }

        .actions-grid {
          display: grid;
          grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
          gap: 20px;
          justify-items: center;
        }

        .action-card {
          width: 100%;
          max-width: 320px;
          min-height: 200px;
          border: 2px solid;
          border-radius: 15px;
          padding: 20px;
          position: relative;
          transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
          backdrop-filter: blur(10px);
          animation: actionSlideIn 0.6s ease-out both;
        }

        .action-card.available {
          cursor: pointer;
        }

        .action-card.available:hover {
          transform: translateY(-8px) scale(1.02);
          box-shadow: 0 12px 40px rgba(0, 0, 0, 0.4), 0 0 50px var(--glow-color, rgba(100, 150, 255, 0.4)) !important;
        }

        .action-card.unavailable {
          opacity: 0.6;
          cursor: not-allowed;
        }

        .action-type-badge {
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
          display: flex;
          align-items: center;
          gap: 6px;
        }

        .action-icon {
          font-size: 14px;
        }

        .action-cost {
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

        .action-cost.unavailable-cost {
          background: linear-gradient(135deg, rgba(120, 120, 120, 0.9) 0%, rgba(80, 80, 80, 1) 100%);
          border-color: rgba(150, 150, 150, 0.6);
          color: rgba(255, 255, 255, 0.8);
          box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
        }

        .action-content {
          margin-top: 35px;
        }

        .action-name {
          color: #ffffff;
          font-size: 18px;
          font-weight: bold;
          margin: 0 0 8px 0;
          text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.8);
          line-height: 1.3;
        }

        .action-source {
          color: rgba(255, 255, 255, 0.7);
          font-size: 12px;
          font-style: italic;
          margin-bottom: 10px;
        }

        .action-description {
          color: rgba(255, 255, 255, 0.9);
          font-size: 14px;
          line-height: 1.5;
          margin: 0 0 10px 0;
          background: rgba(0, 0, 0, 0.3);
          padding: 12px;
          border-radius: 8px;
          border: 1px solid rgba(255, 255, 255, 0.1);
        }

        .action-requirement {
          color: rgba(255, 200, 100, 0.9);
          font-size: 12px;
          line-height: 1.4;
          background: rgba(255, 200, 100, 0.1);
          padding: 8px 12px;
          border-radius: 6px;
          border: 1px solid rgba(255, 200, 100, 0.3);
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

        @keyframes actionSlideIn {
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

          .actions-container {
            padding: 20px;
          }

          .actions-grid {
            grid-template-columns: 1fr;
            gap: 15px;
          }

          .action-card {
            max-width: 100%;
            min-height: 180px;
          }

          .action-name {
            font-size: 16px;
          }

          .action-description {
            font-size: 13px;
            padding: 10px;
          }
        }
      `}</style>
    </div>
  );
};

export default AvailableActionsModal;