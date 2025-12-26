import React from "react";
import { Z_INDEX } from "@/constants/zIndex.ts";

interface CostDisplayProps {
  cost: number;
  size?: "small" | "medium" | "large";
  className?: string;
}

const CostDisplay: React.FC<CostDisplayProps> = ({ cost, size = "medium", className = "" }) => {
  const sizeMap = {
    small: { container: 24, icon: 24, fontSize: "10px" },
    medium: { container: 32, icon: 32, fontSize: "12px" },
    large: { container: 40, icon: 40, fontSize: "14px" },
  };

  const dimensions = sizeMap[size];

  return (
    <div
      className={`relative inline-flex items-center justify-center rounded ${className}`}
      style={{
        width: `${dimensions.container}px`,
        height: `${dimensions.container}px`,
      }}
    >
      <img
        src="/assets/resources/megacredit.png"
        alt="Megacredits"
        className="block"
        style={{
          width: `${dimensions.icon}px`,
          height: `${dimensions.icon}px`,
        }}
      />
      <span
        className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 text-black font-bold text-center font-[Arial,sans-serif] leading-none whitespace-nowrap [text-shadow:0.5px_0.5px_1px_rgba(255,255,255,0.8)]"
        style={{
          zIndex: Z_INDEX.COST_DISPLAY,
          fontSize: dimensions.fontSize,
        }}
      >
        {cost}
      </span>
    </div>
  );
};

export default CostDisplay;
