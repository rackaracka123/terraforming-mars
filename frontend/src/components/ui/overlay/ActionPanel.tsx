import React from 'react';
import VictoryPointsDisplay from '../display/VictoryPointsDisplay.tsx';
import CardsPlayedTracker from '../display/CardsPlayedTracker.tsx';
import TagsOverview from '../display/TagsOverview.tsx';
import AvailableActionsDisplay from '../display/AvailableActionsDisplay.tsx';
import { CardTag, CardType } from '../../../types/cards.ts';
import { useMainContent } from '../../../contexts/MainContentContext.tsx';

interface Card {
  id: string;
  name: string;
  type: CardType;
  cost: number;
  description: string;
}

interface Player {
  id: string;
  name: string;
  victoryPoints: number;
  availableActions: number;
  playedCards: Array<{ type: CardType }> | Card[];
  tags: CardTag[];
}

interface ActionPanelProps {
  currentPlayer: Player | null;
  className?: string;
}

const ActionPanel: React.FC<ActionPanelProps> = ({ currentPlayer, className = '' }) => {
  const { setContentType, setContentData } = useMainContent();

  if (!currentPlayer) {
    return null;
  }

  // Add mock played cards for demonstration
  const mockPlayedCards: Card[] = [
    {
      id: 'mining-guild',
      name: 'Mining Guild',
      type: CardType.CORPORATION,
      cost: 0,
      description: 'You start with 30 M€, 5 steel production, and 1 steel. Each steel and titanium resource on the board gives you 1 M€ production.'
    },
    {
      id: 'power-plant',
      name: 'Power Plant',
      type: CardType.AUTOMATED,
      cost: 4,
      description: 'Increase your energy production 1 step.'
    },
    {
      id: 'research',
      name: 'Research',
      type: CardType.ACTIVE,
      cost: 11,
      description: 'Action: Spend 1 M€ to draw a card.'
    },
    {
      id: 'asteroid',
      name: 'Asteroid',
      type: CardType.EVENT,
      cost: 14,
      description: 'Raise temperature 1 step and gain 2 titanium. Remove up to 2 plants from any player.'
    },
    {
      id: 'mining-rights',
      name: 'Mining Rights',
      type: CardType.AUTOMATED,
      cost: 9,
      description: 'Place a tile on an area with a steel or titanium bonus, adjacent to one of your tiles. Increase steel or titanium production 1 step.'
    },
    {
      id: 'solar-wind-power',
      name: 'Solar Wind Power',
      type: CardType.AUTOMATED,
      cost: 11,
      description: 'Increase your energy production 1 step and gain 2 titanium.'
    }
  ];

  // Use mock cards if current player doesn't have played cards, otherwise use their cards
  const playedCardsToShow = currentPlayer.playedCards?.length ? currentPlayer.playedCards : mockPlayedCards;

  // Calculate available effects (placeholder logic - would need to be implemented based on actual game state)
  const availableEffects = playedCardsToShow?.filter(card => card.type === CardType.ACTIVE)?.length || 0;

  // Create full card data for display
  const fullCardsData: Card[] = playedCardsToShow?.map((card, index) => {
    if ('name' in card) {
      return card as Card;
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

  // Mock card actions only - no standard projects, milestones, or awards
  const mockActions = [
    {
      id: 'research-action',
      name: 'Research',
      type: 'card' as const,
      description: 'Spend 1 M€ to draw a card.',
      source: 'Research Card',
      available: true,
      actionCost: { credits: 1 },
      actionReward: { cards: 1 }
    },
    {
      id: 'mining-guild-action',
      name: 'Mining Guild Action',
      type: 'card' as const,
      description: 'Gain 1 steel for each steel and titanium resource on the board.',
      source: 'Mining Guild',
      available: true,
      actionReward: { steel: 1 }
    },
    {
      id: 'power-generation-action',
      name: 'Power Generation',
      type: 'card' as const,
      description: 'Convert 2 energy to 2 heat.',
      source: 'Power Plant Card',
      available: true,
      actionCost: { energy: 2 },
      actionReward: { heat: 2 }
    },
    {
      id: 'tree-planting-action',
      name: 'Tree Planting',
      type: 'card' as const,
      description: 'Spend 3 M€ to place a greenery tile.',
      source: 'Tree Farming Card',
      available: false,
      requirement: 'Must have available greenery space',
      actionCost: { credits: 3 }
    },
    {
      id: 'water-import-action',
      name: 'Water Import',
      type: 'card' as const,
      description: 'Spend 4 M€ to gain 2 plants.',
      source: 'Water Import Card',
      available: true,
      actionCost: { credits: 4 },
      actionReward: { plants: 2 }
    },
    {
      id: 'development-action',
      name: 'Development',
      type: 'card' as const,
      description: 'Convert plants to energy efficiently.',
      source: 'Development Card',
      available: true,
      actionCost: { plants: 3 },
      actionReward: { energy: 2 }
    }
  ];

  const handleOpenCardsView = () => {
    setContentData({
      cards: fullCardsData,
      playerName: currentPlayer.name
    });
    setContentType('played-cards');
  };

  const handleOpenActionsView = () => {
    setContentData({
      actions: mockActions,
      playerName: currentPlayer.name
    });
    setContentType('available-actions');
  };

  return (
    <div 
      className={`action-panel ${className}`}
      style={{
        position: 'fixed',
        right: '20px',
        bottom: '120px',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'stretch',
        gap: '12px',
        padding: '16px',
        background: 'linear-gradient(135deg, rgba(10, 20, 40, 0.95) 0%, rgba(20, 30, 50, 0.9) 50%, rgba(15, 25, 45, 0.95) 100%)',
        border: '2px solid rgba(100, 150, 200, 0.3)',
        borderRadius: '16px',
        boxShadow: '0 8px 32px rgba(0, 0, 0, 0.6), 0 0 40px rgba(0, 50, 100, 0.4)',
        backdropFilter: 'blur(15px)',
        zIndex: 500,
        clipPath: 'polygon(0 0, calc(100% - 20px) 0, 100% 20px, 100% 100%, 20px 100%, 0 calc(100% - 20px))',
        width: '280px',
      }}
    >
      {/* Background overlay for the futuristic look */}
      <div
        style={{
          position: 'absolute',
          top: 0,
          left: 0,
          right: 0,
          bottom: 0,
          background: 'linear-gradient(45deg, rgba(0, 100, 200, 0.05) 0%, transparent 25%, rgba(100, 150, 255, 0.03) 50%, transparent 75%, rgba(50, 100, 200, 0.05) 100%)',
          borderRadius: 'inherit',
          pointerEvents: 'none',
        }}
      />

      {/* Cards Played Tracker */}
      <CardsPlayedTracker 
        playedCards={playedCardsToShow || []}
        size="medium"
        onClick={handleOpenCardsView}
      />

      {/* Tags Overview */}
      <TagsOverview 
        tags={currentPlayer.tags || []}
        size="medium"
      />

      {/* Victory Points Display */}
      <VictoryPointsDisplay 
        victoryPoints={currentPlayer.victoryPoints}
        size="medium"
      />

      {/* Available Actions Display */}
      <AvailableActionsDisplay 
        availableActions={currentPlayer.availableActions || 0}
        availableEffects={availableEffects}
        size="medium"
        onClick={handleOpenActionsView}
      />

      <style jsx>{`
        .action-panel::before {
          content: '';
          position: absolute;
          top: -2px;
          left: -2px;
          right: -2px;
          bottom: -2px;
          background: linear-gradient(
            45deg,
            rgba(100, 150, 255, 0.3) 0%,
            rgba(50, 100, 200, 0.2) 25%,
            rgba(0, 50, 150, 0.1) 50%,
            rgba(50, 100, 200, 0.2) 75%,
            rgba(100, 150, 255, 0.3) 100%
          );
          border-radius: inherit;
          clip-path: inherit;
          z-index: -1;
          animation: borderGlow 3s ease-in-out infinite alternate;
        }

        @keyframes borderGlow {
          from {
            opacity: 0.6;
          }
          to {
            opacity: 1;
          }
        }

        .action-panel:hover::before {
          animation-duration: 1.5s;
        }

        @media (max-width: 1200px) {
          .action-panel {
            right: 15px;
            bottom: 100px;
            width: 240px;
            gap: 10px;
            padding: 12px;
          }
        }

        @media (max-width: 900px) {
          .action-panel {
            right: 10px;
            bottom: 80px;
            width: 200px;
            gap: 8px;
            padding: 10px;
          }
        }

        @media (max-width: 768px) {
          .action-panel {
            right: 10px;
            bottom: 60px;
            width: 180px;
            gap: 6px;
            padding: 8px;
          }
        }
      `}</style>
    </div>
  );
};

export default ActionPanel;