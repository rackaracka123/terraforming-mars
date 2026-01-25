import React from "react";
import { GamePopoverEmptyProps } from "./types";

const GamePopoverEmpty: React.FC<GamePopoverEmptyProps> = ({ icon, title, description }) => {
  return (
    <div className="flex flex-col items-center justify-center py-10 px-5 text-center">
      <div className="mb-[15px] opacity-60">{icon}</div>
      <div className="text-white text-sm font-medium mb-2">{title}</div>
      <div className="text-white/60 text-xs leading-[1.4]">{description}</div>
    </div>
  );
};

export default GamePopoverEmpty;
