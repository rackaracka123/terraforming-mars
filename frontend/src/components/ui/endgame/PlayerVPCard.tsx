import { FC } from "react";
import type { FinalScoreDto } from "../../../types/generated/api-types";
import { VPSequencePhase, isPhaseAtOrAfter } from "../../../constants/gameConstants";
import VPBadge from "./VPBadge";

/** Hover type for tile VP indicators */
export interface TileHoverType {
  playerId: string;
  type: "greenery" | "city";
}

interface PlayerVPCardProps {
  score: FinalScoreDto;
  placement: number;
  isCurrentPlayer: boolean;
  currentPhase: VPSequencePhase;
  isCountingTiles: boolean;
  /** Live-updating greenery VP during tile counting animation */
  revealedGreeneryVP?: number;
  /** Live-updating city VP during tile counting animation */
  revealedCityVP?: number;
  onHoverTileType?: (hover: TileHoverType | null) => void;
  onHoverCardVP?: (playerId: string | null) => void;
}

/** Compact player VP card showing progressive VP reveal */
const PlayerVPCard: FC<PlayerVPCardProps> = ({
  score,
  placement,
  isCurrentPlayer,
  currentPhase,
  isCountingTiles,
  revealedGreeneryVP,
  revealedCityVP,
  onHoverTileType,
  onHoverCardVP,
}) => {
  const { vpBreakdown } = score;

  // Determine which VP categories are revealed based on phase
  const showTR = isPhaseAtOrAfter(currentPhase, "tr");
  const showMilestones = isPhaseAtOrAfter(currentPhase, "milestones");
  const showAwards = isPhaseAtOrAfter(currentPhase, "awards");
  const showTiles = isPhaseAtOrAfter(currentPhase, "tiles");
  const showCards = isPhaseAtOrAfter(currentPhase, "cards");

  // Use live-updating values during tiles phase if provided, otherwise use final values
  const displayGreeneryVP = revealedGreeneryVP ?? vpBreakdown.greeneryVP;
  const displayCityVP = revealedCityVP ?? vpBreakdown.cityVP;

  const revealedTotal =
    (showTR ? vpBreakdown.terraformRating : 0) +
    (showMilestones ? vpBreakdown.milestoneVP : 0) +
    (showAwards ? vpBreakdown.awardVP : 0) +
    (showTiles ? displayGreeneryVP + displayCityVP : 0) +
    (showCards ? vpBreakdown.cardVP : 0);

  const placementClasses =
    placement === 1
      ? "bg-amber-400 text-black"
      : placement === 2
        ? "bg-gray-300 text-black"
        : placement === 3
          ? "bg-amber-700 text-white"
          : "bg-gray-600 text-white";

  // Determine background/border styling based on state
  // When counting tiles, "light up" the existing accent colors
  const getCardStyles = () => {
    if (isCurrentPlayer) {
      return isCountingTiles
        ? "border-amber-400/70 bg-amber-400/20" // Lit up amber
        : "border-amber-400/50 bg-amber-400/10"; // Default amber
    }
    return isCountingTiles
      ? "border-white/30 bg-white/10" // Lit up white/gray
      : "border-white/10 bg-white/5"; // Default white/gray
  };

  return (
    <div
      className={`
        p-3 rounded-lg border transition-all duration-300
        ${getCardStyles()}
        ${placement === 1 && currentPhase === "complete" ? "winner-glow-animate" : ""}
      `}
    >
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-2">
          <span
            className={`w-6 h-6 rounded-full flex items-center justify-center text-xs font-bold ${placementClasses}`}
          >
            {placement}
          </span>
          <span
            className={`font-orbitron text-sm ${isCurrentPlayer ? "text-amber-400" : "text-white"}`}
          >
            {score.playerName}
          </span>
          {isCurrentPlayer && (
            <span className="px-1.5 py-px rounded-lg text-[8px] font-semibold uppercase tracking-[0.3px] shadow-[0_1px_2px_rgba(0,0,0,0.2)] bg-[linear-gradient(135deg,#00d4ff,#0099cc)] text-white border-2 border-[rgba(0,212,255,0.8)] [text-shadow:0_0_12px_rgba(0,212,255,0.8),0_2px_4px_rgba(0,0,0,0.6)] shadow-[0_0_16px_rgba(0,212,255,0.4),inset_0_1px_0_rgba(255,255,255,0.3)]">
              YOU
            </span>
          )}
        </div>
        <span className="font-orbitron text-xl font-bold text-white">{revealedTotal}</span>
      </div>

      {/* VP breakdown row - uses grid for smooth height animation */}
      <div
        className="grid transition-[grid-template-rows] duration-300 ease-out"
        style={{ gridTemplateRows: showTR ? "1fr" : "0fr" }}
      >
        <div className="overflow-hidden">
          <div className="flex flex-wrap gap-2 text-xs">
            {showTR && (
              <VPBadge icon="terraform-rating" value={vpBreakdown.terraformRating} color="blue" />
            )}
            {showMilestones && vpBreakdown.milestoneVP > 0 && (
              <VPBadge icon="milestone" value={vpBreakdown.milestoneVP} color="purple" />
            )}
            {showAwards && vpBreakdown.awardVP > 0 && (
              <VPBadge icon="award" value={vpBreakdown.awardVP} color="yellow" />
            )}
            {showTiles && (
              <>
                <VPBadge
                  icon="greenery-tile"
                  value={displayGreeneryVP}
                  color="green"
                  onMouseEnter={() =>
                    onHoverTileType?.({
                      playerId: score.playerId,
                      type: "greenery",
                    })
                  }
                  onMouseLeave={() => onHoverTileType?.(null)}
                />
                <VPBadge
                  icon="city-tile"
                  value={displayCityVP}
                  color="gray"
                  onMouseEnter={() => onHoverTileType?.({ playerId: score.playerId, type: "city" })}
                  onMouseLeave={() => onHoverTileType?.(null)}
                />
              </>
            )}
            {showCards && (
              <VPBadge
                icon="card"
                value={vpBreakdown.cardVP}
                color="indigo"
                onMouseEnter={() => onHoverCardVP?.(score.playerId)}
                onMouseLeave={() => onHoverCardVP?.(null)}
              />
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default PlayerVPCard;
