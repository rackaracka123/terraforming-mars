import React, { useState, useEffect } from "react";
import {
  CardDto,
  ResourceTypeCredits,
  ResourceTypeCreditsProduction,
} from "@/types/generated/api-types.ts";
import { UnplayableReason } from "@/utils/cardPlayabilityUtils.ts";
import GameIcon from "../display/GameIcon.tsx";

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
  const [isPendingTileSelection, setIsPendingTileSelection] = useState(false);

  useEffect(() => {
    if (isVisible && card && reason) {
      // Check if this is a pending tile selection message
      const isTileSelection = reason.message === "Pending tile selection";

      // Mount first
      setShouldRender(true);
      setLastValidCard(card);
      setLastValidReason(reason);
      setIsPendingTileSelection(isTileSelection);

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
        setIsPendingTileSelection(false);
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

  // Extract the icon type from requirement data - GameIcon handles all the mapping
  const getRequirementIconType = (reason: UnplayableReason): string | null => {
    if (reason.type === "cost") return ResourceTypeCredits;
    if (reason.type === "global-param") return reason.requirement?.type || null;
    if (reason.type === "tag")
      return reason.requirement?.tag?.toLowerCase() || null;
    if (reason.type === "production" || reason.type === "resource") {
      return reason.requirement?.resource || null;
    }
    return null;
  };

  // Get the amount to display inside the icon (for credits)
  const getRequirementAmount = (
    reason: UnplayableReason,
  ): number | undefined => {
    if (reason.type === "cost" && typeof reason.requiredValue === "number") {
      return reason.requiredValue;
    }
    if (
      reason.requirement?.resource === ResourceTypeCreditsProduction &&
      typeof reason.requiredValue === "number"
    ) {
      return reason.requiredValue;
    }

    return undefined;
  };

  const iconType = getRequirementIconType(displayReason);
  const iconAmount = getRequirementAmount(displayReason);

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
            fill="rgba(40, 20, 5, 0.85)"
            stroke="rgba(255, 152, 0, 0.8)"
            strokeWidth="2"
            opacity={opacity}
            className="transition-opacity duration-500 [filter:drop-shadow(0_0_8px_rgba(255,152,0,0.9))]"
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
      className={`hexagonal-shield-overlay fixed top-[calc(50%+10px)] left-1/2 -translate-x-1/2 -translate-y-1/2 w-[108vw] h-[96vh] z-[1500] pointer-events-none flex justify-center items-center transition-opacity duration-300 ${overlayClass === "visible" ? "opacity-100" : "opacity-0"} ${createHighlightClasses()}`}
    >
      {/* Hexagonal shield background */}
      <div className="relative w-full h-full flex justify-center items-center max-w-[800px] max-h-[600px]">
        <svg
          className="absolute top-0 left-0 w-full h-full opacity-90 [filter:blur(0.5px)_drop-shadow(0_0_12px_rgba(255,152,0,0.8))]"
          viewBox={`0 0 ${svgWidth} ${svgHeight}`}
        >
          {generateHexagonPattern()}
        </svg>

        {/* Shield text overlay - directly on hexagons */}
        <div className="absolute top-0 left-0 w-full h-full flex items-center justify-center z-[3] pointer-events-none">
          <div
            className={`text-center max-w-[80%] transition-all duration-300 ${overlayClass === "visible" ? "opacity-100 scale-100" : "opacity-0 scale-90"}`}
          >
            {/* If pending tile selection, ONLY show that message - ignore all other requirements */}
            {isPendingTileSelection ? (
              <div className="flex items-center gap-3 bg-black/70 py-3 px-5 rounded-xl border-2 border-[rgba(255,152,0,0.6)] backdrop-blur-[8px] shadow-[0_0_24px_rgba(255,152,0,0.6)]">
                <span className="text-white text-lg font-medium [text-shadow:2px_2px_4px_rgba(0,0,0,0.9)] leading-[1.3]">
                  {displayReason.message}
                </span>
              </div>
            ) : displayReason.type === "multiple" &&
              displayReason.failedRequirements ? (
              <div className="flex flex-col gap-4 items-center">
                {displayReason.failedRequirements.map((req, index) => {
                  const reqIconType = getRequirementIconType(req);
                  const reqIconAmount = getRequirementAmount(req);
                  const showMessage =
                    req.type !== "cost" &&
                    req.requirement?.resource !== ResourceTypeCreditsProduction;
                  return (
                    <div
                      key={index}
                      className="flex items-center gap-3 bg-black/70 py-3 px-5 rounded-xl border-2 border-[rgba(255,152,0,0.6)] backdrop-blur-[8px] shadow-[0_0_24px_rgba(255,152,0,0.6)]"
                    >
                      {showMessage && (
                        <span className="text-white text-lg font-medium [text-shadow:2px_2px_4px_rgba(0,0,0,0.9)] leading-[1.3]">
                          {req.message}
                        </span>
                      )}
                      {reqIconType && (
                        <div className="flex-shrink-0">
                          <GameIcon
                            iconType={reqIconType}
                            size="medium"
                            amount={reqIconAmount}
                          />
                        </div>
                      )}
                    </div>
                  );
                })}
              </div>
            ) : (
              <div className="flex items-center gap-3 bg-black/70 py-3 px-5 rounded-xl border-2 border-[rgba(255,152,0,0.6)] backdrop-blur-[8px] shadow-[0_0_24px_rgba(255,152,0,0.6)]">
                {displayReason.type !== "cost" && (
                  <span className="text-white text-lg font-medium [text-shadow:2px_2px_4px_rgba(0,0,0,0.9)] leading-[1.3]">
                    {displayReason.message}
                  </span>
                )}
                {iconType && (
                  <div className="flex-shrink-0">
                    <GameIcon
                      iconType={iconType}
                      size="medium"
                      amount={iconAmount}
                    />
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default HexagonalShieldOverlay;
