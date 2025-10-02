import React from "react";

interface MegaCreditIconProps {
  value: number;
  size?: "small" | "medium" | "large";
}

const MegaCreditIcon: React.FC<MegaCreditIconProps> = ({
  value,
  size = "medium",
}) => {
  const sizeClasses = {
    small: "w-6 h-6 text-[12px]",
    medium: "w-8 h-8 text-[16px]",
    large: "w-10 h-10 text-[20px]",
  };

  return (
    <div
      className={`relative inline-flex items-center justify-center ${sizeClasses[size]}`}
    >
      <img
        src="/assets/resources/megacredit.png"
        alt="MC"
        className="w-full h-full object-contain [filter:drop-shadow(0_2px_4px_rgba(0,0,0,0.6))]"
      />
      <span className="absolute top-0 left-0 right-0 bottom-0 text-black font-black font-[Prototype,Arial_Black,Arial,sans-serif] flex items-center justify-center text-center leading-none [text-shadow:0_0_2px_rgba(255,255,255,0.3)] tracking-[0.5px] [-webkit-font-smoothing:antialiased] [-moz-osx-font-smoothing:grayscale] [text-rendering:optimizeLegibility]">
        {value}
      </span>
    </div>
  );
};

export default MegaCreditIcon;
