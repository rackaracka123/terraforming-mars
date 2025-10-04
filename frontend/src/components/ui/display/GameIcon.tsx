import React from "react";
import {
  ResourceType,
  ResourceTypeCredits,
  ResourceTypeSteel,
  ResourceTypeTitanium,
  ResourceTypePlants,
  ResourceTypeEnergy,
  ResourceTypeHeat,
  ResourceTypeFloaters,
  ResourceTypeMicrobes,
  ResourceTypeAnimals,
  ResourceTypeScience,
  ResourceTypeAsteroid,
  ResourceTypeDisease,
  ResourceTypeCardDraw,
  ResourceTypeCardTake,
  ResourceTypeCardPeek,
  ResourceTypeCityPlacement,
  ResourceTypeOceanPlacement,
  ResourceTypeGreeneryPlacement,
  ResourceTypeGreeneryTile,
  ResourceTypeCityTile,
  ResourceTypeOceanTile,
  ResourceTypeColonyTile,
  ResourceTypeTemperature,
  ResourceTypeOxygen,
  ResourceTypeVenus,
  ResourceTypeTR,
  ResourceTypeFighters,
  ResourceTypeCamps,
  ResourceTypePreservation,
  ResourceTypeData,
  ResourceTypeSpecialized,
  ResourceTypeDelegate,
  ResourceTypeInfluence,
  ResourceTypeSpecialTile,
  ResourceTypeOceans,
} from "@/types/generated/api-types.ts";

interface GameIconProps {
  resourceType: ResourceType;
  amount?: number;
  isAttack?: boolean;
  size?: "small" | "medium" | "large";
  className?: string;
}

