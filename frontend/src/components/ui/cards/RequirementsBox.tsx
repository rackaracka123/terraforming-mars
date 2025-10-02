import React from "react";

interface RequirementsBoxProps {
  requirements?: any[];
}

const RequirementsBox: React.FC<RequirementsBoxProps> = ({ requirements }) => {
  if (!requirements || requirements.length === 0) {
    return null;
  }

  // Resource icon mapping
  const getResourceIcon = (resourceType: string): string | null => {
    const iconMap: { [key: string]: string } = {
      // Global parameters
      oxygen: "/assets/global-parameters/oxygen.png",
      temperature: "/assets/global-parameters/temperature.png",
      ocean: "/assets/tiles/ocean.png",
      oceans: "/assets/tiles/ocean.png",

      // Basic resources
      credits: "/assets/resources/megacredit.png",
      megacredits: "/assets/resources/megacredit.png",
      megacredit: "/assets/resources/megacredit.png",
      mc: "/assets/resources/megacredit.png",
      steel: "/assets/resources/steel.png",
      titanium: "/assets/resources/titanium.png",
      plant: "/assets/resources/plant.png",
      plants: "/assets/resources/plant.png",
      energy: "/assets/resources/power.png",
      power: "/assets/resources/power.png",
      heat: "/assets/resources/heat.png",
      card: "/assets/resources/card.png",
      cards: "/assets/resources/card.png",
      "card-draw": "/assets/resources/card.png",
      "card-take": "/assets/resources/card.png",
      "card-peek": "/assets/resources/card.png",
      tr: "/assets/resources/tr.png",
      terraformrating: "/assets/resources/tr.png",
      "terraform-rating": "/assets/resources/tr.png",
      venus: "/assets/global-parameters/venus.png",

      // Special resources
      microbe: "/assets/resources/microbe.png",
      microbes: "/assets/resources/microbe.png",
      animal: "/assets/resources/animal.png",
      animals: "/assets/resources/animal.png",
      science: "/assets/resources/science.png",
      floater: "/assets/resources/floater.png",
      floaters: "/assets/resources/floater.png",
      asteroid: "/assets/resources/asteroid.png",
      asteroids: "/assets/resources/asteroid.png",
      disease: "/assets/resources/disease.png",

      // Production resources
      "credits-production": "/assets/resources/megacredit.png",
      "steel-production": "/assets/resources/steel.png",
      "titanium-production": "/assets/resources/titanium.png",
      "plants-production": "/assets/resources/plant.png",
      "energy-production": "/assets/resources/power.png",
      "heat-production": "/assets/resources/heat.png",

      // Tile placements and counting
      "greenery-placement": "/assets/tiles/greenery.png",
      "greenery-tile": "/assets/tiles/greenery.png",
      greenery: "/assets/tiles/greenery.png",
      "ocean-placement": "/assets/tiles/ocean.png",
      "ocean-tile": "/assets/tiles/ocean.png",
      "city-placement": "/assets/tiles/city.png",
      "city-tile": "/assets/tiles/city.png",
      city: "/assets/tiles/city.png",
      "colony-tile": "/assets/tiles/colony.png",
      effect: "/assets/misc/effect.png",
    };

    const cleanType = resourceType?.toLowerCase().replace(/[_\s]/g, "-");
    return iconMap[cleanType] || iconMap[resourceType?.toLowerCase()] || null;
  };

  // Tag icon mapping
  const getTagIcon = (tagName: string): string | null => {
    const tagMap: { [key: string]: string } = {
      "earth-tag": "/assets/tags/earth.png",
      earth: "/assets/tags/earth.png",
      "science-tag": "/assets/tags/science.png",
      science: "/assets/tags/science.png",
      "plant-tag": "/assets/tags/plant.png",
      plant: "/assets/tags/plant.png",
      "microbe-tag": "/assets/tags/microbe.png",
      microbe: "/assets/tags/microbe.png",
      "animal-tag": "/assets/tags/animal.png",
      animal: "/assets/tags/animal.png",
      "power-tag": "/assets/tags/power.png",
      power: "/assets/tags/power.png",
      energy: "/assets/tags/power.png",
      "space-tag": "/assets/tags/space.png",
      space: "/assets/tags/space.png",
      "building-tag": "/assets/tags/building.png",
      building: "/assets/tags/building.png",
      "city-tag": "/assets/tags/city.png",
      city: "/assets/tags/city.png",
      "jovian-tag": "/assets/tags/jovian.png",
      jovian: "/assets/tags/jovian.png",
      "venus-tag": "/assets/tags/venus.png",
      venus: "/assets/tags/venus.png",
      "event-tag": "/assets/tags/event.png",
      event: "/assets/tags/event.png",
      "mars-tag": "/assets/tags/mars.png",
      mars: "/assets/tags/mars.png",
      "moon-tag": "/assets/tags/moon.png",
      moon: "/assets/tags/moon.png",
      "wild-tag": "/assets/tags/wild.png",
      wild: "/assets/tags/wild.png",
    };

    const cleanTag = tagName?.toLowerCase().replace(/[_\s]/g, "-");
    return tagMap[cleanTag] || tagMap[tagName?.toLowerCase()] || null;
  };

  // Group requirements by type/tag
  const groupRequirements = (requirements: any[]) => {
    const grouped: { [key: string]: any[] } = {};

    requirements.forEach((req) => {
      const key = req.tag || req.affectedTags?.[0] || req.type;
      if (!grouped[key]) {
        grouped[key] = [];
      }
      grouped[key].push(req);
    });

    return Object.values(grouped);
  };

  const renderRequirementGroup = (group: any[], index: number) => {
    const firstReq = group[0];
    const { type, min, max, amount, affectedTags, tag, resource } = firstReq;

    // Determine if it's a tag requirement
    const isTagRequirement = tag || (affectedTags && affectedTags.length > 0);
    // Determine if it's a production requirement
    const isProductionRequirement = type === "production" && resource;
    const key = tag || affectedTags?.[0] || resource || type;

    let icon: string | null = null;
    let displayText = "";
    let showMultipleIcons = false;
    let iconCount = 1;
    if (isProductionRequirement) {
      // Handle production requirements
      icon = getResourceIcon(resource);

      if (min !== undefined && min > 0) {
        if (min === 1) {
          iconCount = 1;
          showMultipleIcons = true;
          displayText = "";
        } else {
          iconCount = 1;
          showMultipleIcons = false;
          displayText = `${min}`;
        }
      } else if (max !== undefined) {
        iconCount = 1;
        showMultipleIcons = false;
        displayText = `Max ${max}`;
      } else if (amount !== undefined) {
        if (amount === 1) {
          iconCount = 1;
          showMultipleIcons = true;
          displayText = "";
        } else {
          iconCount = 1;
          showMultipleIcons = false;
          displayText = `${amount}`;
        }
      } else {
        // Default to 1 production required
        iconCount = 1;
        showMultipleIcons = true;
        displayText = "";
      }
    } else if (isTagRequirement) {
      icon = getTagIcon(key);

      // For tag requirements, determine count and display
      if (min !== undefined && min > 0) {
        if (min === 1) {
          // Single tag - just show the icon
          iconCount = 1;
          showMultipleIcons = true;
          displayText = "";
        } else {
          // Multiple tags - show number before icon
          iconCount = 1;
          showMultipleIcons = false;
          displayText = `${min}`;
        }
      } else if (max !== undefined) {
        iconCount = 1;
        showMultipleIcons = false;
        displayText = `Max ${max}`;
      } else if (amount !== undefined) {
        if (amount === 1) {
          // Single tag - just show the icon
          iconCount = 1;
          showMultipleIcons = true;
          displayText = "";
        } else {
          // Multiple tags - show number before icon
          iconCount = 1;
          showMultipleIcons = false;
          displayText = `${amount}`;
        }
      } else {
        // Default to 1 tag required
        iconCount = 1;
        showMultipleIcons = true;
        displayText = "";
      }
    } else {
      // Handle resource/parameter requirements
      icon = getResourceIcon(type);

      // Add proper units for global parameters
      if (type === "oxygen") {
        if (min !== undefined && min > 0) {
          displayText = `${min}%+`;
        } else if (max !== undefined) {
          displayText = `≤${max}%`;
        } else if (amount !== undefined) {
          displayText = `${amount}%`;
        }
      } else if (type === "temperature") {
        if (min !== undefined) {
          displayText = `${min}°C+`;
        } else if (max !== undefined) {
          displayText = `≤${max}°C`;
        } else if (amount !== undefined) {
          displayText = `${amount}°C`;
        }
      } else {
        // Regular resources
        if (min !== undefined && min > 0) {
          displayText = `${min}+`;
        } else if (max !== undefined) {
          displayText = `≤${max}`;
        } else if (amount !== undefined) {
          displayText = `${amount}`;
        } else {
          displayText = type;
        }
      }
    }

    return (
      <div
        key={index}
        className="flex items-center gap-px px-0.5 py-px [&:has(span)]:px-1 [&:has(span)]:py-0.5"
      >
        {/* Show amount before icon for tag requirements with multiple tags */}
        {isTagRequirement && displayText && !showMultipleIcons && (
          <span className="text-[11px] font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] leading-none">
            {displayText}
          </span>
        )}

        {/* Show amount before icon for production requirements with multiple units */}
        {isProductionRequirement && displayText && !showMultipleIcons && (
          <span className="text-[11px] font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] leading-none">
            {displayText}
          </span>
        )}

        {icon ? (
          <div className="flex items-center gap-px">
            {isProductionRequirement ? (
              // Production requirements with brown background
              <div className="relative flex items-center justify-center">
                <img
                  src="/assets/misc/production.png"
                  alt="production"
                  className="w-5 h-5 object-contain [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.6))] max-md:w-[18px] max-md:h-[18px]"
                />
                {showMultipleIcons ? (
                  Array.from({ length: Math.min(iconCount, 4) }, (_, i) => (
                    <img
                      key={i}
                      src={icon}
                      alt={resource}
                      className="absolute w-3 h-3 object-contain [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.6))] max-md:w-2.5 max-md:h-2.5"
                    />
                  ))
                ) : (
                  <img
                    src={icon}
                    alt={resource}
                    className="absolute w-3 h-3 object-contain [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.6))] max-md:w-2.5 max-md:h-2.5"
                  />
                )}
              </div>
            ) : showMultipleIcons ? (
              // Show multiple icons for single tag requirements
              Array.from({ length: Math.min(iconCount, 4) }, (_, i) => (
                <img
                  key={i}
                  src={icon}
                  alt={key}
                  className="w-5 h-5 object-contain [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.6))] max-md:w-[18px] max-md:h-[18px]"
                />
              ))
            ) : (
              // Show single icon
              <img
                src={icon}
                alt={key}
                className="w-5 h-5 object-contain [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.6))] max-md:w-[18px] max-md:h-[18px]"
              />
            )}
          </div>
        ) : (
          <span className="text-[10px] font-semibold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] capitalize max-md:text-[9px]">
            {key}
          </span>
        )}

        {/* Show amount after icon for non-tag, non-production requirements */}
        {!isTagRequirement && !isProductionRequirement && displayText && (
          <span className="text-[11px] font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] leading-none max-md:text-[10px]">
            {displayText}
          </span>
        )}
      </div>
    );
  };

  const groupedRequirements = groupRequirements(requirements);

  return (
    <div className="absolute bottom-full left-[10%] w-fit min-w-[60px] max-w-[80%] z-[-10] bg-[linear-gradient(135deg,rgba(255,87,34,0.15)_0%,rgba(255,69,0,0.12)_100%)] rounded-[2px] shadow-[0_3px_8px_rgba(0,0,0,0.4)] backdrop-blur-[2px] pt-1 pr-0.5 pb-1 pl-1.5 before:content-[''] before:absolute before:top-0 before:left-0 before:right-0 before:bottom-0 before:border-2 before:border-[rgba(255,87,34,0.9)] before:rounded-[2px] before:pointer-events-none max-md:min-w-[50px] max-md:pt-[3px] max-md:pr-0.5 max-md:pb-2.5 max-md:pl-1">
      <div className="flex items-center justify-start gap-[3px] flex-wrap max-md:gap-1">
        {groupedRequirements.map((group, index) =>
          renderRequirementGroup(group, index),
        )}
      </div>
    </div>
  );
};

export default RequirementsBox;
