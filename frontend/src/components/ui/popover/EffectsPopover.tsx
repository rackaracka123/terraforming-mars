import React, { useEffect, useRef } from "react";
import { PlayerEffectDto } from "../../../types/generated/api-types.ts";
import BehaviorSection from "../cards/BehaviorSection.tsx";

interface EffectsPopoverProps {
  isVisible: boolean;
  onClose: () => void;
  effects: PlayerEffectDto[];
  playerName?: string;
  onOpenDetails?: () => void;
  anchorRef: React.RefObject<HTMLElement>;
}

const EffectsPopover: React.FC<EffectsPopoverProps> = ({
  isVisible,
  onClose,
  effects,
  playerName: _playerName = "Player",
  onOpenDetails,
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

  // No conversion needed - PlayerEffectDto now contains CardBehaviorDto directly

  return (
    <div
      className="fixed bottom-[85px] right-[30px] w-[320px] max-h-[400px] bg-[linear-gradient(145deg,rgba(20,30,45,0.98)_0%,rgba(30,40,60,0.95)_100%)] border-2 border-[rgba(255,150,255,0.6)] rounded-xl shadow-[0_15px_40px_rgba(0,0,0,0.8),0_0_30px_rgba(255,150,255,0.3)] [backdrop-filter:blur(15px)] z-[10001] animate-[popoverSlideUp_0.3s_ease-out] flex flex-col overflow-hidden isolate pointer-events-auto max-[768px]:w-[280px] max-[768px]:right-[15px] max-[768px]:bottom-[70px]"
      ref={popoverRef}
    >
      <div className="absolute -bottom-2 right-[50px] w-0 h-0 border-l-[8px] border-l-transparent border-r-[8px] border-r-transparent border-t-[8px] border-t-[rgba(255,150,255,0.6)] max-[768px]:right-[40px]" />

      <div className="flex items-center justify-between py-[15px] px-5 bg-[linear-gradient(90deg,rgba(50,20,50,0.9)_0%,rgba(60,30,60,0.7)_100%)] border-b border-b-[rgba(255,150,255,0.3)]">
        <div className="flex items-center gap-2.5">
          <h3 className="m-0 text-white text-base font-bold [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)]">
            Card Effects
          </h3>
        </div>
        <div className="flex items-center gap-2">
          <div className="text-white/80 text-xs bg-[rgba(255,150,255,0.2)] py-1 px-2 rounded-md border border-[rgba(255,150,255,0.3)]">
            {effects.length} active
          </div>
          {onOpenDetails && (
            <button
              className="bg-[linear-gradient(135deg,rgba(150,100,255,0.8)_0%,rgba(120,80,200,0.9)_100%)] border border-[rgba(150,100,255,0.6)] rounded-md text-white text-[11px] font-semibold py-1 px-2.5 cursor-pointer transition-all duration-200 [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] pointer-events-auto relative z-[1] hover:bg-[linear-gradient(135deg,rgba(150,100,255,1)_0%,rgba(120,80,200,1)_100%)] hover:border-[rgba(150,100,255,0.8)] hover:-translate-y-px hover:shadow-[0_2px_8px_rgba(150,100,255,0.3)]"
              onClick={() => {
                onOpenDetails();
                onClose();
              }}
              title="Open detailed effects view"
            >
              Details
            </button>
          )}
        </div>
      </div>

      <div className="flex-1 overflow-y-auto [scrollbar-width:thin] [scrollbar-color:rgba(255,150,255,0.5)_rgba(50,75,125,0.3)] [&::-webkit-scrollbar]:w-1.5 [&::-webkit-scrollbar-track]:bg-[rgba(50,75,125,0.3)] [&::-webkit-scrollbar-track]:rounded [&::-webkit-scrollbar-thumb]:bg-[rgba(255,150,255,0.5)] [&::-webkit-scrollbar-thumb]:rounded [&::-webkit-scrollbar-thumb:hover]:bg-[rgba(255,150,255,0.7)]">
        {effects.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-10 px-5 text-center">
            <img
              src="/assets/misc/asterisc.png"
              alt="No effects"
              className="w-10 h-10 mb-[15px] opacity-60"
            />
            <div className="text-white text-sm font-medium mb-2">
              No card effects active
            </div>
            <div className="text-white/60 text-xs leading-[1.4]">
              Play cards with ongoing effects to gain bonuses
            </div>
          </div>
        ) : (
          <div className="p-2 flex flex-col gap-2">
            {effects.map((effect, index) => (
              <div
                key={`${effect.cardId}-${effect.behaviorIndex}`}
                className="flex items-center gap-3 py-2.5 px-[15px] bg-[linear-gradient(135deg,rgba(60,30,90,0.4)_0%,rgba(40,20,70,0.3)_100%)] border border-[rgba(255,150,255,0.3)] rounded-lg transition-all duration-300 animate-[effectSlideIn_0.4s_ease-out_both] max-[768px]:py-2 max-[768px]:px-3"
                style={{
                  animationDelay: `${index * 0.05}s`,
                }}
              >
                <div className="flex flex-col gap-2 flex-1">
                  <div className="text-white/70 text-[11px] font-medium uppercase tracking-[0.5px] [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] leading-[1.2] opacity-80 flex items-center gap-2 max-[768px]:text-[11px]">
                    {effect.cardName}
                  </div>

                  <div className="relative w-full min-h-[32px] [&>div]:!relative [&>div]:!bottom-auto [&>div]:!left-auto [&>div]:!right-auto [&>div]:w-full [&>div:hover]:!transform-none [&>div:hover]:!shadow-none [&>div:hover]:!filter-none">
                    <BehaviorSection
                      behaviors={[effect.behavior]}
                      greyOutAll={false}
                    />
                  </div>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default EffectsPopover;
