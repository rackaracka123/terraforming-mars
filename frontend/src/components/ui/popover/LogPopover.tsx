import React, { useState, useEffect, useRef, useMemo, useCallback } from "react";
import { createPortal } from "react-dom";
import {
  StateDiffDto,
  CardDto,
  GameDto,
  VPConditionDto,
  CalculatedOutputDto,
} from "@/types/generated/api-types.ts";
import { apiService } from "@/services/apiService.ts";
import { globalWebSocketManager } from "@/services/globalWebSocketManager.ts";
import SimpleGameCard from "@/components/ui/cards/SimpleGameCard.tsx";
import GameIcon from "@/components/ui/display/GameIcon.tsx";
import VictoryPointIcon from "@/components/ui/display/VictoryPointIcon.tsx";
import BehaviorSection from "@/components/ui/cards/BehaviorSection";
import { GamePopover, GamePopoverEmpty } from "../GamePopover";

interface LogPopoverProps {
  isVisible: boolean;
  onClose: () => void;
  anchorRef: React.RefObject<HTMLElement>;
  gameId: string;
  gameState?: GameDto;
}

interface CardPreviewPortalProps {
  card: CardDto;
  show: boolean;
}

const CardPreviewPortal: React.FC<CardPreviewPortalProps> = ({ card, show }) => {
  if (!show) return null;

  return createPortal(
    <div
      className="fixed z-[10002] pointer-events-none"
      style={{ right: "400px", top: "50%", transform: "translateY(-50%)" }}
    >
      <div className="transform scale-90 origin-center">
        <SimpleGameCard card={card} isSelected={false} onSelect={() => {}} animationDelay={0} />
      </div>
    </div>,
    document.body
  );
};

const TagIcon: React.FC<{ tag: string }> = ({ tag }) => {
  return <GameIcon iconType={tag} size="small" />;
};

const resourceTypeToIconType: Record<string, string> = {
  credits: "credit",
  steel: "steel",
  titanium: "titanium",
  plants: "plant",
  energy: "energy",
  heat: "heat",
  "credits-production": "credit-production",
  "steel-production": "steel-production",
  "titanium-production": "titanium-production",
  "plants-production": "plant-production",
  "energy-production": "energy-production",
  "heat-production": "heat-production",
  tr: "tr",
  oxygen: "oxygen",
  temperature: "temperature",
  "ocean-placement": "ocean-placement",
  "greenery-placement": "greenery-placement",
  "city-placement": "city-placement",
};

const TILE_PLACEMENT_TYPES = ["ocean-placement", "greenery-placement", "city-placement"];
const GLOBAL_PARAMETER_TYPES = ["temperature", "oxygen"];

const STANDARD_PROJECT_BEHAVIORS: Record<string, { outputs: Array<{ type: string; amount: number; target: string }> }> = {
  "Standard Project: Power Plant": {
    outputs: [{ type: "energy-production", amount: 1, target: "self" }],
  },
  "Standard Project: Asteroid": {
    outputs: [{ type: "temperature", amount: 1, target: "global" }],
  },
  "Standard Project: Aquifer": {
    outputs: [{ type: "ocean-placement", amount: 1, target: "global" }],
  },
  "Standard Project: Greenery": {
    outputs: [{ type: "greenery-placement", amount: 1, target: "global" }],
  },
  "Standard Project: City": {
    outputs: [{ type: "city-placement", amount: 1, target: "global" }],
  },
  "Standard Project: Sell Patents": {
    outputs: [{ type: "credit", amount: 1, target: "self" }],
  },
  "Convert Heat": {
    outputs: [{ type: "temperature", amount: 1, target: "global" }],
  },
  "Convert Plants": {
    outputs: [{ type: "greenery-placement", amount: 1, target: "global" }],
  },
};

const BEHAVIOR_OUTPUT_TYPES = [...TILE_PLACEMENT_TYPES, ...GLOBAL_PARAMETER_TYPES];

