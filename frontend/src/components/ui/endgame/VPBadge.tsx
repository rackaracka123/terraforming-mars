import { FC } from "react";
import GameIcon from "../display/GameIcon";

const COLOR_CLASSES: Record<string, string> = {
  blue: "bg-blue-500/20 text-blue-400",
  purple: "bg-purple-500/20 text-purple-400",
  yellow: "bg-amber-500/20 text-amber-400",
  green: "bg-green-500/20 text-green-400 hover:bg-green-500/30 cursor-pointer",
  gray: "bg-gray-500/20 text-gray-400 hover:bg-gray-500/30 cursor-pointer",
  indigo: "bg-indigo-500/20 text-indigo-400 hover:bg-indigo-500/30 cursor-pointer",
};

interface VPBadgeProps {
  icon: string;
  value: number;
  color: string;
  onMouseEnter?: () => void;
  onMouseLeave?: () => void;
}

/** Small VP badge with icon */
const VPBadge: FC<VPBadgeProps> = ({ icon, value, color, onMouseEnter, onMouseLeave }) => {
  return (
    <div
      className={`flex items-center gap-1 px-2 py-0.5 rounded transition-colors ${COLOR_CLASSES[color] ?? COLOR_CLASSES.gray}`}
      onMouseEnter={onMouseEnter}
      onMouseLeave={onMouseLeave}
    >
      <GameIcon iconType={icon} size="small" />
      <span>{value}</span>
    </div>
  );
};

export default VPBadge;
