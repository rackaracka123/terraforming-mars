import React from "react";
import { getIconPath } from "@/utils/iconStore.ts";

interface VictoryPointIconProps {
  value?: number | string; // Deprecated: for backward compatibility
  vpConditions?: any[]; // New: VP conditions array
  size?: "small" | "medium" | "large";
}

const VictoryPointIcon: React.FC<VictoryPointIconProps> = ({
  value,
  vpConditions,
  size = "medium",
}) => {
  const vpIconPath = getIconPath("mars");

  const sizeClasses = {
    small: "w-8 h-8 text-[calc(32px*0.7)]",
    medium: "w-10 h-10 text-[calc(40px*0.7)]",
    large: "w-12 h-12 text-[calc(48px*0.7)]",
  };

  // If vpConditions is provided, use the new system
  if (vpConditions && Array.isArray(vpConditions) && vpConditions.length > 0) {
    // Handle multiple VP conditions - for now, render each separately or combine them
    const totalConditions = vpConditions.length;

    if (totalConditions === 1) {
      const condition = vpConditions[0];

      if (condition.condition === "fixed") {
        // Fixed VP amount
        if (condition.amount === 0) return null;
        return (
          <div
            className={`relative inline-flex items-center justify-center ${sizeClasses[size]}`}
          >
            <img
              src={vpIconPath || ""}
              alt="VP"
              className="w-full h-full object-contain brightness-[0.7] [filter:drop-shadow(0_2px_4px_rgba(0,0,0,0.6))]"
            />
            <span className="absolute top-0 left-0 right-0 bottom-0 text-black font-bold font-[Prototype,Arial_Black,Arial,sans-serif] flex items-center justify-center text-center leading-none [text-shadow:-1px_-1px_0_#d2691e,1px_-1px_0_#d2691e,-1px_1px_0_#d2691e,1px_1px_0_#d2691e,0_0_3px_rgba(210,105,30,0.5)] tracking-[0.3px] [-webkit-font-smoothing:antialiased] [-moz-osx-font-smoothing:grayscale] [text-rendering:optimizeLegibility]">
              {condition.amount}
            </span>
          </div>
        );
      } else if (condition.condition === "per" && condition.per) {
        // Per condition - display as fraction with icon
        const perCondition = condition.per;
        let resourceIcon = null;
        let displayText = "";

        // Get the resource icon - check tag first, then type
        const resourceType = perCondition.tag || perCondition.type;
        if (resourceType) {
          resourceIcon = getIconPath(resourceType);
          // If per.amount is 1, show slash but omit the number (e.g., "1/" instead of "1/1")
          if ((perCondition.amount || 1) === 1) {
            displayText = `${condition.amount}/`;
          } else {
            displayText = `${condition.amount}/${perCondition.amount}`;
          }
        }

        // Calculate text size based on content length
        const textLength = displayText.length;
        const textSizeClass =
          textLength <= 3
            ? "text-[calc(100%*0.6)]" // Smaller size for single row layout
            : "text-[calc(100%*0.45)]"; // Even smaller for longer text

        return (
          <div
            className={`relative inline-flex items-center justify-center ${sizeClasses[size]}`}
          >
            <img
              src={vpIconPath || ""}
              alt="VP"
              className="w-full h-full object-contain brightness-[0.7] [filter:drop-shadow(0_2px_4px_rgba(0,0,0,0.6))]"
            />
            <div className="absolute inset-0 flex flex-row items-center justify-center gap-0.5 p-1">
              <span
                className={`text-black font-bold font-[Prototype,Arial_Black,Arial,sans-serif] leading-none [text-shadow:-1px_-1px_0_#d2691e,1px_-1px_0_#d2691e,-1px_1px_0_#d2691e,1px_1px_0_#d2691e,0_0_3px_rgba(210,105,30,0.5)] [-webkit-font-smoothing:antialiased] [-moz-osx-font-smoothing:grayscale] [text-rendering:optimizeLegibility] ${textSizeClass} tracking-[-0.5px]`}
              >
                {displayText}
              </span>
              {resourceIcon && (
                <img
                  src={resourceIcon}
                  alt={resourceType}
                  className="w-[40%] h-[40%] object-contain [filter:drop-shadow(0_2px_4px_rgba(0,0,0,0.8))_drop-shadow(0_0_2px_rgba(0,0,0,0.6))]"
                />
              )}
            </div>
          </div>
        );
      } else if (condition.condition === "once") {
        // Once condition - similar to fixed but different styling?
        if (condition.amount === 0) return null;
        return (
          <div
            className={`relative inline-flex items-center justify-center ${sizeClasses[size]}`}
          >
            <img
              src={vpIconPath || ""}
              alt="VP"
              className="w-full h-full object-contain brightness-[0.7] [filter:drop-shadow(0_2px_4px_rgba(0,0,0,0.6))]"
            />
            <span className="absolute top-0 left-0 right-0 bottom-0 text-black font-bold font-[Prototype,Arial_Black,Arial,sans-serif] flex items-center justify-center text-center leading-none [text-shadow:-1px_-1px_0_#d2691e,1px_-1px_0_#d2691e,-1px_1px_0_#d2691e,1px_1px_0_#d2691e,0_0_3px_rgba(210,105,30,0.5)] tracking-[0.3px] [-webkit-font-smoothing:antialiased] [-moz-osx-font-smoothing:grayscale] [text-rendering:optimizeLegibility]">
              {condition.amount}
            </span>
          </div>
        );
      }
    } else {
      // Multiple conditions - sum up fixed ones and show first per condition
      let totalFixed = 0;
      let firstPerCondition = null;

      for (const condition of vpConditions) {
        if (condition.condition === "fixed" || condition.condition === "once") {
          totalFixed += condition.amount;
        } else if (condition.condition === "per" && !firstPerCondition) {
          firstPerCondition = condition;
        }
      }

      // For now, just show the total fixed VP or the first per condition
      if (firstPerCondition && firstPerCondition.per) {
        const perCondition = firstPerCondition.per;
        let resourceIcon = null;
        let displayText = "";

        // Get the resource icon - check tag first, then type
        const resourceType = perCondition.tag || perCondition.type;
        if (resourceType) {
          resourceIcon = getIconPath(resourceType);
          // If per.amount is 1, show slash but omit the number (e.g., "1/" instead of "1/1")
          if ((perCondition.amount || 1) === 1) {
            displayText = `${firstPerCondition.amount}/`;
          } else {
            displayText = `${firstPerCondition.amount}/${perCondition.amount}`;
          }
        }

        // Calculate text size based on content length
        const textLength = displayText.length;
        const textSizeClass =
          textLength <= 3
            ? "text-[calc(100%*0.6)]" // Smaller size for single row layout
            : "text-[calc(100%*0.45)]"; // Even smaller for longer text

        return (
          <div
            className={`relative inline-flex items-center justify-center ${sizeClasses[size]}`}
          >
            <img
              src={vpIconPath || ""}
              alt="VP"
              className="w-full h-full object-contain brightness-[0.7] [filter:drop-shadow(0_2px_4px_rgba(0,0,0,0.6))]"
            />
            <div className="absolute inset-0 flex flex-row items-center justify-center gap-0.5 p-1">
              <span
                className={`text-black font-bold font-[Prototype,Arial_Black,Arial,sans-serif] leading-none [text-shadow:-1px_-1px_0_#d2691e,1px_-1px_0_#d2691e,-1px_1px_0_#d2691e,1px_1px_0_#d2691e,0_0_3px_rgba(210,105,30,0.5)] [-webkit-font-smoothing:antialiased] [-moz-osx-font-smoothing:grayscale] [text-rendering:optimizeLegibility] ${textSizeClass} tracking-[-0.5px]`}
              >
                {displayText}
              </span>
              {resourceIcon && (
                <img
                  src={resourceIcon}
                  alt={resourceType}
                  className="w-[40%] h-[40%] object-contain [filter:drop-shadow(0_2px_4px_rgba(0,0,0,0.8))_drop-shadow(0_0_2px_rgba(0,0,0,0.6))]"
                />
              )}
            </div>
          </div>
        );
      } else if (totalFixed > 0) {
        return (
          <div
            className={`relative inline-flex items-center justify-center ${sizeClasses[size]}`}
          >
            <img
              src={vpIconPath || ""}
              alt="VP"
              className="w-full h-full object-contain brightness-[0.7] [filter:drop-shadow(0_2px_4px_rgba(0,0,0,0.6))]"
            />
            <span className="absolute top-0 left-0 right-0 bottom-0 text-black font-bold font-[Prototype,Arial_Black,Arial,sans-serif] flex items-center justify-center text-center leading-none [text-shadow:-1px_-1px_0_#d2691e,1px_-1px_0_#d2691e,-1px_1px_0_#d2691e,1px_1px_0_#d2691e,0_0_3px_rgba(210,105,30,0.5)] tracking-[0.3px] [-webkit-font-smoothing:antialiased] [-moz-osx-font-smoothing:grayscale] [text-rendering:optimizeLegibility]">
              {totalFixed}
            </span>
          </div>
        );
      }
    }

    return null; // No valid conditions
  }

  // Fallback to old system for backward compatibility
  if (value === 0 || !value) {
    return null;
  }

  return (
    <div
      className={`relative inline-flex items-center justify-center ${sizeClasses[size]}`}
    >
      <img
        src={vpIconPath || ""}
        alt="VP"
        className="w-full h-full object-contain brightness-[0.7] [filter:drop-shadow(0_2px_4px_rgba(0,0,0,0.6))]"
      />
    </div>
  );
};

export default VictoryPointIcon;
