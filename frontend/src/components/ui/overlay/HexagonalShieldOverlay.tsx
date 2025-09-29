import React, { useState, useEffect } from "react";
import { CardDto } from "../../../types/generated/api-types.ts";
import { UnplayableReason } from "../../../utils/cardPlayabilityUtils.ts";

interface HexagonalShieldOverlayProps {
  card: CardDto | null;
  reason: UnplayableReason | null;
  isVisible: boolean;
}

const HexagonalShieldOverlay: React.FC<HexagonalShieldOverlayProps> = ({
  card,
  reason,
  isVisible,
}) => {
  const [shouldRender, setShouldRender] = useState(false);
  const [isAnimatingOut, setIsAnimatingOut] = useState(false);
  const [isAnimatingIn, setIsAnimatingIn] = useState(false);
  const [lastValidCard, setLastValidCard] = useState<CardDto | null>(null);
  const [lastValidReason, setLastValidReason] =
    useState<UnplayableReason | null>(null);

  useEffect(() => {
    if (isVisible && card && reason) {
      // Mount first
      setShouldRender(true);
      setLastValidCard(card);
      setLastValidReason(reason);

      // Next tick → trigger fade-in
      requestAnimationFrame(() => {
        setIsAnimatingOut(false);
        setIsAnimatingIn(true);
      });
      return undefined;
    } else if (shouldRender) {
      // Start fade-out
      setIsAnimatingOut(true);
      setIsAnimatingIn(false);

      const timer = setTimeout(() => {
        setShouldRender(false);
        setIsAnimatingOut(false);
        setLastValidCard(null);
        setLastValidReason(null);
      }, 300);
      return () => clearTimeout(timer);
    }
    return undefined;
  }, [isVisible, card, reason, shouldRender]);

  // Don't render if not supposed to be visible and not animating
  if (!shouldRender) {
    return null;
  }

  const overlayClass = isAnimatingOut
    ? "hidden"
    : isAnimatingIn
      ? "visible"
      : "hidden";
  // Use last valid data during fade-out animation
  const displayCard = lastValidCard;
  const displayReason = lastValidReason;

  // Safety check - if we don't have valid data, don't render
  if (!displayCard || !displayReason) {
    return null;
  }

  // Get the appropriate icon for the requirement type
  const getRequirementIcon = (reason: UnplayableReason): string | null => {
    const iconMap: { [key: string]: string } = {
      // Cost
      cost: "/assets/resources/megacredit.png",
      credits: "/assets/resources/megacredit.png",
      megacredits: "/assets/resources/megacredit.png",

      // Global parameters
      temperature: "/assets/global-parameters/temperature.png",
      oxygen: "/assets/global-parameters/oxygen.png",
      oceans: "/assets/tiles/ocean.png",

      // Resources
      steel: "/assets/resources/steel.png",
      titanium: "/assets/resources/titanium.png",
      plants: "/assets/resources/plant.png",
      energy: "/assets/resources/power.png",
      heat: "/assets/resources/heat.png",

      // Tags
      science: "/assets/tags/science.png",
      "power-tag": "/assets/tags/power.png",
      space: "/assets/tags/space.png",
      building: "/assets/tags/building.png",
      city: "/assets/tags/city.png",
      earth: "/assets/tags/earth.png",
      jovian: "/assets/tags/jovian.png",
      venus: "/assets/tags/venus.png",
      microbe: "/assets/tags/microbe.png",
      animal: "/assets/tags/animal.png",
      plant: "/assets/tags/plant.png",
      event: "/assets/tags/event.png",
      mars: "/assets/tags/mars.png",
      moon: "/assets/tags/moon.png",
    };

    // For cost type, return megacredit icon
    if (reason.type === "cost") {
      return iconMap.cost;
    }

    // For global-param type, use the specific parameter type
    if (reason.type === "global-param" && reason.requirement?.type) {
      return iconMap[reason.requirement.type];
    }

    // For tag type, use the tag name
    if (reason.type === "tag" && reason.requirement?.tag) {
      const tagName = reason.requirement.tag.toLowerCase();
      // Handle special case for power tag to avoid duplicate key conflict
      if (tagName === "power") {
        return iconMap["power-tag"];
      }
      return iconMap[tagName];
    }

    // For production type, use the resource icon
    if (reason.type === "production" && reason.requirement?.resource) {
      const resourceName = reason.requirement.resource.toLowerCase();
      // For production, power refers to energy resource, not power tag
      if (resourceName === "power") {
        return iconMap["energy"];
      }
      // For production, plant refers to plants resource
      if (resourceName === "plant") {
        return iconMap["plants"];
      }
      return iconMap[resourceName];
    }

    return null;
  };

  const icon = getRequirementIcon(displayReason);

  // Determine which global parameters are affected
  const getAffectedParameters = (reason: UnplayableReason): string[] => {
    if (reason.type === "multiple" && reason.failedRequirements) {
      // Get all global parameters from failed requirements
      return reason.failedRequirements
        .filter((req) => req.type === "global-param" && req.requirement?.type)
        .map((req) => req.requirement.type);
    } else if (reason.type === "global-param" && reason.requirement?.type) {
      return [reason.requirement.type];
    }
    return [];
  };

  const affectedParameters = displayReason
    ? getAffectedParameters(displayReason)
    : [];

  // Create CSS class string for highlighting multiple parameters
  const createHighlightClasses = () => {
    if (affectedParameters.length === 0) return "";
    if (affectedParameters.length === 1)
      return `highlight-${affectedParameters[0]}`;

    // Multiple parameters - create multiple classes
    return affectedParameters.map((param) => `highlight-${param}`).join(" ");
  };

  // Hex grid configuration
  const maxCols = 7; // Reduced columns for fewer, bigger hexagons
  const totalRows = 5; // Reduced rows for fewer, bigger hexagons
  const hexSize = 192; // 160 * 1.2 = 192 for 20% larger hexagons
  const hexWidth = hexSize * Math.sqrt(3);
  const hexHeight = hexSize * 2;
  const vertSpacing = hexHeight * 0.75;

  // Calculate padding to ensure no hexagons are clipped
  const hexRadius = hexSize;
  const padding = hexRadius;

  // Calculate SVG viewBox dimensions with padding for complete hexagons
  const gridWidth = (maxCols - 1) * hexWidth + hexWidth; // Full width of grid
  const gridHeight = (totalRows - 1) * vertSpacing + hexHeight; // Full height of grid
  const svgWidth = gridWidth + padding * 2;
  const svgHeight = gridHeight + padding * 2;

  // Generate hexagon pattern for the shield in a circular shape
  const generateHexagonPattern = () => {
    const hexagons = [];
    const centerX = svgWidth / 2;
    const centerY = svgHeight / 2;

    // Define how many hexagons per row to create a circular shape
    // Pattern: fewer hexagons but much bigger - asymmetric design
    const hexagonsPerRow = [4, 5, 6, 5, 4]; // 5 rows with fewer, bigger hexagons
    const actualRows = hexagonsPerRow.length;

    for (let row = 0; row < actualRows; row++) {
      const colsInThisRow = hexagonsPerRow[row];

      for (let col = 0; col < colsInThisRow; col++) {
        // All rows offset by half hex width, then additional adjustments
        let offsetX = hexWidth / 2; // Base offset for all rows
        if (row === 1 || row === 3) {
          offsetX += hexWidth / 2; // Additional half hex width for rows 1 and 3
        }
        if (row === 0 || row === 2 || row === 4) {
          offsetX += hexWidth / 2; // Additional half hex width for rows 0, 2, and 4
        }

        // Center each row based on how many hexagons it has
        const maxRowWidth = Math.max(...hexagonsPerRow); // Find the widest row
        const rowStartX = ((maxRowWidth - colsInThisRow) * hexWidth) / 2; // Center relative to widest row

        // Add padding offset to ensure hexagons are not clipped
        const x = padding + rowStartX + col * hexWidth + offsetX;
        const y = padding + row * vertSpacing + hexSize; // Add hexSize to account for hex height

        // Calculate distance from center of SVG for circular fade effect
        const distanceFromCenter = Math.sqrt(
          Math.pow(x - centerX, 2) + Math.pow(y - centerY, 2),
        );

        // Create circular fade with maximum radius
        const maxRadius = Math.min(svgWidth, svgHeight) / 2;
        const normalizedDistance = distanceFromCenter / maxRadius;

        // Create circular gradient opacity (more opaque in center, transparent at edges)
        const opacity = Math.max(0, 1 - normalizedDistance * 0.8);

        hexagons.push(
          <polygon
            key={`hex-${row}-${col}`}
            points={generateHexagonPoints(x, y, hexSize)}
            fill="rgba(80, 40, 0, 0.3)"
            stroke="rgba(255, 152, 0, 0.8)"
            strokeWidth="2"
            opacity={opacity}
            className="hex-tile"
          />,
        );
      }
    }
    return hexagons;
  };

  // Generate hexagon points for SVG polygon
  const generateHexagonPoints = (
    centerX: number,
    centerY: number,
    size: number,
  ) => {
    const points = [];
    for (let i = 0; i < 6; i++) {
      const angle = (i * 60 - 90) * (Math.PI / 180); // Start from top, rotate 60° each step
      const x = centerX + size * Math.cos(angle);
      const y = centerY + size * Math.sin(angle);
      points.push(`${x},${y}`);
    }
    return points.join(" ");
  };

  return (
    <div
      className={`hexagonal-shield-overlay ${createHighlightClasses()} ${overlayClass}`}
    >
      {/* Hexagonal shield background */}
      <div className="shield-container">
        <svg
          className="hexagon-pattern"
          viewBox={`0 0 ${svgWidth} ${svgHeight}`}
        >
          {generateHexagonPattern()}
        </svg>

        {/* Shield text overlay - directly on hexagons */}
        <div className="shield-text-overlay">
          <div className="shield-text-content">
            {displayReason.type === "multiple" &&
            displayReason.failedRequirements ? (
              <div className="requirements-list">
                {displayReason.failedRequirements.map((req, index) => {
                  const reqIcon = getRequirementIcon(req);
                  return (
                    <div key={index} className="requirement-item">
                      {reqIcon && (
                        <img
                          src={reqIcon}
                          alt="requirement"
                          className="requirement-icon"
                        />
                      )}
                      <span className="requirement-text">{req.message}</span>
                    </div>
                  );
                })}
              </div>
            ) : (
              <div className="single-requirement">
                {icon && (
                  <img
                    src={icon}
                    alt="requirement"
                    className="requirement-icon"
                  />
                )}
                <span className="requirement-text">
                  {displayReason.message}
                </span>
              </div>
            )}
          </div>
        </div>
      </div>

      <style>{`
        .hexagonal-shield-overlay {
          position: fixed;
          top: calc(50% + 10px);
          left: 50%;
          transform: translate(-50%, -50%);
          width: 108vw;
          height: 96vh;
          z-index: 1500; /* Above Mars but below card hand */
          pointer-events: none;
          display: flex;
          justify-content: center;
          align-items: center;
          opacity: 0;
          transition: opacity 0.3s ease-in-out;
        }

        .hexagonal-shield-overlay.visible {
          opacity: 1;
        }

        .hexagonal-shield-overlay.hidden {
          opacity: 0;
        }

        .shield-container {
          position: relative;
          width: 100%;
          height: 100%;
          display: flex;
          justify-content: center;
          align-items: center;
          max-width: 800px;
          max-height: 600px;
        }

        .hexagon-pattern {
          position: absolute;
          top: 0;
          left: 0;
          width: 100%;
          height: 100%;
          opacity: 0.9;
          filter: blur(0.5px) drop-shadow(0 0 12px rgba(255, 152, 0, 0.8));
        }

        .hex-tile {
          transition: opacity 0.5s ease-in-out;
          filter: drop-shadow(0 0 8px rgba(255, 152, 0, 0.9));
        }

        .visible .hex-tile {
          opacity: 1;
        }

        .hidden .hex-tile {
          opacity: 0;
        }

        .shield-text-overlay {
          position: absolute;
          top: 0;
          left: 0;
          width: 100%;
          height: 100%;
          display: flex;
          align-items: center;
          justify-content: center;
          z-index: 3;
          pointer-events: none;
        }

        .shield-text-content {
          text-align: center;
          max-width: 80%;
          transition: opacity 0.3s ease-out, transform 0.3s ease-out;
        }

        .visible .shield-text-content {
          opacity: 1;
          transform: scale(1);
        }

        .hidden .shield-text-content {
          opacity: 0;
          transform: scale(0.9);
        }

        .requirements-list {
          display: flex;
          flex-direction: column;
          gap: 16px;
          align-items: center;
        }

        .requirement-item,
        .single-requirement {
          display: flex;
          align-items: center;
          gap: 12px;
          background: rgba(0, 0, 0, 0.7);
          padding: 12px 20px;
          border-radius: 12px;
          border: 2px solid rgba(255, 152, 0, 0.6);
          backdrop-filter: blur(8px);
          box-shadow: 0 0 24px rgba(255, 152, 0, 0.6);
        }

        .requirement-icon {
          width: 32px;
          height: 32px;
          border-radius: 6px;
          box-shadow: 0 2px 8px rgba(0, 0, 0, 0.6);
          object-fit: contain;
          flex-shrink: 0;
        }

        .requirement-text {
          color: #ffffff;
          font-size: 18px;
          font-weight: 500;
          text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.9);
          line-height: 1.3;
        }

        /* Responsive design */
        @media (max-width: 1200px) {
          .hexagonal-shield-overlay {
            left: 5%;
            right: 5%;
          }

          .shield-content {
            max-width: 450px;
            padding: 20px 28px;
          }

          .card-name {
            font-size: 18px;
          }
        }

        @media (max-width: 768px) {
          .hexagonal-shield-overlay {
            top: 80px;
            left: 5%;
            right: 5%;
            bottom: 250px;
          }

          .shield-content {
            max-width: 90vw;
            padding: 18px 24px;
          }

          .card-name {
            font-size: 16px;
          }

          .cannot-play {
            font-size: 12px;
          }

          .requirement-message {
            font-size: 14px;
          }

          .requirement-icon-container {
            width: 32px;
            height: 32px;
          }

          .requirement-icon {
            max-width: 32px;
            max-height: 32px;
          }

          .production-background {
            width: 32px;
            height: 32px;
          }
        }

        @media (max-width: 480px) {
          .hexagonal-shield-overlay {
            top: 60px;
            bottom: 200px;
          }

          .shield-content {
            padding: 16px 20px;
          }

          .requirement-info {
            flex-direction: column;
            gap: 8px;
          }

          .requirement-icon-container {
            width: 28px;
            height: 28px;
          }

          .requirement-icon {
            max-width: 28px;
            max-height: 28px;
          }

          .production-background {
            width: 28px;
            height: 28px;
          }
        }

        /* Temperature requirement not met - highlight temperature meter, dim others */
        body:has(.hexagonal-shield-overlay.highlight-temperature) .oxygenMeter,
        body:has(.hexagonal-shield-overlay.highlight-temperature) .oceanCounter,
        body:has(.hexagonal-shield-overlay.highlight-temperature) .currentOxygen,
        body:has(.hexagonal-shield-overlay.highlight-temperature) .currentOceans {
          filter: brightness(0.5) saturate(0.4) grayscale(0.3) !important;
          transition: all 0.3s ease !important;
        }

        body:has(.hexagonal-shield-overlay.highlight-temperature) .temperatureMeter,
        body:has(.hexagonal-shield-overlay.highlight-temperature) .currentTemp {
          filter: brightness(1.4) saturate(1.5) contrast(1.2) !important;
          box-shadow: 0 0 20px rgba(255, 100, 50, 0.8) !important;
          border: 3px solid rgba(255, 100, 50, 0.9) !important;
          border-radius: 8px !important;
          transform: scale(1.05) !important;
          transition: all 0.3s ease !important;
          position: relative;
          z-index: 10;
        }

        /* Oxygen requirement not met - highlight oxygen meter, dim others */
        body:has(.hexagonal-shield-overlay.highlight-oxygen) .temperatureMeter,
        body:has(.hexagonal-shield-overlay.highlight-oxygen) .oceanCounter,
        body:has(.hexagonal-shield-overlay.highlight-oxygen) .currentTemp,
        body:has(.hexagonal-shield-overlay.highlight-oxygen) .currentOceans {
          filter: brightness(0.5) saturate(0.4) grayscale(0.3) !important;
          transition: all 0.3s ease !important;
        }

        body:has(.hexagonal-shield-overlay.highlight-oxygen) .oxygenMeter,
        body:has(.hexagonal-shield-overlay.highlight-oxygen) .currentOxygen {
          filter: brightness(1.4) saturate(1.5) contrast(1.2) !important;
          box-shadow: 0 0 20px rgba(100, 200, 255, 0.8) !important;
          border: 3px solid rgba(100, 200, 255, 0.9) !important;
          border-radius: 8px !important;
          transform: scale(1.05) !important;
          transition: all 0.3s ease !important;
          position: relative;
          z-index: 10;
        }

        /* Oceans requirement not met - highlight oceans meter, dim others */
        body:has(.hexagonal-shield-overlay.highlight-oceans) .temperatureMeter,
        body:has(.hexagonal-shield-overlay.highlight-oceans) .oxygenMeter,
        body:has(.hexagonal-shield-overlay.highlight-oceans) .currentTemp,
        body:has(.hexagonal-shield-overlay.highlight-oceans) .currentOxygen {
          filter: brightness(0.5) saturate(0.4) grayscale(0.3) !important;
          transition: all 0.3s ease !important;
        }

        body:has(.hexagonal-shield-overlay.highlight-oceans) .oceanCounter,
        body:has(.hexagonal-shield-overlay.highlight-oceans) .currentOceans {
          filter: brightness(1.4) saturate(1.5) contrast(1.2) !important;
          box-shadow: 0 0 20px rgba(100, 180, 255, 0.8) !important;
          border: 3px solid rgba(100, 180, 255, 0.9) !important;
          border-radius: 8px !important;
          transform: scale(1.05) !important;
          transition: all 0.3s ease !important;
          position: relative;
          z-index: 10;
        }

        /* Multiple parameter highlighting support */
        /* When multiple parameters are affected, highlight all of them without dimming others */
        body:has(.hexagonal-shield-overlay.highlight-temperature.highlight-oxygen) .temperatureMeter,
        body:has(.hexagonal-shield-overlay.highlight-temperature.highlight-oxygen) .currentTemp,
        body:has(.hexagonal-shield-overlay.highlight-temperature.highlight-oxygen) .oxygenMeter,
        body:has(.hexagonal-shield-overlay.highlight-temperature.highlight-oxygen) .currentOxygen {
          filter: brightness(1.4) saturate(1.5) contrast(1.2) !important;
          border-radius: 8px !important;
          transform: scale(1.05) !important;
          transition: all 0.3s ease !important;
          position: relative;
          z-index: 10;
        }

        body:has(.hexagonal-shield-overlay.highlight-temperature.highlight-oxygen) .temperatureMeter,
        body:has(.hexagonal-shield-overlay.highlight-temperature.highlight-oxygen) .currentTemp {
          box-shadow: 0 0 20px rgba(255, 100, 50, 0.8) !important;
          border: 3px solid rgba(255, 100, 50, 0.9) !important;
        }

        body:has(.hexagonal-shield-overlay.highlight-temperature.highlight-oxygen) .oxygenMeter,
        body:has(.hexagonal-shield-overlay.highlight-temperature.highlight-oxygen) .currentOxygen {
          box-shadow: 0 0 20px rgba(100, 200, 255, 0.8) !important;
          border: 3px solid rgba(100, 200, 255, 0.9) !important;
        }

        /* Temperature + Oceans combination */
        body:has(.hexagonal-shield-overlay.highlight-temperature.highlight-oceans) .temperatureMeter,
        body:has(.hexagonal-shield-overlay.highlight-temperature.highlight-oceans) .currentTemp,
        body:has(.hexagonal-shield-overlay.highlight-temperature.highlight-oceans) .oceanCounter,
        body:has(.hexagonal-shield-overlay.highlight-temperature.highlight-oceans) .currentOceans {
          filter: brightness(1.4) saturate(1.5) contrast(1.2) !important;
          border-radius: 8px !important;
          transform: scale(1.05) !important;
          transition: all 0.3s ease !important;
          position: relative;
          z-index: 10;
        }

        body:has(.hexagonal-shield-overlay.highlight-temperature.highlight-oceans) .temperatureMeter,
        body:has(.hexagonal-shield-overlay.highlight-temperature.highlight-oceans) .currentTemp {
          box-shadow: 0 0 20px rgba(255, 100, 50, 0.8) !important;
          border: 3px solid rgba(255, 100, 50, 0.9) !important;
        }

        body:has(.hexagonal-shield-overlay.highlight-temperature.highlight-oceans) .oceanCounter,
        body:has(.hexagonal-shield-overlay.highlight-temperature.highlight-oceans) .currentOceans {
          box-shadow: 0 0 20px rgba(100, 180, 255, 0.8) !important;
          border: 3px solid rgba(100, 180, 255, 0.9) !important;
        }

        /* Oxygen + Oceans combination */
        body:has(.hexagonal-shield-overlay.highlight-oxygen.highlight-oceans) .oxygenMeter,
        body:has(.hexagonal-shield-overlay.highlight-oxygen.highlight-oceans) .currentOxygen,
        body:has(.hexagonal-shield-overlay.highlight-oxygen.highlight-oceans) .oceanCounter,
        body:has(.hexagonal-shield-overlay.highlight-oxygen.highlight-oceans) .currentOceans {
          filter: brightness(1.4) saturate(1.5) contrast(1.2) !important;
          border-radius: 8px !important;
          transform: scale(1.05) !important;
          transition: all 0.3s ease !important;
          position: relative;
          z-index: 10;
        }

        body:has(.hexagonal-shield-overlay.highlight-oxygen.highlight-oceans) .oxygenMeter,
        body:has(.hexagonal-shield-overlay.highlight-oxygen.highlight-oceans) .currentOxygen {
          box-shadow: 0 0 20px rgba(100, 200, 255, 0.8) !important;
          border: 3px solid rgba(100, 200, 255, 0.9) !important;
        }

        body:has(.hexagonal-shield-overlay.highlight-oxygen.highlight-oceans) .oceanCounter,
        body:has(.hexagonal-shield-overlay.highlight-oxygen.highlight-oceans) .currentOceans {
          box-shadow: 0 0 20px rgba(100, 180, 255, 0.8) !important;
          border: 3px solid rgba(100, 180, 255, 0.9) !important;
        }

        /* All three parameters combination */
        body:has(.hexagonal-shield-overlay.highlight-temperature.highlight-oxygen.highlight-oceans) .temperatureMeter,
        body:has(.hexagonal-shield-overlay.highlight-temperature.highlight-oxygen.highlight-oceans) .currentTemp,
        body:has(.hexagonal-shield-overlay.highlight-temperature.highlight-oxygen.highlight-oceans) .oxygenMeter,
        body:has(.hexagonal-shield-overlay.highlight-temperature.highlight-oxygen.highlight-oceans) .currentOxygen,
        body:has(.hexagonal-shield-overlay.highlight-temperature.highlight-oxygen.highlight-oceans) .oceanCounter,
        body:has(.hexagonal-shield-overlay.highlight-temperature.highlight-oxygen.highlight-oceans) .currentOceans {
          filter: brightness(1.4) saturate(1.5) contrast(1.2) !important;
          border-radius: 8px !important;
          transform: scale(1.05) !important;
          transition: all 0.3s ease !important;
          position: relative;
          z-index: 10;
        }

        body:has(.hexagonal-shield-overlay.highlight-temperature.highlight-oxygen.highlight-oceans) .temperatureMeter,
        body:has(.hexagonal-shield-overlay.highlight-temperature.highlight-oxygen.highlight-oceans) .currentTemp {
          box-shadow: 0 0 20px rgba(255, 100, 50, 0.8) !important;
          border: 3px solid rgba(255, 100, 50, 0.9) !important;
        }

        body:has(.hexagonal-shield-overlay.highlight-temperature.highlight-oxygen.highlight-oceans) .oxygenMeter,
        body:has(.hexagonal-shield-overlay.highlight-temperature.highlight-oxygen.highlight-oceans) .currentOxygen {
          box-shadow: 0 0 20px rgba(100, 200, 255, 0.8) !important;
          border: 3px solid rgba(100, 200, 255, 0.9) !important;
        }

        body:has(.hexagonal-shield-overlay.highlight-temperature.highlight-oxygen.highlight-oceans) .oceanCounter,
        body:has(.hexagonal-shield-overlay.highlight-temperature.highlight-oxygen.highlight-oceans) .currentOceans {
          box-shadow: 0 0 20px rgba(100, 180, 255, 0.8) !important;
          border: 3px solid rgba(100, 180, 255, 0.9) !important;
        }
      `}</style>
    </div>
  );
};

export default HexagonalShieldOverlay;
