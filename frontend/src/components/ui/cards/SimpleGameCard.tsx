import React, { useState } from "react";
import GameIcon from "../display/GameIcon.tsx";
import VictoryPointIcon from "../display/VictoryPointIcon.tsx";
import BehaviorSection from "./BehaviorSection";
import RequirementsBox from "./RequirementsBox.tsx";
import { CardDto, ResourceTypeCredits } from "@/types/generated/api-types.ts";

interface SimpleGameCardProps {
  card: CardDto;
  isSelected: boolean;
  onSelect: (cardId: string) => void;
  animationDelay?: number;
  showCheckbox?: boolean; // Whether to show the selection checkbox (default: false)
  discountAmount?: number; // Optional discount to apply to card cost
}

const SimpleGameCard: React.FC<SimpleGameCardProps> = ({
  card,
  isSelected,
  onSelect,
  animationDelay = 0,
  showCheckbox = false,
  discountAmount = 0,
}) => {
  const [imageError, setImageError] = useState(false);
  const [imageLoaded, setImageLoaded] = useState(false);

  const handleClick = () => {
    onSelect(card.id);
  };

  const handleImageLoad = () => {
    setImageLoaded(true);
  };

  const handleImageError = () => {
    setImageError(true);
  };

  const cardImagePath = `/assets/cards/${card.id}.webp`;

  // Card type specific styles
  const borderColors = {
    automated: "border-[#4caf50]",
    active: "border-[#2196f3]",
    event: "border-[#f44336]",
    corporation: "border-[#ffc107]",
    prelude: "border-[#e91e63]",
  };

  const titleStyles = {
    automated:
      "bg-[linear-gradient(135deg,#0a1a0d_0%,#050f08_100%)] border border-[rgba(76,175,80,0.4)]",
    active:
      "bg-[linear-gradient(135deg,#0a1520_0%,#050a15_100%)] border border-[rgba(33,150,243,0.4)]",
    event:
      "bg-[linear-gradient(135deg,#1a0a0a_0%,#0f0505_100%)] border border-[rgba(244,67,54,0.4)]",
    corporation:
      "bg-[linear-gradient(135deg,#1a1508_0%,#0f0a04_100%)] border border-[rgba(255,193,7,0.4)]",
    prelude:
      "bg-[linear-gradient(135deg,#1a0a15_0%,#0f050a_100%)] border border-[rgba(233,30,99,0.4)]",
  };

  // Card type specific background colors (near-black with barely visible accent tint)
  const cardBackgrounds = {
    automated: "bg-[rgba(2,5,2,0.98)]", // Near-black green tint
    active: "bg-[rgba(2,4,6,0.98)]", // Near-black blue tint
    event: "bg-[rgba(5,2,2,0.98)]", // Near-black red tint
    corporation: "bg-[rgba(5,4,2,0.98)]", // Near-black yellow tint
    prelude: "bg-[rgba(5,2,4,0.98)]", // Near-black pink tint
  };

  // Card type specific glow colors for selected state (halo effect behind card)
  const cardGlows = {
    automated:
      "shadow-[0_4px_20px_rgba(76,175,80,0.3),0_0_40px_rgba(76,175,80,0.2)]", // Green halo
    active:
      "shadow-[0_4px_20px_rgba(33,150,243,0.3),0_0_40px_rgba(33,150,243,0.2)]", // Blue halo
    event:
      "shadow-[0_4px_20px_rgba(244,67,54,0.3),0_0_40px_rgba(244,67,54,0.2)]", // Red halo
    corporation:
      "shadow-[0_4px_20px_rgba(255,193,7,0.3),0_0_40px_rgba(255,193,7,0.2)]", // Yellow halo
    prelude:
      "shadow-[0_4px_20px_rgba(233,30,99,0.3),0_0_40px_rgba(233,30,99,0.2)]", // Pink halo
  };

  // Card type specific checkbox colors (darker background, matching card border)
  const checkboxColors = {
    automated: { bg: "bg-[#1f3322]", border: borderColors.automated },
    active: { bg: "bg-[#152d4a]", border: borderColors.active },
    event: { bg: "bg-[#3a1f1f]", border: borderColors.event },
    corporation: { bg: "bg-[#3a2f0d]", border: borderColors.corporation },
    prelude: { bg: "bg-[#3a152c]", border: borderColors.prelude },
  };

  const cardType = card.type as keyof typeof borderColors;
  const cardBg =
    cardType && cardBackgrounds[cardType]
      ? cardBackgrounds[cardType]
      : "bg-[rgba(0,0,0,0.9)]";
  const cardGlow =
    cardType && cardGlows[cardType]
      ? cardGlows[cardType]
      : "shadow-[0_0_20px_rgba(74,144,226,0.25)]";
  const checkboxColor =
    cardType && checkboxColors[cardType]
      ? checkboxColors[cardType]
      : { bg: "bg-[#4a90e2]", border: "border-[#4a90e2]" };

  return (
    <div
      className={`relative w-[200px] min-h-[280px] ${cardBg} border-none rounded-lg p-4 cursor-pointer transition-all duration-200 opacity-0 translate-y-5 shadow-[0_4px_12px_rgba(0,0,0,0.3)] z-[1] animate-[fadeInUp_0.5s_ease_forwards] ${isSelected ? `brightness-110 ${cardGlow} hover:${cardGlow}` : "hover:shadow-[0_6px_20px_rgba(30,100,200,0.15)]"} max-md:w-[160px] max-md:min-h-[240px] max-md:p-3 group`}
      style={{ animationDelay: `${animationDelay}ms` }}
      onClick={handleClick}
    >
      {/* Requirements box */}
      <RequirementsBox requirements={card.requirements} />

      {/* Futuristic card border */}
      <div
        className={`absolute top-0 left-0 right-0 bottom-0 rounded-[5px] border-2 pointer-events-none z-[1] ${cardType && borderColors[cardType] ? borderColors[cardType] : "border-[#4a90e2]"}`}
      ></div>
      {/* Tags at the very top, peeking out */}
      {((card.tags && card.tags.length > 0) || card.type === "event") && (
        <div className="absolute -top-[15px] right-0 flex gap-0.5 z-[3] items-center justify-center max-md:-top-3 max-md:right-2">
          {/* Show other tags first (limit total to 3 including event tag) */}
          {card.tags &&
            card.tags
              .slice(0, card.type === "event" ? 2 : 3)
              .map((tag, index) => (
                <div
                  key={index}
                  className="flex items-center justify-center shrink-0 [filter:drop-shadow(0_2px_6px_rgba(0,0,0,0.7))]"
                >
                  <GameIcon iconType={tag.toLowerCase()} size="medium" />
                </div>
              ))}
          {/* Show event tag icon last (right-most) if card type is event */}
          {card.type === "event" && (
            <div className="flex items-center justify-center shrink-0 [filter:drop-shadow(0_2px_6px_rgba(0,0,0,0.7))]">
              <GameIcon iconType="event" size="medium" />
            </div>
          )}
        </div>
      )}

      {/* Cost in top-left */}
      {discountAmount > 0 ? (
        // Stacked display when discounted
        <div className="absolute -top-3 -left-3 flex flex-col items-center z-[3] shrink-0 max-md:-top-2.5 max-md:-left-2.5">
          {/* Original cost (faded) */}
          <div className="grayscale-[0.7]">
            <GameIcon
              iconType={ResourceTypeCredits}
              amount={card.cost}
              size="medium"
            />
          </div>
          {/* Downward arrow - centered container */}
          <div className="w-full flex justify-center items-center">
            <svg
              width="12"
              height="8"
              viewBox="0 0 12 8"
              className="opacity-70 my-[6px]"
            >
              <path
                d="M6 8 L0 0 L3 0 L6 5 L9 0 L12 0 Z"
                fill="rgba(76, 175, 80, 0.9)"
              />
            </svg>
          </div>
          {/* Discounted cost (clear) */}
          <div>
            <GameIcon
              iconType={ResourceTypeCredits}
              amount={Math.max(0, card.cost - discountAmount)}
              size="medium"
            />
          </div>
        </div>
      ) : (
        // Single icon when no discount
        <div className="absolute -top-3 -left-3 flex items-center justify-start z-[3] shrink-0 max-md:-top-2.5 max-md:-left-2.5">
          <GameIcon
            iconType={ResourceTypeCredits}
            amount={card.cost}
            size="medium"
          />
        </div>
      )}

      {/* Image area */}
      <div className="absolute top-5 left-4 right-4 h-[35%] bg-white/5 rounded border border-dashed border-white/20 z-[1] overflow-hidden max-md:top-4 max-md:left-3 max-md:right-3">
        {!imageError && (
          <img
            src={cardImagePath}
            alt={card.name}
            className={`w-full h-full object-cover rounded border opacity-0 transition-opacity duration-300 ${cardType && borderColors[cardType] ? borderColors[cardType] : "border-[#4a90e2]"} ${imageLoaded ? "opacity-100" : ""}`}
            onLoad={handleImageLoad}
            onError={handleImageError}
          />
        )}
        {/* Show placeholder only when image fails to load */}
        {imageError && (
          <div className="w-full h-full bg-white/5 rounded border border-dashed border-white/20">
            {/* Keep the current grey dashed border look */}
          </div>
        )}
      </div>

      {/* Card title at 40% from top */}
      <div className="absolute top-[40%] left-2 right-2 z-[3] max-md:px-0.5">
        {/* Victory Points icon overlapping the left side of title bar */}
        <div className="absolute -left-5 top-1/2 -translate-y-[calc(50%-5px)] z-[4] pointer-events-none scale-[1.25] max-md:-left-4">
          <VictoryPointIcon vpConditions={card.vpConditions} size="medium" />
        </div>
        <h3
          className={`${card.name.length > 18 ? "text-[12px]" : card.name.length > 20 ? "text-sm" : "text-base"} font-orbitron font-semibold text-white leading-[1.2] text-center flex items-center justify-center w-full h-[44px] rounded-none p-1 ${card.vpConditions ? "pl-[30px] pr-5" : "px-5"} shadow-[0_3px_6px_rgba(0,0,0,0.4)] my-0 mx-auto bg-[#1a2332] ${card.name.length > 28 ? "max-md:text-[9px]" : card.name.length > 20 ? "max-md:text-xs" : "max-md:text-sm"} max-md:h-[36px] ${card.vpConditions ? "max-md:pl-[25px] max-md:pr-3" : "max-md:px-3"} ${cardType && titleStyles[cardType] ? titleStyles[cardType] : ""}`}
        >
          {card.name}
        </h3>
      </div>

      {/* Content section - takes up roughly half the card height and vertically centers content */}
      <div className="absolute top-[calc(50%+20px)] left-2 right-2 bottom-4 flex items-center justify-center z-[2] max-md:top-[calc(50%+25px)] max-md:left-1.5 max-md:right-1.5 max-md:bottom-3">
        <BehaviorSection behaviors={card.behaviors} />
      </div>

      {/* Selection indicator at bottom center, peeking out (only shown when showCheckbox is true) */}
      {showCheckbox && (
        <div className="absolute -bottom-3 left-1/2 -translate-x-1/2 z-[2] max-md:-bottom-2.5">
          <div
            className={`w-6 h-6 rounded-full bg-[#2a3142] border-2 border-[rgba(100,150,200,0.3)] flex items-center justify-center transition-all duration-300 ${isSelected ? `${checkboxColor.bg} ${checkboxColor.border}` : ""}`}
          >
            {isSelected && (
              <span className="text-white text-sm font-bold">✓</span>
            )}
          </div>
        </div>
      )}

      {/* Hover effect border */}
      <div className="absolute -inset-px rounded-lg bg-[linear-gradient(45deg,transparent,rgba(255,255,255,0.1),transparent)] opacity-0 transition-opacity duration-300 pointer-events-none group-hover:opacity-100" />
    </div>
  );
};

export default SimpleGameCard;
