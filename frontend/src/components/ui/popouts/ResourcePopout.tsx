import React from "react";
import {
  ResourcesDto,
  ProductionDto,
} from "../../../types/generated/api-types.ts";
import ProductionDisplay from "../display/ProductionDisplay.tsx";

interface ResourcePopoutProps {
  resources: ResourcesDto;
  production: ProductionDto;
  playerName: string;
}

interface ResourceDisplayItemProps {
  name: string;
  iconPath: string;
  currentAmount: number;
  productionAmount: number;
  resourceType: string;
}

const ResourceDisplayItem: React.FC<ResourceDisplayItemProps> = ({
  name,
  iconPath,
  currentAmount,
  productionAmount,
  resourceType,
}) => {
  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        gap: "8px",
        padding: "6px 8px",
        backgroundColor: "rgba(255, 255, 255, 0.08)",
        borderRadius: "6px",
        border: "1px solid rgba(255, 255, 255, 0.1)",
        minWidth: "120px",
      }}
    >
      {/* Resource Icon */}
      <img
        src={iconPath}
        alt={name}
        style={{
          width: "24px",
          height: "24px",
          flexShrink: 0,
        }}
      />

      {/* Current Amount */}
      <div
        style={{
          display: "flex",
          alignItems: "center",
          gap: "4px",
          flex: 1,
        }}
      >
        <span
          style={{
            color: "#ffffff",
            fontWeight: "bold",
            fontSize: "14px",
            textShadow: "1px 1px 2px rgba(0, 0, 0, 0.8)",
            fontFamily: "Arial, sans-serif",
            minWidth: "20px",
            textAlign: "center",
          }}
        >
          {currentAmount}
        </span>

        {/* Production Display */}
        {productionAmount !== 0 && (
          <div style={{ display: "flex", alignItems: "center" }}>
            <ProductionDisplay
              amount={productionAmount}
              resourceType={resourceType}
              size="small"
            />
          </div>
        )}
      </div>
    </div>
  );
};

const ResourcePopout: React.FC<ResourcePopoutProps> = ({
  resources,
  production,
  playerName,
}) => {
  const resourceData = [
    {
      name: "Credits",
      iconPath: "/assets/resources/megacredit.png",
      currentAmount: resources.credits,
      productionAmount: production.credits,
      resourceType: "credits",
    },
    {
      name: "Steel",
      iconPath: "/assets/resources/steel.png",
      currentAmount: resources.steel,
      productionAmount: production.steel,
      resourceType: "steel",
    },
    {
      name: "Titanium",
      iconPath: "/assets/resources/titanium.png",
      currentAmount: resources.titanium,
      productionAmount: production.titanium,
      resourceType: "titanium",
    },
    {
      name: "Plants",
      iconPath: "/assets/resources/plant.png",
      currentAmount: resources.plants,
      productionAmount: production.plants,
      resourceType: "plants",
    },
    {
      name: "Energy",
      iconPath: "/assets/resources/power.png",
      currentAmount: resources.energy,
      productionAmount: production.energy,
      resourceType: "energy",
    },
    {
      name: "Heat",
      iconPath: "/assets/resources/heat.png",
      currentAmount: resources.heat,
      productionAmount: production.heat,
      resourceType: "heat",
    },
  ];

  return (
    <div
      className="resource-popout"
      style={{
        position: "absolute",
        zIndex: 1000,
        background:
          "linear-gradient(180deg, rgba(0, 20, 40, 0.95) 0%, rgba(0, 10, 30, 0.95) 50%, rgba(0, 5, 20, 0.95) 100%)",
        backdropFilter: "blur(10px)",
        border: "1px solid rgba(100, 150, 200, 0.3)",
        borderRadius: "12px",
        padding: "12px",
        minWidth: "280px",
        maxWidth: "320px",
        boxShadow: "0 8px 32px rgba(0, 0, 0, 0.4)",
        animation: "slideInRight 300ms ease-out",
      }}
    >
      {/* Header */}
      <div
        style={{
          marginBottom: "12px",
          paddingBottom: "8px",
          borderBottom: "1px solid rgba(100, 150, 200, 0.2)",
        }}
      >
        <div
          style={{
            color: "#64c8ff",
            fontSize: "14px",
            fontWeight: "600",
            textAlign: "center",
            textShadow: "0 0 10px rgba(100, 200, 255, 0.5)",
            letterSpacing: "0.5px",
          }}
        >
          {playerName.toUpperCase()}'S RESOURCES
        </div>
      </div>

      {/* Resources Grid */}
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "1fr 1fr",
          gap: "8px",
        }}
      >
        {resourceData.map((resource) => (
          <ResourceDisplayItem
            key={resource.name}
            name={resource.name}
            iconPath={resource.iconPath}
            currentAmount={resource.currentAmount}
            productionAmount={resource.productionAmount}
            resourceType={resource.resourceType}
          />
        ))}
      </div>
    </div>
  );
};

export default ResourcePopout;
