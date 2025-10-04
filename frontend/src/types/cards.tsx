// Frontend card type definitions for Terraforming Mars
import React from "react";
import {
  CardDto,
  ResourceTypeCityTile,
  ResourceTypeGreeneryTile,
  ResourceTypeOceanTile,
  ResourceTypeAsteroid,
  ResourceTypeEnergy,
  ResourceTypeCardDraw,
  ResourceTypeTemperature,
} from "./generated/api-types.ts";
import GameIcon from "../components/ui/display/GameIcon.tsx";

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

// Standard project with icon for UI display
export interface StandardProjectCard extends CardDto {
  icon?: React.ReactElement;
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

// Standard project cards using CardDto structure
export const STANDARD_PROJECTS: Record<StandardProject, StandardProjectCard> = {
  [StandardProject.SELL_PATENTS]: {
    id: StandardProject.SELL_PATENTS,
    name: "Sell Patents",
    type: "active",
    cost: 0,
    description: "Discard any number of cards from hand for 1 M€ each",
    tags: [],
    behaviors: [],
    icon: <GameIcon resourceType={ResourceTypeCardDraw} size="small" />,
  },
  [StandardProject.POWER_PLANT]: {
    id: StandardProject.POWER_PLANT,
    name: "Power Plant",
    type: "active",
    cost: 11,
    description:
      "Action: Spend 11 M€ to increase your energy production 1 step.",
    tags: ["power"],
    behaviors: [
      {
        triggers: [{ type: "manual" }],
        inputs: [
          {
            type: "credits",
            amount: 11,
            target: "self-player",
          },
        ],
        outputs: [
          {
            type: "energy-production",
            amount: 1,
            target: "self-player",
          },
        ],
      },
    ],
    icon: <GameIcon resourceType={ResourceTypeEnergy} size="small" />,
  },
  [StandardProject.ASTEROID]: {
    id: StandardProject.ASTEROID,
    name: "Asteroid",
    type: "active",
    cost: 14,
    description:
      "Action: Spend 14 M€ to raise temperature 1 step and gain 1 TR.",
    tags: ["space"],
    behaviors: [
      {
        triggers: [{ type: "manual" }],
        inputs: [
          {
            type: "credits",
            amount: 14,
            target: "self-player",
          },
        ],
        outputs: [
          {
            type: ResourceTypeTemperature,
            amount: 1,
            target: "none",
          },
        ],
      },
    ],
    icon: <GameIcon resourceType={ResourceTypeAsteroid} size="small" />,
  },
  [StandardProject.AQUIFER]: {
    id: StandardProject.AQUIFER,
    name: "Aquifer",
    type: "active",
    cost: 18,
    description: "Action: Spend 18 M€ to place an ocean tile and gain 1 TR.",
    tags: [],
    behaviors: [
      {
        triggers: [{ type: "manual" }],
        inputs: [
          {
            type: "credits",
            amount: 18,
            target: "self-player",
          },
        ],
        outputs: [
          {
            type: ResourceTypeOceanTile,
            amount: 1,
            target: "none",
          },
        ],
      },
    ],
    icon: <GameIcon resourceType={ResourceTypeOceanTile} size="small" />,
  },
  [StandardProject.GREENERY]: {
    id: StandardProject.GREENERY,
    name: "Greenery",
    type: "active",
    cost: 23,
    description:
      "Action: Spend 23 M€ to place a greenery tile, raise oxygen 1%, and gain 1 TR.",
    tags: [],
    behaviors: [
      {
        triggers: [{ type: "manual" }],
        inputs: [
          {
            type: "credits",
            amount: 23,
            target: "self-player",
          },
        ],
        outputs: [
          {
            type: ResourceTypeGreeneryTile,
            amount: 1,
            target: "none",
          },
        ],
      },
    ],
    icon: <GameIcon resourceType={ResourceTypeGreeneryTile} size="small" />,
  },
  [StandardProject.CITY]: {
    id: StandardProject.CITY,
    name: "City",
    type: "active",
    cost: 25,
    description:
      "Action: Spend 25 M€ to place a city tile and increase M€ production by 1.",
    tags: ["building", "city"],
    behaviors: [
      {
        triggers: [{ type: "manual" }],
        inputs: [
          {
            type: "credits",
            amount: 25,
            target: "self-player",
          },
        ],
        outputs: [
          {
            type: ResourceTypeCityTile,
            amount: 1,
            target: "none",
          },
          {
            type: "credits-production",
            amount: 1,
            target: "self-player",
          },
        ],
      },
    ],
    icon: <GameIcon resourceType={ResourceTypeCityTile} size="small" />,
  },
};
