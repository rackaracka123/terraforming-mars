export type ResourceType =
  | "credits"
  | "steel"
  | "titanium"
  | "plants"
  | "energy"
  | "heat";

export const RESOURCE_COLORS: Record<ResourceType, string> = {
  credits: "#f1c40f", // Gold
  steel: "#d2691e", // Brown/orangy
  titanium: "#95a5a6", // Grey
  plants: "#27ae60", // Green
  energy: "#9b59b6", // Purple
  heat: "#ff4500", // Red/orange
};

export const RESOURCE_ICONS: Record<ResourceType, string> = {
  credits: "/assets/resources/megacredit.png",
  steel: "/assets/resources/steel.png",
  titanium: "/assets/resources/titanium.png",
  plants: "/assets/resources/plant.png",
  energy: "/assets/resources/power.png",
  heat: "/assets/resources/heat.png",
};

export const RESOURCE_NAMES: Record<ResourceType, string> = {
  credits: "Credits",
  steel: "Steel",
  titanium: "Titanium",
  plants: "Plants",
  energy: "Energy",
  heat: "Heat",
};

export const getResourceColor = (resourceType: ResourceType): string => {
  return RESOURCE_COLORS[resourceType];
};

export const getResourceIcon = (resourceType: ResourceType): string => {
  return RESOURCE_ICONS[resourceType];
};

export const getResourceName = (resourceType: ResourceType): string => {
  return RESOURCE_NAMES[resourceType];
};