const CalculatedOutputsDisplay: React.FC<{ outputs: CalculatedOutputDto[]; showAll?: boolean; excludeBehaviors?: boolean }> = ({ outputs, showAll = false, excludeBehaviors = false }) => {
  const outputsToShow = showAll
    ? outputs.filter(o => o.amount !== 0 && (!excludeBehaviors || !BEHAVIOR_OUTPUT_TYPES.includes(o.resourceType)))
    : outputs.filter(o => o.isScaled && o.amount !== 0 && (!excludeBehaviors || !BEHAVIOR_OUTPUT_TYPES.includes(o.resourceType)));

  if (outputsToShow.length === 0) return null;

  return (
    <div className="mt-1 flex flex-wrap items-center gap-2 px-1">
      <span className="text-[10px] text-gray-400 uppercase tracking-wider">Gained:</span>
      {outputsToShow.map((output, index) => {
        const iconType = resourceTypeToIconType[output.resourceType] || output.resourceType;
        return (
          <div key={index} className="flex items-center gap-0.5">
            <GameIcon iconType={iconType} amount={output.amount} size="small" />
          </div>
        );
      })}
    </div>
  );
};

interface StandardProjectBehaviorDisplayProps {
  projectName: string;
}

const StandardProjectBehaviorDisplay: React.FC<StandardProjectBehaviorDisplayProps> = ({ projectName }) => {
  const behavior = STANDARD_PROJECT_BEHAVIORS[projectName];
  if (!behavior) return null;

  return (
    <div className="mt-1 flex items-center gap-1.5 px-1">
      {behavior.outputs.map((output, index) => {
        const iconType = resourceTypeToIconType[output.type] || output.type;
        return (
          <GameIcon
            key={index}
            iconType={iconType}
            amount={output.amount > 1 ? output.amount : undefined}
            size="small"
          />
        );
      })}
    </div>
  );
};

interface LogEntryProps {
  diff: StateDiffDto;
  cardLookup: Map<string, CardDto>;
  playerNames: Map<string, string>;
}

interface PlayerTurnGroup {
  playerId: string;
  entries: StateDiffDto[];
}

interface LogGroup {
  generation: number;
  playerTurns: PlayerTurnGroup[];
}

const PLAYER_COLORS = [
  { border: "rgba(100, 200, 255, 0.4)", bg: "rgba(100, 200, 255, 0.05)", text: "#64c8ff" },
  { border: "rgba(255, 180, 100, 0.4)", bg: "rgba(255, 180, 100, 0.05)", text: "#ffb464" },
  { border: "rgba(180, 255, 100, 0.4)", bg: "rgba(180, 255, 100, 0.05)", text: "#b4ff64" },
  { border: "rgba(255, 100, 180, 0.4)", bg: "rgba(255, 100, 180, 0.05)", text: "#ff64b4" },
  { border: "rgba(180, 100, 255, 0.4)", bg: "rgba(180, 100, 255, 0.05)", text: "#b464ff" },
];

const groupLogsByGeneration = (logs: StateDiffDto[]): LogGroup[] => {
  if (logs.length === 0) return [];

  const groups: LogGroup[] = [];
  let currentGeneration = 1;

  for (const log of logs) {
    if (log.changes?.generation) {
      currentGeneration = log.changes.generation.new;
    }

    let genGroup = groups.find(g => g.generation === currentGeneration);
    if (!genGroup) {
      genGroup = { generation: currentGeneration, playerTurns: [] };
      groups.push(genGroup);
    }

    const lastTurn = genGroup.playerTurns[genGroup.playerTurns.length - 1];
    if (lastTurn && lastTurn.playerId === log.playerId) {
      lastTurn.entries.push(log);
    } else {
      genGroup.playerTurns.push({ playerId: log.playerId, entries: [log] });
    }
  }

  return groups;
};

interface GenerationDividerProps {
  generation: number;
  entryCount: number;
}

const GenerationDivider: React.FC<GenerationDividerProps> = ({ generation, entryCount }) => (
  <div className="sticky top-0 z-10 flex items-center gap-2 py-2 px-3 bg-[#1a1a2e] border-b border-[rgba(100,200,255,0.3)]">
    <div className="flex items-center gap-1.5">
      <span className="text-[10px] font-bold uppercase tracking-wider text-[#64c8ff]">
        Generation {generation}
      </span>
    </div>
    <div className="flex-1 h-px bg-gradient-to-r from-[rgba(100,200,255,0.3)] to-transparent" />
    <span className="text-[9px] text-gray-500">{entryCount} actions</span>
  </div>
);

