// Frontend card type definitions for Terraforming Mars

export enum CardType {
  AUTOMATED = "automated",
  ACTIVE = "active",
  EVENT = "event",
  CORPORATION = "corporation",
  PRELUDE = "prelude",
}

export enum CardTag {
  BUILDING = "building",
  SPACE = "space",
  POWER = "power",
  SCIENCE = "science",
  MICROBE = "microbe",
  ANIMAL = "animal",
  PLANT = "plant",
  EARTH = "earth",
  JOVIAN = "jovian",
  CITY = "city",
  VENUS = "venus",
  MARS = "mars",
  MOON = "moon",
  WILD = "wild",
  EVENT = "event",
  CLONE = "clone",
}

// Standard Projects enum matching backend definition
export enum StandardProject {
  SELL_PATENTS = "SELL_PATENTS",
  POWER_PLANT = "POWER_PLANT",
  ASTEROID = "ASTEROID",
  AQUIFER = "AQUIFER",
  GREENERY = "GREENERY",
  CITY = "CITY",
}

// Standard project metadata interface
export interface StandardProjectMetadata {
  id: StandardProject;
  name: string;
  cost: number;
  description: string;
  requiresTilePlacement: boolean;
  grantsTR: boolean;
  effects: {
    production?: { type: string; amount: number }[];
    immediate?: { type: string; amount: number }[];
    globalParameters?: { type: string; amount: number }[];
  };
  icon?: string;
}

// Standard project costs (matching backend)
export const STANDARD_PROJECT_COSTS: Record<StandardProject, number> = {
  [StandardProject.SELL_PATENTS]: 0,
  [StandardProject.POWER_PLANT]: 11,
  [StandardProject.ASTEROID]: 14,
  [StandardProject.AQUIFER]: 18,
  [StandardProject.GREENERY]: 23,
  [StandardProject.CITY]: 25,
};

// Standard project metadata
export const STANDARD_PROJECTS: Record<
  StandardProject,
  StandardProjectMetadata
> = {
  [StandardProject.SELL_PATENTS]: {
    id: StandardProject.SELL_PATENTS,
    name: "Sell Patents",
    cost: 0,
    description: "Discard any number of cards from hand for 1 M€ each",
    requiresTilePlacement: false,
    grantsTR: false,
    effects: {
      immediate: [{ type: "credits", amount: 1 }],
    },
    icon: "/assets/misc/card.png",
  },
  [StandardProject.POWER_PLANT]: {
    id: StandardProject.POWER_PLANT,
    name: "Power Plant",
    cost: 11,
    description: "Spend 11 M€ to increase energy production by 1",
    requiresTilePlacement: false,
    grantsTR: false,
    effects: {
      production: [{ type: "energy", amount: 1 }],
    },
    icon: "/assets/resources/power.png",
  },
  [StandardProject.ASTEROID]: {
    id: StandardProject.ASTEROID,
    name: "Asteroid",
    cost: 14,
    description: "Spend 14 M€ to raise temperature 1 step and gain 1 TR",
    requiresTilePlacement: false,
    grantsTR: true,
    effects: {
      globalParameters: [{ type: "temperature", amount: 2 }],
      immediate: [{ type: "tr", amount: 1 }],
    },
    icon: "/assets/misc/asteroid.png",
  },
  [StandardProject.AQUIFER]: {
    id: StandardProject.AQUIFER,
    name: "Aquifer",
    cost: 18,
    description: "Spend 18 M€ to place an ocean tile and gain 1 TR",
    requiresTilePlacement: true,
    grantsTR: true,
    effects: {
      globalParameters: [{ type: "oceans", amount: 1 }],
      immediate: [{ type: "tr", amount: 1 }],
    },
    icon: "/assets/resources/ocean.png",
  },
  [StandardProject.GREENERY]: {
    id: StandardProject.GREENERY,
    name: "Greenery",
    cost: 23,
    description:
      "Spend 23 M€ to place a greenery tile, raise oxygen 1%, and gain 1 TR",
    requiresTilePlacement: true,
    grantsTR: true,
    effects: {
      globalParameters: [{ type: "oxygen", amount: 1 }],
      immediate: [{ type: "tr", amount: 1 }],
    },
    icon: "/assets/resources/plant.png",
  },
  [StandardProject.CITY]: {
    id: StandardProject.CITY,
    name: "City",
    cost: 25,
    description:
      "Spend 25 M€ to place a city tile and increase M€ production by 1",
    requiresTilePlacement: true,
    grantsTR: false,
    effects: {
      production: [{ type: "credits", amount: 1 }],
    },
    icon: "/assets/misc/city.png",
  },
};
