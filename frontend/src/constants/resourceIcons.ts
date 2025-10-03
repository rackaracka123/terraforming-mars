/**
 * Resource and game element icon paths
 */

export const RESOURCE_ICONS = {
  credits: "/assets/resources/megacredit.png",
  steel: "/assets/resources/steel.png",
  titanium: "/assets/resources/titanium.png",
  plants: "/assets/resources/plants.png",
  energy: "/assets/resources/power.png",
  heat: "/assets/resources/heat.png",
  oxygen: "/assets/resources/oxygen.png",
  ocean: "/assets/resources/ocean.png",
  tr: "/assets/resources/tr.png",
} as const;

export const GLOBAL_PARAM_ICONS = {
  temperature: "/assets/resources/heat.png",
  oxygen: "/assets/resources/oxygen.png",
  oceans: "/assets/resources/ocean.png",
} as const;

export const MISC_ICONS = {
  production: "/assets/misc/production.png",
} as const;

export type ResourceType = keyof typeof RESOURCE_ICONS;
export type GlobalParamType = keyof typeof GLOBAL_PARAM_ICONS;