const GameIcon: React.FC<GameIconProps> = ({
  resourceType,
  amount,
  isAttack = false,
  size = "medium",
  className = "",
}) => {
  const isProduction = resourceType.endsWith("-production");
  const baseType = isProduction
    ? resourceType.replace("-production", "")
    : resourceType;
  const isCredits = baseType === ResourceTypeCredits;

  const getIconUrl = (type: string): string | null => {
    const iconMap: Record<string, string> = {
      [ResourceTypeCredits]: "/assets/resources/megacredit.png",
      [ResourceTypeSteel]: "/assets/resources/steel.png",
      [ResourceTypeTitanium]: "/assets/resources/titanium.png",
      [ResourceTypePlants]: "/assets/resources/plant.png",
      [ResourceTypeEnergy]: "/assets/resources/power.png",
      [ResourceTypeHeat]: "/assets/resources/heat.png",
      [ResourceTypeFloaters]: "/assets/resources/floater.png",
      [ResourceTypeMicrobes]: "/assets/resources/microbe.png",
      [ResourceTypeAnimals]: "/assets/resources/animal.png",
      [ResourceTypeScience]: "/assets/resources/science.png",
      [ResourceTypeAsteroid]: "/assets/resources/asteroid.png",
      [ResourceTypeDisease]: "/assets/resources/disease.png",
      [ResourceTypeCardDraw]: "/assets/resources/card.png",
      [ResourceTypeCardTake]: "/assets/resources/card.png",
      [ResourceTypeCardPeek]: "/assets/resources/card.png",
      [ResourceTypeCityPlacement]: "/assets/tiles/city.png",
      [ResourceTypeOceanPlacement]: "/assets/tiles/ocean.png",
      [ResourceTypeGreeneryPlacement]: "/assets/tiles/greenery.png",
      [ResourceTypeGreeneryTile]: "/assets/tiles/greenery.png",
      [ResourceTypeCityTile]: "/assets/tiles/city.png",
      [ResourceTypeOceanTile]: "/assets/tiles/ocean.png",
      [ResourceTypeColonyTile]: "/assets/tiles/colony.png",
      [ResourceTypeTemperature]: "/assets/global-parameters/temperature.png",
      [ResourceTypeOxygen]: "/assets/global-parameters/oxygen.png",
      [ResourceTypeVenus]: "/assets/global-parameters/venus.png",
      [ResourceTypeTR]: "/assets/resources/tr.png",
      [ResourceTypeFighters]: "/assets/resources/fighter.png",
      [ResourceTypeCamps]: "/assets/resources/camp.png",
      [ResourceTypePreservation]: "/assets/resources/preservation.png",
      [ResourceTypeData]: "/assets/resources/data.png",
      [ResourceTypeSpecialized]: "/assets/resources/specialized-robot.png",
      [ResourceTypeDelegate]: "/assets/resources/director.png",
      [ResourceTypeInfluence]: "/assets/misc/influence.png",
      [ResourceTypeSpecialTile]: "/assets/tiles/special.png",
      [ResourceTypeOceans]: "/assets/tiles/ocean.png",
    };

    return iconMap[type] || null;
  };

  const iconUrl = getIconUrl(baseType);

  if (!iconUrl) {
    return (
      <span className="text-xs font-semibold text-white [text-shadow:1px_1px_2px_rgba(0,0,0,0.6)]">
        {resourceType}
      </span>
    );
  }

  const sizeMap = {
    small: { icon: 24, fontSize: "10px" },
    medium: { icon: 32, fontSize: "12px" },
    large: { icon: 40, fontSize: "14px" },
  };

  const dimensions = sizeMap[size];

  const attackGlow = isAttack
    ? "[filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))_drop-shadow(0_0_1px_rgba(244,67,54,0.9))_drop-shadow(0_0_2px_rgba(244,67,54,0.7))] animate-[attackPulse_2s_ease-in-out_infinite]"
    : "";

  const renderCoreIcon = () => {
    if (isCredits && amount !== undefined) {
      return (
        <div
          className={`relative inline-flex items-center justify-center ${attackGlow}`}
          style={{
            width: `${dimensions.icon}px`,
            height: `${dimensions.icon}px`,
          }}
        >
          <img
            src={iconUrl}
            alt={baseType}
            className="w-full h-full object-contain [filter:drop-shadow(0_2px_4px_rgba(0,0,0,0.6))]"
          />
          <span
            className="absolute top-0 left-0 right-0 bottom-0 text-black font-black font-[Prototype,Arial_Black,Arial,sans-serif] flex items-center justify-center text-center leading-none [text-shadow:0_0_2px_rgba(255,255,255,0.3)] tracking-[0.5px] [-webkit-font-smoothing:antialiased] [-moz-osx-font-smoothing:grayscale] [text-rendering:optimizeLegibility]"
            style={{ fontSize: dimensions.fontSize }}
          >
            {amount}
          </span>
        </div>
      );
    }

    return (
      <div
        className={`relative inline-flex items-center justify-center ${attackGlow}`}
        style={{
          width: `${dimensions.icon}px`,
          height: `${dimensions.icon}px`,
        }}
      >
        <img
          src={iconUrl}
          alt={baseType}
          className="w-full h-full object-contain [filter:drop-shadow(0_1px_2px_rgba(0,0,0,0.5))]"
        />
        {amount !== undefined && amount > 1 && !isCredits && (
          <span
            className="absolute bottom-0 right-0 text-white font-bold [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)] bg-black/50 rounded-full px-1 min-w-[16px] text-center leading-none"
            style={{ fontSize: dimensions.fontSize }}
          >
            {amount}
          </span>
        )}
      </div>
    );
  };

  if (isProduction) {
    return (
      <div
        className={`inline-flex items-center justify-center bg-[linear-gradient(135deg,rgba(160,110,60,0.4)_0%,rgba(139,89,42,0.35)_100%)] border border-[rgba(160,110,60,0.5)] rounded px-1.5 py-[3px] shadow-[0_1px_3px_rgba(0,0,0,0.2)] ${className}`}
      >
        {renderCoreIcon()}
      </div>
    );
  }

  return renderCoreIcon();
};

export default GameIcon;
