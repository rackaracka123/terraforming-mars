import React, { useEffect } from 'react';
import CostDisplay from '../display/CostDisplay.tsx';
import ProductionDisplay from '../display/ProductionDisplay.tsx';
import { Z_INDEX } from '../../../constants/zIndex.ts';

interface Milestone {
  id: string;
  name: string;
  description: string;
  reward: string;
  cost: number;
  claimed: boolean;
  claimedBy?: string;
  available: boolean;
}

interface StandardProject {
  id: string;
  name: string;
  cost: number;
  description: string;
  available: boolean;
  effects: {
    production?: { type: string; amount: number }[];
    immediate?: { type: string; amount: number }[];
    tiles?: string[];
  };
  icon?: string;
}

interface Award {
  id: string;
  name: string;
  description: string;
  fundingCost: number;
  funded: boolean;
  fundedBy?: string;
  winner?: string;
  available: boolean;
}

interface ModalPopupProps {
  type: 'milestones' | 'projects' | 'awards' | null;
  onClose: () => void;
  onAction?: (actionType: string, itemId: string) => void;
}

const ModalPopup: React.FC<ModalPopupProps> = ({ type, onClose, onAction }) => {
  // Mock data for different modal types
  const mockMilestones: Milestone[] = [
    {
      id: 'terraformer',
      name: 'Terraformer',
      description: 'Have a terraform rating of at least 35',
      reward: '5 VP',
      cost: 8,
      claimed: false,
      available: true
    },
    {
      id: 'mayor',
      name: 'Mayor',
      description: 'Own at least 3 city tiles',
      reward: '5 VP',
      cost: 8,
      claimed: true,
      claimedBy: 'Alice Chen',
      available: false
    },
    {
      id: 'gardener',
      name: 'Gardener',
      description: 'Own at least 3 greenery tiles',
      reward: '5 VP',
      cost: 8,
      claimed: false,
      available: true
    },
    {
      id: 'builder',
      name: 'Builder',
      description: 'Have at least 8 building tags in play',
      reward: '5 VP',
      cost: 8,
      claimed: false,
      available: false
    },
    {
      id: 'planner',
      name: 'Planner',
      description: 'Have at least 16 cards in hand',
      reward: '5 VP',
      cost: 8,
      claimed: false,
      available: false
    }
  ];

  const mockProjects: StandardProject[] = [
    {
      id: 'sell-patents',
      name: 'Sell Patents',
      cost: 0,
      description: 'Discard any number of cards from hand and gain that many M€',
      available: true,
      effects: {
        immediate: [{ type: 'credits', amount: 1 }]
      },
      icon: '/assets/resources/megacredit.png'
    },
    {
      id: 'power-plant',
      name: 'Power Plant',
      cost: 11,
      description: 'Increase your energy production 1 step',
      available: true,
      effects: {
        production: [{ type: 'energy', amount: 1 }]
      },
      icon: '/assets/resources/power.png'
    },
    {
      id: 'asteroid',
      name: 'Asteroid',
      cost: 14,
      description: 'Raise temperature 1 step',
      available: true,
      effects: {
        immediate: [{ type: 'temperature', amount: 1 }]
      },
      icon: '/assets/resources/asteroid.png'
    },
    {
      id: 'aquifer',
      name: 'Aquifer',
      cost: 18,
      description: 'Place an ocean tile',
      available: true,
      effects: {
        tiles: ['ocean']
      },
      icon: '/assets/tiles/ocean.png'
    },
    {
      id: 'greenery',
      name: 'Greenery',
      cost: 23,
      description: 'Place a greenery tile and raise oxygen 1 step',
      available: true,
      effects: {
        tiles: ['greenery'],
        immediate: [{ type: 'oxygen', amount: 1 }]
      },
      icon: '/assets/tiles/greenery.png'
    },
    {
      id: 'city',
      name: 'City',
      cost: 25,
      description: 'Place a city tile',
      available: true,
      effects: {
        tiles: ['city']
      },
      icon: '/assets/tiles/city.png'
    }
  ];

  const mockAwards: Award[] = [
    {
      id: 'landlord',
      name: 'Landlord',
      description: 'Most tiles on Mars',
      fundingCost: 8,
      funded: true,
      fundedBy: 'Bob Martinez',
      winner: 'Alice Chen',
      available: false
    },
    {
      id: 'banker',
      name: 'Banker',
      description: 'Highest M€ production',
      fundingCost: 8,
      funded: false,
      available: true
    },
    {
      id: 'scientist',
      name: 'Scientist',
      description: 'Most science tags',
      fundingCost: 8,
      funded: true,
      fundedBy: 'Carol Kim',
      available: false
    },
    {
      id: 'thermalist',
      name: 'Thermalist',
      description: 'Most heat resource cubes',
      fundingCost: 8,
      funded: false,
      available: true
    },
    {
      id: 'miner',
      name: 'Miner',
      description: 'Most steel and titanium resource cubes',
      fundingCost: 8,
      funded: false,
      available: true
    }
  ];

  // Handle ESC key to close modal
  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        onClose();
      }
    };

    if (type) {
      document.addEventListener('keydown', handleKeyDown);
      return () => document.removeEventListener('keydown', handleKeyDown);
    }
  }, [type, onClose]);

  if (!type) return null;

  const handleAction = (actionType: string, itemId: string) => {
    onAction?.(actionType, itemId);
    // For demo purposes, just log the action
    console.log(`Action: ${actionType} on ${itemId}`);
  };

  const renderMilestones = () => (
    <div className="modal-content">
      <div className="modal-header">
        <h2>Milestones</h2>
        <p>Claim milestones to earn victory points</p>
      </div>
      <div className="items-grid">
        {mockMilestones.map((milestone) => (
          <div 
            key={milestone.id} 
            className={`item-card milestone-card ${milestone.claimed ? 'claimed' : ''} ${!milestone.available ? 'unavailable' : ''}`}
          >
            <div className="item-header">
              <div className="item-name">{milestone.name}</div>
              <div className="item-cost">{milestone.cost} M€</div>
            </div>
            <div className="item-description">{milestone.description}</div>
            <div className="item-reward">Reward: {milestone.reward}</div>
            {milestone.claimed && milestone.claimedBy && (
              <div className="claimed-by">Claimed by {milestone.claimedBy}</div>
            )}
            <div className="item-actions">
              <button
                className="action-btn claim-btn"
                disabled={milestone.claimed || !milestone.available}
                onClick={() => handleAction('claim-milestone', milestone.id)}
              >
                {milestone.claimed ? 'Claimed' : 'Claim'}
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );

  const renderEffects = (effects: StandardProject['effects']) => {
    const elements = [];
    
    if (effects.production) {
      effects.production.forEach((prod, idx) => {
        elements.push(
          <div key={`prod-${idx}`} className="effect-item">
            <ProductionDisplay amount={prod.amount} resourceType={prod.type} size="small" />
          </div>
        );
      });
    }
    
    if (effects.immediate) {
      effects.immediate.forEach((imm, idx) => {
        if (imm.type === 'credits') {
          elements.push(
            <div key={`imm-${idx}`} className="effect-item">
              <CostDisplay cost={imm.amount} size="small" />
              <span className="effect-label">per card</span>
            </div>
          );
        } else if (imm.type === 'temperature') {
          elements.push(
            <div key={`imm-${idx}`} className="effect-item">
              <img src="/assets/resources/heat.png" alt="Temperature" className="effect-icon" />
              <span className="effect-amount">+{imm.amount}°</span>
            </div>
          );
        } else if (imm.type === 'oxygen') {
          elements.push(
            <div key={`imm-${idx}`} className="effect-item">
              <span className="oxygen-symbol">O₂</span>
              <span className="effect-amount">+{imm.amount}%</span>
            </div>
          );
        }
      });
    }
    
    if (effects.tiles) {
      effects.tiles.forEach((tile, idx) => {
        const tileIcons: { [key: string]: string } = {
          ocean: '/assets/tiles/ocean.png',
          greenery: '/assets/tiles/greenery.png',
          city: '/assets/tiles/city.png'
        };
        elements.push(
          <div key={`tile-${idx}`} className="effect-item">
            <img src={tileIcons[tile]} alt={tile} className="effect-tile" />
          </div>
        );
      });
    }
    
    return elements;
  };

  const renderProjects = () => (
    <div className="modal-content">
      <div className="modal-header">
        <h2>Standard Projects</h2>
        <p>Standard actions available every turn</p>
      </div>
      <div className="items-grid">
        {mockProjects.map((project) => (
          <div 
            key={project.id} 
            className={`item-card project-card ${!project.available ? 'unavailable' : ''}`}
          >
            <div className="project-header">
              <div className="project-icon-name">
                {project.icon && (
                  <img src={project.icon} alt={project.name} className="project-icon" />
                )}
                <div className="item-name">{project.name}</div>
              </div>
              <CostDisplay cost={project.cost} size="medium" />
            </div>
            <div className="item-description">{project.description}</div>
            <div className="project-effects">
              {renderEffects(project.effects)}
            </div>
            <div className="item-actions">
              <button
                className="action-btn play-btn"
                disabled={!project.available}
                onClick={() => handleAction('play-project', project.id)}
              >
                Play
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );

  const renderAwards = () => (
    <div className="modal-content">
      <div className="modal-header">
        <h2>Awards</h2>
        <p>Fund awards and compete for victory points</p>
      </div>
      <div className="items-grid">
        {mockAwards.map((award) => (
          <div 
            key={award.id} 
            className={`item-card award-card ${award.funded ? 'funded' : ''} ${!award.available ? 'unavailable' : ''}`}
          >
            <div className="item-header">
              <div className="item-name">{award.name}</div>
              <div className="item-cost">{award.fundingCost} M€</div>
            </div>
            <div className="item-description">{award.description}</div>
            <div className="award-info">
              <div className="award-rewards">1st place: 5 VP, 2nd place: 2 VP</div>
              {award.funded && award.fundedBy && (
                <div className="funded-by">Funded by {award.fundedBy}</div>
              )}
              {award.winner && (
                <div className="current-winner">Leading: {award.winner}</div>
              )}
            </div>
            <div className="item-actions">
              <button
                className="action-btn fund-btn"
                disabled={award.funded || !award.available}
                onClick={() => handleAction('fund-award', award.id)}
              >
                {award.funded ? 'Funded' : 'Fund'}
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  );

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-popup" onClick={(e) => e.stopPropagation()}>
        <button className="close-btn" onClick={onClose}>×</button>
        
        {type === 'milestones' && renderMilestones()}
        {type === 'projects' && renderProjects()}
        {type === 'awards' && renderAwards()}

        <style jsx>{`
          .modal-overlay {
            position: fixed;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            background: rgba(0, 0, 0, 0.8);
            backdrop-filter: blur(5px);
            display: flex;
            align-items: center;
            justify-content: center;
            z-index: 2000;
            padding: 20px;
          }

          .modal-popup {
            background: linear-gradient(
              135deg,
              rgba(10, 20, 40, 0.98) 0%,
              rgba(20, 30, 50, 0.96) 50%,
              rgba(15, 25, 45, 0.98) 100%
            );
            border: 2px solid rgba(100, 150, 255, 0.5);
            border-radius: 20px;
            max-width: 1000px;
            max-height: 80vh;
            width: 100%;
            overflow-y: auto;
            backdrop-filter: blur(20px);
            box-shadow: 
              0 20px 60px rgba(0, 0, 0, 0.8),
              0 0 40px rgba(100, 150, 255, 0.3);
            position: relative;
          }

          .close-btn {
            position: absolute;
            top: 15px;
            right: 20px;
            background: none;
            border: none;
            font-size: 24px;
            color: rgba(255, 255, 255, 0.7);
            cursor: pointer;
            width: 30px;
            height: 30px;
            display: flex;
            align-items: center;
            justify-content: center;
            border-radius: 50%;
            transition: all 0.2s ease;
            z-index: 10;
          }

          .close-btn:hover {
            background: rgba(255, 255, 255, 0.1);
            color: rgba(255, 255, 255, 1);
          }

          .modal-content {
            padding: 30px;
          }

          .modal-header {
            text-align: center;
            margin-bottom: 30px;
            border-bottom: 1px solid rgba(100, 150, 255, 0.3);
            padding-bottom: 20px;
          }

          .modal-header h2 {
            font-size: 28px;
            color: #ffffff;
            margin-bottom: 8px;
            text-shadow: 0 2px 4px rgba(0, 0, 0, 0.8);
          }

          .modal-header p {
            font-size: 14px;
            color: rgba(255, 255, 255, 0.7);
            margin: 0;
          }

          .items-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
            gap: 20px;
          }

          .item-card {
            background: linear-gradient(
              135deg,
              rgba(30, 50, 80, 0.6) 0%,
              rgba(20, 40, 70, 0.5) 100%
            );
            border: 2px solid rgba(255, 255, 255, 0.2);
            border-radius: 12px;
            padding: 20px;
            transition: all 0.3s ease;
            position: relative;
          }

          .milestone-card {
            border-left-color: #ff6b35;
          }

          .project-card {
            border-left-color: #4a90e2;
          }

          .award-card {
            border-left-color: #f39c12;
          }

          .item-card:hover:not(.unavailable) {
            transform: translateY(-2px);
            box-shadow: 
              0 8px 25px rgba(0, 0, 0, 0.4),
              0 0 20px rgba(100, 150, 255, 0.3);
          }

          .item-card.claimed,
          .item-card.funded {
            border-color: rgba(150, 255, 150, 0.5);
            background: linear-gradient(
              135deg,
              rgba(30, 60, 30, 0.6) 0%,
              rgba(20, 50, 20, 0.5) 100%
            );
          }

          .item-card.unavailable {
            opacity: 0.5;
            border-color: rgba(255, 150, 150, 0.3);
          }

          .item-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 15px;
          }
          
          .project-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 15px;
          }
          
          .project-icon-name {
            display: flex;
            align-items: center;
            gap: 10px;
          }
          
          .project-icon {
            width: 24px;
            height: 24px;
            filter: drop-shadow(0 1px 3px rgba(0, 0, 0, 0.8));
          }
          
          .project-effects {
            display: flex;
            flex-wrap: wrap;
            gap: 10px;
            margin: 15px 0;
            padding: 10px;
            background: rgba(0, 0, 0, 0.2);
            border-radius: 8px;
            border: 1px solid rgba(255, 255, 255, 0.1);
          }
          
          .effect-item {
            display: flex;
            align-items: center;
            gap: 4px;
          }
          
          .effect-icon {
            width: 18px;
            height: 18px;
          }
          
          .effect-tile {
            width: 20px;
            height: 20px;
            border-radius: 3px;
          }
          
          .effect-amount {
            color: #ffffff;
            font-size: 12px;
            font-weight: bold;
            text-shadow: 1px 1px 2px rgba(0, 0, 0, 0.8);
          }
          
          .effect-label {
            color: rgba(255, 255, 255, 0.7);
            font-size: 10px;
            font-style: italic;
          }
          
          .oxygen-symbol {
            color: #7dd3fc;
            font-size: 14px;
            font-weight: bold;
            text-shadow: 1px 1px 2px rgba(0, 0, 0, 0.8);
          }

          .item-name {
            font-size: 18px;
            font-weight: bold;
            color: #ffffff;
            text-shadow: 0 1px 3px rgba(0, 0, 0, 0.8);
          }

          .item-cost {
            font-size: 16px;
            font-weight: bold;
            color: #f1c40f;
            background: rgba(241, 196, 15, 0.2);
            padding: 4px 8px;
            border-radius: 6px;
          }

          .item-description {
            font-size: 14px;
            color: rgba(255, 255, 255, 0.9);
            line-height: 1.5;
            margin-bottom: 12px;
          }

          .item-reward,
          .item-requirements {
            font-size: 12px;
            color: rgba(150, 255, 150, 0.9);
            margin-bottom: 8px;
            font-weight: 500;
          }

          .item-requirements {
            color: rgba(255, 200, 100, 0.9);
          }

          .award-info {
            margin-bottom: 12px;
          }

          .award-rewards {
            font-size: 12px;
            color: rgba(150, 255, 150, 0.9);
            margin-bottom: 4px;
          }

          .claimed-by,
          .funded-by,
          .current-winner {
            font-size: 11px;
            color: rgba(100, 200, 255, 0.8);
            font-style: italic;
          }

          .item-actions {
            display: flex;
            justify-content: flex-end;
            margin-top: 15px;
          }

          .action-btn {
            padding: 8px 16px;
            border: none;
            border-radius: 6px;
            font-weight: bold;
            cursor: pointer;
            transition: all 0.2s ease;
            font-size: 14px;
          }

          .claim-btn {
            background: linear-gradient(135deg, #ff6b35 0%, #ff8c42 100%);
            color: white;
          }

          .claim-btn:hover:not(:disabled) {
            background: linear-gradient(135deg, #e55a2b 0%, #ff6b35 100%);
            transform: translateY(-1px);
          }

          .play-btn {
            background: linear-gradient(135deg, #4a90e2 0%, #5ba0f2 100%);
            color: white;
          }

          .play-btn:hover:not(:disabled) {
            background: linear-gradient(135deg, #357abd 0%, #4a90e2 100%);
            transform: translateY(-1px);
          }

          .fund-btn {
            background: linear-gradient(135deg, #f39c12 0%, #f1c40f 100%);
            color: white;
          }

          .fund-btn:hover:not(:disabled) {
            background: linear-gradient(135deg, #d68910 0%, #f39c12 100%);
            transform: translateY(-1px);
          }

          .action-btn:disabled {
            background: rgba(100, 100, 100, 0.5) !important;
            color: rgba(255, 255, 255, 0.5);
            cursor: not-allowed;
            transform: none !important;
          }

          @media (max-width: 800px) {
            .modal-popup {
              margin: 10px;
              max-height: 90vh;
            }

            .modal-content {
              padding: 20px;
            }

            .items-grid {
              grid-template-columns: 1fr;
              gap: 15px;
            }

            .modal-header h2 {
              font-size: 24px;
            }
          }
        `}</style>
      </div>
    </div>
  );
};

export default ModalPopup;