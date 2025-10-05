import React, { useRef, useState } from "react";
import {
  PlayerDto,
  PlayerActionDto,
  GameDto,
  CardDto,
  ResourceTypeCredits,
  ResourceTypeSteel,
  ResourceTypeTitanium,
  ResourceTypePlants,
  ResourceTypeEnergy,
  ResourceTypeHeat,
} from "@/types/generated/api-types.ts";
import ActionsPopover from "../popover/ActionsPopover.tsx";
import EffectsPopover from "../popover/EffectsPopover.tsx";
import TagsPopover from "../popover/TagsPopover.tsx";
import StoragesPopover from "../popover/StoragesPopover.tsx";
import GameIcon from "../display/GameIcon.tsx";
// Modal components are now imported and managed in GameInterface

interface ResourceData {
  id: string;
  name: string;
  current: number;
  production: number;
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

  // Map resource ID to ResourceType constant
  const getResourceType = (resourceId: string): string => {
    const resourceTypeMap: Record<string, string> = {
      credits: ResourceTypeCredits,
      steel: ResourceTypeSteel,
      titanium: ResourceTypeTitanium,
      plants: ResourceTypePlants,
      energy: ResourceTypeEnergy,
      heat: ResourceTypeHeat,
    };
    return resourceTypeMap[resourceId] || resourceId;
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
    }));
  }, [playedCards]);

  // Count storage cards (cards with resource storage)
  const storageCardsCount = React.useMemo(() => {
    if (!currentPlayer?.resourceStorage) return 0;
    return Object.keys(currentPlayer.resourceStorage).length;
  }, [currentPlayer?.resourceStorage]);

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
      color: "#f1c40f", // Gold - OK already
    },
    {
      id: "steel",
      name: "Steel",
      current: currentPlayer.resources.steel,
      production: currentPlayer.production.steel,
      color: "#d2691e", // Brown/orangy
    },
    {
      id: "titanium",
      name: "Titanium",
      current: currentPlayer.resources.titanium,
      production: currentPlayer.production.titanium,
      color: "#95a5a6", // Grey
    },
    {
      id: "plants",
      name: "Plants",
      current: currentPlayer.resources.plants,
      production: currentPlayer.production.plants,
      color: "#27ae60", // Green - OK already
    },
    {
      id: "energy",
      name: "Energy",
      current: currentPlayer.resources.energy,
      production: currentPlayer.production.energy,
      color: "#9b59b6", // Purple
    },
    {
      id: "heat",
      name: "Heat",
      current: currentPlayer.resources.heat,
      production: currentPlayer.production.heat,
      color: "#ff4500", // Red/orange
    },
  ];

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
              className="flex flex-col items-center gap-1.5 bg-space-black-darker/90 border-2 rounded-xl p-2 transition-all duration-200 cursor-pointer relative overflow-hidden hover:-translate-y-0.5"
              style={
                {
                  "--resource-color": resource.color,
                  borderColor: resource.color,
                  boxShadow: `0 0 10px ${resource.color}40`,
                } as React.CSSProperties
              }
              onMouseEnter={(e) => {
                e.currentTarget.style.boxShadow = `0 6px 20px rgba(0,0,0,0.4), 0 0 20px ${resource.color}`;
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.boxShadow = `0 0 10px ${resource.color}40`;
              }}
              title={`${resource.name}: ${resource.current} (${resource.production} production)`}
            >
              <div className="inline-flex items-center justify-center bg-[linear-gradient(135deg,rgba(160,110,60,0.4)_0%,rgba(139,89,42,0.35)_100%)] border border-[rgba(160,110,60,0.5)] rounded px-2 py-1 shadow-[0_1px_3px_rgba(0,0,0,0.2)] mb-1 min-w-[28px]">
                <span className="text-sm font-bold text-white [text-shadow:0_1px_2px_rgba(0,0,0,0.8)] leading-none">
                  {resource.production}
                </span>
              </div>

              {resource.id === "credits" ? (
                <GameIcon
                  iconType={ResourceTypeCredits}
                  amount={resource.current}
                  size="medium"
                />
              ) : (
                <div className="flex items-center gap-1.5">
                  <GameIcon
                    iconType={getResourceType(resource.id)}
                    size="medium"
                  />
                  <div className="text-lg font-bold text-white [text-shadow:0_1px_3px_rgba(0,0,0,0.8)]">
                    {resource.current}
                  </div>
                </div>
              )}
            </div>
          ))}
        </div>
      </div>

      {/* Action Buttons Section */}
      <div className="flex-1 flex items-center justify-end gap-3 -translate-y-[30px] pointer-events-auto relative">
        <button
          ref={actionsButtonRef}
          className={`flex flex-col items-center gap-1 bg-space-black-darker/90 border-2 rounded-xl py-2.5 px-2 cursor-pointer transition-all duration-200 min-w-[60px] hover:-translate-y-0.5 ${
            (currentPlayer?.actions?.length || 0) === 0
              ? "border-[#969696] opacity-70 hover:opacity-80"
              : (currentPlayer?.actions?.length || 0) <= 1
                ? "border-[#ffc800]"
                : "border-[#ff6464]"
          }`}
          style={{
            boxShadow:
              (currentPlayer?.actions?.length || 0) === 0
                ? "0 0 10px #96969640"
                : (currentPlayer?.actions?.length || 0) <= 1
                  ? "0 0 10px #ffc80040"
                  : "0 0 10px #ff646440",
          }}
          onMouseEnter={(e) => {
            const color =
              (currentPlayer?.actions?.length || 0) === 0
                ? "#969696"
                : (currentPlayer?.actions?.length || 0) <= 1
                  ? "#ffc800"
                  : "#ff6464";
            e.currentTarget.style.boxShadow = `0 6px 20px rgba(0,0,0,0.4), 0 0 20px ${color}`;
          }}
          onMouseLeave={(e) => {
            const color =
              (currentPlayer?.actions?.length || 0) === 0
                ? "#969696"
                : (currentPlayer?.actions?.length || 0) <= 1
                  ? "#ffc800"
                  : "#ff6464";
            e.currentTarget.style.boxShadow = `0 0 10px ${color}40`;
          }}
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
          className="flex flex-col items-center gap-1 bg-space-black-darker/90 border-2 border-[#ff96ff] rounded-xl py-2.5 px-2 cursor-pointer transition-all duration-200 min-w-[60px] hover:-translate-y-0.5"
          style={{ boxShadow: "0 0 10px #ff96ff40" }}
          onMouseEnter={(e) => {
            e.currentTarget.style.boxShadow =
              "0 6px 20px rgba(0,0,0,0.4), 0 0 20px #ff96ff";
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.boxShadow = "0 0 10px #ff96ff40";
          }}
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
          className="flex flex-col items-center gap-1 bg-space-black-darker/90 border-2 border-[#64ff96] rounded-xl py-2.5 px-2 cursor-pointer transition-all duration-200 min-w-[60px] hover:-translate-y-0.5"
          style={{ boxShadow: "0 0 10px #64ff9640" }}
          onMouseEnter={(e) => {
            e.currentTarget.style.boxShadow =
              "0 6px 20px rgba(0,0,0,0.4), 0 0 20px #64ff96";
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.boxShadow = "0 0 10px #64ff9640";
          }}
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
          className="flex flex-col items-center gap-1 bg-space-black-darker/90 border-2 border-[#6496c8] rounded-xl py-2.5 px-2 cursor-pointer transition-all duration-200 min-w-[60px] hover:-translate-y-0.5"
          style={{ boxShadow: "0 0 10px #6496c840" }}
          onMouseEnter={(e) => {
            e.currentTarget.style.boxShadow =
              "0 6px 20px rgba(0,0,0,0.4), 0 0 20px #6496c8";
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.boxShadow = "0 0 10px #6496c840";
          }}
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
          className="flex flex-col items-center gap-1 bg-space-black-darker/90 border-2 border-[#9664ff] rounded-xl py-2.5 px-2 cursor-pointer transition-all duration-200 min-w-[60px] hover:-translate-y-0.5"
          style={{ boxShadow: "0 0 10px #9664ff40" }}
          onMouseEnter={(e) => {
            e.currentTarget.style.boxShadow =
              "0 6px 20px rgba(0,0,0,0.4), 0 0 20px #9664ff";
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.boxShadow = "0 0 10px #9664ff40";
          }}
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
          className="flex flex-col items-center gap-1 bg-space-black-darker/90 border-2 border-[#ffc864] rounded-xl py-2.5 px-2 cursor-pointer transition-all duration-200 min-w-[60px] hover:-translate-y-0.5"
          style={{ boxShadow: "0 0 10px #ffc86440" }}
          onMouseEnter={(e) => {
            e.currentTarget.style.boxShadow =
              "0 6px 20px rgba(0,0,0,0.4), 0 0 20px #ffc864";
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.boxShadow = "0 0 10px #ffc86440";
          }}
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
