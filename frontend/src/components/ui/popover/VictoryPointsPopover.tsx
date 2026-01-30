import React from "react";
import { VPGranterDto } from "@/types/generated/api-types.ts";
import { GamePopover } from "../GamePopover";

interface VictoryPointsPopoverProps {
  isVisible: boolean;
  onClose: () => void;
  vpGranters: VPGranterDto[];
  totalVP: number;
  anchorRef: React.RefObject<HTMLElement>;
}

const VictoryPointsPopover: React.FC<VictoryPointsPopoverProps> = ({
  isVisible,
  onClose,
  vpGranters,
  totalVP,
  anchorRef,
}) => {
  return (
    <GamePopover
      isVisible={isVisible}
      onClose={onClose}
      position={{ type: "anchor", anchorRef, placement: "above" }}
      theme="victoryPoints"
      header={{ title: `${totalVP} VP` }}
      width="auto"
      maxHeight="none"
      zIndex={2000}
    >
      {vpGranters.length === 0 ? (
        <div className="py-8 px-5 text-center">
          <div className="text-white/40 text-sm">No VP sources</div>
        </div>
      ) : (
        <div className="p-3 flex flex-col gap-2">
          {vpGranters.map((granter) => (
            <div key={granter.cardId} className="flex items-center justify-between gap-3 p-2">
              <span className="text-sm text-white/80">{granter.cardName}</span>
              <span className="text-base font-bold text-white text-shadow-glow">
                {granter.computedValue} VP
              </span>
            </div>
          ))}
        </div>
      )}
    </GamePopover>
  );
};

export default VictoryPointsPopover;