interface PlayerTurnSectionProps {
  playerName: string;
  entries: StateDiffDto[];
  colorIndex: number;
  cardLookup: Map<string, CardDto>;
  playerNames: Map<string, string>;
}

const PlayerTurnSection: React.FC<PlayerTurnSectionProps> = ({
  playerName,
  entries,
  colorIndex,
  cardLookup,
  playerNames,
}) => {
  const color = PLAYER_COLORS[colorIndex % PLAYER_COLORS.length];

  return (
    <div
      className="rounded-lg mb-2 overflow-hidden"
      style={{
        borderLeft: `3px solid ${color.border}`,
        backgroundColor: color.bg,
      }}
    >
      <div
        className="flex items-center gap-2 py-1.5 px-3"
        style={{ borderBottom: `1px solid ${color.border}` }}
      >
        <div
          className="w-1.5 h-1.5 rounded-full"
          style={{ backgroundColor: color.text }}
        />
        <span
          className="text-[10px] font-semibold uppercase tracking-wider"
          style={{ color: color.text }}
        >
          {playerName}
        </span>
        <span className="text-[9px] text-gray-500">
          {entries.length} action{entries.length !== 1 ? "s" : ""}
        </span>
      </div>
      <div className="flex flex-col">
        {entries.map((diff) => (
          <LogEntry
            key={diff.sequenceNumber}
            diff={diff}
            cardLookup={cardLookup}
            playerNames={playerNames}
          />
        ))}
      </div>
    </div>
  );
};

