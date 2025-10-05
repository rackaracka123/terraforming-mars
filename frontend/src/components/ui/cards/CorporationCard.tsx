import React from "react";
import MegaCreditIcon from "../display/MegaCreditIcon.tsx";
import ProductionDisplay from "../display/ProductionDisplay.tsx";

import {
  CardBehaviorDto,
  ResourceTriggerManual,
} from "../../../types/generated/api-types.ts";

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
}

const CorporationCard: React.FC<CorporationCardProps> = ({
  corporation,
  isSelected,
  onSelect,
}) => {
  const renderResource = (type: string, amount: number) => {
    const iconMap: { [key: string]: string } = {
      steel: "/assets/resources/steel.png",
      titanium: "/assets/resources/titanium.png",
      plants: "/assets/resources/plant.png",
      energy: "/assets/resources/power.png",
      heat: "/assets/resources/heat.png",
    };

    // Use MegaCreditIcon for credits
    if (type === "credits") {
      return <MegaCreditIcon value={amount} size="large" />;
    }

    const icon = iconMap[type];
    if (!icon) return null;

    // Regular resource display
    return (
      <div className="inline-flex items-center gap-2">
        <img
          src={icon}
          alt={type}
          className="w-8 h-8 [filter:drop-shadow(0_2px_4px_rgba(0,0,0,0.6))]"
        />
        <span className="text-white font-bold text-lg [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)]">
          {amount}
        </span>
      </div>
    );
  };

  const renderResourceIcon = (
    resourceType: string,
    amount: number,
    isProduction: boolean,
  ) => {
    if (isProduction) {
      return (
        <ProductionDisplay
          amount={amount}
          resourceType={resourceType}
          size="medium"
        />
      );
    } else if (resourceType === "credits") {
      return <MegaCreditIcon value={amount} size="medium" />;
    } else {
      const iconMap: { [key: string]: string } = {
        steel: "/assets/resources/steel.png",
        titanium: "/assets/resources/titanium.png",
        plants: "/assets/resources/plant.png",
        energy: "/assets/resources/power.png",
        heat: "/assets/resources/heat.png",
      };

      const icon = iconMap[resourceType];
      if (!icon) return null;

      return (
        <div className="inline-flex items-center gap-1">
          <img
            src={icon}
            alt={resourceType}
            className="w-6 h-6 [filter:drop-shadow(0_2px_4px_rgba(0,0,0,0.6))]"
          />
          <span className="text-white font-bold text-sm [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)]">
            {amount}
          </span>
        </div>
      );
    }
  };

  const renderBehavior = (behavior: CardBehaviorDto) => {
    const isManualAction = behavior.triggers?.some(
      (t) => t.type === ResourceTriggerManual,
    );
    const hasInputs = behavior.inputs && behavior.inputs.length > 0;
    const hasOutputs = behavior.outputs && behavior.outputs.length > 0;

    if (!hasOutputs && !hasInputs) return null;

    return (
      <div className="flex items-center gap-2 flex-wrap">
        {/* Render inputs */}
        {hasInputs &&
          behavior.inputs!.map((input, idx) => {
            const resourceType = input.type.replace("-production", "");
            const isProduction = input.type.includes("-production");
            return (
              <div key={`input-${idx}`}>
                {renderResourceIcon(resourceType, input.amount, isProduction)}
              </div>
            );
          })}

        {/* Render arrow for action behaviors */}
        {isManualAction && hasInputs && hasOutputs && (
          <img
            src="/assets/misc/arrow.png"
            alt="converts to"
            className="w-6 h-6 [filter:drop-shadow(0_2px_4px_rgba(0,0,0,0.6))]"
          />
        )}

        {/* Render outputs */}
        {hasOutputs &&
          behavior.outputs!.map((output, idx) => {
            const resourceType = output.type.replace("-production", "");
            const isProduction = output.type.includes("-production");
            return (
              <div key={`output-${idx}`}>
                {renderResourceIcon(resourceType, output.amount, isProduction)}
              </div>
            );
          })}
      </div>
    );
  };

  return (
    <div
      className={`relative bg-[linear-gradient(135deg,rgba(30,50,80,0.6)_0%,rgba(20,40,70,0.5)_100%)] border-2 border-white/20 rounded-xl p-3 cursor-pointer transition-all duration-300 ease-[ease] hover:-translate-y-0.5 hover:shadow-[0_8px_25px_rgba(0,0,0,0.4),0_0_20px_rgba(100,150,255,0.3)] hover:border-[rgba(100,150,255,0.5)] ${isSelected ? "border-[rgba(150,255,150,0.8)] shadow-[0_8px_25px_rgba(0,0,0,0.4),0_0_30px_rgba(150,255,150,0.4)] bg-[linear-gradient(135deg,rgba(30,60,30,0.6)_0%,rgba(20,50,20,0.5)_100%)]" : ""}`}
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

      {/* Corporation name */}
      <div className="text-center mb-3">
        <h3 className="text-lg font-bold text-white m-0 [text-shadow:0_2px_4px_rgba(0,0,0,0.8)]">
          {corporation.name}
        </h3>
      </div>

      {/* Starting resources and production - compact, no headers */}
      {(corporation.startingProduction || corporation.startingResources) && (
        <div className="flex flex-wrap gap-2 justify-center mb-3 pb-3 border-b border-white/20">
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
                  <ProductionDisplay
                    key={type}
                    amount={amount}
                    resourceType={type}
                    size="medium"
                  />
                ) : null,
            )}
        </div>
      )}

      {/* Behaviors - compact, no headers */}
      {corporation.behaviors && corporation.behaviors.length > 0 && (
        <div className="mb-3 pb-3 border-b border-white/20">
          <div className="flex flex-col gap-2">
            {corporation.behaviors.map((behavior, index) => {
              // Skip the first auto behavior without conditions (starting bonuses shown above)
              const hasCondition = behavior.triggers?.some(
                (t) => t.condition !== undefined,
              );
              const isAuto = behavior.triggers?.some((t) => t.type === "auto");

              if (isAuto && !hasCondition && index === 0) {
                return null;
              }

              return (
                <div key={index} className="flex justify-center">
                  {renderBehavior(behavior)}
                </div>
              );
            })}
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
    </div>
  );
};

export default CorporationCard;
