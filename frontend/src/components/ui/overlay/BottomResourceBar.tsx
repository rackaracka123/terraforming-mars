import React, { useState } from 'react';
import { CardType, CardTag } from '../../../types/cards.ts';
// Modal components are now imported and managed in GameInterface

interface ResourceData {
  id: string;
  name: string;
  current: number;
  production: number;
  icon: string;
  color: string;
}

interface BottomResourceBarProps {
  currentPlayer?: {
    id: string;
    name: string;
    victoryPoints: number;
    availableActions: number;
    playedCards: Array<{ type: CardType }> | any[];
    tags: any[];
  } | null;
  onOpenCardEffectsModal?: () => void;
  onOpenActionsModal?: () => void;
  onOpenCardsPlayedModal?: () => void;
  onOpenTagsModal?: () => void;
  onOpenVictoryPointsModal?: () => void;
}

const BottomResourceBar: React.FC<BottomResourceBarProps> = ({ 
  currentPlayer,
  onOpenCardEffectsModal,
  onOpenActionsModal,
  onOpenCardsPlayedModal,
  onOpenTagsModal,
  onOpenVictoryPointsModal
}) => {
  const [cardsExpanded, setCardsExpanded] = useState(false);

  // Helper function to create image with embedded number
  const createImageWithNumber = (imageSrc: string, number: number, className: string = '') => {
    return (
      <div className={`image-with-number ${className}`}>
        <img src={imageSrc} alt="" className="base-image" />
        <span className="embedded-number">{number}</span>
      </div>
    );
  };

  // Mock resource data with dedicated asset paths
  const mockResources: ResourceData[] = [
    { id: 'credits', name: 'Credits', current: 45, production: 12, icon: '/assets/resources/megacredit.png', color: '#f1c40f' },
    { id: 'steel', name: 'Steel', current: 8, production: 3, icon: '/assets/resources/steel.png', color: '#95a5a6' },
    { id: 'titanium', name: 'Titanium', current: 4, production: 1, icon: '/assets/resources/titanium.png', color: '#e74c3c' },
    { id: 'plants', name: 'Plants', current: 12, production: 5, icon: '/assets/resources/plant.png', color: '#27ae60' },
    { id: 'energy', name: 'Energy', current: 6, production: 2, icon: '/assets/resources/power.png', color: '#3498db' },
    { id: 'heat', name: 'Heat', current: 9, production: 1, icon: '/assets/resources/heat.png', color: '#e67e22' }
  ];

  // Resource click handlers
  const handleResourceClick = (resource: ResourceData) => {
    console.log(`Clicked on ${resource.name}: ${resource.current} (${resource.production} production)`);
    alert(`Clicked on ${resource.name}: ${resource.current} (${resource.production} production)`);
    
    // Special handling for different resources
    switch (resource.id) {
      case 'plants':
        if (resource.current >= 8) {
          console.log('Can convert plants to greenery tile');
          alert('Can convert plants to greenery tile!');
        }
        break;
      case 'heat':
        if (resource.current >= 8) {
          console.log('Can convert heat to raise temperature');
          alert('Can convert heat to raise temperature!');
        }
        break;
      case 'energy':
        console.log('Energy converts to heat at end of turn');
        alert('Energy converts to heat at end of turn');
        break;
      default:
        console.log(`${resource.name} resource info displayed`);
    }
  };

  const mockCardCount = 12;

  // Add mock played cards for demonstration
  const mockPlayedCards = [
    {
      id: 'mining-guild',
      name: 'Mining Guild',
      type: CardType.CORPORATION,
      cost: 0,
      description: 'You start with 30 M€, 5 steel production, and 1 steel. Each steel and titanium resource on the board gives you 1 M€ production.',
      tags: [CardTag.BUILDING, CardTag.SCIENCE],
      victoryPoints: 0,
      playOrder: 1
    },
    {
      id: 'power-plant',
      name: 'Power Plant',
      type: CardType.AUTOMATED,
      cost: 4,
      description: 'Increase your energy production 1 step.',
      tags: [CardTag.POWER, CardTag.BUILDING],
      victoryPoints: 1,
      playOrder: 2
    },
    {
      id: 'research',
      name: 'Research',
      type: CardType.ACTIVE,
      cost: 11,
      description: 'Action: Spend 1 M€ to draw a card.',
      tags: [CardTag.SCIENCE],
      victoryPoints: 1,
      playOrder: 3
    },
    {
      id: 'asteroid',
      name: 'Asteroid',
      type: CardType.EVENT,
      cost: 14,
      description: 'Raise temperature 1 step and gain 2 titanium. Remove up to 2 plants from any player.',
      tags: [CardTag.SPACE, CardTag.EVENT],
      victoryPoints: 0,
      playOrder: 4
    },
    {
      id: 'mining-rights',
      name: 'Mining Rights',
      type: CardType.AUTOMATED,
      cost: 9,
      description: 'Place a tile on an area with a steel or titanium bonus, adjacent to one of your tiles. Increase steel or titanium production 1 step.',
      tags: [CardTag.BUILDING, CardTag.MARS],
      victoryPoints: 1,
      playOrder: 5
    },
    {
      id: 'solar-wind-power',
      name: 'Solar Wind Power',
      type: CardType.AUTOMATED,
      cost: 11,
      description: 'Increase your energy production 1 step and gain 2 titanium.',
      tags: [CardTag.SPACE, CardTag.POWER, CardTag.SCIENCE],
      victoryPoints: 1,
      playOrder: 6
    }
  ];

  // Mock actions data compatible with ActionsModal
  const mockActions = [
    {
      id: 'research-action',
      name: 'Research',
      type: 'card' as const,
      cost: 1,
      description: 'Spend 1 M€ to draw a card.',
      requirement: undefined,
      available: true,
      source: 'Research Card',
      sourceType: CardType.ACTIVE,
      resourceType: 'cards',
      immediate: true
    },
    {
      id: 'mining-guild-action',
      name: 'Mining Guild Production',
      type: 'corporation' as const,
      cost: 0,
      description: 'Gain 1 steel for each steel and titanium resource on the board.',
      requirement: undefined,
      available: true,
      source: 'Mining Guild',
      sourceType: CardType.CORPORATION,
      resourceType: 'steel',
      immediate: true
    },
    {
      id: 'power-generation-action',
      name: 'Power Generation',
      type: 'card' as const,
      cost: 0,
      description: 'Convert 2 energy to 2 heat.',
      requirement: 'Must have at least 2 energy',
      available: true,
      source: 'Power Plant Card',
      sourceType: CardType.AUTOMATED,
      resourceType: 'heat',
      immediate: true
    },
    {
      id: 'tree-planting-action',
      name: 'Tree Planting',
      type: 'standard' as const,
      cost: 23,
      description: 'Spend 23 M€ to place a greenery tile and increase oxygen 1 step.',
      requirement: 'Must have available greenery space',
      available: false,
      source: 'Standard Project',
      sourceType: undefined,
      resourceType: undefined,
      immediate: true
    },
    {
      id: 'water-import-action',
      name: 'Aquifer Pumping',
      type: 'standard' as const,
      cost: 18,
      description: 'Spend 18 M€ to place an ocean tile.',
      requirement: 'Must have available ocean space',
      available: true,
      source: 'Standard Project',
      sourceType: undefined,
      resourceType: undefined,
      immediate: true
    },
    {
      id: 'development-action',
      name: 'Energy to Heat',
      type: 'card' as const,
      cost: 0,
      description: 'Convert plants to energy efficiently.',
      requirement: undefined,
      available: true,
      source: 'Development Card',
      sourceType: CardType.ACTIVE,
      resourceType: 'energy',
      immediate: true
    }
  ];

  // Mock card effects data
  const mockCardEffects = [
    {
      id: 'mining-guild-effect',
      cardId: 'mining-guild',
      cardName: 'Mining Guild',
      cardType: CardType.CORPORATION,
      effectType: 'ongoing' as const,
      name: 'Steel and Titanium Bonus',
      description: 'Gain 1 M€ production for each steel and titanium resource on the board.',
      isActive: true,
      category: 'production' as const,
      resource: 'credits',
      value: 5
    },
    {
      id: 'research-effect',
      cardId: 'research',
      cardName: 'Research',
      cardType: CardType.ACTIVE,
      effectType: 'activated' as const,
      name: 'Draw Cards',
      description: 'Spend 1 M€ to draw a card.',
      isActive: true,
      category: 'conversion' as const,
      resource: 'cards',
      value: 1,
      cooldown: false,
      usesRemaining: undefined
    },
    {
      id: 'power-plant-effect',
      cardId: 'power-plant',
      cardName: 'Power Plant',
      cardType: CardType.AUTOMATED,
      effectType: 'immediate' as const,
      name: 'Energy Production',
      description: 'Increased energy production by 1 step.',
      isActive: false,
      category: 'production' as const,
      resource: 'energy',
      value: 1
    }
  ];

  // Mock milestones data
  const mockMilestones = [
    {
      id: 'terraformer',
      name: 'Terraformer',
      description: 'Having a terraform rating of at least 35',
      points: 5,
      claimed: true
    },
    {
      id: 'mayor',
      name: 'Mayor',
      description: 'Having at least 3 cities',
      points: 5,
      claimed: false
    }
  ];

  // Mock awards data
  const mockAwards = [
    {
      id: 'landlord',
      name: 'Landlord',
      description: 'Most tiles in play',
      points: 5,
      position: 1
    }
  ];

  const playedCardsToShow = currentPlayer?.playedCards?.length ? currentPlayer.playedCards : mockPlayedCards;
  const availableEffects = playedCardsToShow?.filter((card: any) => card.type === CardType.ACTIVE)?.length || 0;

  // Modal handlers
  const handleOpenCardsModal = () => {
    console.log('Opening cards modal');
    onOpenCardsPlayedModal?.();
  };

  const handleOpenActionsModal = () => {
    console.log('Opening actions modal');
    onOpenActionsModal?.();
  };

  const handleOpenTagsModal = () => {
    console.log('Opening tags modal');
    onOpenTagsModal?.();
  };

  const handleOpenVictoryPointsModal = () => {
    console.log('Opening victory points modal');
    onOpenVictoryPointsModal?.();
  };

  const handleOpenCardEffectsModal = () => {
    console.log('Opening card effects modal');
    onOpenCardEffectsModal?.();
  };

  // Use mock cards if current player doesn't have played cards, otherwise use their cards
  const fullCardsData = playedCardsToShow?.map((card, index) => {
    if ('name' in card) {
      return card;
    } else {
      // Create mock card data
      return {
        id: `card-${index}`,
        name: `${card.type.charAt(0).toUpperCase() + card.type.slice(1)} Card`,
        type: card.type,
        cost: Math.floor(Math.random() * 20) + 1,
        description: `This is a ${card.type} card with various effects and abilities.`
      };
    }
  }) || [];

  // Modal escape handling is now managed in GameInterface

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
              onClick={() => handleResourceClick(resource)}
              title={`${resource.name}: ${resource.current} (${resource.production} production)`}
            >
              <div className="resource-production">
                {createImageWithNumber('/assets/misc/production.png', resource.production, 'production-display')}
              </div>
              
              <div className="resource-main">
                <div className="resource-icon">
                  {resource.id === 'credits' ? 
                    createImageWithNumber(resource.icon, resource.current, 'credits-display') :
                    <img src={resource.icon} alt={resource.name} className="resource-icon-img" />
                  }
                </div>
                {resource.id !== 'credits' && (
                  <div className="resource-current">{resource.current}</div>
                )}
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

      {/* Action Buttons Section */}
      <div className="action-buttons-section">
        <button 
          className="action-button cards-button"
          onClick={handleOpenCardsModal}
          title="View Played Cards"
        >
          <div className="button-icon">🃏</div>
          <div className="button-count">{playedCardsToShow?.length || 0}</div>
          <div className="button-label">Played</div>
        </button>

        <button 
          className="action-button tags-button"
          onClick={handleOpenTagsModal}
          title="View Tags"
        >
          <div className="button-icon">🏷️</div>
          <div className="button-count">{currentPlayer?.tags?.length || 0}</div>
          <div className="button-label">Tags</div>
        </button>

        <button 
          className="action-button vp-button"
          onClick={handleOpenVictoryPointsModal}
          title="View Victory Points"
        >
          <div className="button-icon">🏆</div>
          <div className="button-count">{currentPlayer?.victoryPoints || 0}</div>
          <div className="button-label">VP</div>
        </button>

        <button 
          className="action-button actions-button"
          onClick={handleOpenActionsModal}
          title="View Available Actions"
        >
          <div className="button-icon">⚡</div>
          <div className="button-count">{availableEffects}</div>
          <div className="button-label">Actions</div>
        </button>

        <button 
          className="action-button effects-button"
          onClick={handleOpenCardEffectsModal}
          title="View Card Effects"
        >
          <div className="button-icon">✨</div>
          <div className="button-count">{mockCardEffects.length}</div>
          <div className="button-label">Effects</div>
        </button>
      </div>

      <style>{`
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
          /* z-index removed - natural DOM order places this above game content */
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
          gap: 6px;
          background: linear-gradient(
            135deg,
            rgba(30, 60, 90, 0.4) 0%,
            rgba(20, 40, 70, 0.3) 100%
          );
          border: 2px solid var(--resource-color);
          border-radius: 12px;
          padding: 8px 6px;
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

        .resource-production {
          display: flex;
          align-items: center;
          justify-content: center;
          margin-bottom: 4px;
        }
        
        .resource-main {
          display: flex;
          align-items: center;
          gap: 6px;
        }
        
        .resource-icon {
          width: 32px;
          height: 32px;
          display: flex;
          align-items: center;
          justify-content: center;
          filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.5));
        }
        
        .resource-icon-img {
          width: 100%;
          height: 100%;
          object-fit: contain;
          image-rendering: crisp-edges;
        }

        .resource-current {
          font-size: 18px;
          font-weight: bold;
          color: #ffffff;
          text-shadow: 0 1px 3px rgba(0, 0, 0, 0.8);
        }
        
        .image-with-number {
          position: relative;
          display: inline-block;
        }
        
        .base-image {
          display: block;
          width: 100%;
          height: 100%;
          object-fit: contain;
        }
        
        .embedded-number {
          position: absolute;
          top: 50%;
          left: 50%;
          transform: translate(-50%, -50%);
          font-weight: bold;
          text-shadow: 0 1px 2px rgba(0, 0, 0, 0.8);
          pointer-events: none;
          line-height: 1;
        }
        
        .production-display {
          width: 24px;
          height: 24px;
        }
        
        .production-display .embedded-number {
          font-size: 12px;
          color: #ffffff;
        }
        
        .credits-display {
          width: 32px;
          height: 32px;
        }
        
        .credits-display .embedded-number {
          font-size: 14px;
          color: #000000;
          font-weight: 900;
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
          /* z-index removed - isolation provides natural stacking */
          isolation: isolate;
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

        .action-buttons-section {
          flex: 1;
          display: flex;
          align-items: center;
          justify-content: flex-end;
          gap: 12px;
        }

        .action-button {
          display: flex;
          flex-direction: column;
          align-items: center;
          gap: 4px;
          background: linear-gradient(
            135deg,
            rgba(30, 60, 90, 0.6) 0%,
            rgba(20, 40, 70, 0.5) 100%
          );
          border: 2px solid rgba(100, 150, 200, 0.4);
          border-radius: 12px;
          padding: 10px 8px;
          cursor: pointer;
          transition: all 0.3s ease;
          min-width: 60px;
          backdrop-filter: blur(5px);
        }

        .action-button:hover {
          transform: translateY(-2px);
          border-color: rgba(100, 150, 200, 0.8);
          background: linear-gradient(
            135deg,
            rgba(30, 60, 90, 0.8) 0%,
            rgba(20, 40, 70, 0.7) 100%
          );
          box-shadow: 
            0 4px 15px rgba(0, 0, 0, 0.3),
            0 0 15px rgba(100, 150, 200, 0.3);
        }

        .button-icon {
          font-size: 18px;
          filter: drop-shadow(0 1px 2px rgba(0, 0, 0, 0.5));
        }

        .button-count {
          font-size: 14px;
          font-weight: bold;
          color: #ffffff;
          text-shadow: 0 1px 2px rgba(0, 0, 0, 0.8);
          line-height: 1;
        }

        .button-label {
          font-size: 10px;
          font-weight: 500;
          color: rgba(255, 255, 255, 0.9);
          text-transform: uppercase;
          letter-spacing: 0.5px;
          text-shadow: 0 1px 2px rgba(0, 0, 0, 0.8);
        }

        .cards-button {
          border-color: rgba(150, 100, 255, 0.4);
        }

        .cards-button:hover {
          border-color: rgba(150, 100, 255, 0.8);
          box-shadow: 
            0 4px 15px rgba(0, 0, 0, 0.3),
            0 0 15px rgba(150, 100, 255, 0.3);
        }

        .tags-button {
          border-color: rgba(100, 255, 150, 0.4);
        }

        .tags-button:hover {
          border-color: rgba(100, 255, 150, 0.8);
          box-shadow: 
            0 4px 15px rgba(0, 0, 0, 0.3),
            0 0 15px rgba(100, 255, 150, 0.3);
        }

        .vp-button {
          border-color: rgba(255, 200, 100, 0.4);
        }

        .vp-button:hover {
          border-color: rgba(255, 200, 100, 0.8);
          box-shadow: 
            0 4px 15px rgba(0, 0, 0, 0.3),
            0 0 15px rgba(255, 200, 100, 0.3);
        }

        .actions-button {
          border-color: rgba(255, 100, 100, 0.4);
        }

        .actions-button:hover {
          border-color: rgba(255, 100, 100, 0.8);
          box-shadow: 
            0 4px 15px rgba(0, 0, 0, 0.3),
            0 0 15px rgba(255, 100, 100, 0.3);
        }

        .effects-button {
          border-color: rgba(255, 150, 255, 0.4);
        }

        .effects-button:hover {
          border-color: rgba(255, 150, 255, 0.8);
          box-shadow: 
            0 4px 15px rgba(0, 0, 0, 0.3),
            0 0 15px rgba(255, 150, 255, 0.3);
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
            width: 18px;
            height: 18px;
          }

          .resource-current {
            font-size: 14px;
          }
        }

        @media (max-width: 1024px) {
          .bottom-resource-bar {
            height: 100px;
            padding: 12px 25px;
          }

          .resources-grid {
            gap: 12px;
            max-width: 450px;
          }

          .resource-item {
            padding: 10px 7px;
          }

          .cards-indicator {
            padding: 12px 18px;
          }

          .cards-icon {
            font-size: 20px;
          }

          .cards-count {
            font-size: 18px;
          }

          .action-buttons-section {
            gap: 10px;
            padding: 0 15px;
          }

          .action-button {
            min-width: 55px;
            padding: 8px 6px;
          }

          .button-icon {
            font-size: 16px;
          }

          .button-count {
            font-size: 13px;
          }

          .button-label {
            font-size: 9px;
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
          .action-buttons-section,
          .game-info-section {
            width: 100%;
            align-items: center;
          }

          .action-buttons-section {
            gap: 8px;
            padding: 0 10px;
          }

          .action-button {
            min-width: 50px;
            padding: 6px 4px;
          }

          .button-icon {
            font-size: 14px;
          }

          .button-count {
            font-size: 12px;
          }

          .button-label {
            font-size: 8px;
          }

          .cards-preview {
            position: static;
            transform: none;
            margin-bottom: 0;
            margin-top: 10px;
          }

          .cards-grid {
            grid-template-columns: repeat(2, 1fr);
            width: 100%;
          }
        }

        @media (max-width: 600px) {
          .bottom-resource-bar {
            padding: 12px;
            gap: 12px;
          }

          .resources-grid {
            grid-template-columns: repeat(2, 1fr);
            gap: 8px;
          }

          .resource-item {
            padding: 8px 5px;
          }

          .resource-icon {
            width: 18px;
            height: 18px;
          }

          .resource-current {
            font-size: 14px;
          }

          .resource-production {
            font-size: 11px;
          }

          .cards-indicator {
            padding: 10px 15px;
          }

          .phase-label {
            font-size: 10px;
          }

          .actions-left {
            font-size: 12px;
          }

          .action-buttons-section {
            gap: 6px;
          }

          .action-button {
            min-width: 45px;
            padding: 5px 3px;
          }

          .button-icon {
            font-size: 12px;
          }

          .button-count {
            font-size: 11px;
          }

          .button-label {
            font-size: 7px;
          }
        }
      `}</style>

      {/* Modal components are now rendered in GameInterface */}
    </div>
  );
};

export default BottomResourceBar;