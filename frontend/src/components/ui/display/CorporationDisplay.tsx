import React from "react";
import { CardDto } from "@/types/generated/api-types.ts";
import { getCorporationLogo } from "@/utils/corporationLogos.tsx";

interface CorporationDisplayProps {
  corporation: CardDto;
}

const CorporationDisplay: React.FC<CorporationDisplayProps> = ({
  corporation,
}) => {
  return (
    <div
      className="fixed bottom-[150px] left-[30px] z-[999] pointer-events-auto"
      title={`${corporation.name}\n${corporation.description}`}
    >
      <div className="bg-[linear-gradient(135deg,rgba(15,30,55,0.5)_0%,rgba(10,20,45,0.4)_100%)] border-2 border-[#050a10] rounded-lg p-1.5 shadow-[0_2px_8px_rgba(0,0,0,0.3)] backdrop-blur-space transition-all duration-300 hover:-translate-y-1 hover:shadow-[0_4px_12px_rgba(0,0,0,0.4)]">
        {/* Corporation Logo Only */}
        <div className="rounded-lg p-1 bg-black/30 flex items-center justify-center [&>*]:box-content [&>*]:[filter:drop-shadow(0_2px_4px_rgba(0,0,0,0.8))] [&>*]:rounded-[4px] [&>*>*]:rounded-[4px] [&>*>*]:[filter:drop-shadow(0_2px_4px_rgba(0,0,0,0.8))]">
          {getCorporationLogo(corporation.name.toLowerCase())}
        </div>
      </div>
    </div>
  );
};

export default CorporationDisplay;
