import React from "react";
import { GameModalEmptyProps } from "./types";

const GameModalEmpty: React.FC<GameModalEmptyProps> = ({ icon, title, description }) => {
  return (
    <div className="flex flex-col items-center justify-center py-[60px] px-5 text-center min-h-[300px]">
      <div className="mb-5 opacity-60">{icon}</div>
      <h3 className="text-white text-2xl m-0 mb-2.5">{title}</h3>
      <p className="text-white/70 text-base m-0">{description}</p>
    </div>
  );
};

export default GameModalEmpty;
