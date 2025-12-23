import { FC } from "react";
import { AwardDto, VPBreakdownDto } from "../../../types/generated/api-types";
import GameIcon from "../display/GameIcon";
import { getAwardIconType } from "../../../utils/achievementIcons";
import { VP_VALUES, ANIMATION_TIMINGS } from "../../../constants/gameConstants";
import { useSequentialAnimation, isItemVisible } from "../../../hooks/useSequentialAnimation";

interface AwardResult {
  awardType: string;
  firstPlaceIds: string[];
  secondPlaceIds: string[];
}

interface AwardSectionProps {
  /** All awards with funding status */
  awards: AwardDto[];
  /** Award results showing 1st and 2nd place for each funded award */
  awardResults?: AwardResult[];
  /** Current player's ID */
  playerId: string;
  /** Player's VP breakdown (for award VP) */
  vpBreakdown: VPBreakdownDto;
  /** Whether to animate the display */
  isAnimating: boolean;
  /** Callback when animation completes */
  onAnimationComplete?: () => void;
}

/**
 * AwardSection - Displays award badges with placements and VP
 */
const AwardSection: FC<AwardSectionProps> = ({
  awards,
  awardResults = [],
  playerId,
  vpBreakdown,
  isAnimating,
  onAnimationComplete,
}) => {
  const animatedIndex = useSequentialAnimation(
    awards.length,
    ANIMATION_TIMINGS.SECTION_TRANSITION,
    isAnimating,
    onAnimationComplete,
  );

  const getPlayerPlacement = (awardType: string): "first" | "second" | null => {
    const result = awardResults.find((r) => r.awardType === awardType);
    if (!result) return null;
    if (result.firstPlaceIds.includes(playerId)) return "first";
    if (result.secondPlaceIds.includes(playerId)) return "second";
    return null;
  };

  return (
    <div className="section-slide-in-animate flex flex-col items-center gap-4 p-6">
      <h3 className="font-orbitron text-lg text-white/80 uppercase tracking-wider">Awards</h3>

      <div className="flex flex-wrap justify-center gap-3">
        {awards.map((award, index) => {
          const isRevealed = isItemVisible(index, animatedIndex, isAnimating);
          const placement = getPlayerPlacement(award.type);
          const isFunded = award.isFunded;

          return (
            <div
              key={award.type}
              className={`
                relative flex flex-col items-center p-4 rounded-lg border-2 transition-all duration-300
                ${isRevealed ? "opacity-100 translate-y-0" : "opacity-0 translate-y-4"}
                ${
                  placement === "first" && isRevealed
                    ? "border-amber-400 bg-amber-400/20 winner-glow-animate"
                    : placement === "second" && isRevealed
                      ? "border-gray-300 bg-gray-300/20"
                      : isFunded
                        ? "border-gray-500 bg-gray-500/20"
                        : "border-gray-700 bg-gray-800/50"
                }
              `}
            >
              {/* Award icon */}
              <div className="mb-2">
                <GameIcon iconType={getAwardIconType(award.type)} size="medium" />
              </div>

              {/* Award name */}
              <span
                className={`font-orbitron text-sm ${
                  placement === "first"
                    ? "text-amber-400"
                    : placement === "second"
                      ? "text-gray-300"
                      : "text-white/70"
                }`}
              >
                {award.name}
              </span>

              {/* Placement indicator */}
              {isFunded && placement && isRevealed && (
                <span
                  className={`text-xs mt-1 font-bold ${
                    placement === "first" ? "text-amber-400" : "text-gray-300"
                  }`}
                >
                  {placement === "first" ? "1st Place" : "2nd Place"}
                </span>
              )}

              {/* Not funded indicator */}
              {!isFunded && <span className="text-xs text-white/40 mt-1">Not funded</span>}

              {/* VP badge */}
              {placement && isRevealed && (
                <div
                  className={`absolute -top-2 -right-2 font-bold text-xs px-2 py-1 rounded-full float-up-animate ${
                    placement === "first" ? "bg-amber-400 text-black" : "bg-gray-300 text-black"
                  }`}
                >
                  +{placement === "first" ? VP_VALUES.AWARD_FIRST : VP_VALUES.AWARD_SECOND} VP
                </div>
              )}
            </div>
          );
        })}
      </div>

      {/* Total award VP */}
      <div className="flex items-center gap-2 mt-4">
        <span className="text-white/60">Total:</span>
        <span className="text-2xl font-orbitron font-bold text-amber-400">
          {vpBreakdown.awardVP} VP
        </span>
      </div>
    </div>
  );
};

export default AwardSection;