const LogEntry: React.FC<LogEntryProps> = ({ diff, cardLookup, playerNames }) => {
  const [showTooltip, setShowTooltip] = useState(false);

  const isCardPlay = diff.sourceType === "card_play";
  const isCardAction = diff.sourceType === "card_action";
  const isStandardProject = diff.sourceType === "standard_project";
  const isResourceConvert = diff.sourceType === "resource_convert";
  const isGameEvent = diff.sourceType === "game_event";
  const isCardSource = isCardPlay || isCardAction;
  const isBehaviorSource = isStandardProject || isResourceConvert || isGameEvent;
  const card = isCardSource ? (cardLookup.get(diff.source) || null) : null;

  const playerName = playerNames.get(diff.playerId) || "Unknown";

  const cardTags = card?.tags || [];
  const vpConditions: VPConditionDto[] = card?.vpConditions || [];

  // For card actions, only show manual action behaviors
  const behaviorsToShow = useMemo(() => {
    if (!card?.behaviors) return [];
    if (isCardAction) {
      return card.behaviors.filter(b => b.triggers?.some(t => t.type === "manual"));
    }
    return card.behaviors;
  }, [card, isCardAction]);

  // Determine the choice display mode
  const choiceDisplayInfo = useMemo(() => {
    if (diff.choiceIndex === undefined || diff.choiceIndex === null) {
      return { hasChoices: false, type: "none" as const };
    }

    // Check if the card has a single behavior with choices array (e.g., Artificial Photosynthesis)
    if (behaviorsToShow.length === 1 && behaviorsToShow[0].choices && behaviorsToShow[0].choices.length > 0) {
      return { hasChoices: true, type: "within-behavior" as const, choices: behaviorsToShow[0].choices };
    }

    // Check if the card has multiple behaviors (OR between behaviors)
    if (isCardPlay && behaviorsToShow.length > 1) {
      return { hasChoices: true, type: "between-behaviors" as const };
    }

    return { hasChoices: false, type: "none" as const };
  }, [diff.choiceIndex, behaviorsToShow, isCardPlay]);

  return (
    <div
      className="relative flex flex-col gap-1 py-2 px-3 hover:bg-white/5 rounded cursor-pointer transition-colors border-b border-[rgba(100,200,255,0.2)] last:border-b-0"
      onMouseEnter={() => setShowTooltip(true)}
      onMouseLeave={() => setShowTooltip(false)}
    >
      <div className="flex items-center gap-2">
        <span className="text-xs text-[#64c8ff] font-medium shrink-0">{playerName}</span>
        <span className="text-sm text-white truncate font-medium">{diff.source}</span>
        {isCardAction && (
          <span className="bg-[linear-gradient(135deg,rgba(255,100,100,0.8)_0%,rgba(200,60,60,0.9)_100%)] text-white/90 text-[8px] font-semibold uppercase tracking-[0.3px] py-0.5 px-1.5 rounded-lg border border-[rgba(255,100,100,0.6)] shrink-0">
            action
          </span>
        )}
        {isCardPlay && cardTags.length > 0 && (
          <div className="flex items-center gap-1 shrink-0">
            {cardTags.map((tag, i) => (
              <TagIcon key={i} tag={tag} />
            ))}
          </div>
        )}
        {isCardPlay && vpConditions.length > 0 && (
          <div className="shrink-0">
            <VictoryPointIcon vpConditions={vpConditions} size="small" />
          </div>
        )}
      </div>

      {choiceDisplayInfo.hasChoices && choiceDisplayInfo.type === "within-behavior" ? (
        // Single behavior with choices array - render each choice separately
        <div className="mt-1 flex flex-col gap-1">
          {choiceDisplayInfo.choices!.map((choice, choiceIndex) => {
            const isChosen = choiceIndex === diff.choiceIndex;
            // Create a synthetic behavior from the choice for display
            const syntheticBehavior = {
              ...behaviorsToShow[0],
              choices: undefined,
              inputs: choice.inputs,
              outputs: choice.outputs,
            };
            return (
              <div
                key={choiceIndex}
                className={`[&>div]:!relative [&>div]:!bottom-auto [&>div]:!left-auto [&>div]:!right-auto [&>div]:w-full [&>div:hover]:!transform-none [&>div:hover]:!shadow-none [&>div:hover]:!filter-none scale-90 origin-left ${!isChosen ? "opacity-40 grayscale" : ""}`}
              >
                <BehaviorSection behaviors={[syntheticBehavior]} greyOutAll={!isChosen} />
              </div>
            );
          })}
        </div>
      ) : choiceDisplayInfo.hasChoices && choiceDisplayInfo.type === "between-behaviors" ? (
        // Multiple behaviors - render each behavior with highlighting based on behavior index
        <div className="mt-1 flex flex-col gap-1">
          {behaviorsToShow.map((behavior, index) => {
            const isChosen = index === diff.choiceIndex;
            return (
              <div
                key={index}
                className={`[&>div]:!relative [&>div]:!bottom-auto [&>div]:!left-auto [&>div]:!right-auto [&>div]:w-full [&>div:hover]:!transform-none [&>div:hover]:!shadow-none [&>div:hover]:!filter-none scale-90 origin-left ${!isChosen ? "opacity-40 grayscale" : ""}`}
              >
                <BehaviorSection behaviors={[behavior]} greyOutAll={!isChosen} />
              </div>
            );
          })}
        </div>
      ) : behaviorsToShow.length > 0 && (
        <div className="mt-1 [&>div]:!relative [&>div]:!bottom-auto [&>div]:!left-auto [&>div]:!right-auto [&>div]:w-full [&>div:hover]:!transform-none [&>div:hover]:!shadow-none [&>div:hover]:!filter-none scale-90 origin-left">
          <BehaviorSection behaviors={behaviorsToShow} />
        </div>
      )}

      {isBehaviorSource && (
        <StandardProjectBehaviorDisplay projectName={diff.source} />
      )}

      {diff.calculatedOutputs && diff.calculatedOutputs.length > 0 && (
        <CalculatedOutputsDisplay
          outputs={diff.calculatedOutputs}
          showAll={!isCardSource || isCardAction}
          excludeBehaviors={isBehaviorSource}
        />
      )}

      {!card && !isCardSource && !isBehaviorSource && (
        <div className="text-xs text-gray-400">{diff.description}</div>
      )}

      {card && <CardPreviewPortal card={card} show={showTooltip} />}
    </div>
  );
};

