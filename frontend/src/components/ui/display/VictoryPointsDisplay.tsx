import React from "react";

interface VictoryPointsDisplayProps {
  victoryPoints: number;
  size?: "small" | "medium" | "large";
  className?: string;
}

const VictoryPointsDisplay: React.FC<VictoryPointsDisplayProps> = ({
  victoryPoints,
  size = "medium",
  className = "",
}) => {
  const sizeMap = {
    small: { fontSize: "16px", padding: "8px 12px" },
    medium: { fontSize: "24px", padding: "12px 16px" },
    large: { fontSize: "32px", padding: "16px 20px" },
  };

  const dimensions = sizeMap[size];

  return (
    <div
      className={`inline-flex items-center justify-center bg-[linear-gradient(135deg,rgba(20,40,60,0.9)_0%,rgba(30,50,70,0.8)_100%)] border-2 border-[rgba(255,215,0,0.6)] rounded-lg shadow-[0_4px_15px_rgba(0,0,0,0.5),0_0_20px_rgba(255,215,0,0.3)] backdrop-blur-[10px] ${className}`}
      style={{
        padding: dimensions.padding,
      }}
    >
      <div className="flex items-center gap-2">
        <img
          src="/assets/resources/tr.png"
          alt="Victory Points"
          className="w-5 h-5 brightness-[1.2]"
        />
        <span
          className="text-white font-bold font-[Courier_New,monospace] [text-shadow:2px_2px_4px_rgba(0,0,0,0.8)] leading-none"
          style={{
            fontSize: dimensions.fontSize,
          }}
        >
          {victoryPoints}
        </span>
      </div>
    </div>
  );
};

export default VictoryPointsDisplay;
