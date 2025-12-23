import { FC, useEffect, useState } from "react";
import AnimatedNumber from "../display/AnimatedNumber";
import GameIcon from "../display/GameIcon";

export type TileHighlightType = "greenery" | "city" | "adjacent" | null;

interface TileSectionProps {
  /** VP from greenery tiles */
  greeneryVP: number;
  /** VP from city adjacency */
  cityVP: number;
  /** Whether to animate the display */
  isAnimating: boolean;
  /** Current highlight phase */
  highlightPhase: "greenery" | "city" | null;
  /** Callback to highlight tiles on the 3D board */
  onHighlightTiles?: (type: TileHighlightType) => void;
  /** Callback when animation completes */
  onAnimationComplete?: () => void;
}

/**
 * TileSection - Displays greenery and city VP with board highlighting coordination
 */
const TileSection: FC<TileSectionProps> = ({
  greeneryVP,
  cityVP,
  isAnimating,
  highlightPhase,
  onHighlightTiles,
  onAnimationComplete,
}) => {
  const [greeneryDone, setGreeneryDone] = useState(false);
  const [cityDone, setCityDone] = useState(false);

  // Trigger board highlighting when phase changes
  useEffect(() => {
    if (highlightPhase) {
      onHighlightTiles?.(highlightPhase);
    }
  }, [highlightPhase, onHighlightTiles]);

  // Handle greenery phase completion
  const handleGreeneryComplete = () => {
    setGreeneryDone(true);
    onAnimationComplete?.();
  };

  // Handle city phase completion
  const handleCityComplete = () => {
    setCityDone(true);
    onAnimationComplete?.();
  };

  // Show greenery when: not animating, or in greenery phase, or greenery is done
  const showGreenery = !isAnimating || highlightPhase === "greenery" || greeneryDone;
  // Show city when: not animating, or in city phase, or city is done
  const showCity = !isAnimating || highlightPhase === "city" || cityDone;

  return (
    <div className="section-slide-in-animate flex flex-col items-center gap-6 p-6">
      <h3 className="font-orbitron text-lg text-white/80 uppercase tracking-wider">
        Tile Victory Points
      </h3>

      <div className="flex flex-wrap justify-center gap-8">
        {/* Greenery VP */}
        <div
          className={`
            flex flex-col items-center gap-3 p-6 rounded-lg border-2 transition-all duration-500
            ${
              showGreenery
                ? "opacity-100 translate-y-0 border-green-500 bg-green-500/10"
                : "opacity-0 translate-y-4 border-transparent"
            }
            ${highlightPhase === "greenery" ? "tile-highlight-animate" : ""}
          `}
          style={{ color: "#22c55e" }}
        >
          <GameIcon iconType="greenery-tile" size="large" />

          <span className="font-orbitron text-sm text-green-400 uppercase">Greenery Tiles</span>

          <div className="text-4xl font-orbitron font-bold text-green-400">
            {isAnimating && highlightPhase === "greenery" ? (
              <AnimatedNumber
                value={greeneryVP}
                duration={2000}
                onComplete={handleGreeneryComplete}
              />
            ) : (
              greeneryVP
            )}
            <span className="text-xl text-white/60 ml-2">VP</span>
          </div>

          <p className="text-xs text-white/50 text-center">1 VP per greenery tile you own</p>
        </div>

        {/* City VP */}
        <div
          className={`
            flex flex-col items-center gap-3 p-6 rounded-lg border-2 transition-all duration-500
            ${
              showCity
                ? "opacity-100 translate-y-0 border-gray-400 bg-gray-400/10"
                : "opacity-0 translate-y-4 border-transparent"
            }
            ${highlightPhase === "city" ? "tile-highlight-animate" : ""}
          `}
          style={{ color: "#9ca3af" }}
        >
          <GameIcon iconType="city-tile" size="large" />

          <span className="font-orbitron text-sm text-gray-300 uppercase">City Adjacency</span>

          <div className="text-4xl font-orbitron font-bold text-gray-300">
            {isAnimating && highlightPhase === "city" ? (
              <AnimatedNumber value={cityVP} duration={3000} onComplete={handleCityComplete} />
            ) : (
              cityVP
            )}
            <span className="text-xl text-white/60 ml-2">VP</span>
          </div>

          <p className="text-xs text-white/50 text-center">
            1 VP per greenery adjacent to your cities
          </p>
        </div>
      </div>

      {/* Total tile VP */}
      <div className="flex items-center gap-2 mt-2">
        <span className="text-white/60">Total:</span>
        <span className="text-2xl font-orbitron font-bold text-white">
          {greeneryVP + cityVP} VP
        </span>
      </div>
    </div>
  );
};

export default TileSection;
