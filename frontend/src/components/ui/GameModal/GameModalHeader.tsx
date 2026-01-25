import React from "react";
import { GameModalHeaderProps } from "./types";

const GameModalHeader: React.FC<GameModalHeaderProps> = ({
  title,
  subtitle,
  stats,
  controls,
  showCloseButton = true,
  onClose,
}) => {
  return (
    <div className="flex items-start justify-between py-[25px] px-[30px] bg-black/40 border-b border-[var(--modal-accent)]/60 flex-shrink-0 max-md:p-5">
      <div className="flex flex-col gap-[15px]">
        <h1 className="m-0 font-orbitron text-white text-[28px] font-bold text-shadow-glow tracking-wider">
          {title}
        </h1>
        {subtitle && <p className="text-white/70 text-sm m-0">{subtitle}</p>}
        {stats && <div className="flex gap-5 items-center">{stats}</div>}
      </div>

      <div className="flex gap-5 items-start max-md:flex-col max-md:gap-2.5">
        {controls}

        {showCloseButton && onClose && (
          <button
            className="bg-[linear-gradient(135deg,rgba(255,80,80,0.8)_0%,rgba(200,40,40,0.9)_100%)] border-2 border-[rgba(255,120,120,0.6)] rounded-full w-[45px] h-[45px] text-white text-2xl font-bold cursor-pointer flex items-center justify-center transition-all duration-300 shadow-[0_4px_15px_rgba(0,0,0,0.4)] flex-shrink-0 hover:scale-110 hover:shadow-[0_6px_25px_rgba(255,80,80,0.5)]"
            onClick={onClose}
          >
            ×
          </button>
        )}
      </div>
    </div>
  );
};

export default GameModalHeader;
