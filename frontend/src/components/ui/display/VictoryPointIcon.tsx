import React from "react";

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
  // Helper function to get tag icon for VP conditions
  const getTagIcon = (tag: string) => {
    const iconMap: { [key: string]: string } = {
      jovian: "/assets/tags/jovian.png",
      science: "/assets/tags/science.png",
      space: "/assets/tags/space.png",
      microbe: "/assets/tags/microbe.png",
      animal: "/assets/tags/animal.png",
      plant: "/assets/tags/plant.png",
      earth: "/assets/tags/earth.png",
      building: "/assets/tags/building.png",
      power: "/assets/tags/power.png",
      city: "/assets/tags/city.png",
      venus: "/assets/tags/venus.png",
      event: "/assets/tags/event.png",
      wild: "/assets/tags/wild.png",
    };
    return iconMap[tag.toLowerCase()] || null;
  };

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
              src="/assets/mars.png"
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
        let tagIcon = null;
        let displayText = "";

        if (perCondition.tag) {
          tagIcon = getTagIcon(perCondition.tag);
          displayText = `${condition.amount}/${perCondition.amount || 1}`;
        } else if (perCondition.type) {
          // Handle other per types (city-tile, ocean-tile, etc.)
          displayText = `${condition.amount}/${perCondition.amount || 1}`;
        }

        return (
          <div
            className={`relative inline-flex items-center justify-center gap-0.5 ${sizeClasses[size]}`}
          >
            <img
              src="/assets/mars.png"
              alt="VP"
              className="w-full h-full object-contain brightness-[0.7] [filter:drop-shadow(0_2px_4px_rgba(0,0,0,0.6))]"
            />
            <span className="absolute top-0 left-0 right-0 bottom-0 text-black font-bold font-[Prototype,Arial_Black,Arial,sans-serif] flex items-center justify-center text-center leading-none [text-shadow:-1px_-1px_0_#d2691e,1px_-1px_0_#d2691e,-1px_1px_0_#d2691e,1px_1px_0_#d2691e,0_0_3px_rgba(210,105,30,0.5)] [-webkit-font-smoothing:antialiased] [-moz-osx-font-smoothing:grayscale] [text-rendering:optimizeLegibility] text-[calc(100%*1.8)] tracking-[-0.5px]">
              {displayText}
            </span>
            {tagIcon && (
              <img
                src={tagIcon}
                alt={perCondition.tag}
                className="w-4/5 h-4/5 object-contain [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.4))]"
              />
            )}
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
              src="/assets/mars.png"
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
        let tagIcon = null;
        let displayText = "";

        if (perCondition.tag) {
          tagIcon = getTagIcon(perCondition.tag);
          displayText = `${firstPerCondition.amount}/${perCondition.amount || 1}`;
        }

        return (
          <div
            className={`relative inline-flex items-center justify-center gap-0.5 ${sizeClasses[size]}`}
          >
            <img
              src="/assets/mars.png"
              alt="VP"
              className="w-full h-full object-contain brightness-[0.7] [filter:drop-shadow(0_2px_4px_rgba(0,0,0,0.6))]"
            />
            <span className="absolute top-0 left-0 right-0 bottom-0 text-black font-bold font-[Prototype,Arial_Black,Arial,sans-serif] flex items-center justify-center text-center leading-none [text-shadow:-1px_-1px_0_#d2691e,1px_-1px_0_#d2691e,-1px_1px_0_#d2691e,1px_1px_0_#d2691e,0_0_3px_rgba(210,105,30,0.5)] [-webkit-font-smoothing:antialiased] [-moz-osx-font-smoothing:grayscale] [text-rendering:optimizeLegibility] text-[calc(100%*1.8)] tracking-[-0.5px]">
              {displayText}
            </span>
            {tagIcon && (
              <img
                src={tagIcon}
                alt={perCondition.tag}
                className="w-4/5 h-4/5 object-contain [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.4))]"
              />
            )}
          </div>
        );
      } else if (totalFixed > 0) {
        return (
          <div
            className={`relative inline-flex items-center justify-center ${sizeClasses[size]}`}
          >
            <img
              src="/assets/mars.png"
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

  // Helper function to format victory point text
  const formatVictoryPoints = (val: number | string): string => {
    if (typeof val === "number") {
      return val.toString();
    }

    // Handle special Terraforming Mars syntax
    return val
      .replace(/1 point per animal/gi, "1/🐾")
      .replace(/1 point per jupiter card/gi, "1/♃")
      .replace(/1 point per jovian/gi, "1/♃")
      .replace(/1 point per science/gi, "1/🧪")
      .replace(/1 point per space/gi, "1/🚀")
      .replace(/1 point per microbe/gi, "1/🦠")
      .replace(/1 point per plant/gi, "1/🌱")
      .replace(/1 point per city/gi, "1/🏙️")
      .replace(/1 point per earth/gi, "1/🌍")
      .replace(/1 point per energy/gi, "1/⚡")
      .replace(/1 point per building/gi, "1/🏢")
      .replace(/1 point per venus/gi, "1/♀");
  };

  return (
    <div
      className={`relative inline-flex items-center justify-center ${sizeClasses[size]}`}
    >
      <img
        src="/assets/mars.png"
        alt="VP"
        className="w-full h-full object-contain brightness-[0.7] [filter:drop-shadow(0_2px_4px_rgba(0,0,0,0.6))]"
      />
      <span className="absolute top-0 left-0 right-0 bottom-0 text-black font-bold font-[Prototype,Arial_Black,Arial,sans-serif] flex items-center justify-center text-center leading-none [text-shadow:-1px_-1px_0_#d2691e,1px_-1px_0_#d2691e,-1px_1px_0_#d2691e,1px_1px_0_#d2691e,0_0_3px_rgba(210,105,30,0.5)] tracking-[0.3px] [-webkit-font-smoothing:antialiased] [-moz-osx-font-smoothing:grayscale] [text-rendering:optimizeLegibility]">
        {formatVictoryPoints(value)}
      </span>
    </div>
  );
};

export default VictoryPointIcon;
