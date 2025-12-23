import { FC, useEffect, useState } from "react";
import { MilestoneDto } from "../../../types/generated/api-types";
import GameIcon from "../display/GameIcon";

interface MilestoneSectionProps {
  /** All milestones with claim status */
  milestones: MilestoneDto[];
  /** Current player's ID */
  playerId: string;
  /** Whether to animate the display */
  isAnimating: boolean;
  /** Callback when animation completes */
  onAnimationComplete?: () => void;
}

const MILESTONE_VP = 5;

/**
 * MilestoneSection - Displays milestone badges with VP awards
 */
const MilestoneSection: FC<MilestoneSectionProps> = ({
  milestones,
  playerId,
  isAnimating,
  onAnimationComplete,
}) => {
  const [animatedIndex, setAnimatedIndex] = useState(-1);
  const playerMilestones = milestones.filter((m) => m.claimedBy === playerId);
  const totalMilestoneVP = playerMilestones.length * MILESTONE_VP;

  useEffect(() => {
    if (!isAnimating) return;

    // Animate milestones one by one
    let currentIndex = 0;
    const interval = setInterval(() => {
      if (currentIndex < milestones.length) {
        setAnimatedIndex(currentIndex);
        currentIndex++;
      } else {
        clearInterval(interval);
        onAnimationComplete?.();
      }
    }, 400);

    return () => clearInterval(interval);
  }, [isAnimating, milestones.length, onAnimationComplete]);

  const getMilestoneIconType = (type: string): string => {
    switch (type) {
      case "terraformer":
        return "tr";
      case "mayor":
        return "city-tile";
      case "gardener":
        return "greenery-tile";
      case "builder":
        return "building";
      case "planner":
        return "card";
      default:
        return "milestone";
    }
  };

  return (
    <div className="section-slide-in-animate flex flex-col items-center gap-4 p-6">
      <h3 className="font-orbitron text-lg text-white/80 uppercase tracking-wider">Milestones</h3>

      <div className="flex flex-wrap justify-center gap-3">
        {milestones.map((milestone, index) => {
          const isPlayerMilestone = milestone.claimedBy === playerId;
          const isRevealed = !isAnimating || index <= animatedIndex;
          const isClaimed = milestone.isClaimed;

          return (
            <div
              key={milestone.type}
              className={`
                relative flex flex-col items-center p-4 rounded-lg border-2 transition-all duration-300
                ${isRevealed ? "opacity-100 translate-y-0" : "opacity-0 translate-y-4"}
                ${
                  isPlayerMilestone && isRevealed
                    ? "border-amber-400 bg-amber-400/20 winner-glow-animate"
                    : isClaimed
                      ? "border-gray-500 bg-gray-500/20"
                      : "border-gray-700 bg-gray-800/50"
                }
              `}
              style={{ animationDelay: `${index * 100}ms` }}
            >
              {/* Milestone icon */}
              <div className="mb-2">
                <GameIcon iconType={getMilestoneIconType(milestone.type)} size="medium" />
              </div>

              {/* Milestone name */}
              <span
                className={`font-orbitron text-sm ${
                  isPlayerMilestone ? "text-amber-400" : "text-white/70"
                }`}
              >
                {milestone.name}
              </span>

              {/* Claimed by indicator */}
              {isClaimed && (
                <span className="text-xs text-white/50 mt-1">
                  {isPlayerMilestone ? "You" : "Claimed"}
                </span>
              )}

              {/* VP badge for player's milestones */}
              {isPlayerMilestone && isRevealed && (
                <div className="absolute -top-2 -right-2 bg-amber-400 text-black font-bold text-xs px-2 py-1 rounded-full float-up-animate">
                  +{MILESTONE_VP} VP
                </div>
              )}
            </div>
          );
        })}
      </div>

      {/* Total milestone VP */}
      <div className="flex items-center gap-2 mt-4">
        <span className="text-white/60">Total:</span>
        <span className="text-2xl font-orbitron font-bold text-amber-400">
          {totalMilestoneVP} VP
        </span>
      </div>
    </div>
  );
};

export default MilestoneSection;
