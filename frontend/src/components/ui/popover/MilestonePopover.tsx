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

interface MilestonePopoverProps {
  isVisible: boolean;
  onClose: () => void;
  gameState?: GameDto;
  anchorRef: React.RefObject<HTMLButtonElement | null>;
}

const MilestonePopover: React.FC<MilestonePopoverProps> = ({
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

  // Determine if milestones can be claimed
  const isGameActive = gameState?.status === GameStatusActive;
  const isActionPhase = gameState?.currentPhase === GamePhaseAction;
  const isCurrentPlayerTurn = gameState?.currentTurn === gameState?.viewingPlayerId;

  const canClaimMilestones =
    isGameActive && isActionPhase && isCurrentPlayerTurn && canPerformActions(gameState);

  // Get milestones from backend player state
  const milestones = gameState?.currentPlayer?.milestones ?? [];

  // Calculate counts
  const claimedCount = milestones.filter((m) => m.isClaimed).length;
  const availableCount = milestones.filter((m) => m.available && !m.isClaimed).length;

  // Helper to get player name from ID
  const getPlayerName = (playerId: string | undefined): string => {
    if (!playerId || !gameState) return "Unknown";
    if (playerId === gameState.currentPlayer.id) return gameState.currentPlayer.name;
    const otherPlayer = gameState.otherPlayers.find((p) => p.id === playerId);
    return otherPlayer?.name ?? "Unknown";
  };

  const handleClaimMilestone = (milestoneId: string) => {
    if (!canClaimMilestones) return;
    void webSocketService.claimMilestone(milestoneId);
  };

  return (
    <div
      ref={popoverRef}
      className="fixed top-[60px] left-[20px] w-[500px] max-h-[calc(100vh-80px)] bg-space-black-darker/98 border-2 border-[#ff6b35] rounded-xl overflow-hidden shadow-[0_10px_40px_rgba(0,0,0,0.8),0_0_20px_rgba(255,107,53,0.5)] backdrop-blur-space z-[3000] animate-[popoverSlideDown_0.3s_ease-out]"
    >
      {/* Header */}
      <div className="flex items-center justify-between px-4 py-3 bg-black/40 border-b border-[#ff6b35]">
        <div className="flex items-center gap-3">
          <h2 className="m-0 font-orbitron text-white text-base font-bold text-shadow-glow">
            Milestones
          </h2>
          <div className="flex gap-2 text-xs">
            <span className="bg-[#ff6b35]/20 border border-[#ff6b35]/30 rounded px-2 py-0.5 text-white/80">
              {claimedCount}/3 Claimed
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

      {/* Milestones List */}
      <div className="max-h-[calc(100vh-140px)] overflow-y-auto [scrollbar-width:thin] [scrollbar-color:rgba(255,107,53,0.5)_rgba(10,10,15,0.3)] p-2">
        {milestones.map((milestone) => {
          const isClaimed = milestone.isClaimed;
          const isAvailable = milestone.available && !isClaimed;
          const isExecutable = canClaimMilestones && isAvailable;
          const progressMet =
            milestone.progress !== undefined &&
            milestone.required !== undefined &&
            milestone.progress >= milestone.required;

          return (
            <div
              key={milestone.type}
              className={`relative mb-2 last:mb-0 border rounded-lg p-3 transition-all duration-200 ${
                isClaimed
                  ? "border-[#ff6b35] bg-[#ff6b35]/30"
                  : isAvailable
                    ? "border-[#ff6b35] bg-[#ff6b35]/20 hover:bg-[#ff6b35]/30"
                    : "border-[#ff6b35]/30 bg-[#ff6b35]/10 opacity-60"
              }`}
              onClick={() => isExecutable && handleClaimMilestone(milestone.type)}
            >
              {/* Error indicator for unavailable milestones */}
              {!isAvailable && !isClaimed && milestone.errors && milestone.errors.length > 0 && (
                <div className="absolute top-2 right-2 z-[4] bg-[linear-gradient(135deg,#e74c3c,#c0392b)] text-white text-[9px] font-bold px-2 py-1 rounded border border-[rgba(231,76,60,0.8)] shadow-[0_2px_8px_rgba(231,76,60,0.4)] flex items-center gap-1">
                  <span>⚠</span>
                  <span className="max-w-[140px] truncate">
                    {milestone.errors[0].message}
                    {milestone.errors.length > 1 && ` (+${milestone.errors.length - 1})`}
                  </span>
                </div>
              )}
              <div className="flex items-start justify-between gap-3 mb-2">
                {/* Left: Name, Cost, Progress */}
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2 mb-2">
                    <h3 className="text-white text-sm font-bold font-orbitron m-0">
                      {milestone.name}
                    </h3>
                    {isClaimed && (
                      <span className="text-[10px] text-[#ff6b35] bg-[#ff6b35]/30 px-1.5 py-0.5 rounded border border-[#ff6b35]/50">
                        Claimed
                      </span>
                    )}
                  </div>

                  <div className="flex items-center gap-2">
                    <GameIcon
                      iconType={ResourceTypeCredit}
                      amount={milestone.claimCost}
                      size="small"
                    />
                    <span className="text-white/60 text-xs">→</span>
                    <span className="text-amber-400 text-xs font-semibold">5 VP</span>
                  </div>

                  {/* Progress bar */}
                  {milestone.progress !== undefined &&
                    milestone.required !== undefined &&
                    !isClaimed && (
                      <div className="mt-2">
                        <div className="flex justify-between text-xs mb-1">
                          <span className={progressMet ? "text-green-400" : "text-amber-400"}>
                            Progress: {milestone.progress}/{milestone.required}
                          </span>
                          {progressMet && <span className="text-green-400">Ready!</span>}
                        </div>
                        <div className="h-1.5 bg-black/40 rounded-full overflow-hidden">
                          <div
                            className={`h-full rounded-full transition-all ${progressMet ? "bg-green-500" : "bg-amber-500"}`}
                            style={{
                              width: `${Math.min(100, (milestone.progress / milestone.required) * 100)}%`,
                            }}
                          />
                        </div>
                      </div>
                    )}
                </div>

                {/* Right: Claim Button */}
                {canClaimMilestones && !isClaimed && (
                  <button
                    className={`flex-shrink-0 px-3 py-1.5 rounded text-xs font-semibold transition-all cursor-pointer ${
                      isAvailable
                        ? "bg-[#ff6b35]/80 hover:bg-[#ff6b35] text-white shadow-sm hover:shadow-md"
                        : "bg-gray-600/50 text-gray-400"
                    }`}
                    onClick={(e) => {
                      e.stopPropagation();
                      if (isExecutable) handleClaimMilestone(milestone.type);
                    }}
                    disabled={!isAvailable}
                  >
                    Claim
                  </button>
                )}
              </div>

              <p className="text-white/70 text-xs leading-relaxed m-0 text-left">
                {milestone.description}
              </p>

              {/* Claimed by info */}
              {isClaimed && milestone.claimedBy && (
                <div className="mt-2 text-xs text-blue-400/80 italic">
                  Claimed by {getPlayerName(milestone.claimedBy)}
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

export default MilestonePopover;
