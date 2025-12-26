import { FC } from "react";
import type { FinalScoreDto } from "../../../types/generated/api-types";
import { VP_VALUES } from "../../../constants/gameConstants";

interface AwardData {
  type: string;
  name: string;
  isFunded: boolean;
}

interface AwardResult {
  awardType: string;
  firstPlaceIds: string[];
  secondPlaceIds: string[];
}

interface AwardCompactProps {
  awards: AwardData[];
  scores: FinalScoreDto[];
  awardResults?: AwardResult[];
  playerId: string;
}

/** Compact award display for end game overlay */
const AwardCompact: FC<AwardCompactProps> = ({
  awards,
  scores,
  awardResults,
  playerId,
}) => {
  const fundedAwards = awards.filter((a) => a.isFunded);
  if (fundedAwards.length === 0) return null;

  const getPlayerName = (id: string) =>
    scores.find((s) => s.playerId === id)?.playerName ?? "Unknown";

  return (
    <div className="border-t border-white/10 pt-3">
      <h3 className="text-xs text-white/50 uppercase tracking-wider mb-2">
        Awards
      </h3>
      <div className="space-y-1">
        {fundedAwards.map((a) => {
          const result = awardResults?.find((r) => r.awardType === a.type);
          const isFirst = result?.firstPlaceIds?.includes(playerId) ?? false;
          const isSecond = result?.secondPlaceIds?.includes(playerId) ?? false;
          const firstPlaceWinner = result?.firstPlaceIds?.[0];

          return (
            <div
              key={a.type}
              className="flex items-center justify-between text-sm"
            >
              <span className="text-white/80">{a.name}</span>
              <div className="flex items-center gap-2">
                {isFirst && (
                  <>
                    <span className="text-amber-400">
                      {getPlayerName(playerId)}
                    </span>
                    <span className="text-amber-400 text-xs">
                      1st +{VP_VALUES.AWARD_FIRST} VP
                    </span>
                  </>
                )}
                {isSecond && (
                  <>
                    <span className="text-gray-300">
                      {getPlayerName(playerId)}
                    </span>
                    <span className="text-gray-300 text-xs">
                      2nd +{VP_VALUES.AWARD_SECOND} VP
                    </span>
                  </>
                )}
                {!isFirst && !isSecond && firstPlaceWinner && (
                  <>
                    <span className="text-amber-400 text-xs">
                      {getPlayerName(firstPlaceWinner)}
                    </span>
                    <span className="text-gray-500">-</span>
                  </>
                )}
                {!isFirst && !isSecond && !firstPlaceWinner && (
                  <span className="text-gray-500">-</span>
                )}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default AwardCompact;
