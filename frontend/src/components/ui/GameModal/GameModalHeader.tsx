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
            className="text-white/70 hover:text-white text-xl leading-none transition-colors cursor-pointer"
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
