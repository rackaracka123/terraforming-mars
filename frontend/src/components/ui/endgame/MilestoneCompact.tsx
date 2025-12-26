import { FC } from "react";
import type { FinalScoreDto } from "../../../types/generated/api-types";
import { VP_VALUES } from "../../../constants/gameConstants";

interface MilestoneData {
  type: string;
  name: string;
  isClaimed: boolean;
  claimedBy?: string;
}

interface MilestoneCompactProps {
  milestones: MilestoneData[];
  scores: FinalScoreDto[];
}

/** Compact milestone display for end game overlay */
const MilestoneCompact: FC<MilestoneCompactProps> = ({
  milestones,
  scores,
}) => {
  const claimedMilestones = milestones.filter((m) => m.isClaimed);
  if (claimedMilestones.length === 0) return null;

  const getPlayerName = (id?: string) =>
    scores.find((s) => s.playerId === id)?.playerName ?? "Unknown";

  return (
    <div className="border-t border-white/10 pt-3">
      <h3 className="text-xs text-white/50 uppercase tracking-wider mb-2">
        Milestones
      </h3>
      <div className="space-y-1">
        {claimedMilestones.map((m) => (
          <div
            key={m.type}
            className="flex items-center justify-between text-sm"
          >
            <span className="text-white/80">{m.name}</span>
            <div className="flex items-center gap-2">
              <span className="text-amber-400">
                {getPlayerName(m.claimedBy)}
              </span>
              <span className="text-green-400 text-xs">
                +{VP_VALUES.MILESTONE} VP
              </span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default MilestoneCompact;
