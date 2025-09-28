import React, { useState, useEffect, useMemo, useCallback } from "react";
import {
  GameDto,
  ProductionPhaseDto,
  OtherPlayerDto,
  PlayerDto,
} from "@/types/generated/api-types.ts";
import {
  RESOURCE_COLORS,
  RESOURCE_ICONS,
  RESOURCE_NAMES,
  ResourceType,
} from "@/utils/resourceColors.ts";
import { globalWebSocketManager } from "@/services/globalWebSocketManager.ts";
import styles from "./ProductionPhaseModal.module.css";

interface ProductionPhaseModalProps {
  isOpen: boolean;
  gameState: GameDto | null;
  onClose: () => void;
}

const ProductionPhaseModal: React.FC<ProductionPhaseModalProps> = ({
  isOpen,
  gameState,
  onClose,
}) => {
  const [hasSubmittedCardSelection, setHasSubmittedCardSelection] =
    useState(false);
  const [currentPlayerIndex, setCurrentPlayerIndex] = useState(0);
  const [animationStep, setAnimationStep] = useState<
    "energyConversion" | "production"
  >("energyConversion");
  const [isAnimating, setIsAnimating] = useState(false);
  const [showSummary, setShowSummary] = useState(false);
  const [resourceAnimationState, setResourceAnimationState] = useState<
    "initial" | "fadeInResources" | "showProduction" | "fadeOut" | "fadeIn"
  >("initial");
  const [energyAnimationState, setEnergyAnimationState] = useState<
    "initial" | "fadeOut" | "fadeIn"
  >("initial");

  // Handle closing the modal and sending card selection
  const handleClose = useCallback(async () => {
    // Only send card selection once per production phase
    if (
      !hasSubmittedCardSelection &&
      gameState?.currentPlayer?.productionPhase
    ) {
      try {
        // Send empty card selection (player not buying any cards)
        await globalWebSocketManager.selectCards([]);
        setHasSubmittedCardSelection(true);
        // Don't call onClose here - let the game state update handle closing the modal
        // The modal will close automatically when productionPhase is removed from the player
      } catch (error) {
        console.error("Failed to submit card selection:", error);
        // On error, still close the modal manually
        onClose();
      }
    } else {
      // If already submitted or no production data, just close normally
      onClose();
    }
  }, [hasSubmittedCardSelection, gameState, onClose]);

  // Reset submission flag when modal opens with new production data
  useEffect(() => {
    if (isOpen && gameState?.currentPlayer?.productionPhase) {
      setHasSubmittedCardSelection(false);
    }
  }, [isOpen, gameState?.currentPlayer?.productionPhase]);

  // Extract production data from game state
  const modalProductionData = useMemo(() => {
    if (!gameState || !gameState.currentPlayer?.productionPhase) {
      return null;
    }

    // Gather all players' production data
    const allPlayers: (PlayerDto | OtherPlayerDto)[] = [
      gameState.currentPlayer,
      ...gameState.otherPlayers,
    ];

    // Filter players that have production selection data
    const playersWithProduction = allPlayers.filter(
      (player) => player.productionPhase,
    );

    if (playersWithProduction.length === 0) {
      return null;
    }

    // Transform into the format the modal expects
    const playersData = playersWithProduction.map((player) => {
      const productionPhase = player.productionPhase as ProductionPhaseDto;
      // For the current player, we have full data
      if ("cards" in player) {
        const currentPlayer = player as PlayerDto;
        return {
          playerId: currentPlayer.id,
          playerName: currentPlayer.name,
          production: currentPlayer.resourceProduction,
          terraformRating: currentPlayer.terraformRating,
          ...productionPhase,
        };
      } else {
        // For other players, we have limited data
        const otherPlayer = player as OtherPlayerDto;
        return {
          playerId: otherPlayer.id,
          playerName: otherPlayer.name,
          production: otherPlayer.resourceProduction,
          terraformRating: otherPlayer.terraformRating,
          ...productionPhase,
        };
      }
    });

    return {
      playersData,
      generation: gameState.generation || 1,
    };
  }, [gameState]);

  // Resource configuration from utility
  const resourceIcons = RESOURCE_ICONS;
  const resourceNames = RESOURCE_NAMES;

  // Auto-advance through animation steps within each player (but not between players)
  useEffect(() => {
    if (!isAnimating) return;

    const timer = setTimeout(
      () => {
        if (animationStep === "energyConversion") {
          setAnimationStep("production");
        } else {
          // Production phase is done for this player, stop animating
          setIsAnimating(false);
        }
      },
      animationStep === "energyConversion" ? 4500 : 4000,
    ); // 4.5 seconds for energy (includes 1.5s hold time), 4 seconds for production (with pause)

    return () => clearTimeout(timer);
  }, [currentPlayerIndex, animationStep, isAnimating]);

  // Handle player selection
  const handlePlayerSelect = (playerIndex: number) => {
    if (playerIndex !== currentPlayerIndex || showSummary) {
      setShowSummary(false);
      setCurrentPlayerIndex(playerIndex);
      setAnimationStep("energyConversion");
      setIsAnimating(true);
      setResourceAnimationState("initial");
      setEnergyAnimationState("initial");
    }
  };

  // Energy conversion animation sequence
  useEffect(() => {
    if (!isAnimating || animationStep !== "energyConversion") return;

    let timeoutId: NodeJS.Timeout;

    if (energyAnimationState === "initial") {
      // Wait 2 seconds to show change indicators longer, then start fade out
      timeoutId = setTimeout(() => {
        setEnergyAnimationState("fadeOut");
      }, 2000);
    } else if (energyAnimationState === "fadeOut") {
      // After 400ms of fade out, fade in new values
      timeoutId = setTimeout(() => {
        setEnergyAnimationState("fadeIn");
      }, 400);
    }
    // Note: fadeIn state persists until next phase (no further transitions needed)

    return () => clearTimeout(timeoutId);
  }, [energyAnimationState, animationStep, isAnimating]);

  // Auto-start animation when component mounts
  useEffect(() => {
    setCurrentPlayerIndex(0);
    setAnimationStep("energyConversion");
    setIsAnimating(true);
    setResourceAnimationState("initial");
    setEnergyAnimationState("initial");
  }, []);

  // Enhanced resource animation sequence
  useEffect(() => {
    if (!isAnimating || animationStep !== "production") return;

    let timeoutId: NodeJS.Timeout;

    if (resourceAnimationState === "initial") {
      // Wait 500ms, then fade in other resources
      timeoutId = setTimeout(() => {
        setResourceAnimationState("fadeInResources");
      }, 500);
    } else if (resourceAnimationState === "fadeInResources") {
      // Wait 500ms for fade in to complete, then show production indicators
      timeoutId = setTimeout(() => {
        setResourceAnimationState("showProduction");
      }, 500);
    } else if (resourceAnimationState === "showProduction") {
      // Wait 1500ms showing production indicators, then fade out old values
      timeoutId = setTimeout(() => {
        setResourceAnimationState("fadeOut");
      }, 1500);
    } else if (resourceAnimationState === "fadeOut") {
      // After 400ms of fade out, fade in new values
      timeoutId = setTimeout(() => {
        setResourceAnimationState("fadeIn");
      }, 400);
    }

    return () => clearTimeout(timeoutId);
  }, [resourceAnimationState, animationStep, isAnimating]);

  // Reset resource animation state when changing players or phases
  useEffect(() => {
    if (animationStep === "production") {
      setResourceAnimationState("initial");
    }
  }, [currentPlayerIndex, animationStep]);

  // Handle ESC key to close modal
  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        void handleClose();
      }
    };

    if (isOpen) {
      document.addEventListener("keydown", handleKeyDown);
      return () => document.removeEventListener("keydown", handleKeyDown);
    }

    return () => {};
  }, [isOpen, handleClose]);

  // Normal visibility check
  if (!isOpen) return null;

  if (!modalProductionData) return null;

  const currentPlayerData = modalProductionData.playersData[currentPlayerIndex];
  if (!currentPlayerData) return null;

  const renderResourceAnimation = (
    resourceType: ResourceType,
    beforeAmount: number,
    afterAmount: number,
  ) => {
    let displayBeforeAmount = beforeAmount;
    let displayAfterAmount = afterAmount;
    let change = afterAmount - beforeAmount;
    let shouldAnimate;

    // Handle energy-to-heat conversion values
    if (animationStep === "energyConversion") {
      if (resourceType === "energy") {
        displayBeforeAmount = currentPlayerData.beforeResources.energy;
        displayAfterAmount = 0; // Energy goes to 0
        change = -currentPlayerData.beforeResources.energy;
        shouldAnimate = true;
      } else if (resourceType === "heat") {
        displayBeforeAmount = currentPlayerData.beforeResources.heat;
        displayAfterAmount =
          currentPlayerData.beforeResources.heat +
          currentPlayerData.energyConverted;
        change = currentPlayerData.energyConverted;
        shouldAnimate = true;
      } else {
        // Other resources don't change during energy conversion
        shouldAnimate = false;
      }
    } else {
      // During production phase, energy starts from 0 and heat starts from post-conversion value
      if (resourceType === "energy") {
        displayBeforeAmount = 0;
        displayAfterAmount = currentPlayerData.production.energy;
        change = currentPlayerData.production.energy;
        // Energy should animate if it has production and after showing production indicators
        shouldAnimate =
          change !== 0 &&
          resourceAnimationState !== "initial" &&
          resourceAnimationState !== "fadeInResources" &&
          resourceAnimationState !== "showProduction";
      } else if (resourceType === "heat") {
        displayBeforeAmount =
          currentPlayerData.beforeResources.heat +
          currentPlayerData.energyConverted;
        displayAfterAmount = currentPlayerData.afterResources.heat;
        change = displayAfterAmount - displayBeforeAmount;
        // Heat should animate if it has additional production and after showing production indicators
        shouldAnimate =
          change !== 0 &&
          resourceAnimationState !== "initial" &&
          resourceAnimationState !== "fadeInResources" &&
          resourceAnimationState !== "showProduction";
      } else {
        // Other resources animate normally during production phase after showing production indicators
        shouldAnimate =
          change !== 0 &&
          resourceAnimationState !== "initial" &&
          resourceAnimationState !== "fadeInResources" &&
          resourceAnimationState !== "showProduction";
      }
    }

    const getAnimationStateClass = () => {
      switch (resourceAnimationState) {
        case "initial":
          return styles.initialState;
        case "fadeOut":
          return styles.fadeOutState;
        case "fadeIn":
          return styles.fadeInState;
        default:
          return "";
      }
    };

    const renderAmounts = () => {
      // During production phase initial and fadeInResources state, show clean values
      if (
        animationStep === "production" &&
        (resourceAnimationState === "initial" ||
          resourceAnimationState === "fadeInResources")
      ) {
        return (
          <div className={styles.enhancedResourceAmounts}>
            <div
              className={`${styles.beforeAmount} ${displayBeforeAmount < 0 ? styles.negative : ""}`}
            >
              {displayBeforeAmount}
            </div>
          </div>
        );
      }

      // During showProduction state, show change indicators for all resources with changes
      if (
        animationStep === "production" &&
        resourceAnimationState === "showProduction" &&
        change !== 0
      ) {
        return (
          <div className={styles.enhancedResourceAmounts}>
            <div
              className={`${styles.beforeAmount} ${displayBeforeAmount < 0 ? styles.negative : ""}`}
            >
              {displayBeforeAmount}
            </div>
            <div
              className={`${styles.changeIndicator} ${change < 0 ? styles.negativeChange : ""}`}
            >
              {change > 0 ? `+${change}` : change}
            </div>
          </div>
        );
      }

      // Handle energy conversion animation for energy and heat
      if (animationStep === "energyConversion" && shouldAnimate) {
        if (energyAnimationState === "fadeIn") {
          return (
            <div
              className={`${styles.enhancedResourceAmounts} ${styles.fadeInState}`}
            >
              <div
                className={`${styles.finalValue} ${displayAfterAmount < 0 ? styles.negative : ""}`}
              >
                {displayAfterAmount}
              </div>
            </div>
          );
        } else if (energyAnimationState === "fadeOut") {
          return (
            <div
              className={`${styles.enhancedResourceAmounts} ${styles.fadeOutState}`}
            >
              <div
                className={`${styles.beforeAmount} ${displayBeforeAmount < 0 ? styles.negative : ""}`}
              >
                {displayBeforeAmount}
              </div>
              <div
                className={`${styles.changeIndicator} ${change < 0 ? styles.negativeChange : ""}`}
              >
                {change > 0 ? `+${change}` : change}
              </div>
            </div>
          );
        } else {
          // Initial state for energy conversion
          return (
            <div className={styles.enhancedResourceAmounts}>
              <div
                className={`${styles.beforeAmount} ${displayBeforeAmount < 0 ? styles.negative : ""}`}
              >
                {displayBeforeAmount}
              </div>
              <div
                className={`${styles.changeIndicator} ${change < 0 ? styles.negativeChange : ""}`}
              >
                {change > 0 ? `+${change}` : change}
              </div>
            </div>
          );
        }
      }

      // For resources that don't animate, just show the current value
      if (!shouldAnimate) {
        return (
          <div className={styles.enhancedResourceAmounts}>
            <div
              className={`${styles.beforeAmount} ${displayBeforeAmount < 0 ? styles.negative : ""}`}
            >
              {displayBeforeAmount}
            </div>
          </div>
        );
      }

      if (resourceAnimationState === "fadeIn") {
        return (
          <div
            className={`${styles.enhancedResourceAmounts} ${getAnimationStateClass()}`}
          >
            <div
              className={`${styles.finalValue} ${displayAfterAmount < 0 ? styles.negative : ""}`}
            >
              {displayAfterAmount}
            </div>
          </div>
        );
      }

      return (
        <div
          className={`${styles.enhancedResourceAmounts} ${getAnimationStateClass()}`}
        >
          <div className={styles.beforeAmount}>{displayBeforeAmount}</div>
          <div
            className={`${styles.changeIndicator} ${change < 0 ? styles.negativeChange : ""}`}
          >
            {change > 0 ? `+${change}` : change}
          </div>
        </div>
      );
    };

    return (
      <div
        key={resourceType}
        className={`${styles.resourceItem} ${shouldAnimate ? styles.active : ""} ${
          (animationStep === "energyConversion" && !shouldAnimate) ||
          (animationStep === "production" &&
            resourceAnimationState === "initial" &&
            !shouldAnimate &&
            resourceType !== "energy" &&
            resourceType !== "heat")
            ? styles.dimmed
            : ""
        } ${
          animationStep === "energyConversion" &&
          resourceType === "energy" &&
          energyAnimationState === "initial"
            ? styles.overheating
            : ""
        }`}
        style={
          {
            "--resource-color": RESOURCE_COLORS[resourceType],
          } as React.CSSProperties
        }
      >
        <div className={styles.resourceIcon}>
          <img
            src={resourceIcons[resourceType]}
            alt={resourceNames[resourceType]}
          />
        </div>
        {renderAmounts()}
        <div className={styles.resourceLabel}>
          {resourceNames[resourceType]}
        </div>
      </div>
    );
  };

  const renderResourceDelta = (
    resourceType: ResourceType,
    deltaAmount: number,
  ) => {
    if (deltaAmount === 0) return null;

    return (
      <div
        key={resourceType}
        className={styles.summaryResourceItem}
        style={
          {
            "--resource-color": RESOURCE_COLORS[resourceType],
          } as React.CSSProperties
        }
      >
        <div className={styles.resourceIcon}>
          <img
            src={RESOURCE_ICONS[resourceType]}
            alt={RESOURCE_NAMES[resourceType]}
          />
        </div>
        <div className={styles.deltaAmount}>
          <span className={deltaAmount > 0 ? styles.positive : styles.negative}>
            {deltaAmount > 0 ? `+${deltaAmount}` : deltaAmount}
          </span>
        </div>
        <div className={styles.resourceLabel}>
          {RESOURCE_NAMES[resourceType]}
        </div>
      </div>
    );
  };

  const renderSummaryView = () => {
    return (
      <div className={styles.summaryView}>
        <div className={styles.summaryTitle}>
          Resource Changes Summary - Generation {modalProductionData.generation}
        </div>
        <div className={styles.playersGrid}>
          {modalProductionData.playersData.map((player) => (
            <div key={player.playerId} className={styles.playerSummary}>
              <div className={styles.playerName}>{player.playerName}</div>
              <div className={styles.resourceDeltas}>
                {renderResourceDelta(
                  "credits" as ResourceType,
                  player.resourceDelta.credits,
                )}
                {renderResourceDelta(
                  "steel" as ResourceType,
                  player.resourceDelta.steel,
                )}
                {renderResourceDelta(
                  "titanium" as ResourceType,
                  player.resourceDelta.titanium,
                )}
                {renderResourceDelta(
                  "plants" as ResourceType,
                  player.resourceDelta.plants,
                )}
                {renderResourceDelta(
                  "energy" as ResourceType,
                  player.resourceDelta.energy,
                )}
                {renderResourceDelta(
                  "heat" as ResourceType,
                  player.resourceDelta.heat,
                )}
              </div>
              <div className={styles.creditsBreakdown}>
                Credits: {player.terraformRating} (TR) +{" "}
                {player.production.credits} (Production) ={" "}
                {player.creditsIncome}
              </div>
            </div>
          ))}
        </div>
      </div>
    );
  };

  const renderProductionPhase = () => {
    return (
      <div className={styles.productionPhase}>
        <div className={styles.productionTitle}>
          {animationStep === "energyConversion"
            ? "Energy → Heat Conversion"
            : "Resource Production"}
        </div>
        <div className={styles.resourcesGrid}>
          {renderResourceAnimation(
            "credits" as ResourceType,
            currentPlayerData.beforeResources.credits,
            currentPlayerData.afterResources.credits,
          )}
          {renderResourceAnimation(
            "steel" as ResourceType,
            currentPlayerData.beforeResources.steel,
            currentPlayerData.afterResources.steel,
          )}
          {renderResourceAnimation(
            "titanium" as ResourceType,
            currentPlayerData.beforeResources.titanium,
            currentPlayerData.afterResources.titanium,
          )}
          {renderResourceAnimation(
            "plants" as ResourceType,
            currentPlayerData.beforeResources.plants,
            currentPlayerData.afterResources.plants,
          )}
          {renderResourceAnimation(
            "energy" as ResourceType,
            currentPlayerData.beforeResources.energy,
            currentPlayerData.afterResources.energy,
          )}
          {renderResourceAnimation(
            "heat" as ResourceType,
            currentPlayerData.beforeResources.heat,
            currentPlayerData.afterResources.heat,
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
    <div className={styles.modalOverlay} onClick={handleClose}>
      <div className={styles.modalPopup} onClick={(e) => e.stopPropagation()}>
        <div className={styles.modalHeader}>
          <h2>Production Phase</h2>
          <div className={styles.generationInfo}>
            Generation {modalProductionData.generation}
          </div>
          <button className={styles.closeBtn} onClick={handleClose}>
            ×
          </button>
        </div>

        <div className={styles.playerProgress}>
          {modalProductionData.playersData.map((player, index) => (
            <button
              key={player.playerId}
              className={`${styles.playerTab} ${
                index === currentPlayerIndex && !showSummary
                  ? styles.current
                  : ""
              }`}
              onClick={() => handlePlayerSelect(index)}
            >
              {player.playerName}
            </button>
          ))}
          <button
            className={`${styles.playerTab} ${styles.summaryTab} ${
              showSummary ? styles.current : ""
            }`}
            onClick={() => setShowSummary(true)}
          >
            Summary
          </button>
        </div>

        <div className={styles.currentPlayerSection}>
          {showSummary ? (
            <div className={styles.animationContent}>{renderSummaryView()}</div>
          ) : (
            <>
              <div className={styles.playerHeader}>
                <div className={styles.playerName}>
                  {currentPlayerData.playerName}
                </div>
                <div className={styles.animationStep}>
                  {animationStep === "energyConversion"
                    ? "Energy → Heat Conversion"
                    : "Resource Production"}
                </div>
              </div>

              <div className={styles.animationContent}>
                {renderProductionPhase()}
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
};

export default ProductionPhaseModal;
