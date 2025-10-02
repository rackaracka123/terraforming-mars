import React, { useState } from "react";
import MegaCreditIcon from "../display/MegaCreditIcon.tsx";
import VictoryPointIcon from "../display/VictoryPointIcon.tsx";
import BehaviorSection from "./BehaviorSection.tsx";
import RequirementsBox from "./RequirementsBox.tsx";
import { CardDto } from "@/types/generated/api-types.ts";

interface SimpleGameCardProps {
  card: CardDto;
  isSelected: boolean;
  onSelect: (cardId: string) => void;
  animationDelay?: number;
  showCheckbox?: boolean; // Whether to show the selection checkbox (default: false)
}

const SimpleGameCard: React.FC<SimpleGameCardProps> = ({
  card,
  isSelected,
  onSelect,
  animationDelay = 0,
  showCheckbox = false,
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
      "bg-[linear-gradient(135deg,#2d4a2f_0%,#1f3322_100%)] border border-[rgba(76,175,80,0.6)]",
    active:
      "bg-[linear-gradient(135deg,#1e3a5f_0%,#152d4a_100%)] border border-[rgba(33,150,243,0.6)]",
    event:
      "bg-[linear-gradient(135deg,#4a2b2b_0%,#3a1f1f_100%)] border border-[rgba(244,67,54,0.6)]",
    corporation:
      "bg-[linear-gradient(135deg,#4a3d1a_0%,#3a2f0d_100%)] border border-[rgba(255,193,7,0.6)]",
    prelude:
      "bg-[linear-gradient(135deg,#4a1e3a_0%,#3a152c_100%)] border border-[rgba(233,30,99,0.6)]",
  };

  const cardType = card.type as keyof typeof borderColors;

  // Get tag icon mapping from tags folder
  const getTagIcon = (tag: string) => {
    const iconMap: { [key: string]: string } = {
      power: "/assets/tags/power.png",
      science: "/assets/tags/science.png",
      space: "/assets/tags/space.png",
      building: "/assets/tags/building.png",
      city: "/assets/tags/city.png",
      jovian: "/assets/tags/jovian.png",
      earth: "/assets/tags/earth.png",
      microbe: "/assets/tags/microbe.png",
      animal: "/assets/tags/animal.png",
      plant: "/assets/tags/plant.png",
      event: "/assets/tags/event.png",
      venus: "/assets/tags/venus.png",
      wild: "/assets/tags/wild.png",
      mars: "/assets/tags/mars.png",
      moon: "/assets/tags/moon.png",
      clone: "/assets/tags/clone.png",
      crime: "/assets/tags/crime.png",
    };
    return iconMap[tag.toLowerCase()] || null;
  };

  return (
    <div
      className={`relative w-[200px] min-h-[280px] bg-[linear-gradient(135deg,#1a2332_0%,#0f1419_100%)] border-none rounded-lg p-4 cursor-pointer transition-all duration-300 opacity-0 translate-y-5 shadow-[0_4px_12px_rgba(0,0,0,0.3)] z-[1] animate-[fadeInUp_0.5s_ease_forwards] hover:-translate-y-1 hover:scale-[1.02] hover:shadow-[0_8px_32px_rgba(30,100,200,0.2)] ${isSelected ? "bg-[linear-gradient(135deg,#1a2332_0%,#203040_100%)] shadow-[0_0_20px_rgba(74,144,226,0.4)]" : ""} max-md:w-[160px] max-md:min-h-[240px] max-md:p-3 group`}
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
      {card.tags && card.tags.length > 0 && (
        <div className="absolute -top-[15px] right-5 flex gap-0.5 z-[3] items-center justify-center max-md:-top-3 max-md:right-4">
          {card.tags.slice(0, 3).map((tag, index) => {
            const iconSrc = getTagIcon(tag);
            return iconSrc ? (
              <div
                key={index}
                className="flex items-center justify-center shrink-0"
              >
                <img
                  src={iconSrc}
                  alt={tag}
                  className="w-[30px] h-[30px] object-contain [filter:drop-shadow(0_2px_6px_rgba(0,0,0,0.7))]"
                />
              </div>
            ) : null;
          })}
        </div>
      )}

      {/* Cost in top-left */}
      <div className="absolute -top-3 -left-3 flex items-center justify-start z-[3] shrink-0 max-md:-top-2.5 max-md:-left-2.5">
        <MegaCreditIcon value={card.cost} size="medium" />
      </div>

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
        <h3
          className={`text-base font-semibold text-white leading-[1.2] text-center flex items-center justify-center w-[90%] h-10 rounded-none p-1 px-2 shadow-[0_3px_6px_rgba(0,0,0,0.4)] my-0 mx-auto bg-[#1a2332] max-md:text-sm max-md:h-8 ${cardType && titleStyles[cardType] ? titleStyles[cardType] : ""}`}
        >
          {card.name}
        </h3>
      </div>

      {/* Selection indicator at bottom center, peeking out (only shown when showCheckbox is true) */}
      {showCheckbox && (
        <div className="absolute -bottom-3 left-1/2 -translate-x-1/2 z-[2] max-md:-bottom-2.5">
          <div
            className={`w-6 h-6 rounded-full bg-[#2a3142] border-2 border-[rgba(100,150,200,0.3)] flex items-center justify-center transition-all duration-300 ${isSelected ? "bg-[#4a90e2] border-[#4a90e2]" : ""}`}
          >
            {isSelected && (
              <span className="text-white text-sm font-bold">✓</span>
            )}
          </div>
        </div>
      )}

      {/* Behavior section */}
      <BehaviorSection behaviors={card.behaviors} />

      {/* Victory Points icon in bottom right */}
      <div className="absolute bottom-2 right-2 z-[3] pointer-events-none">
        <VictoryPointIcon vpConditions={card.vpConditions} size="large" />
      </div>

      {/* Hover effect border */}
      <div className="absolute -inset-px rounded-lg bg-[linear-gradient(45deg,transparent,rgba(255,255,255,0.1),transparent)] opacity-0 transition-opacity duration-300 pointer-events-none group-hover:opacity-100" />
    </div>
  );
};

export default SimpleGameCard;
