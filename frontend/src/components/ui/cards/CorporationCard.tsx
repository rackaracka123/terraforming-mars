import React from "react";
import MegaCreditIcon from "../display/MegaCreditIcon.tsx";
import ProductionDisplay from "../display/ProductionDisplay.tsx";

import { CardBehaviorDto } from "../../../types/generated/api-types.ts";

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
      credits: "/assets/resources/megacredit.png",
      steel: "/assets/resources/steel.png",
      titanium: "/assets/resources/titanium.png",
      plants: "/assets/resources/plant.png",
      energy: "/assets/resources/power.png",
      heat: "/assets/resources/heat.png",
    };

    const icon = iconMap[type];
    if (!icon) return null;

    // Use MegaCreditIcon for credits
    if (type === "credits") {
      return <MegaCreditIcon value={amount} size="large" />;
    }

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

  const renderBehaviorIcons = (behavior: CardBehaviorDto) => {
    const iconMap: { [key: string]: string } = {
      credits: "/assets/resources/megacredit.png",
      steel: "/assets/resources/steel.png",
      titanium: "/assets/resources/titanium.png",
      plants: "/assets/resources/plant.png",
      energy: "/assets/resources/power.png",
      heat: "/assets/resources/heat.png",
    };

    if (!behavior.outputs || behavior.outputs.length === 0) {
      return null;
    }

    return (
      <div className="flex flex-wrap gap-3 items-center">
        {behavior.outputs.map((output, idx) => {
          const resourceType = output.type.replace("-production", "");
          const isProduction = output.type.includes("-production");

          if (isProduction) {
            return (
              <ProductionDisplay
                key={idx}
                amount={output.amount}
                resourceType={resourceType}
                size="large"
              />
            );
          } else if (iconMap[resourceType]) {
            return (
              <div key={idx} className="flex flex-col items-center gap-1">
                <img
                  src={iconMap[resourceType]}
                  alt={resourceType}
                  className="w-10 h-10 [filter:drop-shadow(0_2px_4px_rgba(0,0,0,0.6))]"
                />
                <span className="text-white font-bold text-sm [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)]">
                  {output.amount}
                </span>
              </div>
            );
          }
          return null;
        })}
      </div>
    );
  };

  const renderEffectWithIcons = (effect: string) => {
    const text = effect;

    // Icon patterns to match
    const patterns = [
      { regex: /(\d+)\s*(M€|megacredits?|credits?)/gi, type: "credits" },
      { regex: /(\d+)\s*steel/gi, type: "steel" },
      { regex: /(\d+)\s*titanium/gi, type: "titanium" },
      { regex: /(\d+)\s*(plants?)/gi, type: "plants" },
      { regex: /(\d+)\s*(energy|power)/gi, type: "energy" },
      { regex: /(\d+)\s*heat/gi, type: "heat" },
    ];

    const iconMap: { [key: string]: string } = {
      credits: "/assets/resources/megacredit.png",
      steel: "/assets/resources/steel.png",
      titanium: "/assets/resources/titanium.png",
      plants: "/assets/resources/plant.png",
      energy: "/assets/resources/power.png",
      heat: "/assets/resources/heat.png",
    };

    // Check if effect contains any resource patterns
    let hasIcons = false;
    for (const pattern of patterns) {
      if (pattern.regex.test(text)) {
        hasIcons = true;
        break;
      }
    }

    if (!hasIcons) {
      // No icons found, return plain text
      return <div className="text-xs text-white/80 mt-1 italic">{effect}</div>;
    }

    // Extract resources with amounts
    const resources: { type: string; amount: number }[] = [];
    for (const pattern of patterns) {
      const matches = text.matchAll(pattern.regex);
      for (const match of matches) {
        resources.push({
          type: pattern.type,
          amount: parseInt(match[1]),
        });
      }
    }

    return (
      <div className="flex flex-col gap-2">
        <div className="flex flex-wrap gap-3 items-center">
          {resources.map((res, idx) => (
            <div key={idx} className="flex flex-col items-center gap-1">
              <img
                src={iconMap[res.type]}
                alt={res.type}
                className="w-10 h-10 [filter:drop-shadow(0_2px_4px_rgba(0,0,0,0.6))]"
              />
              <span className="text-white font-bold text-sm [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)]">
                {res.amount}
              </span>
            </div>
          ))}
        </div>
        <div className="text-xs text-white/80 italic">{effect}</div>
      </div>
    );
  };

  return (
    <div
      className={`relative bg-[linear-gradient(135deg,rgba(30,50,80,0.6)_0%,rgba(20,40,70,0.5)_100%)] border-2 border-white/20 rounded-xl p-5 cursor-pointer transition-all duration-300 ease-[ease] hover:-translate-y-0.5 hover:shadow-[0_8px_25px_rgba(0,0,0,0.4),0_0_20px_rgba(100,150,255,0.3)] hover:border-[rgba(100,150,255,0.5)] ${isSelected ? "border-[rgba(150,255,150,0.8)] shadow-[0_8px_25px_rgba(0,0,0,0.4),0_0_30px_rgba(150,255,150,0.4)] bg-[linear-gradient(135deg,rgba(30,60,30,0.6)_0%,rgba(20,50,20,0.5)_100%)]" : ""}`}
      onClick={() => onSelect(corporation.id)}
    >
      <div className="flex items-center mb-5 gap-4">
        {corporation.logoPath && (
          <img
            src={corporation.logoPath}
            alt={corporation.name}
            className="w-16 h-16 rounded-lg object-cover [filter:drop-shadow(0_2px_6px_rgba(0,0,0,0.6))]"
          />
        )}
        <div className="flex-1">
          <h3 className="text-2xl font-bold text-white m-0 mb-3 [text-shadow:0_2px_4px_rgba(0,0,0,0.8)]">
            {corporation.name}
          </h3>
          <div className="flex items-center justify-center bg-[rgba(241,196,15,0.25)] py-4 px-5 rounded-xl">
            <MegaCreditIcon
              value={corporation.startingMegaCredits}
              size="large"
            />
          </div>
        </div>
      </div>

      <div className="text-sm text-white/90 leading-[1.5] mb-[15px]">
        {corporation.description}
      </div>

      {corporation.behaviors && corporation.behaviors.length > 0 && (
        <div className="mt-4 mb-4">
          <h4 className="text-sm font-semibold text-white/90 m-0 mb-2 uppercase tracking-wider [text-shadow:0_1px_2px_rgba(0,0,0,0.6)]">
            Special Effects
          </h4>
          <div className="bg-[rgba(100,150,255,0.15)] border-l-4 border-[rgba(100,150,255,0.5)] p-3 rounded">
            {corporation.behaviors.map((behavior, index) => {
              // Skip the first auto behavior without conditions (starting bonuses are shown separately)
              const hasCondition = behavior.triggers?.some(
                (t) => t.condition !== undefined,
              );
              const isAuto = behavior.triggers?.some((t) => t.type === "auto");

              if (isAuto && !hasCondition && index === 0) {
                return null; // Skip starting bonuses
              }

              return (
                <div key={index} className="mb-3 last:mb-0">
                  <div className="flex flex-col gap-2">
                    {renderBehaviorIcons(behavior)}
                    <div className="text-xs text-white/80 italic">
                      {corporation.description}
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      )}

      {(corporation.startingProduction || corporation.startingResources) && (
        <div className="mt-5 pt-5 border-t border-white/20">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            {corporation.startingProduction && (
              <div>
                <h4 className="text-sm font-semibold text-white/90 m-0 mb-3 uppercase tracking-wider [text-shadow:0_1px_2px_rgba(0,0,0,0.6)]">
                  Starting Production
                </h4>
                <div className="flex flex-wrap gap-3">
                  {Object.entries(corporation.startingProduction).map(
                    ([type, amount]) =>
                      amount > 0 ? (
                        <ProductionDisplay
                          key={type}
                          amount={amount}
                          resourceType={type}
                          size="large"
                        />
                      ) : null,
                  )}
                </div>
              </div>
            )}

            {corporation.startingResources && (
              <div>
                <h4 className="text-sm font-semibold text-white/90 m-0 mb-3 uppercase tracking-wider [text-shadow:0_1px_2px_rgba(0,0,0,0.6)]">
                  Starting Resources
                </h4>
                <div className="flex flex-wrap gap-3">
                  {Object.entries(corporation.startingResources).map(
                    ([type, amount]) =>
                      amount > 0 ? (
                        <div key={type}>{renderResource(type, amount)}</div>
                      ) : null,
                  )}
                </div>
              </div>
            )}
          </div>
        </div>
      )}

      {corporation.expansion && (
        <div className="absolute top-2.5 right-2.5 bg-[rgba(100,150,255,0.3)] text-white/80 py-1 px-2 rounded text-[10px] uppercase tracking-[0.5px]">
          {corporation.expansion}
        </div>
      )}
    </div>
  );
};

export default CorporationCard;
