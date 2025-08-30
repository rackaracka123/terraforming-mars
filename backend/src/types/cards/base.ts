// Base card type definitions for Terraforming Mars

import { CardRequirement } from '../requirements';

export enum CardType {
  AUTOMATED = 'automated',
  ACTIVE = 'active',
  EVENT = 'event',
  CORPORATION = 'corporation',
  PRELUDE = 'prelude'
}

export enum CardTag {
  BUILDING = 'building',
  SPACE = 'space',
  POWER = 'power',
  SCIENCE = 'science',
  MICROBE = 'microbe',
  ANIMAL = 'animal',
  PLANT = 'plant',
  EARTH = 'earth',
  JOVIAN = 'jovian',
  CITY = 'city',
  VENUS = 'venus',
  MARS = 'mars',
  MOON = 'moon',
  WILD = 'wild',
  EVENT = 'event',
  CLONE = 'clone'
}

export interface CardDefinition {
  id: string;
  name: string;
  cost: number;
  type: CardType;
  tags: CardTag[];
  victoryPoints?: number;
  requirements?: CardRequirement[];
  effects: any[]; // Use any[] to avoid circular import, will be typed as Effect[] when imported
  description: string;
  flavor?: string;
  expansion?: string;
  renderData?: RenderData;
}

export interface Card {
  id: string;
  definition: CardDefinition;
  resources?: number; // For cards that store resources
  playedBy?: string;
  isUsed?: boolean; // For active cards
}


// Rendering data for UI
export interface RenderData {
  cost?: number;
  description?: string;
  effects?: RenderEffect[];
}

export interface RenderEffect {
  type: string;
  amount?: number;
  asterix?: boolean;
}

