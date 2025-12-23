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

  const revealedTotal =
    (showTR ? vpBreakdown.terraformRating : 0) +
    (showMilestones ? vpBreakdown.milestoneVP : 0) +
    (showAwards ? vpBreakdown.awardVP : 0) +
    (showTiles ? vpBreakdown.greeneryVP + vpBreakdown.cityVP : 0) +
    (showCards ? vpBreakdown.cardVP : 0);

  const placementClasses =
    placement === 1
      ? "bg-amber-400 text-black"
      : placement === 2
        ? "bg-gray-300 text-black"
        : placement === 3
          ? "bg-amber-700 text-white"
          : "bg-gray-600 text-white";

  return (
    <div
      className={`
        p-3 rounded-lg border transition-all duration-300
        ${isCurrentPlayer ? "border-amber-400/50 bg-amber-400/10" : "border-white/10 bg-white/5"}
        ${isCountingTiles ? "ring-2 ring-green-400/50" : ""}
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
        </div>
        <span className="font-orbitron text-xl font-bold text-white">{revealedTotal}</span>
      </div>

      {/* VP breakdown row */}
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
              value={vpBreakdown.greeneryVP}
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
              value={vpBreakdown.cityVP}
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
  );
};

export default PlayerVPCard;
