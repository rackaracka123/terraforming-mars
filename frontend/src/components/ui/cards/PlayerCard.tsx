import React, { useEffect, useState, useRef } from "react";
import { createPortal } from "react-dom";
import {
  PlayerDto,
  OtherPlayerDto,
  TriggeredEffectDto,
  ResourceType,
  ResourceTypeCredit,
} from "@/types/generated/api-types.ts";
import GameIcon from "../display/GameIcon.tsx";

interface EffectNotification {
  id: string;
  effect: TriggeredEffectDto;
  visible: boolean;
}

interface PlayerCardProps {
  player: PlayerDto | OtherPlayerDto;
  playerColor: string;
  isCurrentPlayer: boolean;
  isCurrentTurn: boolean;
  isActionPhase: boolean;
  onSkipAction?: () => void;
  totalPlayers?: number;
  hasPendingTilePlacement?: boolean;
  triggeredEffects?: TriggeredEffectDto[];
}

const PlayerCard: React.FC<PlayerCardProps> = ({
  player,
  playerColor,
  isCurrentPlayer,
  isCurrentTurn,
  isActionPhase,
  onSkipAction,
  totalPlayers = 1,
  hasPendingTilePlacement = false,
  triggeredEffects = [],
}) => {
  const isPassed = player.passed;
  const isDisconnected = !player.isConnected;
  const hasUnlimitedActions = player.availableActions === -1;
  const actionsRemaining = player.availableActions;

  // Determine button text - PASS for unlimited actions or 2 actions remaining, otherwise SKIP
  const buttonText = hasUnlimitedActions || actionsRemaining === 2 ? "PASS" : "SKIP";

  // Notification state for triggered effects
  const [notifications, setNotifications] = useState<EffectNotification[]>([]);

  // Track processed effect batches to avoid duplicates
  const processedBatchesRef = useRef<Set<string>>(new Set());

  // Filter effects for this player and manage notifications
  useEffect(() => {
    const playerEffects = triggeredEffects.filter((e) => e.playerId === player.id);
    if (playerEffects.length === 0) return;

    // Create a unique batch ID based on the effects
    const batchId = playerEffects
      .map((e) => `${e.cardName}-${e.outputs.map((o) => `${o.type}:${o.amount}`).join(",")}`)
      .join("|");

    // Skip if we've already processed this exact batch
    if (processedBatchesRef.current.has(batchId)) return;
    processedBatchesRef.current.add(batchId);

    // Add new notifications
    const newNotificationIds: string[] = [];
    const newNotifications = playerEffects.map((effect, i) => {
      const id = `${Date.now()}-${i}-${Math.random()}`;
      newNotificationIds.push(id);
      return {
        id,
        effect,
        visible: true,
      };
    });

    setNotifications((prev) => [...prev, ...newNotifications]);

    // Auto-dismiss after 3 seconds (not tied to effect cleanup)
    setTimeout(() => {
      setNotifications((prev) =>
        prev.map((n) => (newNotificationIds.includes(n.id) ? { ...n, visible: false } : n)),
      );
      // Remove from DOM after fade out
      setTimeout(() => {
        setNotifications((prev) => prev.filter((n) => !newNotificationIds.includes(n.id)));
        // Clean up processed batch after removal
        processedBatchesRef.current.delete(batchId);
      }, 300);
    }, 3000);
  }, [triggeredEffects, player.id]);

  // Helper to check if a resource type is credits or credit-production
  const isCreditsType = (type: string): boolean => {
    return (
      type === ResourceTypeCredit ||
      type === "credits" ||
      type === "credit-production" ||
      type === "credits-production"
    );
  };

  // Ref for positioning notifications relative to the card
  const cardRef = useRef<HTMLDivElement>(null);
  const [cardRect, setCardRect] = useState<DOMRect | null>(null);

  // Update card position when notifications change
  useEffect(() => {
    if (notifications.length > 0 && cardRef.current) {
      setCardRect(cardRef.current.getBoundingClientRect());
    }
  }, [notifications.length]);

  return (
    <div
      ref={cardRef}
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
        {isCurrentPlayer && isCurrentTurn && isActionPhase && !hasPendingTilePlacement && (
          <button
            className="absolute right-3 top-1/2 -translate-y-1/2 bg-[linear-gradient(135deg,#00d4ff,#0099cc)] text-white border border-[rgba(0,212,255,0.8)] py-1.5 px-2.5 rounded cursor-pointer text-[10px] font-semibold uppercase tracking-[0.4px] transition-all duration-200 shadow-[0_6px_20px_rgba(0,212,255,0.4),0_0_16px_rgba(0,212,255,0.3),inset_0_1px_0_rgba(255,255,255,0.3)] [text-shadow:0_0_8px_rgba(0,212,255,0.8),0_2px_4px_rgba(0,0,0,0.6)] hover:bg-[linear-gradient(135deg,#00b8e6,#0088bb)] hover:-translate-y-[calc(50%+3px)] hover:shadow-[0_8px_28px_rgba(0,212,255,0.6),0_0_24px_rgba(0,212,255,0.5),inset_0_1px_0_rgba(255,255,255,0.4)] hover:[text-shadow:0_0_12px_rgba(0,212,255,1),0_2px_6px_rgba(0,0,0,0.7)]"
            onClick={onSkipAction}
          >
            {buttonText}
          </button>
        )}
      </div>

      {/* Triggered effect notifications - rendered via portal to avoid clipping */}
      {notifications.length > 0 &&
        cardRect &&
        createPortal(
          <div
            className="fixed flex flex-row gap-1 z-[9999] pointer-events-none"
            style={{
              left: `${cardRect.right + 10}px`,
              top: `${cardRect.top + cardRect.height / 2}px`,
              transform: "translateY(-50%)",
            }}
          >
            <style>{`
            @keyframes notificationEnter {
              0% {
                opacity: 0;
                transform: translateX(50px);
              }
              100% {
                opacity: 1;
                transform: translateX(0);
              }
            }
            @keyframes notificationExit {
              0% {
                opacity: 1;
                transform: translateX(0);
              }
              100% {
                opacity: 0;
                transform: translateX(-50px);
              }
            }
          `}</style>
            {/* Group notifications by card name */}
            {(() => {
              const grouped = new Map<
                string,
                {
                  ids: string[];
                  outputs: (typeof notifications)[0]["effect"]["outputs"];
                  visible: boolean;
                }
              >();

              for (const { id, effect, visible } of notifications) {
                const existing = grouped.get(effect.cardName);
                if (existing) {
                  existing.ids.push(id);
                  existing.outputs.push(...effect.outputs);
                  existing.visible = existing.visible && visible;
                } else {
                  grouped.set(effect.cardName, {
                    ids: [id],
                    outputs: [...effect.outputs],
                    visible,
                  });
                }
              }

              return Array.from(grouped.entries()).map(([cardName, { ids, outputs, visible }]) => (
                <div
                  key={ids.join("-")}
                  className="flex items-center gap-1.5 px-2 py-1.5 rounded-lg bg-[rgba(20,30,50,0.95)] border border-[rgba(100,150,255,0.3)] shadow-lg whitespace-nowrap"
                  style={{
                    animation: visible
                      ? "notificationEnter 0.3s ease-out forwards"
                      : "notificationExit 0.3s ease-in forwards",
                  }}
                >
                  {/* Card name */}
                  <span className="text-white text-xs font-medium">{cardName}</span>

                  {/* Output icons */}
                  <div className="flex items-center gap-2">
                    {outputs.map((output, i) => (
                      <div key={i} className="flex items-center">
                        {isCreditsType(output.type) ? (
                          // Credits: embed amount inside icon
                          <GameIcon
                            iconType={output.type as ResourceType}
                            amount={Math.abs(output.amount)}
                            size="small"
                          />
                        ) : (
                          // Other resources: show icon with amount on right
                          <>
                            <GameIcon iconType={output.type as ResourceType} size="small" />
                            {output.amount !== 0 && (
                              <span className="text-white text-xs font-bold ml-0.5">
                                {output.amount > 0 ? `+${output.amount}` : output.amount}
                              </span>
                            )}
                          </>
                        )}
                      </div>
                    ))}
                  </div>
                </div>
              ));
            })()}
          </div>,
          document.body,
        )}
    </div>
  );
};

export default PlayerCard;
