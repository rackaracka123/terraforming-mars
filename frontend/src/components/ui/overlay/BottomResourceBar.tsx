import React from "react";
import { CardType, CardTag } from "../../../types/cards.ts";
import { PlayerDto } from "../../../types/generated/api-types.ts";
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
  currentPlayer?: PlayerDto | null;
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
  onOpenVictoryPointsModal,
}) => {
  // Helper function to create image with embedded number
  const createImageWithNumber = (
    imageSrc: string,
    number: number,
    className: string = "",
  ) => {
    return (
      <div className={`image-with-number ${className}`}>
        <img src={imageSrc} alt="" className="base-image" />
        <span className="embedded-number">{number}</span>
      </div>
    );
  };

  // Early return if no player data available
  if (!currentPlayer?.resources || !currentPlayer?.production) {
    return null;
  }

  // Create resources from current player data
  const playerResources: ResourceData[] = [
    {
      id: "credits",
      name: "Credits",
      current: currentPlayer.resources.credits,
      production: currentPlayer.production.credits,
      icon: "/assets/resources/megacredit.png",
      color: "#f1c40f",
    },
    {
      id: "steel",
      name: "Steel",
      current: currentPlayer.resources.steel,
      production: currentPlayer.production.steel,
      icon: "/assets/resources/steel.png",
      color: "#95a5a6",
    },
    {
      id: "titanium",
      name: "Titanium",
      current: currentPlayer.resources.titanium,
      production: currentPlayer.production.titanium,
      icon: "/assets/resources/titanium.png",
      color: "#e74c3c",
    },
    {
      id: "plants",
      name: "Plants",
      current: currentPlayer.resources.plants,
      production: currentPlayer.production.plants,
      icon: "/assets/resources/plant.png",
      color: "#27ae60",
    },
    {
      id: "energy",
      name: "Energy",
      current: currentPlayer.resources.energy,
      production: currentPlayer.production.energy,
      icon: "/assets/resources/power.png",
      color: "#3498db",
    },
    {
      id: "heat",
      name: "Heat",
      current: currentPlayer.resources.heat,
      production: currentPlayer.production.heat,
      icon: "/assets/resources/heat.png",
      color: "#e67e22",
    },
  ];

  // Resource click handlers
  const handleResourceClick = (resource: ResourceData) => {
    // Show resource information
    alert(
      `Clicked on ${resource.name}: ${resource.current} (${resource.production} production)`,
    );

    // Special handling for different resources
    switch (resource.id) {
      case "plants":
        if (resource.current >= 8) {
          alert("Can convert plants to greenery tile!");
        }
        break;
      case "heat":
        if (resource.current >= 8) {
          alert("Can convert heat to raise temperature!");
        }
        break;
      case "energy":
        alert("Energy converts to heat at end of turn");
        break;
      default:
        // Resource info displayed
        break;
    }
  };

  // Add mock played cards for demonstration
  const mockPlayedCards = [
    {
      id: "mining-guild",
      name: "Mining Guild",
      type: CardType.CORPORATION,
      cost: 0,
      description:
        "You start with 30 M€, 5 steel production, and 1 steel. Each steel and titanium resource on the board gives you 1 M€ production.",
      tags: [CardTag.BUILDING, CardTag.SCIENCE],
      victoryPoints: 0,
      playOrder: 1,
    },
    {
      id: "power-plant",
      name: "Power Plant",
      type: CardType.AUTOMATED,
      cost: 4,
      description: "Increase your energy production 1 step.",
      tags: [CardTag.POWER, CardTag.BUILDING],
      victoryPoints: 1,
      playOrder: 2,
    },
    {
      id: "research",
      name: "Research",
      type: CardType.ACTIVE,
      cost: 11,
      description: "Action: Spend 1 M€ to draw a card.",
      tags: [CardTag.SCIENCE],
      victoryPoints: 1,
      playOrder: 3,
    },
    {
      id: "asteroid",
      name: "Asteroid",
      type: CardType.EVENT,
      cost: 14,
      description:
        "Raise temperature 1 step and gain 2 titanium. Remove up to 2 plants from any player.",
      tags: [CardTag.SPACE, CardTag.EVENT],
      victoryPoints: 0,
      playOrder: 4,
    },
    {
      id: "mining-rights",
      name: "Mining Rights",
      type: CardType.AUTOMATED,
      cost: 9,
      description:
        "Place a tile on an area with a steel or titanium bonus, adjacent to one of your tiles. Increase steel or titanium production 1 step.",
      tags: [CardTag.BUILDING, CardTag.MARS],
      victoryPoints: 1,
      playOrder: 5,
    },
    {
      id: "solar-wind-power",
      name: "Solar Wind Power",
      type: CardType.AUTOMATED,
      cost: 11,
      description:
        "Increase your energy production 1 step and gain 2 titanium.",
      tags: [CardTag.SPACE, CardTag.POWER, CardTag.SCIENCE],
      victoryPoints: 1,
      playOrder: 6,
    },
  ];

  // Mock card effects data
  const mockCardEffects = [
    {
      id: "mining-guild-effect",
      cardId: "mining-guild",
      cardName: "Mining Guild",
      cardType: CardType.CORPORATION,
      effectType: "ongoing" as const,
      name: "Steel and Titanium Bonus",
      description:
        "Gain 1 M€ production for each steel and titanium resource on the board.",
      isActive: true,
      category: "production" as const,
      resource: "credits",
      value: 5,
    },
    {
      id: "research-effect",
      cardId: "research",
      cardName: "Research",
      cardType: CardType.ACTIVE,
      effectType: "triggered" as const,
      name: "Research Bonus",
      description: "Get bonus when drawing cards.",
      isActive: true,
      category: "bonus" as const,
      resource: "cards",
      value: 1,
    },
    {
      id: "power-plant-effect",
      cardId: "power-plant",
      cardName: "Power Plant",
      cardType: CardType.AUTOMATED,
      effectType: "immediate" as const,
      name: "Energy Production",
      description: "Increased energy production by 1 step.",
      isActive: false,
      category: "production" as const,
      resource: "energy",
      value: 1,
    },
  ];

  const playedCardsToShow = currentPlayer?.playedCards?.length
    ? mockPlayedCards // Use mock cards for display until we have proper card data structure
    : mockPlayedCards;
  const availableEffects =
    playedCardsToShow?.filter((card) => card.type === CardType.ACTIVE)
      ?.length || 0;

  // Modal handlers
  const handleOpenCardsModal = () => {
    // Opening cards modal
    onOpenCardsPlayedModal?.();
  };

  const handleOpenActionsModal = () => {
    // Opening actions modal
    onOpenActionsModal?.();
  };

  const handleOpenTagsModal = () => {
    // Opening tags modal
    onOpenTagsModal?.();
  };

  const handleOpenVictoryPointsModal = () => {
    // Opening victory points modal
    onOpenVictoryPointsModal?.();
  };

  const handleOpenCardEffectsModal = () => {
    // Opening card effects modal
    onOpenCardEffectsModal?.();
  };

  // Modal escape handling is now managed in GameInterface

  return (
    <div className="bottom-resource-bar">
      {/* Resource Grid */}
      <div className="resources-section">
        <div className="resources-grid">
          {playerResources.map((resource) => (
            <div
              key={resource.id}
              className="resource-item"
              style={
                { "--resource-color": resource.color } as React.CSSProperties
              }
              onClick={() => handleResourceClick(resource)}
              title={`${resource.name}: ${resource.current} (${resource.production} production)`}
            >
              <div className="resource-production">
                {createImageWithNumber(
                  "/assets/misc/production.png",
                  resource.production,
                  "production-display",
                )}
              </div>

              <div className="resource-main">
                <div className="resource-icon">
                  {resource.id === "credits" ? (
                    createImageWithNumber(
                      resource.icon,
                      resource.current,
                      "credits-display",
                    )
                  ) : (
                    <img
                      src={resource.icon}
                      alt={resource.name}
                      className="resource-icon-img"
                    />
                  )}
                </div>
                {resource.id !== "credits" && (
                  <div className="resource-current">{resource.current}</div>
                )}
              </div>
            </div>
          ))}
        </div>
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
          <div className="button-count">{0}</div>
          <div className="button-label">Tags</div>
        </button>

        <button
          className="action-button vp-button"
          onClick={handleOpenVictoryPointsModal}
          title="View Victory Points"
        >
          <div className="button-icon">🏆</div>
          <div className="button-count">
            {currentPlayer?.victoryPoints || 0}
          </div>
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

          .action-buttons-section {
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
