import React from "react";
import { PlayerDto, OtherPlayerDto } from "@/types/generated/api-types.ts";
import GameIcon from "../display/GameIcon.tsx";

interface PlayerCardProps {
  player: PlayerDto | OtherPlayerDto;
  playerColor: string;
  isCurrentPlayer: boolean;
  isActivePlayer: boolean;
  isCurrentTurn: boolean;
  isActionPhase: boolean;
  onSkipAction?: () => void;
  actionsUsed?: number;
  totalActions?: number;
  totalPlayers?: number; // Added to determine solo vs last player
}

const PlayerCard: React.FC<PlayerCardProps> = ({
  player,
  playerColor,
  isCurrentPlayer,
  isActivePlayer,
  isCurrentTurn,
  isActionPhase,
  onSkipAction,
  actionsUsed = 0,
  totalActions = 2,
  totalPlayers = 1,
}) => {
  const isPassed = player.passed;
  const isDisconnected = !player.isConnected;
  const hasUnlimitedActions = player.availableActions === -1;

  // For unlimited actions, calculate actionsRemaining and button text differently
  const actionsRemaining = hasUnlimitedActions
    ? -1
    : totalActions - actionsUsed;

  // Determine button text - always PASS for unlimited actions, otherwise SKIP if actions used
  const buttonText = hasUnlimitedActions
    ? "PASS"
    : actionsUsed > 0
      ? "SKIP"
      : "PASS";

  return (
    <div
      className={`relative w-full h-[60px] overflow-visible pointer-events-auto ${isCurrentTurn ? "mb-1.5" : "mb-2"}`}
    >
      {/* Main player card with angled edge */}
      <div
        className={`relative h-full bg-[linear-gradient(180deg,rgba(15,35,60,0.2)_0%,rgba(10,25,45,0.2)_50%,rgba(5,15,35,0.3)_100%)] backdrop-blur-[2px] border-l-[6px] pl-2 pr-2 transition-all duration-300 flex items-center [clip-path:polygon(0_0,calc(100%-8px)_0,100%_100%,0_100%)] max-w-[220px] z-[2] opacity-100 shadow-[0_2px_8px_rgba(0,0,0,0.3),-2px_0_6px_var(--player-color),-1px_0_4px_var(--player-color),0_1px_8px_rgba(100,200,255,0.05),inset_0_1px_0_rgba(255,255,255,0.05)] ${isDisconnected ? "opacity-20" : ""} ${!isCurrentTurn ? "opacity-70" : ""} ${isCurrentTurn ? "border-l-8 bg-[linear-gradient(180deg,rgba(20,45,70,0.4)_0%,rgba(15,35,55,0.4)_50%,rgba(10,25,45,0.5)_100%)] shadow-[0_6px_24px_rgba(0,0,0,0.5),-6px_0_20px_var(--player-color),-3px_0_12px_var(--player-color),0_2px_20px_rgba(0,212,255,0.2),inset_0_2px_0_rgba(255,255,255,0.2)] animate-[activePlayerGlow_2s_ease-in-out_infinite_alternate] before:content-[''] before:absolute before:-top-[2px] before:-left-[2px] before:-right-[2px] before:-bottom-[2px] before:bg-[linear-gradient(135deg,rgba(0,212,255,0.3),rgba(0,150,255,0.2))] before:[clip-path:polygon(0_0,calc(100%-10px)_0,100%_100%,0_100%)] before:z-[-1] before:rounded-[inherit] before:animate-[activePlayerGlowBg_2s_ease-in-out_infinite_alternate]" : ""}`}
        style={{ "--player-color": playerColor } as React.CSSProperties}
      >
        <div className="flex flex-col items-start justify-center w-full gap-1">
          <div className="flex gap-1 flex-wrap justify-start items-center relative z-[2]">
            {isCurrentPlayer && (
              <span className="px-1.5 py-px rounded-lg text-[8px] font-semibold uppercase tracking-[0.3px] shadow-[0_1px_2px_rgba(0,0,0,0.2)] bg-[linear-gradient(135deg,#00d4ff,#0099cc)] text-white border-2 border-[rgba(0,212,255,0.8)] [text-shadow:0_0_12px_rgba(0,212,255,0.8),0_2px_4px_rgba(0,0,0,0.6)] shadow-[0_0_16px_rgba(0,212,255,0.4),inset_0_1px_0_rgba(255,255,255,0.3)]">
                YOU
              </span>
            )}
            {isPassed && (
              <span className="px-1.5 py-px rounded-lg text-[8px] font-semibold uppercase tracking-[0.3px] shadow-[0_1px_2px_rgba(0,0,0,0.2)] bg-[linear-gradient(135deg,#95a5a6,#7f8c8d)] text-white border border-[rgba(149,165,166,0.5)]">
                PASSED
              </span>
            )}
            {isDisconnected && (
              <span className="px-1.5 py-px rounded-lg text-[8px] font-semibold uppercase tracking-[0.3px] shadow-[0_1px_2px_rgba(0,0,0,0.2)] bg-[linear-gradient(135deg,#e74c3c,#c0392b)] text-white border border-[rgba(231,76,60,0.5)] opacity-100 relative z-[3]">
                DISCONNECTED
              </span>
            )}
            {isCurrentTurn && isActionPhase && (
              <span className="px-1.5 py-px rounded-lg text-[8px] font-semibold uppercase tracking-[0.3px] shadow-[0_1px_2px_rgba(0,0,0,0.2)] bg-[linear-gradient(135deg,rgba(0,212,255,0.3),rgba(0,150,200,0.4))] text-[#00d4ff] border border-[rgba(0,212,255,0.6)] text-[9px] [text-shadow:0_0_10px_rgba(0,212,255,0.6),0_1px_2px_rgba(0,0,0,0.5)] shadow-[0_0_8px_rgba(0,212,255,0.2)]">
                {hasUnlimitedActions
                  ? totalPlayers === 1
                    ? "Solo"
                    : "Last player"
                  : `${actionsRemaining} ${actionsRemaining === 1 ? "action" : "actions"} left`}
              </span>
            )}
          </div>
          <span className="text-sm font-semibold text-white [text-shadow:0_1px_4px_rgba(0,0,0,0.8),0_0_8px_rgba(100,200,255,0.3)] tracking-[0.4px] shrink-0">
            {player.name}
          </span>
        </div>
        {/* TR Display - always visible */}
        <div className="absolute right-16 top-1/2 -translate-y-1/2 flex items-center gap-1 bg-[linear-gradient(135deg,rgba(20,45,70,0.6),rgba(15,35,55,0.6))] border border-[rgba(100,200,255,0.3)] rounded px-1.5 py-1 shadow-[0_2px_8px_rgba(0,0,0,0.4)]">
          <GameIcon iconType="tr" size="small" />
          <span className="text-xs font-bold text-white [text-shadow:0_1px_2px_rgba(0,0,0,0.8)] min-w-[16px] text-center">
            {player.terraformRating}
          </span>
        </div>
        {isActivePlayer && isCurrentTurn && isActionPhase && (
          <button
            className="absolute right-3 top-1/2 -translate-y-1/2 bg-[linear-gradient(135deg,#00d4ff,#0099cc)] text-white border border-[rgba(0,212,255,0.8)] py-1.5 px-2.5 rounded cursor-pointer text-[10px] font-semibold uppercase tracking-[0.4px] transition-all duration-200 shadow-[0_6px_20px_rgba(0,212,255,0.4),0_0_16px_rgba(0,212,255,0.3),inset_0_1px_0_rgba(255,255,255,0.3)] [text-shadow:0_0_8px_rgba(0,212,255,0.8),0_2px_4px_rgba(0,0,0,0.6)] hover:bg-[linear-gradient(135deg,#00b8e6,#0088bb)] hover:-translate-y-[calc(50%+3px)] hover:shadow-[0_8px_28px_rgba(0,212,255,0.6),0_0_24px_rgba(0,212,255,0.5),inset_0_1px_0_rgba(255,255,255,0.4)] hover:[text-shadow:0_0_12px_rgba(0,212,255,1),0_2px_6px_rgba(0,0,0,0.7)]"
            onClick={onSkipAction}
          >
            {buttonText}
          </button>
        )}
      </div>
    </div>
  );
};

export default PlayerCard;
