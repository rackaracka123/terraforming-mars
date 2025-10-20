import React from "react";
import GameIcon from "../display/GameIcon.tsx";
import BehaviorSection from "./BehaviorSection";

import {
  CardBehaviorDto,
  ResourceTypeCredits,
  ResourceTypeSteel,
  ResourceTypeTitanium,
  ResourceTypePlants,
  ResourceTypeEnergy,
  ResourceTypeHeat,
} from "../../../types/generated/api-types.ts";
import { getCorporationLogo } from "@/utils/corporationLogos.tsx";

interface Corporation {
  id: string;
  name: string;
  description: string;
  startingMegaCredits: number;
  startingProduction?: {
    credits?: number;
    steel?: number;
    titanium?: number;
    plants?: number;
    energy?: number;
    heat?: number;
  };
  startingResources?: {
    credits?: number;
    steel?: number;
    titanium?: number;
    plants?: number;
    energy?: number;
    heat?: number;
  };
  behaviors?: CardBehaviorDto[];
  specialEffects?: string[];
  expansion?: string;
  logoPath?: string;
}

interface CorporationCardProps {
  corporation: Corporation;
  isSelected: boolean;
  onSelect: (corporationId: string) => void;
  showCheckbox?: boolean; // Whether to show the selection checkbox (default: false)
}

const CorporationCard: React.FC<CorporationCardProps> = ({
  corporation,
  isSelected,
  onSelect,
  showCheckbox = false,
}) => {
  const renderResource = (type: string, amount: number) => {
    const resourceTypeMap: { [key: string]: string } = {
      credits: ResourceTypeCredits,
      steel: ResourceTypeSteel,
      titanium: ResourceTypeTitanium,
      plants: ResourceTypePlants,
      energy: ResourceTypeEnergy,
      heat: ResourceTypeHeat,
    };

    const resourceType = resourceTypeMap[type];
    if (!resourceType) return null;

    // Use GameIcon for credits (shows amount inside icon)
    if (type === "credits") {
      return <GameIcon iconType={resourceType} amount={amount} size="large" />;
    }

    // Regular resource display with icon and number
    return (
      <div className="inline-flex items-center gap-2">
        <GameIcon iconType={resourceType} size="large" />
        <span className="text-white font-bold text-lg [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)]">
          {amount}
        </span>
      </div>
    );
  };

  const renderProduction = (type: string, amount: number) => {
    const resourceTypeMap: { [key: string]: string } = {
      credits: ResourceTypeCredits,
      steel: ResourceTypeSteel,
      titanium: ResourceTypeTitanium,
      plants: ResourceTypePlants,
      energy: ResourceTypeEnergy,
      heat: ResourceTypeHeat,
    };

    const resourceType = resourceTypeMap[type];
    if (!resourceType) return null;

    // Use GameIcon with -production suffix for automatic brown background
    return (
      <GameIcon
        iconType={`${resourceType}-production`}
        amount={amount}
        size="medium"
      />
    );
  };

  // Filter out the first auto behavior without conditions (starting bonuses shown above)
  const filterBehaviors = (behaviors: CardBehaviorDto[] | undefined) => {
    if (!behaviors || behaviors.length === 0) return [];

    return behaviors.filter((behavior, index) => {
      const hasCondition = behavior.triggers?.some(
        (t) => t.condition !== undefined,
      );
      const isAuto = behavior.triggers?.some((t) => t.type === "auto");

      // Skip the first auto behavior without conditions (starting bonuses)
      if (isAuto && !hasCondition && index === 0) {
        return false;
      }

      return true;
    });
  };

  return (
    <div
      className={`w-[400px] h-[380px] relative bg-[rgba(5,4,2,0.98)] border-2 rounded-xl p-3 cursor-pointer transition-all duration-300 ease-[ease] ${
        isSelected
          ? "border-[#ffc107] shadow-[0_4px_20px_rgba(255,193,7,0.3),0_0_40px_rgba(255,193,7,0.2)]"
          : "border-[rgba(255,193,7,0.3)] hover:-translate-y-0.5 hover:shadow-[0_8px_25px_rgba(0,0,0,0.4),0_0_15px_rgba(255,193,7,0.15)] hover:border-[rgba(255,193,7,0.5)]"
      }`}
      onClick={() => onSelect(corporation.id)}
    >
      {/* Logo centered at top */}
      {corporation.logoPath && (
        <div className="flex justify-center mb-2">
          <img
            src={corporation.logoPath}
            alt={corporation.name}
            className="w-20 h-20 rounded-lg object-cover [filter:drop-shadow(0_2px_6px_rgba(0,0,0,0.6))]"
          />
        </div>
      )}

      <div className="mb-3 p-3 bg-black/30 rounded-lg flex justify-center [&>*]:box-content [&>*]:[filter:drop-shadow(0_2px_4px_rgba(0,0,0,0.8))] [&>*]:rounded-[4px] [&>*>*]:rounded-[4px] [&>*>*]:[filter:drop-shadow(0_2px_4px_rgba(0,0,0,0.8))]">
        {getCorporationLogo(corporation.name.toLowerCase())}
      </div>

      {/* Starting resources and production - compact, no headers */}
      {(corporation.startingProduction || corporation.startingResources) && (
        <div className="flex flex-wrap gap-2 justify-center items-center mb-3 pb-3 border-b border-white/20">
          {corporation.startingResources &&
            Object.entries(corporation.startingResources).map(
              ([type, amount]) =>
                amount > 0 ? (
                  <div key={type}>{renderResource(type, amount)}</div>
                ) : null,
            )}
          {corporation.startingProduction &&
            Object.entries(corporation.startingProduction).map(
              ([type, amount]) =>
                amount > 0 ? (
                  <div key={type}>{renderProduction(type, amount)}</div>
                ) : null,
            )}
        </div>
      )}

      {/* Behaviors - using BehaviorSection component */}
      {filterBehaviors(corporation.behaviors).length > 0 && (
        <div className="mb-3 border-b border-white/20 pb-3">
          <div className="relative [&>div]:static [&>div]:!bottom-auto [&>div]:!left-auto [&>div]:!right-auto">
            <BehaviorSection
              behaviors={filterBehaviors(corporation.behaviors)}
            />
          </div>
        </div>
      )}

      {/* Description at bottom */}
      <div className="text-xs text-white/80 leading-[1.4] text-center">
        {corporation.description}
      </div>

      {/* Expansion badge */}
      {corporation.expansion && (
        <div className="absolute top-2 right-2 bg-[rgba(100,150,255,0.3)] text-white/80 py-0.5 px-1.5 rounded text-[9px] uppercase tracking-[0.5px]">
          {corporation.expansion}
        </div>
      )}

      {/* Selection indicator at bottom center (only shown when showCheckbox is true) */}
      {showCheckbox && (
        <div className="absolute -bottom-3 left-1/2 -translate-x-1/2 z-[2]">
          <div
            className={`w-6 h-6 rounded-full bg-[#1a1508] border-2 border-[rgba(255,193,7,0.3)] flex items-center justify-center transition-all duration-300 ${isSelected ? "bg-[#3a2f0d] border-[#ffc107]" : ""}`}
          >
            {isSelected && (
              <span className="text-white text-sm font-bold">✓</span>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export default CorporationCard;
