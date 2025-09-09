import React, { useState, useEffect } from "react";
import { ProductionPhaseStartedPayload } from "@/types/generated/api-types.ts";
import styles from "./ProductionPhaseModal.module.css";

interface ProductionPhaseModalProps {
  isOpen: boolean;
  productionData: ProductionPhaseStartedPayload | null;
  onClose: () => void;
}

const ProductionPhaseModal: React.FC<ProductionPhaseModalProps> = ({
  isOpen,
  productionData,
  onClose,
}) => {
  const [currentPlayerIndex, setCurrentPlayerIndex] = useState(0);
  const [animationStep, setAnimationStep] = useState<"energy" | "production">(
    "energy",
  );
  const [isAnimating, setIsAnimating] = useState(false);

  // Resource icons mapping
  const resourceIcons = {
    credits: "/assets/resources/megacredit.png",
    steel: "/assets/resources/steel.png",
    titanium: "/assets/resources/titanium.png",
    plants: "/assets/resources/plant.png",
    energy: "/assets/resources/power.png",
    heat: "/assets/resources/heat.png",
  };

  const resourceNames = {
    credits: "Credits",
    steel: "Steel",
    titanium: "Titanium",
    plants: "Plants",
    energy: "Energy",
    heat: "Heat",
  };

  // Auto-advance through players and animation steps
  useEffect(() => {
    if (!isOpen || !productionData || !isAnimating) return;

    const timer = setTimeout(() => {
      if (animationStep === "energy") {
        setAnimationStep("production");
      } else {
        // Move to next player or close
        if (currentPlayerIndex < productionData.playersData.length - 1) {
          setCurrentPlayerIndex(currentPlayerIndex + 1);
          setAnimationStep("energy");
        } else {
          // Animation complete - send ready message and close
          setTimeout(() => {
            setIsAnimating(false);

            // Send production-phase-ready message to server
            try {
              // webSocketService.productionPhaseReady();
            } catch (error) {
              void error;
            }

            onClose();
          }, 1500);
        }
      }
    }, 2500); // 2.5 seconds per animation step

    return () => clearTimeout(timer);
  }, [
    isOpen,
    productionData,
    currentPlayerIndex,
    animationStep,
    isAnimating,
    onClose,
  ]);

  // Start animation when modal opens
  useEffect(() => {
    if (isOpen && productionData) {
      setCurrentPlayerIndex(0);
      setAnimationStep("energy");
      setIsAnimating(true);
    }
  }, [isOpen, productionData]);

  // Handle ESC key to close modal
  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        onClose();
      }
    };

    if (isOpen) {
      document.addEventListener("keydown", handleKeyDown);
      return () => document.removeEventListener("keydown", handleKeyDown);
    }

    return () => { };
  }, [isOpen, onClose]);

  if (!isOpen || !productionData) return null;

  const currentPlayerData = productionData.playersData[currentPlayerIndex];
  if (!currentPlayerData) return null;

  const renderResourceAnimation = (
    resourceType: keyof typeof resourceIcons,
    beforeAmount: number,
    afterAmount: number,
    isActive: boolean,
  ) => {
    const change = afterAmount - beforeAmount;
    if (change === 0 && resourceType !== "energy") return null;

    return (
      <div
        key={resourceType}
        className={`${styles.resourceItem} ${isActive ? styles.active : ""}`}
        style={
          {
            "--player-color": 'red',//currentPlayerData.playerColor,
          } as React.CSSProperties
        }
      >
        <div className={styles.resourceIcon}>
          <img
            src={resourceIcons[resourceType]}
            alt={resourceNames[resourceType]}
          />
        </div>
        <div className={styles.resourceAmounts}>
          <div className={styles.beforeAmount}>{beforeAmount}</div>
          <div className={styles.arrow}>→</div>
          <div className={styles.afterAmount}>{afterAmount}</div>
          {change > 0 && (
            <div className={styles.changeIndicator}>+{change}</div>
          )}
        </div>
        <div className={styles.resourceLabel}>
          {resourceNames[resourceType]}
        </div>
      </div>
    );
  };

  const renderEnergyConversion = () => {
    const energyConverted = currentPlayerData.energyConverted;
    if (energyConverted === 0) return null;

    return (
      <div className={styles.energyConversion}>
        <div className={styles.conversionTitle}>Energy → Heat Conversion</div>
        <div className={styles.conversionAnimation}>
          <div className={styles.conversionItem}>
            <img src={resourceIcons.energy} alt="Energy" />
            <span>{energyConverted}</span>
          </div>
          <div className={styles.conversionArrow}>⚡→🔥</div>
          <div className={styles.conversionItem}>
            <img src={resourceIcons.heat} alt="Heat" />
            <span>+{energyConverted}</span>
          </div>
        </div>
      </div>
    );
  };

  const renderProductionPhase = () => {
    return (
      <div className={styles.productionPhase}>
        <div className={styles.productionTitle}>Resource Production</div>
        <div className={styles.resourcesGrid}>
          {renderResourceAnimation(
            "credits",
            currentPlayerData.beforeResources.credits,
            currentPlayerData.afterResources.credits,
            animationStep === "production",
          )}
          {renderResourceAnimation(
            "steel",
            currentPlayerData.beforeResources.steel,
            currentPlayerData.afterResources.steel,
            animationStep === "production",
          )}
          {renderResourceAnimation(
            "titanium",
            currentPlayerData.beforeResources.titanium,
            currentPlayerData.afterResources.titanium,
            animationStep === "production",
          )}
          {renderResourceAnimation(
            "plants",
            currentPlayerData.beforeResources.plants,
            currentPlayerData.afterResources.plants,
            animationStep === "production",
          )}
          {renderResourceAnimation(
            "energy",
            0, // Energy starts at 0 after conversion
            currentPlayerData.production.energy,
            animationStep === "production",
          )}
          {renderResourceAnimation(
            "heat",
            currentPlayerData.beforeResources.heat +
            currentPlayerData.energyConverted,
            currentPlayerData.afterResources.heat,
            animationStep === "production",
          )}
        </div>
        <div className={styles.productionInfo}>
          <div className={styles.creditsBreakdown}>
            Credits Income: {currentPlayerData.terraformRating} (TR) +{" "}
            {currentPlayerData.production.credits} (Production) ={" "}
            {currentPlayerData.creditsIncome}
          </div>
        </div>
      </div>
    );
  };

  return (
    <div className={styles.modalOverlay} onClick={onClose}>
      <div className={styles.modalPopup} onClick={(e) => e.stopPropagation()}>
        <div className={styles.modalHeader}>
          <h2>Production Phase</h2>
          <div className={styles.generationInfo}>
            Generation {productionData.generation}
          </div>
          <button className={styles.closeBtn} onClick={onClose}>
            ×
          </button>
        </div>

        <div className={styles.playerProgress}>
          {productionData.playersData.map((player, index) => (
            <div
              key={player.playerId}
              className={`${styles.progressDot} ${index === currentPlayerIndex ? styles.current : ""
                } ${index < currentPlayerIndex ? styles.completed : ""}`}
              style={
                {
                  "--player-color": 'red',//currentPlayerData.playerColor
                } as React.CSSProperties
              }
            />
          ))}
        </div>

        <div className={styles.currentPlayerSection}>
          <div
            className={styles.playerHeader}
            style={
              {
                "--player-color": 'red',//currentPlayerData.playerColor,
              } as React.CSSProperties
            }
          >
            <div className={styles.playerName}>
              {currentPlayerData.playerName}
            </div>
            <div className={styles.animationStep}>
              {animationStep === "energy"
                ? "Energy Conversion"
                : "Resource Production"}
            </div>
          </div>

          <div className={styles.animationContent}>
            {animationStep === "energy"
              ? renderEnergyConversion()
              : renderProductionPhase()}
          </div>
        </div>
      </div>
    </div>
  );
};

export default ProductionPhaseModal;
