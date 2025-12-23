import { FC, useEffect, useState } from "react";
import AnimatedNumber from "../display/AnimatedNumber";
import GameIcon from "../display/GameIcon";

interface CardVPSectionProps {
  /** Total VP from cards */
  cardVP: number;
  /** Number of VP-contributing cards */
  vpCardCount: number;
  /** Whether to animate the display */
  isAnimating: boolean;
  /** Callback when animation completes */
  onAnimationComplete?: () => void;
}

/**
 * CardVPSection - Displays VP earned from played cards
 */
const CardVPSection: FC<CardVPSectionProps> = ({
  cardVP,
  vpCardCount,
  isAnimating,
  onAnimationComplete,
}) => {
  const [showCards, setShowCards] = useState(!isAnimating);
  const [cardIndex, setCardIndex] = useState(0);

  useEffect(() => {
    if (!isAnimating) {
      setShowCards(true);
      return;
    }

    // Animate card stack appearing
    const showTimeout = setTimeout(() => {
      setShowCards(true);
    }, 500);

    // Animate cards "popping"
    const interval = setInterval(() => {
      setCardIndex((prev) => {
        if (prev >= Math.min(vpCardCount, 5)) {
          clearInterval(interval);
          return prev;
        }
        return prev + 1;
      });
    }, 400);

    return () => {
      clearTimeout(showTimeout);
      clearInterval(interval);
    };
  }, [isAnimating, vpCardCount]);

  return (
    <div className="section-slide-in-animate flex flex-col items-center gap-4 p-6">
      <h3 className="font-orbitron text-lg text-white/80 uppercase tracking-wider">
        Card Victory Points
      </h3>

      <div className="flex items-center gap-8">
        {/* Card stack visualization */}
        <div
          className={`
            relative w-24 h-32 transition-all duration-500
            ${showCards ? "opacity-100 scale-100" : "opacity-0 scale-90"}
          `}
        >
          {/* Stacked cards effect */}
          {[...Array(Math.min(5, vpCardCount))].map((_, i) => (
            <div
              key={i}
              className={`
                absolute inset-0 rounded-lg border-2 border-blue-400/50 bg-gradient-to-br from-blue-900/80 to-blue-950/80
                transition-all duration-300
                ${i <= cardIndex && isAnimating ? "card-pop-animate" : ""}
              `}
              style={{
                transform: `translateY(${i * -4}px) translateX(${i * 2}px) rotate(${i * -2}deg)`,
                zIndex: 5 - i,
              }}
            >
              {/* Card content indicator */}
              <div className="absolute inset-2 flex items-center justify-center opacity-50">
                <GameIcon iconType="card" size="medium" />
              </div>
            </div>
          ))}

          {/* VP badge on top card */}
          {showCards && cardVP > 0 && (
            <div className="absolute -top-3 -right-3 z-10 bg-purple-500 text-white font-bold text-sm px-2 py-1 rounded-full shadow-lg">
              VP
            </div>
          )}
        </div>

        {/* VP count */}
        <div className="flex flex-col items-center gap-2">
          <div className="text-5xl font-orbitron font-bold text-purple-400">
            {isAnimating ? (
              <AnimatedNumber
                value={cardVP}
                duration={4000}
                onComplete={onAnimationComplete}
              />
            ) : (
              cardVP
            )}
          </div>
          <span className="text-xl font-orbitron text-white/60">VP</span>
        </div>
      </div>

      {/* Card count */}
      <div className="flex items-center gap-4 mt-2">
        <div className="flex items-center gap-2">
          <GameIcon iconType="card" size="small" />
          <span className="text-white/60">
            {vpCardCount} card{vpCardCount !== 1 ? "s" : ""} with VP
          </span>
        </div>
      </div>

      <p className="text-xs text-white/50 text-center max-w-xs">
        Victory points from card effects, conditions, and bonuses
      </p>
    </div>
  );
};

export default CardVPSection;
