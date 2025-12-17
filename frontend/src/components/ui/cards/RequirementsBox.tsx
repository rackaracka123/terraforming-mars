import React from "react";
import GameIcon from "../display/GameIcon.tsx";
import { CardTag, ResourceType } from "@/types/generated/api-types.ts";

interface RequirementsBoxProps {
  requirements?: any[];
}

const RequirementsBox: React.FC<RequirementsBoxProps> = ({ requirements }) => {
  if (!requirements || requirements.length === 0) {
    return null;
  }

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

    let resourceType: ResourceType | null = null;
    let cardTag: CardTag | null = null;
    let displayText = "";
    let showMultipleIcons = false;
    let iconCount = 1;

    if (isTagRequirement) {
      cardTag = key;
    } else if (isProductionRequirement) {
      // Production requirements - use resource as-is (already contains -production)
      resourceType = resource;
    } else {
      // Regular resource/parameter requirements
      resourceType = type;
    }

    // Determine count and display text
    if (isTagRequirement || isProductionRequirement) {
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
        iconCount = 1;
        showMultipleIcons = true;
        displayText = "";
      }
    } else {
      // Regular resource requirements - handle parameter-specific formatting

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
        {/* Show amount before icon for tag/production requirements with multiple units */}
        {(isTagRequirement || isProductionRequirement) && displayText && !showMultipleIcons && (
          <span className="text-[11px] font-bold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] leading-none">
            {displayText}
          </span>
        )}

        {resourceType || cardTag ? (
          <div className="flex items-center gap-px">
            {showMultipleIcons ? (
              // Show multiple icons for single tag/resource requirements
              Array.from({ length: Math.min(iconCount, 4) }, (_, i) => (
                <GameIcon
                  key={i}
                  iconType={cardTag ? `${cardTag}-tag` : (resourceType as string)}
                  size="small"
                />
              ))
            ) : (
              // Show single icon
              <GameIcon
                iconType={cardTag ? `${cardTag}-tag` : (resourceType as string)}
                size="small"
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
        {groupedRequirements.map((group, index) => renderRequirementGroup(group, index))}
      </div>
    </div>
  );
};

export default RequirementsBox;
