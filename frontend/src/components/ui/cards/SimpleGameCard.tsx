import React from "react";
import styles from "./SimpleGameCard.module.css";
import MegaCreditIcon from "../display/MegaCreditIcon.tsx";
import VictoryPointIcon from "../display/VictoryPointIcon.tsx";
import ProductionIcon from "../display/ProductionIcon.tsx";
import { CardDto } from "../../../types/generated/api-types.ts";

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
  const handleClick = () => {
    onSelect(card.id);
  };

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

  // Helper function to get requirement components
  const getRequirementComponents = (requirements: any) => {
    const components = [];

    // Check if requirements exists before accessing properties
    if (!requirements) {
      return components;
    }

    if (requirements.minTemperature !== undefined) {
      components.push({
        type: "temperature",
        value: `${requirements.minTemperature}°C`,
        icon: "/assets/global-parameters/temperature.png",
      });
    }
    if (requirements.maxTemperature !== undefined) {
      components.push({
        type: "max-temperature",
        value: `max ${requirements.maxTemperature}°C`,
        icon: "/assets/global-parameters/temperature.png",
      });
    }
    if (requirements.minOxygen !== undefined) {
      components.push({
        type: "oxygen",
        value: `${requirements.minOxygen}%`,
        icon: "/assets/global-parameters/oxygen.png",
      });
    }
    if (requirements.maxOxygen !== undefined) {
      components.push({
        type: "max-oxygen",
        value: `max ${requirements.maxOxygen}%`,
        icon: "/assets/global-parameters/oxygen.png",
      });
    }
    if (requirements.minOceans !== undefined) {
      components.push({
        type: "oceans",
        value: requirements.minOceans.toString(),
        icon: "/assets/global-parameters/ocean.png",
      });
    }
    if (requirements.maxOceans !== undefined) {
      components.push({
        type: "max-oceans",
        value: `max ${requirements.maxOceans}`,
        icon: "/assets/global-parameters/ocean.png",
      });
    }
    // Handle production requirements
    if (requirements.requiredProduction) {
      if (requirements.requiredProduction.energy > 0) {
        components.push({
          type: "energy-production",
          value: requirements.requiredProduction.energy.toString(),
          isProduction: true,
          resourceIcon: "/assets/resources/power.png",
        });
      }
      if (requirements.requiredProduction.steel > 0) {
        components.push({
          type: "steel-production",
          value: requirements.requiredProduction.steel.toString(),
          isProduction: true,
          resourceIcon: "/assets/resources/steel.png",
        });
      }
      if (requirements.requiredProduction.titanium > 0) {
        components.push({
          type: "titanium-production",
          value: requirements.requiredProduction.titanium.toString(),
          isProduction: true,
          resourceIcon: "/assets/resources/titanium.png",
        });
      }
      if (requirements.requiredProduction.plants > 0) {
        components.push({
          type: "plants-production",
          value: requirements.requiredProduction.plants.toString(),
          isProduction: true,
          resourceIcon: "/assets/resources/plants.png",
        });
      }
      if (requirements.requiredProduction.heat > 0) {
        components.push({
          type: "heat-production",
          value: requirements.requiredProduction.heat.toString(),
          isProduction: true,
          resourceIcon: "/assets/resources/heat.png",
        });
      }
      if (requirements.requiredProduction.credits > 0) {
        components.push({
          type: "credits-production",
          value: requirements.requiredProduction.credits.toString(),
          isProduction: true,
          resourceIcon: "/assets/resources/megacredit.png",
        });
      }
    }

    return components;
  };

  const requirementComponents = getRequirementComponents(card.requirements);
  const shouldShowRequirements = requirementComponents.length > 0;

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
        {/* Card header with cost, type icon, and name */}
        <div className={styles.cardHeader}>
          <div className={styles.topRow}>
            {/* Cost in top-left */}
            <div className={styles.costDisplay}>
              <MegaCreditIcon value={card.cost} size="medium" />
            </div>

            {/* Requirements in center */}
            {shouldShowRequirements && (
              <div className={styles.requirementsRow}>
                {requirementComponents.map((req, index) => (
                  <div key={index} className={styles.requirementBox}>
                    {req.isProduction ? (
                      <ProductionIcon
                        resourceIcon={req.resourceIcon}
                        value={req.value}
                        size="small"
                      />
                    ) : (
                      <>
                        <img
                          src={req.icon}
                          alt={req.type}
                          className={styles.requirementIcon}
                        />
                        <span className={styles.requirementValue}>
                          {req.value}
                        </span>
                      </>
                    )}
                  </div>
                ))}
              </div>
            )}

            {/* Card tag icons in top-right */}
            {card.tags && card.tags.length > 0 && (
              <div className={styles.tagIcons}>
                {card.tags.slice(0, 3).map((tag, index) => {
                  const iconSrc = getTagIcon(tag);
                  return iconSrc ? (
                    <div key={index} className={styles.cardTypeIcon}>
                      <img
                        src={iconSrc}
                        alt={tag}
                        className={styles.typeIcon}
                      />
                    </div>
                  ) : null;
                })}
              </div>
            )}
          </div>
          {/* Card name centered with type background */}
          <h3
            className={`${styles.cardName} ${card.type ? styles[card.type] || "" : ""}`}
          >
            {card.name}
          </h3>
        </div>

        {/* Card description */}
        {card.description && (
          <div className={styles.description}>{card.description}</div>
        )}
      </div>

      {/* Victory Points icon in bottom right */}
      <div className={styles.victoryPointsCorner}>
        <VictoryPointIcon value={card.victoryPoints} size="large" />
      </div>

      {/* Hover effect border */}
      <div className={styles.hoverBorder} />
    </div>
  );
};

export default SimpleGameCard;
