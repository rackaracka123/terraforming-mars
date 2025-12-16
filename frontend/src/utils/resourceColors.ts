import { ResourceType } from "@/types/generated/api-types.ts";

export const RESOURCE_COLORS: Record<ResourceType, string> = {
  credit: "#f1c40f", // Gold
  steel: "#d2691e", // Brown/orangy
  titanium: "#95a5a6", // Grey
  plant: "#27ae60", // Green
  energy: "#9b59b6", // Purple
  heat: "#ff4500", // Red/orange
};

export const RESOURCE_NAMES: Record<ResourceType, string> = {
  credit: "Credits",
  steel: "Steel",
  titanium: "Titanium",
  plant: "Plants",
  energy: "Energy",
  heat: "Heat",
};

export const getResourceColor = (resourceType: ResourceType): string => {
  return RESOURCE_COLORS[resourceType];
};

export const getResourceName = (resourceType: ResourceType): string => {
  return RESOURCE_NAMES[resourceType];
};
