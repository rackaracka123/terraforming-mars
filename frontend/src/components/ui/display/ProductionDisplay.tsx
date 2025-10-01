import React from "react";

interface ProductionDisplayProps {
  amount: number;
  resourceType?: string;
  size?: "small" | "medium" | "large";
  className?: string;
}

const ProductionDisplay: React.FC<ProductionDisplayProps> = ({
  amount,
  resourceType,
  size = "medium",
  className = "",
}) => {
  const sizeMap = {
    small: { productionIcon: 18, resourceIcon: 12, fontSize: "10px", gap: 4 },
    medium: { productionIcon: 22, resourceIcon: 16, fontSize: "12px", gap: 6 },
    large: { productionIcon: 26, resourceIcon: 20, fontSize: "14px", gap: 8 },
  };

  const dimensions = sizeMap[size];

  const getResourceIcon = (type: string) => {
    const iconMap: { [key: string]: string } = {
      credits: "/assets/resources/megacredit.png",
      steel: "/assets/resources/steel.png",
      titanium: "/assets/resources/titanium.png",
      plants: "/assets/resources/plant.png",
      energy: "/assets/resources/power.png",
      heat: "/assets/resources/heat.png",
    };
    return iconMap[type] || "";
  };

  return (
    <div
      className={`inline-flex items-center ${className}`}
      style={{
        gap: `${dimensions.gap}px`,
      }}
    >
      <div
        className="relative inline-flex items-center justify-center"
        style={{
          width: `${dimensions.productionIcon}px`,
          height: `${dimensions.productionIcon}px`,
        }}
      >
        <img
          src="/assets/misc/production.png"
          alt="Production"
          style={{
            width: `${dimensions.productionIcon}px`,
            height: `${dimensions.productionIcon}px`,
          }}
        />
        {resourceType && (
          <img
            src={getResourceIcon(resourceType)}
            alt={resourceType}
            className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2"
            style={{
              width: `${dimensions.resourceIcon}px`,
              height: `${dimensions.resourceIcon}px`,
            }}
          />
        )}
      </div>
      <span
        className="text-white font-bold font-[Arial,sans-serif] leading-none [text-shadow:1px_1px_2px_rgba(0,0,0,0.8)]"
        style={{
          fontSize: dimensions.fontSize,
        }}
      >
        {amount}
      </span>
    </div>
  );
};

export default ProductionDisplay;
