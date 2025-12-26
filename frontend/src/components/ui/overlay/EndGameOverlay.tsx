import { FC, useState, useCallback, useEffect, useRef } from "react";
import type { GameDto, CardVPDetailDto } from "../../../types/generated/api-types";
import {
  ANIMATION_TIMINGS,
  VPSequencePhase,
  TileHighlightType,
} from "../../../constants/gameConstants";
import VPBarChart from "../endgame/VPBarChart";
import CardVPOverlay from "../endgame/CardVPOverlay";
import PlayerVPCard, { TileHoverType } from "../endgame/PlayerVPCard";
import MilestoneCompact from "../endgame/MilestoneCompact";
import AwardCompact from "../endgame/AwardCompact";
import CardVPHoverModal from "../endgame/CardVPHoverModal";
import { PRIMARY_BUTTON_CLASS } from "./overlayStyles";

/** VP indicator to show floating above a tile */
export interface TileVPIndicator {
  coordinate: string;
  amount: number;
  type: "greenery" | "city-adjacency";
  playerId: string;
  isAnimating: boolean;
  showVPText: boolean; // Whether to show the floating +X text
}

/** State for per-player tile counting */
interface TileCountingState {
  currentPlayerIndex: number;
  phase: "greenery" | "cities" | "done";
  currentTileIndex: number;
  activeTileCoord: string | null;
  adjacentCoords: string[];
}

/** State for card VP cycling */
interface CardVPState {
  currentPlayerIndex: number;
  currentCardIndex: number;
  allCardsWithVP: CardVPDetailDto[];
  totalCardVP: number;
}

interface EndGameOverlayProps {
  /** Current game state */
  game: GameDto;
  /** Current player's ID */
  playerId: string;
  /** Callback to control tile highlighting on 3D board */
  onTileHighlight?: (type: TileHighlightType) => void;
  /** Callback to send VP indicators to 3D board */
  onVPIndicators?: (indicators: TileVPIndicator[]) => void;
  /** Callback when user clicks to return to menu */
  onReturnToMenu?: () => void;
}

const TILE_VP_DELAY_MS = ANIMATION_TIMINGS.TILE_REVEAL;

/**
 * EndGameOverlay - Mars-Centered End Game VP Display
 *
 * Right sidebar showing VP breakdown while Mars board displays
 * tile-by-tile VP counting with floating numbers.
 *
 * Phases:
 * 1. intro - "Game Complete" title
 * 2. tr - Terraform Rating per player
 * 3. milestones - Milestone VP badges
 * 4. awards - Award placements (gold/silver)
 * 5. tiles - Per-player greenery + city VP on Mars board
 * 6. cards - Card VP with center overlay (math breakdown)
 * 7. summary - Vertical bar chart
 * 8. rankings - Winner announcement
 */
