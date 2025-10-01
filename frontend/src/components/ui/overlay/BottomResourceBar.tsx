import React, { useRef, useState } from "react";
import {
  PlayerDto,
  PlayerActionDto,
  GameDto,
  CardDto,
} from "@/types/generated/api-types.ts";
import ActionsPopover from "../popover/ActionsPopover.tsx";
import EffectsPopover from "../popover/EffectsPopover.tsx";
import TagsPopover from "../popover/TagsPopover.tsx";
import StoragesPopover from "../popover/StoragesPopover.tsx";
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
  gameState?: GameDto;
  playedCards?: CardDto[];
  onOpenCardEffectsModal?: () => void;
  onOpenCardsPlayedModal?: () => void;
  onOpenVictoryPointsModal?: () => void;
  onOpenActionsModal?: () => void;
  onActionSelect?: (action: PlayerActionDto) => void;
}

const BottomResourceBar: React.FC<BottomResourceBarProps> = ({
  currentPlayer,
  gameState,
  playedCards = [],
  onOpenCardEffectsModal,
  onOpenCardsPlayedModal,
  onOpenVictoryPointsModal,
  onOpenActionsModal,
  onActionSelect,
}) => {
  const [showActionsPopover, setShowActionsPopover] = useState(false);
  const [showEffectsPopover, setShowEffectsPopover] = useState(false);
  const [showTagsPopover, setShowTagsPopover] = useState(false);
  const [showStoragesPopover, setShowStoragesPopover] = useState(false);
  const actionsButtonRef = useRef<HTMLButtonElement>(null);
  const effectsButtonRef = useRef<HTMLButtonElement>(null);
  const tagsButtonRef = useRef<HTMLButtonElement>(null);
  const storagesButtonRef = useRef<HTMLButtonElement>(null);
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

  // Get tag icon mapping
  const getTagIcon = (tag: string): string => {
    const iconMap: { [key: string]: string } = {
      space: "/assets/tags/space.png",
      earth: "/assets/tags/earth.png",
      science: "/assets/tags/science.png",
      power: "/assets/tags/power.png",
      building: "/assets/tags/building.png",
      microbe: "/assets/tags/microbe.png",
      animal: "/assets/tags/animal.png",
      plant: "/assets/tags/plant.png",
      event: "/assets/tags/event.png",
      city: "/assets/tags/city.png",
      venus: "/assets/tags/venus.png",
      jovian: "/assets/tags/jovian.png",
      wildlife: "/assets/tags/wildlife.png",
      wild: "/assets/tags/wild.png",
      mars: "/assets/tags/mars.png",
      moon: "/assets/tags/moon.png",
      clone: "/assets/tags/clone.png",
      crime: "/assets/tags/crime.png",
    };
    return iconMap[tag.toLowerCase()] || "/assets/tags/empty.png";
  };

  // Count tags from played cards
  const tagCounts = React.useMemo(() => {
    if (!playedCards || playedCards.length === 0) return [];

    const counts: { [key: string]: number } = {};

    // Count tags from played cards
    playedCards.forEach((card) => {
      if (card.tags) {
        card.tags.forEach((tag) => {
          const tagKey = tag.toLowerCase();
          counts[tagKey] = (counts[tagKey] || 0) + 1;
        });
      }
    });

    // All possible tags
    const allTags = [
      "space",
      "earth",
      "science",
      "power",
      "building",
      "microbe",
      "animal",
      "plant",
      "event",
      "city",
      "venus",
      "jovian",
      "wild",
      "mars",
      "moon",
      "clone",
      "crime",
    ];

    return allTags.map((tag) => ({
      tag,
      count: counts[tag] || 0,
      icon: getTagIcon(tag),
    }));
  }, [playedCards]);

  // Count storage cards (cards with resource storage)
  const storageCardsCount = React.useMemo(() => {
    if (!currentPlayer?.resourceStorage) return 0;
    return Object.keys(currentPlayer.resourceStorage).length;
  }, [currentPlayer?.resourceStorage]);

  // Early return if no player data available
  if (!currentPlayer?.resources || !currentPlayer?.resourceProduction) {
    return null;
  }

  // Create resources from current player data
  const playerResources: ResourceData[] = [
    {
      id: "credits",
      name: "Credits",
      current: currentPlayer.resources.credits,
      production: currentPlayer.resourceProduction.credits,
      icon: "/assets/resources/megacredit.png",
      color: "#f1c40f",
    },
    {
      id: "steel",
      name: "Steel",
      current: currentPlayer.resources.steel,
      production: currentPlayer.resourceProduction.steel,
      icon: "/assets/resources/steel.png",
      color: "#95a5a6",
    },
    {
      id: "titanium",
      name: "Titanium",
      current: currentPlayer.resources.titanium,
      production: currentPlayer.resourceProduction.titanium,
      icon: "/assets/resources/titanium.png",
      color: "#e74c3c",
    },
    {
      id: "plants",
      name: "Plants",
      current: currentPlayer.resources.plants,
      production: currentPlayer.resourceProduction.plants,
      icon: "/assets/resources/plant.png",
      color: "#27ae60",
    },
    {
      id: "energy",
      name: "Energy",
      current: currentPlayer.resources.energy,
      production: currentPlayer.resourceProduction.energy,
      icon: "/assets/resources/power.png",
      color: "#3498db",
    },
    {
      id: "heat",
      name: "Heat",
      current: currentPlayer.resources.heat,
      production: currentPlayer.resourceProduction.heat,
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

  // Get actual played cards count from game state
  const playedCardsCount = currentPlayer?.playedCards?.length || 0;

  // Modal handlers
  const handleOpenCardsModal = () => {
    // Opening cards modal
    onOpenCardsPlayedModal?.();
  };

  const handleOpenActionsPopover = () => {
    setShowActionsPopover(!showActionsPopover);
  };

  const handleOpenEffectsPopover = () => {
    setShowEffectsPopover(!showEffectsPopover);
  };

  const handleOpenStoragesPopover = () => {
    setShowStoragesPopover(!showStoragesPopover);
  };

  const handleOpenTagsPopover = () => {
    setShowTagsPopover(!showTagsPopover);
  };

  const handleOpenVictoryPointsModal = () => {
    // Opening victory points modal
    onOpenVictoryPointsModal?.();
  };

  // Modal escape handling is now managed in GameInterface

  return (
    <div className="fixed bottom-0 left-0 right-0 h-12 flex items-end justify-between px-[30px] pb-2 z-[1000] pointer-events-auto">
      {/* Background bar */}
      <div className="absolute inset-0 bg-space-black-darker/95 backdrop-blur-space border-t-2 border-space-blue-400 shadow-[0_-8px_32px_rgba(0,0,0,0.6),0_0_20px_rgba(30,60,150,0.3)] -z-10" />

      {/* Resource Grid */}
      <div className="flex-[2] -translate-y-[30px] pointer-events-auto relative">
        <div className="grid grid-cols-6 gap-[15px] max-w-[500px]">
          {playerResources.map((resource) => (
            <div
              key={resource.id}
              className="flex flex-col items-center gap-1.5 bg-gradient-to-br from-[rgba(30,60,90,0.8)] to-[rgba(20,40,70,0.7)] border-2 rounded-xl p-2 transition-all duration-150 cursor-pointer relative overflow-hidden hover:-translate-y-0.5 hover:shadow-[0_6px_20px_rgba(0,0,0,0.4)] before:absolute before:inset-0 before:opacity-10 before:transition-opacity before:duration-150 hover:before:opacity-20"
              style={
                {
                  "--resource-color": resource.color,
                  borderColor: resource.color,
                } as React.CSSProperties
              }
              onClick={() => handleResourceClick(resource)}
              title={`${resource.name}: ${resource.current} (${resource.production} production)`}
            >
              <div className="flex items-center justify-center mb-1">
                {createImageWithNumber(
                  "/assets/misc/production.png",
                  resource.production,
                  "production-display",
                )}
              </div>

              <div className="flex items-center gap-1.5">
                <div className="w-8 h-8 flex items-center justify-center [filter:drop-shadow(0_2px_4px_rgba(0,0,0,0.5))]">
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
                      className="w-full h-full object-contain [image-rendering:crisp-edges]"
                    />
                  )}
                </div>
                {resource.id !== "credits" && (
                  <div className="text-lg font-bold text-white [text-shadow:0_1px_3px_rgba(0,0,0,0.8)]">
                    {resource.current}
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Action Buttons Section */}
      <div className="flex-1 flex items-center justify-end gap-3 -translate-y-[30px] pointer-events-auto relative">
        <button
          ref={actionsButtonRef}
          className={`flex flex-col items-center gap-1 bg-gradient-to-br from-[rgba(30,60,90,0.6)] to-[rgba(20,40,70,0.5)] border-2 rounded-xl py-2.5 px-2 cursor-pointer transition-all duration-150 min-w-[60px] backdrop-blur-[5px] hover:-translate-y-0.5 hover:bg-gradient-to-br hover:from-[rgba(30,60,90,0.8)] hover:to-[rgba(20,40,70,0.7)] hover:shadow-[0_4px_15px_rgba(0,0,0,0.3),0_0_15px_rgba(255,100,100,0.3)] ${
            (currentPlayer?.actions?.length || 0) === 0
              ? "border-[rgba(150,150,150,0.4)] bg-gradient-to-br from-[rgba(40,40,40,0.6)] to-[rgba(30,30,30,0.5)] opacity-70 hover:border-[rgba(150,150,150,0.6)] hover:opacity-80"
              : (currentPlayer?.actions?.length || 0) <= 1
                ? "border-[rgba(255,200,0,0.6)] bg-gradient-to-br from-[rgba(60,50,0,0.6)] to-[rgba(40,30,0,0.5)] hover:border-[rgba(255,200,0,0.9)] hover:shadow-[0_4px_15px_rgba(0,0,0,0.3),0_0_15px_rgba(255,200,0,0.4)]"
                : "border-[rgba(255,100,100,0.4)] hover:border-[rgba(255,100,100,0.8)]"
          }`}
          onClick={handleOpenActionsPopover}
          title={`Card Actions: ${currentPlayer?.actions?.length || 0}`}
        >
          <div className="text-lg [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))]">
            ⚡
          </div>
          <div
            className={`text-sm font-bold text-white [text-shadow:0_1px_2px_rgba(0,0,0,0.8)] leading-none ${(currentPlayer?.actions?.length || 0) === 0 ? "text-white/60" : ""}`}
          >
            {currentPlayer?.actions?.length || 0}
          </div>
          <div className="text-[10px] font-medium text-white/90 uppercase tracking-[0.5px] [text-shadow:0_1px_2px_rgba(0,0,0,0.8)]">
            Actions
          </div>
        </button>

        <button
          ref={effectsButtonRef}
          className="flex flex-col items-center gap-1 bg-gradient-to-br from-[rgba(30,60,90,0.6)] to-[rgba(20,40,70,0.5)] border-2 border-[rgba(255,150,255,0.4)] rounded-xl py-2.5 px-2 cursor-pointer transition-all duration-150 min-w-[60px] backdrop-blur-[5px] hover:-translate-y-0.5 hover:border-[rgba(255,150,255,0.8)] hover:bg-gradient-to-br hover:from-[rgba(30,60,90,0.8)] hover:to-[rgba(20,40,70,0.7)] hover:shadow-[0_4px_15px_rgba(0,0,0,0.3),0_0_15px_rgba(255,150,255,0.3)]"
          onClick={handleOpenEffectsPopover}
          title="View Card Effects"
        >
          <div className="text-lg [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))]">
            ✨
          </div>
          <div className="text-sm font-bold text-white [text-shadow:0_1px_2px_rgba(0,0,0,0.8)] leading-none">
            {currentPlayer?.effects?.length || 0}
          </div>
          <div className="text-[10px] font-medium text-white/90 uppercase tracking-[0.5px] [text-shadow:0_1px_2px_rgba(0,0,0,0.8)]">
            Effects
          </div>
        </button>

        <button
          ref={tagsButtonRef}
          className="flex flex-col items-center gap-1 bg-gradient-to-br from-[rgba(30,60,90,0.6)] to-[rgba(20,40,70,0.5)] border-2 border-[rgba(100,255,150,0.4)] rounded-xl py-2.5 px-2 cursor-pointer transition-all duration-150 min-w-[60px] backdrop-blur-[5px] hover:-translate-y-0.5 hover:border-[rgba(100,255,150,0.8)] hover:bg-gradient-to-br hover:from-[rgba(30,60,90,0.8)] hover:to-[rgba(20,40,70,0.7)] hover:shadow-[0_4px_15px_rgba(0,0,0,0.3),0_0_15px_rgba(100,255,150,0.3)]"
          onClick={handleOpenTagsPopover}
          title="View Tags"
        >
          <div className="text-lg [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))]">
            🏷️
          </div>
          <div className="text-sm font-bold text-white [text-shadow:0_1px_2px_rgba(0,0,0,0.8)] leading-none">
            {tagCounts.reduce((sum, tag) => sum + tag.count, 0)}
          </div>
          <div className="text-[10px] font-medium text-white/90 uppercase tracking-[0.5px] [text-shadow:0_1px_2px_rgba(0,0,0,0.8)]">
            Tags
          </div>
        </button>

        <button
          ref={storagesButtonRef}
          className="flex flex-col items-center gap-1 bg-gradient-to-br from-[rgba(30,60,90,0.6)] to-[rgba(20,40,70,0.5)] border-2 border-[rgba(100,150,200,0.4)] rounded-xl py-2.5 px-2 cursor-pointer transition-all duration-150 min-w-[60px] backdrop-blur-[5px] hover:-translate-y-0.5 hover:border-[rgba(100,150,200,0.8)] hover:bg-gradient-to-br hover:from-[rgba(30,60,90,0.8)] hover:to-[rgba(20,40,70,0.7)] hover:shadow-[0_4px_15px_rgba(0,0,0,0.3),0_0_15px_rgba(100,150,200,0.3)]"
          onClick={handleOpenStoragesPopover}
          title="View Card Storages"
        >
          <div className="text-lg [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))]">
            💾
          </div>
          <div className="text-sm font-bold text-white [text-shadow:0_1px_2px_rgba(0,0,0,0.8)] leading-none">
            {storageCardsCount}
          </div>
          <div className="text-[10px] font-medium text-white/90 uppercase tracking-[0.5px] [text-shadow:0_1px_2px_rgba(0,0,0,0.8)]">
            Storages
          </div>
        </button>

        <button
          className="flex flex-col items-center gap-1 bg-gradient-to-br from-[rgba(30,60,90,0.6)] to-[rgba(20,40,70,0.5)] border-2 border-[rgba(150,100,255,0.4)] rounded-xl py-2.5 px-2 cursor-pointer transition-all duration-150 min-w-[60px] backdrop-blur-[5px] hover:-translate-y-0.5 hover:border-[rgba(150,100,255,0.8)] hover:bg-gradient-to-br hover:from-[rgba(30,60,90,0.8)] hover:to-[rgba(20,40,70,0.7)] hover:shadow-[0_4px_15px_rgba(0,0,0,0.3),0_0_15px_rgba(150,100,255,0.3)]"
          onClick={handleOpenCardsModal}
          title="View Played Cards"
        >
          <div className="text-lg [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))]">
            🃏
          </div>
          <div className="text-sm font-bold text-white [text-shadow:0_1px_2px_rgba(0,0,0,0.8)] leading-none">
            {playedCardsCount}
          </div>
          <div className="text-[10px] font-medium text-white/90 uppercase tracking-[0.5px] [text-shadow:0_1px_2px_rgba(0,0,0,0.8)]">
            Played
          </div>
        </button>

        <button
          className="flex flex-col items-center gap-1 bg-gradient-to-br from-[rgba(30,60,90,0.6)] to-[rgba(20,40,70,0.5)] border-2 border-[rgba(255,200,100,0.4)] rounded-xl py-2.5 px-2 cursor-pointer transition-all duration-150 min-w-[60px] backdrop-blur-[5px] hover:-translate-y-0.5 hover:border-[rgba(255,200,100,0.8)] hover:bg-gradient-to-br hover:from-[rgba(30,60,90,0.8)] hover:to-[rgba(20,40,70,0.7)] hover:shadow-[0_4px_15px_rgba(0,0,0,0.3),0_0_15px_rgba(255,200,100,0.3)]"
          onClick={handleOpenVictoryPointsModal}
          title="View Victory Points"
        >
          <div className="text-lg [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))]">
            🏆
          </div>
          <div className="text-sm font-bold text-white [text-shadow:0_1px_2px_rgba(0,0,0,0.8)] leading-none">
            {currentPlayer?.victoryPoints || 0}
          </div>
          <div className="text-[10px] font-medium text-white/90 uppercase tracking-[0.5px] [text-shadow:0_1px_2px_rgba(0,0,0,0.8)]">
            VP
          </div>
        </button>
      </div>

      <style>{`
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

      `}</style>

      {/* Actions Popover */}
      <ActionsPopover
        isVisible={showActionsPopover}
        onClose={() => setShowActionsPopover(false)}
        actions={currentPlayer?.actions || []}
        playerName={currentPlayer?.name}
        onActionSelect={(action) => {
          onActionSelect?.(action);
          setShowActionsPopover(false);
        }}
        onOpenDetails={onOpenActionsModal}
        anchorRef={actionsButtonRef as React.RefObject<HTMLElement>}
        gameState={gameState}
      />

      {/* Effects Popover */}
      <EffectsPopover
        isVisible={showEffectsPopover}
        onClose={() => setShowEffectsPopover(false)}
        effects={currentPlayer?.effects || []}
        playerName={currentPlayer?.name}
        onOpenDetails={onOpenCardEffectsModal}
        anchorRef={effectsButtonRef as React.RefObject<HTMLElement>}
      />

      {/* Tags Popover */}
      <TagsPopover
        isVisible={showTagsPopover}
        onClose={() => setShowTagsPopover(false)}
        tagCounts={tagCounts}
        anchorRef={tagsButtonRef as React.RefObject<HTMLElement>}
      />

      {/* Storages Popover */}
      <StoragesPopover
        isVisible={showStoragesPopover}
        onClose={() => setShowStoragesPopover(false)}
        player={currentPlayer}
        anchorRef={storagesButtonRef as React.RefObject<HTMLElement>}
      />

      {/* Modal components are now rendered in GameInterface */}
    </div>
  );
};

export default BottomResourceBar;
