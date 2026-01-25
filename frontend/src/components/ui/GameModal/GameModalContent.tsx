import React from "react";
import { GameModalContentProps } from "./types";

const paddingClasses = {
  none: "",
  normal: "py-[25px] px-[30px] max-md:p-5",
  large: "py-[30px] px-[40px] max-md:p-6",
};

const GameModalContent: React.FC<GameModalContentProps> = ({
  children,
  padding = "normal",
  className = "",
}) => {
  return (
    <div
      className={`flex-1 overflow-y-auto [scrollbar-width:thin] [scrollbar-color:var(--modal-accent)_rgba(30,60,150,0.3)] [&::-webkit-scrollbar]:w-2 [&::-webkit-scrollbar-track]:bg-[rgba(30,60,150,0.3)] [&::-webkit-scrollbar-track]:rounded [&::-webkit-scrollbar-thumb]:bg-[var(--modal-accent)]/70 [&::-webkit-scrollbar-thumb]:rounded [&::-webkit-scrollbar-thumb:hover]:bg-[var(--modal-accent)] ${paddingClasses[padding]} ${className}`}
    >
      {children}
    </div>
  );
};

export default GameModalContent;
