import React from "react";
import { PlayerEffectDto } from "../../../types/generated/api-types.ts";
import BehaviorSection from "../cards/BehaviorSection";
import GameIcon from "../display/GameIcon.tsx";
import { GamePopover, GamePopoverEmpty, GamePopoverItem } from "../GamePopover";

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
  onOpenDetails,
  anchorRef,
}) => {
  return (
    <GamePopover
      isVisible={isVisible}
      onClose={onClose}
      position={{ type: "anchor", anchorRef, placement: "above" }}
      theme="effects"
      header={{
        title: "Card Effects",
        badge: `${effects.length} active`,
        rightContent: onOpenDetails ? (
          <button
            className="bg-space-black-darker/90 border-2 border-[var(--popover-accent)] rounded-lg text-white text-[11px] font-semibold py-1 px-2.5 cursor-pointer transition-all duration-200 text-shadow-dark pointer-events-auto relative z-[1] hover:bg-space-black-darker/95 hover:-translate-y-px hover:shadow-[0_2px_8px_rgba(var(--popover-accent-rgb),0.4)]"
            onClick={() => {
              onOpenDetails();
              onClose();
            }}
            title="Open detailed effects view"
          >
            Details
          </button>
        ) : undefined,
      }}
      arrow={{ enabled: true, position: "right", offset: 30 }}
      width={320}
      maxHeight={400}
    >
      {effects.length === 0 ? (
        <GamePopoverEmpty
          icon={<GameIcon iconType="asterisk" size="medium" />}
          title="No card effects active"
          description="Play cards with ongoing effects to gain bonuses"
        />
      ) : (
        <div className="p-2 flex flex-col gap-2">
          {effects.map((effect, index) => (
            <GamePopoverItem
              key={`${effect.cardId}-${effect.behaviorIndex}`}
              state="available"
              hoverEffect="glow"
              animationDelay={index * 0.05}
            >
              <div className="flex flex-col gap-2 flex-1">
                <div className="text-white/70 text-[11px] font-medium uppercase tracking-[0.5px] [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] leading-[1.2] opacity-80 flex items-center gap-2 max-[768px]:text-[11px]">
                  {effect.cardName}
                </div>

                <div className="relative w-full min-h-[32px] [&>div]:!relative [&>div]:!bottom-auto [&>div]:!left-auto [&>div]:!right-auto [&>div]:w-full [&>div:hover]:!transform-none [&>div:hover]:!shadow-none [&>div:hover]:!filter-none">
                  <BehaviorSection behaviors={[effect.behavior]} greyOutAll={false} />
                </div>
              </div>
            </GamePopoverItem>
          ))}
        </div>
      )}
    </GamePopover>
  );
};

export default EffectsPopover;
