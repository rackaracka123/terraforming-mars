import React from "react";
import styles from "./SimpleGameCard.module.css";

interface SimpleGameCardProps {
  card: {
    id: string;
    name: string;
    cost: number;
    tags?: string[];
    description?: string;
    requirements?: string;
  };
  isSelected: boolean;
  isFree?: boolean;
  onSelect: (cardId: string) => void;
  animationDelay?: number;
}

const SimpleGameCard: React.FC<SimpleGameCardProps> = ({
  card,
  isSelected,
  isFree = false,
  onSelect,
  animationDelay = 0,
}) => {
  const handleClick = () => {
    onSelect(card.id);
  };

  return (
    <div
      className={`${styles.cardContainer} ${isSelected ? styles.selected : ""} ${styles.fadeIn}`}
      style={{ animationDelay: `${animationDelay}ms` }}
      onClick={handleClick}
    >
      {/* Selection indicator */}
      <div className={styles.selectionIndicator}>
        <div className={styles.checkbox}>
          {isSelected && <span className={styles.checkmark}>✓</span>}
        </div>
      </div>

      {/* Card content */}
      <div className={styles.cardContent}>
        {/* Card header with name and cost */}
        <div className={styles.cardHeader}>
          <h3 className={styles.cardName}>{card.name}</h3>
          <div className={styles.costDisplay}>
            {isFree ? (
              <span className={styles.freeLabel}>FREE</span>
            ) : (
              <>
                <img
                  src="/assets/resources/megacredit.png"
                  alt="MC"
                  className={styles.costIcon}
                />
                <span className={styles.costAmount}>{card.cost}</span>
              </>
            )}
          </div>
        </div>

        {/* Tags if present */}
        {card.tags && card.tags.length > 0 && (
          <div className={styles.tags}>
            {card.tags.map((tag, index) => (
              <span key={index} className={styles.tag}>
                {tag}
              </span>
            ))}
          </div>
        )}

        {/* Requirements if present */}
        {card.requirements && (
          <div className={styles.requirements}>
            <span className={styles.requirementLabel}>Requires:</span>
            <span className={styles.requirementText}>{card.requirements}</span>
          </div>
        )}

        {/* Card description */}
        {card.description && (
          <div className={styles.description}>{card.description}</div>
        )}
      </div>

      {/* Hover effect border */}
      <div className={styles.hoverBorder} />
    </div>
  );
};

export default SimpleGameCard;