const LogPopover: React.FC<LogPopoverProps> = ({
  isVisible,
  onClose,
  anchorRef,
  gameId,
  gameState,
}) => {
  const [logs, setLogs] = useState<StateDiffDto[]>([]);
  const [cards, setCards] = useState<CardDto[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const lastSequenceRef = useRef<number>(0);

  const cardLookup = useMemo(() => {
    const lookup = new Map<string, CardDto>();
    cards.forEach((card) => {
      lookup.set(card.name, card);
    });
    return lookup;
  }, [cards]);

  const playerNames = useMemo(() => {
    const names = new Map<string, string>();
    if (gameState?.currentPlayer) {
      names.set(gameState.currentPlayer.id, gameState.currentPlayer.name);
    }
    if (gameState?.otherPlayers) {
      gameState.otherPlayers.forEach((p) => {
        names.set(p.id, p.name);
      });
    }
    return names;
  }, [gameState?.currentPlayer, gameState?.otherPlayers]);

  useEffect(() => {
    const loadCards = async () => {
      try {
        const response = await apiService.listCards(0, 10000);
        setCards(response.cards);
      } catch (error) {
        console.error("Failed to load cards for log panel:", error);
      }
    };
    void loadCards();
  }, []);

  // Handle incoming log updates via WebSocket
  const handleLogUpdate = useCallback((newLogs: StateDiffDto[]) => {
    setLogs((prev) => {
      // Deduplicate by sequence number
      const existingSeqs = new Set(prev.map((l) => l.sequenceNumber));
      const uniqueNewLogs = newLogs.filter((l) => !existingSeqs.has(l.sequenceNumber));
      if (uniqueNewLogs.length === 0) return prev;
      return [...prev, ...uniqueNewLogs];
    });
    if (newLogs.length > 0) {
      const maxSeq = Math.max(...newLogs.map((l) => l.sequenceNumber));
      if (maxSeq > lastSequenceRef.current) {
        lastSequenceRef.current = maxSeq;
      }
    }
    setIsLoading(false);
  }, []);

  // Subscribe to WebSocket log updates
  useEffect(() => {
    globalWebSocketManager.on("log-update", handleLogUpdate);
    return () => {
      globalWebSocketManager.off("log-update", handleLogUpdate);
    };
  }, [handleLogUpdate]);

  // Clear logs when game changes
  useEffect(() => {
    setLogs([]);
    lastSequenceRef.current = 0;
    setIsLoading(true);
  }, [gameId]);

  const groupedLogs = useMemo(() => groupLogsByGeneration(logs), [logs]);

  const playerColorMap = useMemo(() => {
    const map = new Map<string, number>();
    let colorIndex = 0;
    for (const group of groupedLogs) {
      for (const turn of group.playerTurns) {
        if (!map.has(turn.playerId)) {
          map.set(turn.playerId, colorIndex++);
        }
      }
    }
    return map;
  }, [groupedLogs]);

  return (
    <GamePopover
      isVisible={isVisible}
      onClose={onClose}
      position={{ type: "anchor", anchorRef, placement: "above" }}
      theme="log"
      header={{
        title: "Game Log",
        badge: logs.length > 0 ? `${logs.length} entries` : undefined,
      }}
      arrow={{ enabled: true, position: "right", offset: 30 }}
      width={350}
      maxHeight={400}
    >
      {isLoading && logs.length === 0 ? (
        <div className="text-center text-gray-400 py-8">Loading...</div>
      ) : logs.length === 0 ? (
        <GamePopoverEmpty
          icon={<GameIcon iconType="card" size="medium" />}
          title="No log entries yet"
          description="Actions will appear here as the game progresses"
        />
      ) : (
        <div className="flex flex-col">
          {groupedLogs.map((group) => {
            const totalEntries = group.playerTurns.reduce((sum, t) => sum + t.entries.length, 0);
            return (
              <div key={group.generation} className="flex flex-col">
                <GenerationDivider
                  generation={group.generation}
                  entryCount={totalEntries}
                />
                <div className="p-2 flex flex-col">
                  {group.playerTurns.map((turn, turnIndex) => (
                    <PlayerTurnSection
                      key={`${turn.playerId}-${turnIndex}`}
                      playerName={playerNames.get(turn.playerId) || "Unknown"}
                      entries={turn.entries}
                      colorIndex={playerColorMap.get(turn.playerId) || 0}
                      cardLookup={cardLookup}
                      playerNames={playerNames}
                    />
                  ))}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </GamePopover>
  );
};

export default LogPopover;
