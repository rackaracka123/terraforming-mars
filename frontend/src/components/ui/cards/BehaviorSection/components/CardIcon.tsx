import React from "react";
import { getIconPath } from "@/utils/iconStore.ts";

interface CardIconProps {
  amount: number;
  badgeType: "peek" | "take" | "buy" | "none";
  isAffordable?: boolean;
}

const CardIcon: React.FC<CardIconProps> = ({
  amount,
  badgeType,
  isAffordable = true,
}) => {
  const cardIcon = getIconPath("card-draw");

  if (!cardIcon) return null;

  // Map badge type to white Unicode symbols
  const badgeSymbols = {
    peek: "⦿", // Eye-like circle
    take: "↓", // Down arrow
    buy: "$", // Dollar sign
    none: "",
  };

  const badge = badgeSymbols[badgeType];

  // Yellowish-white glow for card icons (matching GameIcon and BehaviorIcon)
  const glowFilter =
    "drop-shadow(0_1px_3px_rgba(0,0,0,0.6))_drop-shadow(0_0_1px_rgba(255,248,220,0.6))_drop-shadow(0_0_2px_rgba(255,248,220,0.4))";

  const iconClass = isAffordable
    ? `w-[26px] h-[26px] object-contain [filter:${glowFilter}] max-md:w-[22px] max-md:h-[22px]`
    : `w-[26px] h-[26px] object-contain opacity-40 [filter:grayscale(0.7)_drop-shadow(0_1px_2px_rgba(0,0,0,0.5))] max-md:w-[22px] max-md:h-[22px]`;

  return (
    <div className="flex items-center gap-0.5 relative">
      {/* Amount number */}
      <span className="text-[11px] font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] mr-px">
        {amount}
      </span>

      {/* Card icon with badge */}
      <div className="relative inline-block">
        <img src={cardIcon} alt="card" className={iconClass} />

        {/* Badge overlay at bottom-right */}
        {badge && (
          <div
            className="absolute -bottom-[3px] -right-[3px] text-white text-[14px] font-bold leading-none [text-shadow:0_0_3px_rgba(0,0,0,0.9),0_0_5px_rgba(0,0,0,0.7),1px_1px_2px_rgba(0,0,0,1)] max-md:text-[12px]"
            style={{ pointerEvents: "none" }}
          >
            {badge}
          </div>
        )}
      </div>
    </div>
  );
};

export default CardIcon;
