import React, { useEffect, useRef } from "react";
import {
  GameDto,
  GameStatusActive,
  GamePhaseAction,
  ResourceTypeCredit,
} from "@/types/generated/api-types.ts";
import GameIcon from "../display/GameIcon.tsx";
import { webSocketService } from "@/services/webSocketService.ts";
import { canPerformActions } from "@/utils/actionUtils.ts";

interface AwardPopoverProps {
  isVisible: boolean;
  onClose: () => void;
  gameState?: GameDto;
  anchorRef: React.RefObject<HTMLButtonElement | null>;
}

const AwardPopover: React.FC<AwardPopoverProps> = ({
  isVisible,
  onClose,
  gameState,
  anchorRef,
}) => {
  const popoverRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        onClose();
      }
    };

    const handleClickOutside = (event: MouseEvent) => {
      if (
        popoverRef.current &&
        !popoverRef.current.contains(event.target as Node) &&
        anchorRef.current &&
        !anchorRef.current.contains(event.target as Node)
      ) {
        onClose();
      }
    };

    if (isVisible) {
      document.addEventListener("keydown", handleEscape);
      document.addEventListener("mousedown", handleClickOutside);
    }

    return () => {
      document.removeEventListener("keydown", handleEscape);
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [isVisible, onClose, anchorRef]);

  if (!isVisible) return null;

  // Determine if awards can be funded
  const isGameActive = gameState?.status === GameStatusActive;
  const isActionPhase = gameState?.currentPhase === GamePhaseAction;
  const isCurrentPlayerTurn = gameState?.currentTurn === gameState?.viewingPlayerId;

  const canFundAwards =
    isGameActive && isActionPhase && isCurrentPlayerTurn && canPerformActions(gameState);

  // Get awards from backend player state
  const awards = gameState?.currentPlayer?.awards ?? [];

  // Calculate counts
  const fundedCount = awards.filter((a) => a.isFunded).length;
  const availableCount = awards.filter((a) => a.available && !a.isFunded).length;

  // Helper to get player name from ID
  const getPlayerName = (playerId: string | undefined): string => {
    if (!playerId || !gameState) return "Unknown";
    if (playerId === gameState.currentPlayer.id) return gameState.currentPlayer.name;
    const otherPlayer = gameState.otherPlayers.find((p) => p.id === playerId);
    return otherPlayer?.name ?? "Unknown";
  };

  const handleFundAward = (awardId: string) => {
    if (!canFundAwards) return;
    void webSocketService.fundAward(awardId);
  };

  return (
    <div
      ref={popoverRef}
      className="fixed top-[60px] left-[20px] w-[500px] max-h-[calc(100vh-80px)] bg-space-black-darker/98 border-2 border-[#f39c12] rounded-xl overflow-hidden shadow-[0_10px_40px_rgba(0,0,0,0.8),0_0_20px_rgba(243,156,18,0.5)] backdrop-blur-space z-[3000] animate-[popoverSlideDown_0.3s_ease-out]"
    >
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 bg-black/40 border-b border-[#f39c12]">
        <div className="flex items-center gap-3">
          <h2 className="m-0 font-orbitron text-white text-base font-bold text-shadow-glow">
            Awards
          </h2>
          <div className="flex gap-2 text-xs">
            <span className="bg-[#f39c12]/20 border border-[#f39c12]/30 rounded px-2 py-0.5 text-white/80">
              {fundedCount}/3 Funded
            </span>
            {availableCount > 0 && (
              <span className="bg-green-500/20 border border-green-500/30 rounded px-2 py-0.5 text-green-400">
                {availableCount} Available
              </span>
            )}
          </div>
        </div>
        <button
          className="text-white/70 hover:text-white text-xl leading-none transition-colors"
          onClick={onClose}
        >
          ×
        </button>
      </div>

      {/* Awards List */}
      <div className="max-h-[calc(100vh-140px)] overflow-y-auto [scrollbar-width:thin] [scrollbar-color:rgba(243,156,18,0.5)_rgba(10,10,15,0.3)] p-2">
        {awards.map((award) => {
          const isFunded = award.isFunded;
          const isAvailable = award.available && !isFunded;
          const isExecutable = canFundAwards && isAvailable;

          return (
            <div
              key={award.type}
              className={`relative mb-2 last:mb-0 border rounded-lg p-3 transition-all duration-200 ${
                isFunded
                  ? "border-[#f39c12] bg-[#f39c12]/30"
                  : isAvailable
                    ? "border-[#f39c12] bg-[#f39c12]/20 hover:bg-[#f39c12]/30"
                    : "border-[#f39c12]/30 bg-[#f39c12]/10 opacity-60"
              }`}
              onClick={() => isExecutable && handleFundAward(award.type)}
            >
              {/* Error indicator for unavailable awards */}
              {!isAvailable && !isFunded && award.errors && award.errors.length > 0 && (
                <div className="absolute top-2 right-2 z-[4] bg-[linear-gradient(135deg,#e74c3c,#c0392b)] text-white text-[9px] font-bold px-2 py-1 rounded border border-[rgba(231,76,60,0.8)] shadow-[0_2px_8px_rgba(231,76,60,0.4)] flex items-center gap-1">
                  <span>⚠</span>
                  <span className="max-w-[140px] truncate">
                    {award.errors[0].message}
                    {award.errors.length > 1 && ` (+${award.errors.length - 1})`}
                  </span>
                </div>
              )}
              <div className="flex items-start justify-between gap-3 mb-2">
                {/* Left: Name, Cost */}
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-2">
                    <h3 className="text-white text-sm font-bold font-orbitron m-0">{award.name}</h3>
                    {isFunded && (
                      <span className="text-[10px] text-[#f39c12] bg-[#f39c12]/30 px-1.5 py-0.5 rounded border border-[#f39c12]/50">
                        Funded
                      </span>
                    )}
                  </div>

                  <div className="flex items-center gap-2">
                    <GameIcon
                      iconType={ResourceTypeCredit}
                      amount={award.fundingCost}
                      size="small"
                    />
                    <span className="text-white/60 text-xs">→</span>
                    <span className="text-amber-400 text-xs font-semibold">
                      5 VP (1st), 2 VP (2nd)
                    </span>
                  </div>
                </div>

                {/* Right: Fund Button */}
                {canFundAwards && !isFunded && (
                  <button
                    className={`flex-shrink-0 px-3 py-1.5 rounded text-xs font-semibold transition-all cursor-pointer ${
                      isAvailable
                        ? "bg-[#f39c12]/80 hover:bg-[#f39c12] text-white shadow-sm hover:shadow-md"
                        : "bg-gray-600/50 text-gray-400"
                    }`}
                    onClick={(e) => {
                      e.stopPropagation();
                      if (isExecutable) handleFundAward(award.type);
                    }}
                    disabled={!isAvailable}
                  >
                    Fund
                  </button>
                )}
              </div>

              <p className="text-white/70 text-xs leading-relaxed m-0 text-left">
                {award.description}
              </p>

              {/* Funded by info */}
              {isFunded && award.fundedBy && (
                <div className="mt-2 text-xs text-blue-400/80 italic">
                  Funded by {getPlayerName(award.fundedBy)}
                </div>
              )}
            </div>
          );
        })}
      </div>

      <style>{`
        @keyframes popoverSlideDown {
          from {
            opacity: 0;
            transform: translateY(-10px);
          }
          to {
            opacity: 1;
            transform: translateY(0);
          }
        }
      `}</style>
    </div>
  );
};

export default AwardPopover;
