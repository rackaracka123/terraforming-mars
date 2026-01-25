import React from "react";
import { GameModalFooterProps } from "./types";

const GameModalFooter: React.FC<GameModalFooterProps> = ({ children, className = "" }) => {
  return (
    <div
      className={`bg-black/40 border-t border-[var(--modal-accent)]/20 py-5 px-[30px] flex-shrink-0 ${className}`}
    >
      {children}
    </div>
  );
};

export default GameModalFooter;
