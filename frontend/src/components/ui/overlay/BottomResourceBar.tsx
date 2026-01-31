import React, { useRef, useState } from "react";
import {
  PlayerDto,
  PlayerActionDto,
  GameDto,
  CardDto,
  ResourceTypeCredit,
  ResourceTypeSteel,
  ResourceTypeTitanium,
  ResourceTypePlant,
  ResourceTypeEnergy,
  ResourceTypeHeat,
  ResourceTypeGreeneryTile,
  ResourceTypeTemperature,
} from "@/types/generated/api-types.ts";
import ActionsPopover from "../popover/ActionsPopover.tsx";
import EffectsPopover from "../popover/EffectsPopover.tsx";
import TagsPopover from "../popover/TagsPopover.tsx";
import StoragesPopover from "../popover/StoragesPopover.tsx";
import VictoryPointsPopover from "../popover/VictoryPointsPopover.tsx";
import GameIcon from "../display/GameIcon.tsx";
import {
  calculatePlantsForGreenery,
  calculateHeatForTemperature,
} from "@/utils/resourceConversionUtils.ts";
// Modal components are now imported and managed in GameInterface

interface ResourceData {
  id: string;
  name: string;
  current: number;
  production: number;
  color: string;
}

export interface BottomResourceBarCallbacks {
  onOpenCardEffectsModal?: () => void;
  onOpenCardsPlayedModal?: () => void;
  onOpenActionsModal?: () => void;
  onActionSelect?: (action: PlayerActionDto) => void;
  onConvertPlantsToGreenery?: () => void;
  onConvertHeatToTemperature?: () => void;
}

interface BottomResourceBarProps {
  currentPlayer?: PlayerDto | null;
  gameState?: GameDto;
  playedCards?: CardDto[];
  changedPaths?: Set<string>;
  callbacks?: BottomResourceBarCallbacks;
}

