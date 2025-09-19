import React, { useState } from "react";
import styles from "./SimpleGameCard.module.css";
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
}

const SimpleGameCard: React.FC<SimpleGameCardProps> = ({
  card,
  isSelected,
  onSelect,
  animationDelay = 0,
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

  const cardImagePath = `/assets/cards/${card.id}.png`;

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
      className={`${styles.cardContainer} ${isSelected ? styles.selected : ""} ${styles.fadeIn} ${card.type ? styles[card.type] || "" : ""}`}
      style={{ animationDelay: `${animationDelay}ms` }}
      onClick={handleClick}
    >
      {/* Requirements box */}
      <RequirementsBox requirements={card.requirements} />

      {/* Futuristic card border */}
      <div className={styles.cardBorder}></div>
      {/* Tags at the very top, peeking out */}
      {card.tags && card.tags.length > 0 && (
        <div className={styles.tagIcons}>
          {card.tags.slice(0, 3).map((tag, index) => {
            const iconSrc = getTagIcon(tag);
            return iconSrc ? (
              <div key={index} className={styles.cardTypeIcon}>
                <img src={iconSrc} alt={tag} className={styles.typeIcon} />
              </div>
            ) : null;
          })}
        </div>
      )}

      {/* Cost in top-left */}
      <div className={styles.costDisplay}>
        <MegaCreditIcon value={card.cost} size="medium" />
      </div>

      {/* Image area */}
      <div className={styles.imageArea}>
        {!imageError && (
          <img
            src={cardImagePath}
            alt={card.name}
            className={`${styles.cardImage} ${imageLoaded ? styles.imageLoaded : ""}`}
            onLoad={handleImageLoad}
            onError={handleImageError}
          />
        )}
        {/* Show placeholder only when image fails to load */}
        {imageError && (
          <div className={styles.imagePlaceholder}>
            {/* Keep the current grey dashed border look */}
          </div>
        )}
      </div>

      {/* Card title at 40% from top */}
      <div className={styles.titleContainer}>
        <div className={styles.titleBorder}></div>
        <h3
          className={`${styles.cardName} ${card.type ? styles[card.type] || "" : ""}`}
        >
          {card.name}
        </h3>
      </div>

      {/* Selection indicator at bottom center, peeking out */}
      <div className={styles.selectionIndicator}>
        <div className={styles.checkbox}>
          {isSelected && <span className={styles.checkmark}>✓</span>}
        </div>
      </div>

      {/* Behavior section */}
      <BehaviorSection behaviors={card.behaviors} />

      {/* Victory Points icon in bottom right */}
      <div className={styles.victoryPointsCorner}>
        <VictoryPointIcon vpConditions={card.vpConditions} size="large" />
      </div>

      {/* Hover effect border */}
      <div className={styles.hoverBorder} />
    </div>
  );
};

export default SimpleGameCard;
