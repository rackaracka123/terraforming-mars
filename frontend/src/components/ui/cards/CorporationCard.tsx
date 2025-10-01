import React from "react";

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
  const renderResourceIcon = (
    type: string,
    amount: number,
    isProduction: boolean = false,
  ) => {
    const iconMap: { [key: string]: string } = {
      credits: "/assets/resources/megacredit.png",
      steel: "/assets/resources/steel.png",
      titanium: "/assets/resources/titanium.png",
      plants: "/assets/resources/plant.png",
      energy: "/assets/resources/energy.png",
      heat: "/assets/resources/heat.png",
    };

    const icon = iconMap[type];
    if (!icon) return null;

    // For credits, use larger custom display
    if (type === "credits") {
      return (
        <div className="flex items-center gap-2.5 bg-white/15 px-3 py-2 rounded-lg text-sm text-white">
          {isProduction && (
            <div className="inline-flex items-center gap-2.5">
              <div className="relative inline-flex items-center justify-center">
                <img
                  src="/assets/misc/production.png"
                  alt="Production"
                  className="w-8 h-8"
                />
                <img
                  src={icon}
                  alt={type}
                  className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-5 h-5"
                />
              </div>
              <span className="font-bold text-base [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)]">
                {amount}
              </span>
            </div>
          )}
          {!isProduction && (
            <div className="relative inline-flex items-center justify-center">
              <img src={icon} alt={type} className="w-8 h-8" />
              <span className="font-bold text-base [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)]">
                {amount}
              </span>
            </div>
          )}
        </div>
      );
    }

    return (
      <div className="flex items-center gap-2.5 bg-white/15 px-3 py-2 rounded-lg text-sm text-white">
        {isProduction && (
          <div className="inline-flex items-center gap-2.5">
            <div className="relative inline-flex items-center justify-center">
              <img
                src="/assets/misc/production.png"
                alt="Production"
                className="w-8 h-8"
              />
              <img
                src={icon}
                alt={type}
                className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-5 h-5"
              />
            </div>
            <span className="font-bold text-base [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)]">
              {amount}
            </span>
          </div>
        )}
        {!isProduction && <img src={icon} alt={type} className="w-8 h-8" />}
        {!isProduction && (
          <span className="font-bold text-base [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)]">
            {amount}
          </span>
        )}
      </div>
    );
  };

  return (
    <div
      className={`relative bg-[linear-gradient(135deg,rgba(30,50,80,0.6)_0%,rgba(20,40,70,0.5)_100%)] border-2 border-white/20 rounded-xl p-5 cursor-pointer transition-all duration-300 ease-[ease] hover:-translate-y-0.5 hover:shadow-[0_8px_25px_rgba(0,0,0,0.4),0_0_20px_rgba(100,150,255,0.3)] hover:border-[rgba(100,150,255,0.5)] ${isSelected ? "border-[rgba(150,255,150,0.8)] shadow-[0_8px_25px_rgba(0,0,0,0.4),0_0_30px_rgba(150,255,150,0.4)] bg-[linear-gradient(135deg,rgba(30,60,30,0.6)_0%,rgba(20,50,20,0.5)_100%)]" : ""}`}
      onClick={() => onSelect(corporation.id)}
    >
      <div className="flex items-center mb-[15px] gap-[15px]">
        {corporation.logoPath && (
          <img
            src={corporation.logoPath}
            alt={corporation.name}
            className="w-[60px] h-[60px] rounded-lg object-cover"
          />
        )}
        <div className="flex-1">
          <h3 className="text-xl font-bold text-white m-0 mb-2 [text-shadow:0_1px_3px_rgba(0,0,0,0.8)]">
            {corporation.name}
          </h3>
          <div className="flex items-center justify-center bg-[rgba(241,196,15,0.2)] py-3 px-4 rounded-xl">
            <div className="relative inline-flex items-center justify-center">
              <img
                src="/assets/resources/megacredit.png"
                alt="Megacredits"
                className="w-14 h-14"
              />
              <span className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 text-black font-bold text-lg font-[Arial,sans-serif] [text-shadow:0.5px_0.5px_1px_rgba(255,255,255,0.8)] leading-none">
                {corporation.startingMegaCredits}
              </span>
            </div>
          </div>
        </div>
      </div>

      <div className="text-sm text-white/90 leading-[1.5] mb-[15px]">
        {corporation.description}
      </div>

      {(corporation.startingProduction || corporation.startingResources) && (
        <div className="mt-[15px] pt-[15px] border-t border-white/10">
          {corporation.startingProduction && (
            <div>
              <h4 className="text-xs text-white/80 m-0 mb-2 uppercase tracking-[0.5px]">
                Starting Production:
              </h4>
              <div className="flex flex-wrap gap-2">
                {Object.entries(corporation.startingProduction).map(
                  ([type, amount]) =>
                    amount > 0 ? renderResourceIcon(type, amount, true) : null,
                )}
              </div>
            </div>
          )}

          {corporation.startingResources && (
            <div>
              <h4 className="text-xs text-white/80 m-0 mb-2 uppercase tracking-[0.5px]">
                Starting Resources:
              </h4>
              <div className="flex flex-wrap gap-2">
                {Object.entries(corporation.startingResources).map(
                  ([type, amount]) =>
                    amount > 0 ? renderResourceIcon(type, amount, false) : null,
                )}
              </div>
            </div>
          )}
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
