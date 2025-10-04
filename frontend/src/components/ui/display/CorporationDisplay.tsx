import React from "react";
import { CardDto } from "@/types/generated/api-types.ts";

interface CorporationDisplayProps {
  corporation: CardDto;
}

const CorporationDisplay: React.FC<CorporationDisplayProps> = ({
  corporation,
}) => {
  // Extract card number from ID (e.g., "001" -> "001.webp")
  const corporationImagePath = `/assets/cards/${corporation.id}.webp`;

  return (
    <div
      className="fixed bottom-[60px] left-[30px] z-[999] pointer-events-auto"
      title={`${corporation.name}\n${corporation.description}`}
    >
      <div className="flex flex-col items-center gap-2 bg-space-black-darker/95 border-2 border-space-blue-400 rounded-xl p-3 shadow-glow backdrop-blur-space transition-all duration-300 hover:-translate-y-1 hover:shadow-glow-lg">
        {/* Corporation Logo/Card */}
        <div className="w-20 h-28 rounded-lg overflow-hidden shadow-[0_4px_12px_rgba(0,0,0,0.6)]">
          <img
            src={corporationImagePath}
            alt={corporation.name}
            className="w-full h-full object-cover"
            onError={(e) => {
              // Fallback to a placeholder if image doesn't exist
              e.currentTarget.src = "/assets/cards/001.webp";
            }}
          />
        </div>

        {/* Corporation Name */}
        <div className="text-xs font-semibold text-white text-center max-w-[80px] leading-tight text-shadow-glow-strong tracking-wider">
          {corporation.name}
        </div>
      </div>
    </div>
  );
};

export default CorporationDisplay;
