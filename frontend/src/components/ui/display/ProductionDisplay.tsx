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
      className={`production-display ${className}`}
      style={{
        display: "inline-flex",
        alignItems: "center",
        gap: `${dimensions.gap}px`,
      }}
    >
      <div
        style={{
          position: "relative",
          display: "inline-flex",
          alignItems: "center",
          justifyContent: "center",
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
            style={{
              position: "absolute",
              top: "50%",
              left: "50%",
              transform: "translate(-50%, -50%)",
              width: `${dimensions.resourceIcon}px`,
              height: `${dimensions.resourceIcon}px`,
            }}
          />
        )}
      </div>
      <span
        style={{
          color: "#ffffff",
          fontWeight: "bold",
          fontSize: dimensions.fontSize,
          textShadow: "1px 1px 2px rgba(0, 0, 0, 0.8)",
          fontFamily: "Arial, sans-serif",
          lineHeight: "1",
        }}
      >
        {amount}
      </span>
    </div>
  );
};

export default ProductionDisplay;