const BottomResourceBar: React.FC<BottomResourceBarProps> = ({
  currentPlayer,
  gameState,
  playedCards = [],
  changedPaths = new Set(),
  callbacks = {},
}) => {
  const {
    onOpenCardEffectsModal,
    onOpenCardsPlayedModal,
    onOpenActionsModal,
    onActionSelect,
    onConvertPlantsToGreenery,
    onConvertHeatToTemperature,
  } = callbacks;
  const [showActionsPopover, setShowActionsPopover] = useState(false);
  const [showEffectsPopover, setShowEffectsPopover] = useState(false);
  const [showTagsPopover, setShowTagsPopover] = useState(false);
  const [showStoragesPopover, setShowStoragesPopover] = useState(false);
  const [showVPPopover, setShowVPPopover] = useState(false);
  const actionsButtonRef = useRef<HTMLButtonElement>(null);
  const effectsButtonRef = useRef<HTMLButtonElement>(null);
  const tagsButtonRef = useRef<HTMLButtonElement>(null);
  const storagesButtonRef = useRef<HTMLButtonElement>(null);
  const vpButtonRef = useRef<HTMLButtonElement>(null);

  // Helper function to check if a path has changed
  const hasPathChanged = (path: string): boolean => {
    return changedPaths.has(path);
  };

  // Map resource ID to ResourceType constant
  const getResourceType = (resourceId: string): string => {
    const resourceTypeMap: Record<string, string> = {
      credits: ResourceTypeCredit,
      steel: ResourceTypeSteel,
      titanium: ResourceTypeTitanium,
      plants: ResourceTypePlant,
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
      id: "credit",
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
      id: "plant",
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

  // Calculate required resources for conversions
  const requiredPlants = calculatePlantsForGreenery(currentPlayer?.effects);
  const requiredHeat = calculateHeatForTemperature(currentPlayer?.effects);

  // Check if player can afford conversions
  const canConvertPlants = (currentPlayer?.resources.plants ?? 0) >= requiredPlants;
  const canConvertHeat =
    (currentPlayer?.resources.heat ?? 0) >= requiredHeat &&
    (gameState?.globalParameters?.temperature ?? -30) < 8; // Same check as Asteroid standard project

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

  const totalVP = (currentPlayer?.vpGranters || []).reduce((sum, g) => sum + g.computedValue, 0);

  const handleOpenVPPopover = () => {
    setShowVPPopover(!showVPPopover);
  };

  // Modal escape handling is now managed in GameInterface

  return (
    <div className="fixed bottom-0 left-0 right-0 h-12 flex items-end justify-between px-[30px] pb-2 z-[1000] pointer-events-auto">
      {/* Background bar */}
      <div className="absolute inset-0 bg-space-black-darker/95 backdrop-blur-space border-t-2 border-space-blue-400 shadow-[0_-8px_32px_rgba(0,0,0,0.6),0_0_20px_rgba(30,60,150,0.3)] -z-10" />

      {/* Resource Grid */}
      <div className="flex-[2] -translate-y-[30px] pointer-events-auto relative">
        <div className="grid grid-cols-6 gap-[15px] max-w-[500px] items-end">
          {playerResources.map((resource) => {
            const resourceChanged = hasPathChanged(`currentPlayer.resources.${resource.id}`);
            const productionChanged = hasPathChanged(`currentPlayer.production.${resource.id}`);

            // Check if this resource has a conversion button
            const showConversionButton =
              (resource.id === "plant" && canConvertPlants) ||
              (resource.id === "heat" && canConvertHeat);

            // Disable conversion buttons during tile placement
            const isTilePlacementActive = !!currentPlayer?.pendingTileSelection;
            const isConversionDisabled = isTilePlacementActive;

            return (
              <div key={resource.id} className="flex flex-col items-center">
                <div
                  className={`flex flex-col items-center gap-1.5 bg-space-black-darker/90 border-2 rounded-xl p-2 relative overflow-hidden transition-all duration-500 [transition-timing-function:cubic-bezier(0.34,1.56,0.64,1)]`}
                  style={
                    {
                      "--resource-color": resource.color,
                      borderColor: resource.color,
                      boxShadow: `0 0 10px ${resource.color}40`,
                    } as React.CSSProperties
                  }
                >
                  {/* Conversion button area - all resources have this for consistent height */}
                  <div
                    className={`transition-all duration-500 [transition-timing-function:cubic-bezier(0.34,1.56,0.64,1)] ${showConversionButton ? "max-h-[40px] h-auto opacity-100 mb-1" : "max-h-0 h-0 opacity-0 mb-0"}`}
                  >
                    {(resource.id === "plant" || resource.id === "heat") && (
                      <button
                        disabled={isConversionDisabled || !showConversionButton}
                        className={`flex items-center justify-center gap-0.5 px-2 py-1 bg-space-black-darker/95 border rounded transition-all duration-200 ${isConversionDisabled || !showConversionButton ? "opacity-40 cursor-not-allowed" : "cursor-pointer"}`}
                        style={{
                          borderColor: resource.color,
                          boxShadow: "none",
                        }}
                        onClick={(e) => {
                          e.stopPropagation();
                          if (isConversionDisabled || !showConversionButton) return;
                          // Execute conversion immediately
                          if (resource.id === "plant") {
                            void onConvertPlantsToGreenery?.();
                          } else if (resource.id === "heat") {
                            void onConvertHeatToTemperature?.();
                          }
                        }}
                        onMouseEnter={(e) => {
                          if (!isConversionDisabled && showConversionButton) {
                            e.currentTarget.style.boxShadow = `0 0 8px ${resource.color}, 0 0 16px ${resource.color}, 0 3px 6px rgba(0, 0, 0, 0.3)`;
                          }
                        }}
                        onMouseLeave={(e) => {
                          e.currentTarget.style.boxShadow = "none";
                        }}
                      >
                        <span className="text-[15px] font-bold text-white/90">+</span>
                        <GameIcon
                          iconType={
                            resource.id === "plant"
                              ? ResourceTypeGreeneryTile
                              : ResourceTypeTemperature
                          }
                          size="small"
                        />
                      </button>
                    )}
                  </div>

                  <div className="inline-flex items-center justify-center bg-[linear-gradient(135deg,rgba(160,110,60,0.4)_0%,rgba(139,89,42,0.35)_100%)] border border-[rgba(160,110,60,0.5)] rounded px-2 py-1 shadow-[0_1px_3px_rgba(0,0,0,0.2)] mb-1 min-w-[28px]">
                    <span
                      className={`text-sm font-bold text-white [text-shadow:0_1px_2px_rgba(0,0,0,0.8)] leading-none ${productionChanged ? "[animation:valueUpdateShine_0.8s_ease-in-out]" : ""}`}
                    >
                      {resource.production}
                    </span>
                  </div>

                  {resource.id === "credit" ? (
                    <div className="flex items-center gap-1.5 min-w-[52px] justify-center">
                      <GameIcon
                        iconType={ResourceTypeCredit}
                        amount={resource.current}
                        size="medium"
                      />
                    </div>
                  ) : (
                    <div className="flex items-center gap-1.5 min-w-[52px]">
                      <GameIcon iconType={getResourceType(resource.id)} size="medium" />
                      <div
                        className={`text-lg font-bold text-white [text-shadow:0_1px_3px_rgba(0,0,0,0.8)] ${resourceChanged ? "[animation:valueUpdateShine_0.8s_ease-in-out]" : ""}`}
                      >
                        {resource.current}
                      </div>
                    </div>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      </div>

      {/* Action Buttons Section */}
      <div className="flex-1 flex items-center justify-end gap-3 -translate-y-[30px] pointer-events-auto relative">
        <button
          ref={actionsButtonRef}
          className="flex flex-col items-center gap-1 bg-space-black-darker/90 border-2 border-[#ff6464] rounded-xl py-2.5 px-2 cursor-pointer transition-all duration-200 min-w-[60px] hover:-translate-y-0.5"
          style={{ boxShadow: "0 0 10px #ff646440" }}
          onMouseEnter={(e) => {
            e.currentTarget.style.boxShadow = "0 6px 20px rgba(0,0,0,0.4), 0 0 20px #ff6464";
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.boxShadow = "0 0 10px #ff646440";
          }}
          onClick={handleOpenActionsPopover}
        >
          <div
            className="text-base font-bold [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))] flex items-center gap-[2px] h-[32px] w-[32px] justify-center"
            style={{ color: "#ff6464" }}
          >
            <span className="text-[8px] leading-none translate-y-[1px]">●</span>
            <span className="text-[8px] leading-none translate-y-[1px]">●</span>
            <span className="text-[23px] leading-none">→</span>
          </div>
          <div
            className={`text-sm font-bold text-white [text-shadow:0_1px_2px_rgba(0,0,0,0.8)] leading-none ${hasPathChanged("currentPlayer.actions") ? "[animation:valueUpdateShine_0.8s_ease-in-out]" : ""}`}
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
            e.currentTarget.style.boxShadow = "0 6px 20px rgba(0,0,0,0.4), 0 0 20px #ff96ff";
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.boxShadow = "0 0 10px #ff96ff40";
          }}
          onClick={handleOpenEffectsPopover}
        >
          <div
            className="font-bold [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))] flex items-center justify-center h-[32px] w-[32px] relative"
            style={{ color: "#ff96ff" }}
          >
            <div className="absolute w-[26px] h-[26px] rounded-full border-2 border-current" />
            <div className="flex flex-col items-center justify-center relative">
              <span className="text-[10px] leading-none">●</span>
              <span className="text-[10px] leading-none">●</span>
            </div>
          </div>
          <div
            className={`text-sm font-bold text-white [text-shadow:0_1px_2px_rgba(0,0,0,0.8)] leading-none ${hasPathChanged("currentPlayer.effects") ? "[animation:valueUpdateShine_0.8s_ease-in-out]" : ""}`}
          >
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
            e.currentTarget.style.boxShadow = "0 6px 20px rgba(0,0,0,0.4), 0 0 20px #64ff96";
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.boxShadow = "0 0 10px #64ff9640";
          }}
          onClick={handleOpenTagsPopover}
        >
          <div
            className="font-bold [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))] flex items-center justify-center h-[32px] w-[32px] relative"
            style={{ color: "#64ff96" }}
          >
            <div className="absolute w-[26px] h-[26px] rounded-full border-2 border-current" />
            <div className="flex items-center gap-[2px] relative text-[8px] leading-none">
              <span>●</span>
              <span>●</span>
              <span>●</span>
            </div>
          </div>
          <div
            className={`text-sm font-bold text-white [text-shadow:0_1px_2px_rgba(0,0,0,0.8)] leading-none ${hasPathChanged("currentPlayer.playedCards") ? "[animation:valueUpdateShine_0.8s_ease-in-out]" : ""}`}
          >
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
            e.currentTarget.style.boxShadow = "0 6px 20px rgba(0,0,0,0.4), 0 0 20px #6496c8";
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.boxShadow = "0 0 10px #6496c840";
          }}
          onClick={handleOpenStoragesPopover}
        >
          <div
            className="font-bold [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))] flex items-center justify-center h-[32px] w-[32px] relative"
            style={{ color: "#6496c8" }}
          >
            <div className="absolute w-[26px] h-[26px] border-2 border-current" />
            <div className="flex items-center gap-[2px] relative text-[8px] leading-none">
              <span>●</span>
              <span>●</span>
              <span>●</span>
            </div>
          </div>
          <div
            className={`text-sm font-bold text-white [text-shadow:0_1px_2px_rgba(0,0,0,0.8)] leading-none ${hasPathChanged("currentPlayer.resourceStorage") ? "[animation:valueUpdateShine_0.8s_ease-in-out]" : ""}`}
          >
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
            e.currentTarget.style.boxShadow = "0 6px 20px rgba(0,0,0,0.4), 0 0 20px #9664ff";
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.boxShadow = "0 0 10px #9664ff40";
          }}
          onClick={handleOpenCardsModal}
        >
          <div
            className="text-2xl font-bold [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))] flex items-center justify-center h-[32px] w-[32px]"
            style={{ color: "#9664ff" }}
          >
            ↓
          </div>
          <div
            className={`text-sm font-bold text-white [text-shadow:0_1px_2px_rgba(0,0,0,0.8)] leading-none ${hasPathChanged("currentPlayer.playedCards") ? "[animation:valueUpdateShine_0.8s_ease-in-out]" : ""}`}
          >
            {playedCardsCount}
          </div>
          <div className="text-[10px] font-medium text-white/90 uppercase tracking-[0.5px] [text-shadow:0_1px_2px_rgba(0,0,0,0.8)]">
            Played
          </div>
        </button>

        <button
          ref={vpButtonRef}
          className="flex flex-col items-center gap-1 bg-space-black-darker/90 border-2 border-[#ffc864] rounded-xl py-2.5 px-2 cursor-pointer transition-all duration-200 min-w-[60px] hover:-translate-y-0.5"
          style={{ boxShadow: "0 0 10px #ffc86440" }}
          onMouseEnter={(e) => {
            e.currentTarget.style.boxShadow = "0 6px 20px rgba(0,0,0,0.4), 0 0 20px #ffc864";
          }}
          onMouseLeave={(e) => {
            e.currentTarget.style.boxShadow = "0 0 10px #ffc86440";
          }}
          onClick={handleOpenVPPopover}
        >
          <div
            className="font-bold [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))] flex items-center justify-center h-[32px] w-[32px] relative"
            style={{ color: "#ffc864" }}
          >
            <span className="text-3xl absolute">○</span>
            <span className="text-lg absolute">●</span>
          </div>
          <div
            className={`text-sm font-bold text-white [text-shadow:0_1px_2px_rgba(0,0,0,0.8)] leading-none ${
              hasPathChanged("currentPlayer.vpGranters") ||
              hasPathChanged("currentPlayer.terraformRating")
                ? "[animation:valueUpdateShine_0.8s_ease-in-out]"
                : ""
            }`}
          >
            {totalVP}
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

      {/* Victory Points Popover */}
      <VictoryPointsPopover
        isVisible={showVPPopover}
        onClose={() => setShowVPPopover(false)}
        vpGranters={currentPlayer?.vpGranters || []}
        totalVP={totalVP}
        anchorRef={vpButtonRef as React.RefObject<HTMLElement>}
      />
    </div>
  );
};

export default BottomResourceBar;
