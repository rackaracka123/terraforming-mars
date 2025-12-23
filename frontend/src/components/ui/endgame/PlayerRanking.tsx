import { FC, useEffect, useState } from "react";
import { FinalScoreDto } from "../../../types/generated/api-types";

interface PlayerRankingProps {
  /** Final scores for all players */
  scores: FinalScoreDto[];
  /** Whether to animate the podium */
  isAnimating: boolean;
  /** Callback when animation completes */
  onAnimationComplete?: () => void;
}

/**
 * PlayerRanking - Podium display with winner spotlight
 */
const PlayerRanking: FC<PlayerRankingProps> = ({
  scores,
  isAnimating,
  onAnimationComplete,
}) => {
  const [showPodium, setShowPodium] = useState(!isAnimating);
  const [revealedPlaces, setRevealedPlaces] = useState<number[]>([]);

  // Sort by placement
  const sortedScores = [...scores].sort((a, b) => a.placement - b.placement);

  // Get top 3 (or less if fewer players)
  const podiumPlayers = sortedScores.slice(0, 3);

  useEffect(() => {
    if (!isAnimating) {
      setShowPodium(true);
      setRevealedPlaces([1, 2, 3]);
      return;
    }

    // Reveal places in order: 3rd, 2nd, 1st (dramatic reveal)
    const revealOrder = [3, 2, 1];
    let currentIndex = 0;

    const interval = setInterval(() => {
      if (currentIndex === 0) {
        setShowPodium(true);
      }

      if (currentIndex < revealOrder.length) {
        setRevealedPlaces((prev) => [...prev, revealOrder[currentIndex]]);
        currentIndex++;
      } else {
        clearInterval(interval);
        onAnimationComplete?.();
      }
    }, 800);

    return () => clearInterval(interval);
  }, [isAnimating, onAnimationComplete]);

  const getPlacementStyle = (placement: number) => {
    switch (placement) {
      case 1:
        return {
          height: "h-32",
          bg: "bg-gradient-to-t from-amber-600 to-amber-400",
          border: "border-amber-300",
          text: "text-amber-100",
          medalBg: "bg-amber-400",
          medalText: "text-amber-900",
          label: "1st",
        };
      case 2:
        return {
          height: "h-24",
          bg: "bg-gradient-to-t from-gray-500 to-gray-300",
          border: "border-gray-200",
          text: "text-gray-100",
          medalBg: "bg-gray-300",
          medalText: "text-gray-800",
          label: "2nd",
        };
      case 3:
        return {
          height: "h-16",
          bg: "bg-gradient-to-t from-amber-800 to-amber-600",
          border: "border-amber-500",
          text: "text-amber-100",
          medalBg: "bg-amber-700",
          medalText: "text-amber-100",
          label: "3rd",
        };
      default:
        return {
          height: "h-12",
          bg: "bg-gradient-to-t from-gray-700 to-gray-600",
          border: "border-gray-500",
          text: "text-gray-300",
          medalBg: "bg-gray-600",
          medalText: "text-gray-300",
          label: `${placement}th`,
        };
    }
  };

  // Reorder for podium display: [2nd, 1st, 3rd]
  const podiumOrder = [
    podiumPlayers.find((p) => p.placement === 2),
    podiumPlayers.find((p) => p.placement === 1),
    podiumPlayers.find((p) => p.placement === 3),
  ].filter(Boolean) as FinalScoreDto[];

  return (
    <div className="section-slide-in-animate flex flex-col items-center gap-6 p-6">
      <h3 className="font-orbitron text-lg text-white/80 uppercase tracking-wider">
        Final Rankings
      </h3>

      {/* Podium */}
      <div
        className={`
          flex items-end justify-center gap-4 transition-all duration-500
          ${showPodium ? "opacity-100 translate-y-0" : "opacity-0 translate-y-10"}
        `}
      >
        {podiumOrder.map((player) => {
          const style = getPlacementStyle(player.placement);
          const isRevealed = revealedPlaces.includes(player.placement);
          const isWinner = player.placement === 1;

          return (
            <div
              key={player.playerId}
              className={`
                flex flex-col items-center transition-all duration-500
                ${isRevealed ? "opacity-100 scale-100" : "opacity-0 scale-90"}
              `}
            >
              {/* Player info */}
              <div
                className={`
                  flex flex-col items-center mb-2 transition-all duration-300
                  ${isWinner && isRevealed ? "winner-glow-animate" : ""}
                `}
              >
                {/* Medal */}
                <span
                  className={`w-10 h-10 rounded-full ${style.medalBg} ${style.medalText} font-orbitron font-bold text-lg flex items-center justify-center mb-1 shadow-lg`}
                >
                  {style.label}
                </span>

                {/* Player name */}
                <span
                  className={`
                    font-orbitron text-sm font-bold max-w-24 truncate
                    ${isWinner ? "text-amber-400" : "text-white"}
                  `}
                >
                  {player.playerName}
                </span>

                {/* VP score */}
                <span className="text-2xl font-bold text-white">
                  {player.vpBreakdown.totalVP}
                  <span className="text-sm text-white/60 ml-1">VP</span>
                </span>
              </div>

              {/* Podium block */}
              <div
                className={`
                  w-24 ${style.height} ${style.bg} rounded-t-lg border-t-4 ${style.border}
                  flex items-start justify-center pt-2
                  ${isRevealed ? "podium-rise-animate" : ""}
                `}
                style={{
                  animationDelay: isRevealed ? "0ms" : "1000ms",
                }}
              >
                <span className={`font-orbitron font-bold ${style.text}`}>
                  {style.label}
                </span>
              </div>
            </div>
          );
        })}
      </div>

      {/* Remaining players (4th place and beyond) */}
      {sortedScores.length > 3 && (
        <div className="flex flex-col gap-2 mt-4">
          {sortedScores.slice(3).map((player) => (
            <div
              key={player.playerId}
              className="flex items-center gap-4 px-4 py-2 bg-gray-800/50 rounded-lg"
            >
              <span className="font-orbitron text-gray-400 w-8">
                {player.placement}th
              </span>
              <span className="font-orbitron text-white/80 flex-1 truncate">
                {player.playerName}
              </span>
              <span className="font-bold text-white">
                {player.vpBreakdown.totalVP} VP
              </span>
            </div>
          ))}
        </div>
      )}

      {/* Winner announcement */}
      {revealedPlaces.includes(1) && podiumPlayers[0] && (
        <div className="text-center mt-4 score-reveal-animate">
          <p className="font-orbitron text-xl text-amber-400 winner-glow-animate">
            {podiumPlayers[0].playerName} Wins!
          </p>
          <p className="text-white/60 text-sm mt-1">
            with {podiumPlayers[0].vpBreakdown.totalVP} Victory Points
          </p>
        </div>
      )}
    </div>
  );
};

export default PlayerRanking;
