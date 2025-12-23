import { FC } from "react";
import AnimatedNumber from "../display/AnimatedNumber";
import GameIcon from "../display/GameIcon";

interface TRCountSectionProps {
  /** Player's terraform rating value */
  terraformRating: number;
  /** Whether to animate the count */
  isAnimating: boolean;
  /** Callback when animation completes */
  onAnimationComplete?: () => void;
}

/**
 * TRCountSection - Displays animated terraform rating VP count
 */
const TRCountSection: FC<TRCountSectionProps> = ({
  terraformRating,
  isAnimating,
  onAnimationComplete,
}) => {
  return (
    <div className="section-slide-in-animate flex flex-col items-center gap-4 p-6">
      <h3 className="font-orbitron text-lg text-white/80 uppercase tracking-wider">
        Terraform Rating
      </h3>

      <div className="flex items-center gap-4">
        <GameIcon iconType="terraform-rating" size="large" />

        <div className="text-5xl font-orbitron font-bold text-white">
          {isAnimating ? (
            <AnimatedNumber
              value={terraformRating}
              duration={1500}
              onComplete={onAnimationComplete}
              className="text-amber-400"
            />
          ) : (
            <span className="text-amber-400">{terraformRating}</span>
          )}
        </div>

        <span className="text-2xl font-orbitron text-white/60">VP</span>
      </div>

      <p className="text-sm text-white/50">1 VP per Terraform Rating</p>
    </div>
  );
};

export default TRCountSection;
