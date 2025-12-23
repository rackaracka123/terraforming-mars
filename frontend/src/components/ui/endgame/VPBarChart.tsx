import { FC, useEffect, useState } from "react";
import { FinalScoreDto } from "../../../types/generated/api-types";

interface VPBarChartProps {
  /** Final scores for all players */
  scores: FinalScoreDto[];
  /** Whether to animate the bars */
  isAnimating: boolean;
  /** Callback when animation completes */
  onAnimationComplete?: () => void;
  /** Use compact vertical layout for sidebar */
  vertical?: boolean;
}

const VP_CATEGORIES = [
  { key: "terraformRating", label: "TR", color: "#3b82f6" }, // blue
  { key: "cardVP", label: "Cards", color: "#a855f7" }, // purple
  { key: "milestoneVP", label: "Milestones", color: "#f59e0b" }, // amber
  { key: "awardVP", label: "Awards", color: "#6b7280" }, // gray
  { key: "greeneryVP", label: "Greenery", color: "#22c55e" }, // green
  { key: "cityVP", label: "Cities", color: "#94a3b8" }, // slate
] as const;

type VPCategoryKey = (typeof VP_CATEGORIES)[number]["key"];

/**
 * VPBarChart - Stacked horizontal bar chart showing VP breakdown for all players
 */
const VPBarChart: FC<VPBarChartProps> = ({
  scores,
  isAnimating,
  onAnimationComplete,
  vertical = false,
}) => {
  const [animationProgress, setAnimationProgress] = useState(0);

  // Sort scores by total VP descending
  const sortedScores = [...scores].sort(
    (a, b) => b.vpBreakdown.totalVP - a.vpBreakdown.totalVP,
  );

  // Find max VP for scaling
  const maxVP = Math.max(...scores.map((s) => s.vpBreakdown.totalVP), 1);

  useEffect(() => {
    if (!isAnimating) {
      setAnimationProgress(1);
      return;
    }

    // Animate progress from 0 to 1 over 2 seconds
    const startTime = Date.now();
    const duration = 2000;

    const animate = () => {
      const elapsed = Date.now() - startTime;
      const progress = Math.min(elapsed / duration, 1);
      setAnimationProgress(progress);

      if (progress < 1) {
        requestAnimationFrame(animate);
      } else {
        onAnimationComplete?.();
      }
    };

    requestAnimationFrame(animate);
  }, [isAnimating, onAnimationComplete]);

  return (
    <div
      className={`section-slide-in-animate flex flex-col w-full ${vertical ? "gap-3 p-2" : "gap-6 p-6 max-w-2xl"}`}
    >
      <h3
        className={`font-orbitron text-white/80 uppercase tracking-wider text-center ${vertical ? "text-sm" : "text-lg"}`}
      >
        Final Scores
      </h3>

      {/* Legend - more compact in vertical mode */}
      <div
        className={`flex flex-wrap justify-center text-xs ${vertical ? "gap-2" : "gap-4"}`}
      >
        {VP_CATEGORIES.map((category) => (
          <div key={category.key} className="flex items-center gap-1">
            <div
              className={`rounded ${vertical ? "w-2 h-2" : "w-3 h-3"}`}
              style={{ backgroundColor: category.color }}
            />
            <span className="text-white/70">{category.label}</span>
          </div>
        ))}
      </div>

      {/* Bars */}
      <div className={`flex flex-col ${vertical ? "gap-2" : "gap-4"}`}>
        {sortedScores.map((score, index) => {
          const { vpBreakdown } = score;
          const segments: {
            key: VPCategoryKey;
            value: number;
            color: string;
          }[] = VP_CATEGORIES.map((cat) => ({
            key: cat.key,
            value: vpBreakdown[cat.key],
            color: cat.color,
          }));

          // Calculate cumulative widths
          let cumulativeWidth = 0;

          return (
            <div
              key={score.playerId}
              className={`
                flex items-center transition-all duration-500
                ${vertical ? "gap-2" : "gap-4"}
                ${score.isWinner ? "scale-105" : ""}
              `}
              style={{ animationDelay: `${index * 200}ms` }}
            >
              {/* Player name */}
              <div
                className={`
                  text-right font-orbitron truncate
                  ${vertical ? "w-16 text-xs" : "w-24 text-sm"}
                  ${score.isWinner ? "text-amber-400 font-bold" : "text-white/80"}
                `}
              >
                {score.playerName}
              </div>

              {/* Stacked bar */}
              <div
                className={`flex-1 bg-gray-800/50 rounded-lg overflow-hidden relative ${vertical ? "h-6" : "h-8"}`}
              >
                {segments.map((segment) => {
                  const segmentWidth =
                    (segment.value / maxVP) * 100 * animationProgress;
                  const left =
                    (cumulativeWidth / maxVP) * 100 * animationProgress;
                  cumulativeWidth += segment.value;

                  if (segment.value === 0) return null;

                  return (
                    <div
                      key={segment.key}
                      className="absolute top-0 bottom-0 transition-all duration-300"
                      style={{
                        left: `${left}%`,
                        width: `${segmentWidth}%`,
                        backgroundColor: segment.color,
                      }}
                    />
                  );
                })}

                {/* Total VP label */}
                <div
                  className={`absolute right-2 top-1/2 -translate-y-1/2 text-white font-bold z-10 drop-shadow-lg ${vertical ? "text-xs" : "text-sm"}`}
                >
                  {Math.round(vpBreakdown.totalVP * animationProgress)}
                </div>
              </div>

              {/* Placement badge */}
              <div
                className={`
                  rounded-full flex items-center justify-center font-bold
                  ${vertical ? "w-6 h-6 text-xs" : "w-8 h-8 text-sm"}
                  ${
                    score.placement === 1
                      ? "bg-amber-400 text-black"
                      : score.placement === 2
                        ? "bg-gray-300 text-black"
                        : score.placement === 3
                          ? "bg-amber-700 text-white"
                          : "bg-gray-600 text-white"
                  }
                `}
              >
                {score.placement}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default VPBarChart;
