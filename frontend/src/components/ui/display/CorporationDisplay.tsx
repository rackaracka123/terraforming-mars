import React, { useState, useEffect, useRef } from "react";
import { CardDto } from "@/types/generated/api-types.ts";
import { getCorporationLogo } from "@/utils/corporationLogos.tsx";
import GameIcon from "./GameIcon.tsx";
import BehaviorSection from "../cards/BehaviorSection";
import {
  ResourceTypeCredits,
  ResourceTypeSteel,
  ResourceTypeTitanium,
  ResourceTypePlants,
  ResourceTypeEnergy,
  ResourceTypeHeat,
} from "@/types/generated/api-types.ts";

interface CorporationDisplayProps {
  corporation: CardDto;
}

const CorporationDisplay: React.FC<CorporationDisplayProps> = ({
  corporation,
}) => {
  const [isExpanded, setIsExpanded] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  const toggleExpanded = (e: React.MouseEvent) => {
    e.stopPropagation();
    setIsExpanded(!isExpanded);
  };

  // Close when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (
        containerRef.current &&
        !containerRef.current.contains(event.target as Node)
      ) {
        setIsExpanded(false);
      }
    };

    if (isExpanded) {
      document.addEventListener("mousedown", handleClickOutside);
    }

    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [isExpanded]);

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
    if (!resourceType || !amount) return null;

    if (type === "credits") {
      return <GameIcon iconType={resourceType} amount={amount} size="large" />;
    }

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
    if (!resourceType || !amount) return null;

    return (
      <GameIcon
        iconType={`${resourceType}-production`}
        amount={amount}
        size="medium"
      />
    );
  };

  const filterBehaviors = (behaviors: CardDto["behaviors"]) => {
    if (!behaviors || behaviors.length === 0) return [];

    return behaviors.filter((behavior, index) => {
      const hasCondition = behavior.triggers?.some(
        (t) => t.condition !== undefined,
      );
      const isAuto = behavior.triggers?.some((t) => t.type === "auto");

      if (isAuto && !hasCondition && index === 0) {
        return false;
      }

      return true;
    });
  };

  return (
    <div
      ref={containerRef}
      className="fixed bottom-[150px] left-[30px] z-[999] pointer-events-auto"
      title={
        isExpanded ? "" : `${corporation.name}\n${corporation.description}`
      }
    >
      <div
        className={`bg-black/95 rounded-lg backdrop-blur-space transition-all duration-300 cursor-pointer ${
          isExpanded
            ? "w-[350px] p-4 shadow-[0_0_20px_rgba(30,60,150,0.52)] hover:shadow-[0_0_30px_rgba(30,60,150,0.78)]"
            : "p-1.5 shadow-[0_0_15px_rgba(30,60,150,0.39)] hover:-translate-y-1 hover:shadow-[0_0_25px_rgba(30,60,150,0.65)]"
        }`}
        onClick={toggleExpanded}
      >
        {!isExpanded ? (
          /* Collapsed: Logo Only */
          <div className="rounded-lg p-1 bg-black/30 flex items-center justify-center [&>*]:box-content [&>*]:[filter:drop-shadow(0_2px_4px_rgba(0,0,0,0.8))] [&>*]:rounded-[4px] [&>*>*]:rounded-[4px] [&>*>*]:[filter:drop-shadow(0_2px_4px_rgba(0,0,0,0.8))]">
            {getCorporationLogo(corporation.name.toLowerCase())}
          </div>
        ) : (
          /* Expanded: Full Card Details */
          <div className="space-y-3">
            {/* Corporation Logo */}
            <div className="p-3 bg-black/30 rounded-lg flex justify-center [&>*]:box-content [&>*]:[filter:drop-shadow(0_2px_4px_rgba(0,0,0,0.8))] [&>*]:rounded-[4px] [&>*>*]:rounded-[4px] [&>*>*]:[filter:drop-shadow(0_2px_4px_rgba(0,0,0,0.8))]">
              {getCorporationLogo(corporation.name.toLowerCase())}
            </div>

            {/* Corporation Name */}
            <h3 className="text-lg font-bold text-white text-center [text-shadow:0_2px_4px_rgba(0,0,0,0.8)]">
              {corporation.name}
            </h3>

            {/* Starting resources and production */}
            {(corporation.startingProduction ||
              corporation.startingResources) && (
              <div className="p-3 bg-black/20 rounded-lg">
                <h4 className="text-xs font-semibold text-white/70 uppercase tracking-[0.5px] mb-2 text-center">
                  Starting Bonuses
                </h4>
                <div className="flex flex-wrap gap-2 justify-center items-center">
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
              </div>
            )}

            {/* Behaviors */}
            {filterBehaviors(corporation.behaviors).length > 0 && (
              <div className="p-3 bg-black/20 rounded-lg">
                <h4 className="text-xs font-semibold text-white/70 uppercase tracking-[0.5px] mb-2 text-center">
                  Special Abilities
                </h4>
                <div className="relative [&>div]:static [&>div]:!bottom-auto [&>div]:!left-auto [&>div]:!right-auto">
                  <BehaviorSection
                    behaviors={filterBehaviors(corporation.behaviors)}
                  />
                </div>
              </div>
            )}

            {/* Description */}
            <div className="p-3 bg-black/20 rounded-lg">
              <h4 className="text-xs font-semibold text-white/70 uppercase tracking-[0.5px] mb-2 text-center">
                Description
              </h4>
              <p className="text-xs text-white/90 leading-[1.4] text-center">
                {corporation.description}
              </p>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default CorporationDisplay;