const EndGameOverlay: FC<EndGameOverlayProps> = ({
  game,
  playerId,
  onTileHighlight,
  onVPIndicators,
  onReturnToMenu,
}) => {
  const [currentPhase, setCurrentPhase] = useState<VPSequencePhase>("intro");
  const [tileCountingState, setTileCountingState] = useState<TileCountingState | null>(null);
  const [vpIndicators, setVPIndicators] = useState<TileVPIndicator[]>([]);
  const [cardVPState, setCardVPState] = useState<CardVPState | null>(null);
  const [isDrawerHidden, setIsDrawerHidden] = useState(false);
  const timerRef = useRef<NodeJS.Timeout | null>(null);

  // Hover states for showing VP indicators on badge hover
  const [hoveredTileType, setHoveredTileType] = useState<TileHoverType | null>(null);
  const [hoveredCardPlayerId, setHoveredCardPlayerId] = useState<string | null>(null);

  const allScores = game.finalScores ?? [];
  const sortedScores = [...allScores].sort((a, b) => b.vpBreakdown.totalVP - a.vpBreakdown.totalVP);

  // Cleanup timer on unmount
  useEffect(() => {
    return () => {
      if (timerRef.current) {
        clearTimeout(timerRef.current);
      }
    };
  }, []);

  // Send VP indicators to parent for 3D rendering
  useEffect(() => {
    onVPIndicators?.(vpIndicators);
  }, [vpIndicators, onVPIndicators]);

  // Generate VP indicators when hovering on tile badges
  // Works after tiles phase completes, or during tiles phase when not actively counting
  useEffect(() => {
    if (!hoveredTileType) return;

    // Don't override active tile counting animation
    if (currentPhase === "tiles" && tileCountingState) return;

    const score = sortedScores.find((s) => s.playerId === hoveredTileType.playerId);
    if (!score) return;

    const hoverIndicators: TileVPIndicator[] = [];

    if (hoveredTileType.type === "greenery") {
      for (const detail of score.vpBreakdown.greeneryVPDetails ?? []) {
        hoverIndicators.push({
          coordinate: detail.coordinate,
          amount: detail.vp,
          type: "greenery",
          playerId: hoveredTileType.playerId,
          isAnimating: false,
          showVPText: true,
        });
      }
    } else {
      for (const detail of score.vpBreakdown.cityVPDetails ?? []) {
        hoverIndicators.push({
          coordinate: detail.cityCoordinate,
          amount: detail.vp,
          type: "city-adjacency",
          playerId: hoveredTileType.playerId,
          isAnimating: false,
          showVPText: true,
        });
      }
    }

    setVPIndicators(hoverIndicators);

    return () => setVPIndicators([]);
  }, [hoveredTileType, sortedScores, currentPhase, tileCountingState]);

  // Phase advancement
  const advanceToPhase = useCallback((nextPhase: VPSequencePhase) => {
    setCurrentPhase(nextPhase);
  }, []);

  // Skip to end
  const skipToEnd = useCallback(() => {
    if (timerRef.current) {
      clearTimeout(timerRef.current);
    }
    setCurrentPhase("complete");
    setTileCountingState(null);
    setCardVPState(null);
    setVPIndicators([]);
    onTileHighlight?.(null);
  }, [onTileHighlight]);

  // Check if there are claimed milestones or funded awards
  const hasClaimedMilestones = (game.milestones ?? []).some((m) => m.isClaimed);
  const hasFundedAwards = (game.awards ?? []).some((a) => a.isFunded);

  // Auto-advance phases
  useEffect(() => {
    if (currentPhase === "intro") {
      timerRef.current = setTimeout(() => advanceToPhase("tr"), ANIMATION_TIMINGS.PHASE_INTRO);
    } else if (currentPhase === "tr") {
      // Skip milestones/awards phases if none exist
      const nextPhase = hasClaimedMilestones ? "milestones" : hasFundedAwards ? "awards" : "tiles";
      timerRef.current = setTimeout(() => advanceToPhase(nextPhase), ANIMATION_TIMINGS.PHASE_TR);
    } else if (currentPhase === "milestones") {
      // Skip awards phase if no funded awards
      const nextPhase = hasFundedAwards ? "awards" : "tiles";
      timerRef.current = setTimeout(
        () => advanceToPhase(nextPhase),
        ANIMATION_TIMINGS.PHASE_MILESTONES,
      );
    } else if (currentPhase === "awards") {
      timerRef.current = setTimeout(() => advanceToPhase("tiles"), ANIMATION_TIMINGS.PHASE_AWARDS);
    } else if (currentPhase === "tiles" && !tileCountingState) {
      // Start per-player tile counting
      startTileCounting();
    } else if (currentPhase === "cards" && !cardVPState) {
      // Start card VP cycling
      startCardVPCycle();
    } else if (currentPhase === "summary") {
      timerRef.current = setTimeout(
        () => advanceToPhase("rankings"),
        ANIMATION_TIMINGS.PHASE_SUMMARY,
      );
    } else if (currentPhase === "rankings") {
      timerRef.current = setTimeout(
        () => advanceToPhase("complete"),
        ANIMATION_TIMINGS.PHASE_RANKINGS,
      );
    }

    return () => {
      if (timerRef.current) {
        clearTimeout(timerRef.current);
      }
    };
  }, [currentPhase, tileCountingState, hasClaimedMilestones, hasFundedAwards]);

  // Backend provides greeneryVPDetails and cityVPDetails in FinalScoreDto
  // No need to calculate locally - use backend data for accuracy

  // Start tile counting sequence
  const startTileCounting = () => {
    if (sortedScores.length === 0) {
      advanceToPhase("cards");
      return;
    }

    setTileCountingState({
      currentPlayerIndex: 0,
      phase: "greenery",
      currentTileIndex: 0,
      activeTileCoord: null,
      adjacentCoords: [],
    });
  };

  // Process next tile in counting sequence using backend VP details
  useEffect(() => {
    if (!tileCountingState || currentPhase !== "tiles") return;

    const { currentPlayerIndex, phase, currentTileIndex } = tileCountingState;
    const currentPlayer = sortedScores[currentPlayerIndex];
    if (!currentPlayer) {
      advanceToPhase("cards");
      return;
    }

    // Use backend-provided VP details instead of calculating locally
    const greeneryDetails = currentPlayer.vpBreakdown.greeneryVPDetails ?? [];
    const cityDetails = currentPlayer.vpBreakdown.cityVPDetails ?? [];

    if (phase === "greenery") {
      if (currentTileIndex < greeneryDetails.length) {
        const greeneryDetail = greeneryDetails[currentTileIndex];
        const coord = greeneryDetail.coordinate;

        // Add to existing indicators (accumulate within greenery phase)
        onTileHighlight?.("greenery");
        setVPIndicators((prev) => [
          ...prev,
          {
            coordinate: coord,
            amount: greeneryDetail.vp,
            type: "greenery",
            playerId: currentPlayer.playerId,
            isAnimating: true,
            showVPText: true,
          },
        ]);

        // Delay before moving to next tile
        timerRef.current = setTimeout(() => {
          // Stop animation on this tile
          setVPIndicators((prev) =>
            prev.map((ind) => (ind.coordinate === coord ? { ...ind, isAnimating: false } : ind)),
          );
          // Move to next tile after brief display
          setTimeout(() => {
            setTileCountingState((prev) =>
              prev
                ? {
                    ...prev,
                    currentTileIndex: currentTileIndex + 1,
                    activeTileCoord: coord,
                  }
                : null,
            );
          }, ANIMATION_TIMINGS.BRIEF_PAUSE);
        }, TILE_VP_DELAY_MS);
      } else {
        // Greenery phase complete - clear and move to cities phase
        timerRef.current = setTimeout(() => {
          setVPIndicators([]);
          setTileCountingState((prev) =>
            prev ? { ...prev, phase: "cities", currentTileIndex: 0 } : null,
          );
        }, ANIMATION_TIMINGS.PHASE_CLEANUP);
      }
    } else if (phase === "cities") {
      if (currentTileIndex < cityDetails.length) {
        const cityDetail = cityDetails[currentTileIndex];
        const cityCoord = cityDetail.cityCoordinate;
        const adjacentGreeneryCoords = cityDetail.adjacentGreeneries ?? [];

        // Highlight city and adjacent greeneries
        onTileHighlight?.("city");

        // Add city indicator (accumulate within cities phase)
        const cityIndicator: TileVPIndicator = {
          coordinate: cityCoord,
          amount: cityDetail.vp,
          type: "city-adjacency" as const,
          playerId: currentPlayer.playerId,
          isAnimating: true,
          showVPText: true,
        };

        // Adjacent greeneries just get highlight, no VP text
        const greeneryIndicators: TileVPIndicator[] = adjacentGreeneryCoords.map((greenCoord) => ({
          coordinate: greenCoord,
          amount: 1,
          type: "city-adjacency" as const,
          playerId: currentPlayer.playerId,
          isAnimating: false,
          showVPText: false,
        }));

        setVPIndicators((prev) => [...prev, cityIndicator, ...greeneryIndicators]);

        // Delay before moving to next city
        timerRef.current = setTimeout(() => {
          // Stop animation on this city
          setVPIndicators((prev) =>
            prev.map((ind) =>
              ind.coordinate === cityCoord ? { ...ind, isAnimating: false } : ind,
            ),
          );
          // Move to next city after brief display
          setTimeout(() => {
            setTileCountingState((prev) =>
              prev
                ? {
                    ...prev,
                    currentTileIndex: currentTileIndex + 1,
                    activeTileCoord: cityCoord,
                    adjacentCoords: adjacentGreeneryCoords,
                  }
                : null,
            );
          }, ANIMATION_TIMINGS.BRIEF_PAUSE);
        }, TILE_VP_DELAY_MS * 2);
      } else {
        // Cities phase complete - clear and move to next player or finish
        timerRef.current = setTimeout(() => {
          setVPIndicators([]);
          if (currentPlayerIndex + 1 < sortedScores.length) {
            setTileCountingState({
              currentPlayerIndex: currentPlayerIndex + 1,
              phase: "greenery",
              currentTileIndex: 0,
              activeTileCoord: null,
              adjacentCoords: [],
            });
          } else {
            // All players done
            setTileCountingState(null);
            onTileHighlight?.(null);
            timerRef.current = setTimeout(
              () => advanceToPhase("cards"),
              ANIMATION_TIMINGS.POST_TILES_DELAY,
            );
          }
        }, ANIMATION_TIMINGS.PHASE_CLEANUP);
      }
    }
  }, [tileCountingState]);

  // Start card VP cycle - collect all cards with VP from all players
  const startCardVPCycle = () => {
    // Collect all cards with VP from all players
    const allCardsWithVP: CardVPDetailDto[] = [];
    let totalCardVP = 0;

    for (const score of sortedScores) {
      const cardDetails = score.vpBreakdown.cardVPDetails ?? [];
      for (const cardDetail of cardDetails) {
        if (cardDetail.totalVP > 0) {
          allCardsWithVP.push(cardDetail);
          totalCardVP += cardDetail.totalVP;
        }
      }
    }

    if (allCardsWithVP.length === 0) {
      // No cards with VP, skip to summary
      advanceToPhase("summary");
      return;
    }

    setCardVPState({
      currentPlayerIndex: 0,
      currentCardIndex: 0,
      allCardsWithVP,
      totalCardVP,
    });
  };

  // Process card VP cycling
  useEffect(() => {
    if (!cardVPState || currentPhase !== "cards") return;

    const { currentCardIndex, allCardsWithVP } = cardVPState;

    if (currentCardIndex >= allCardsWithVP.length) {
      // All cards shown, advance to summary
      setCardVPState(null);
      timerRef.current = setTimeout(
        () => advanceToPhase("summary"),
        ANIMATION_TIMINGS.PHASE_CLEANUP,
      );
      return;
    }

    timerRef.current = setTimeout(() => {
      setCardVPState((prev) => (prev ? { ...prev, currentCardIndex: currentCardIndex + 1 } : null));
    }, ANIMATION_TIMINGS.CARD_VP_DISPLAY);
  }, [cardVPState, currentPhase, advanceToPhase]);

  // Error state
  if (allScores.length === 0) {
    return (
      <div className="fixed right-0 top-0 bottom-0 w-80 z-[1000] bg-black/90 backdrop-blur-md flex items-center justify-center">
        <div className="text-center p-4">
          <h2 className="font-orbitron text-lg text-red-400 mb-4">No scores available</h2>
          <button onClick={onReturnToMenu} className={PRIMARY_BUTTON_CLASS}>
            Return to Menu
          </button>
        </div>
      </div>
    );
  }

  const winner = sortedScores[0];

  return (
    <>
      {/* Card VP Center Overlay - shown during cards phase */}
      {currentPhase === "cards" && cardVPState && (
        <CardVPOverlay
          cardVPDetails={cardVPState.allCardsWithVP}
          currentCardIndex={cardVPState.currentCardIndex}
          isVisible={true}
          totalCardVP={cardVPState.totalCardVP}
        />
      )}

      {/* Card VP Hover Modal - shown when hovering card badge */}
      {hoveredCardPlayerId && currentPhase !== "cards" && (
        <CardVPHoverModal
          playerScore={sortedScores.find((s) => s.playerId === hoveredCardPlayerId)}
        />
      )}

      {/* Toggle button - visible when drawer is hidden */}
      {isDrawerHidden && (
        <button
          onClick={() => setIsDrawerHidden(false)}
          className="fixed right-0 top-1/2 -translate-y-1/2 z-[1001] bg-black/85 backdrop-blur-md border border-white/10 border-r-0 rounded-l-lg px-2 py-4 text-white/70 hover:text-white hover:bg-black/95 transition-colors"
          title="Show Results"
        >
          <span className="text-lg">◀</span>
        </button>
      )}

      {/* Right Sidebar */}
      <div
        className={`fixed right-0 top-0 bottom-0 w-80 z-[1000] bg-black/85 backdrop-blur-md border-l border-white/10 overflow-y-auto transition-transform duration-300 ${
          isDrawerHidden ? "translate-x-full" : "translate-x-0"
        }`}
      >
        {/* Header */}
        <div className="sticky top-0 bg-black/90 backdrop-blur p-4 border-b border-white/10">
          <div className="flex items-center justify-between">
            {/* Hide button */}
            <button
              onClick={() => setIsDrawerHidden(true)}
              className="text-white/50 hover:text-white p-1 -ml-1"
              title="Hide Results"
            >
              <span className="text-lg">▶</span>
            </button>
            <h1 className="font-orbitron text-xl font-bold text-amber-400 flex-1 text-center">
              {currentPhase === "intro"
                ? "Game Complete"
                : currentPhase === "complete"
                  ? "Final Results"
                  : "Counting VP..."}
            </h1>
            {currentPhase !== "complete" && (
              <button
                onClick={skipToEnd}
                className="text-xs text-white/50 hover:text-white px-2 py-1 border border-white/20 rounded"
              >
                Skip
              </button>
            )}
          </div>

          {/* Phase indicator */}
          <div className="mt-2 text-sm text-white/60">
            {currentPhase === "tr" && "Terraform Rating"}
            {currentPhase === "milestones" && "Milestones"}
            {currentPhase === "awards" && "Awards"}
            {currentPhase === "tiles" && (
              <>
                Tile VP:{" "}
                {tileCountingState && sortedScores[tileCountingState.currentPlayerIndex] && (
                  <span className="text-white">
                    {sortedScores[tileCountingState.currentPlayerIndex].playerName}
                  </span>
                )}
              </>
            )}
            {currentPhase === "cards" && "Card VP"}
            {currentPhase === "summary" && "Final Scores"}
            {currentPhase === "rankings" && "Rankings"}
          </div>
        </div>

        {/* Main content */}
        <div className="p-4 space-y-4">
          {/* Player VP Summary - always visible */}
          <div className="space-y-3">
            {sortedScores.map((score, idx) => (
              <PlayerVPCard
                key={score.playerId}
                score={score}
                placement={idx + 1}
                isCurrentPlayer={score.playerId === playerId}
                currentPhase={currentPhase}
                isCountingTiles={
                  tileCountingState?.currentPlayerIndex === idx && currentPhase === "tiles"
                }
                onHoverTileType={setHoveredTileType}
                onHoverCardVP={setHoveredCardPlayerId}
              />
            ))}
          </div>

          {/* Milestones compact display */}
          {(currentPhase === "milestones" ||
            (currentPhase !== "intro" && currentPhase !== "tr")) && (
            <MilestoneCompact milestones={game.milestones ?? []} scores={sortedScores} />
          )}

          {/* Awards compact display */}
          {(currentPhase === "awards" ||
            ["tiles", "cards", "summary", "rankings", "complete"].includes(currentPhase)) && (
            <AwardCompact
              awards={game.awards ?? []}
              scores={sortedScores}
              awardResults={game.awardResults}
              playerId={playerId}
            />
          )}

          {/* Summary phase - show bar chart */}
          {(currentPhase === "summary" ||
            currentPhase === "rankings" ||
            currentPhase === "complete") && (
            <div className="mt-4">
              <VPBarChart
                scores={allScores}
                isAnimating={currentPhase === "summary"}
                vertical={true}
              />
            </div>
          )}

          {/* Rankings phase */}
          {(currentPhase === "rankings" || currentPhase === "complete") && (
            <div className="text-center py-4">
              <p className="text-amber-400 font-orbitron text-lg winner-glow-animate">
                {winner.playerName} Wins!
              </p>
              <p className="text-white/60 text-sm">{winner.vpBreakdown.totalVP} VP</p>
            </div>
          )}

          {/* Return to Menu */}
          {currentPhase === "complete" && (
            <div className="pt-4">
              <button onClick={onReturnToMenu} className={`${PRIMARY_BUTTON_CLASS} w-full`}>
                Return to Menu
              </button>
            </div>
          )}
        </div>
      </div>
    </>
  );
};

export default EndGameOverlay;
