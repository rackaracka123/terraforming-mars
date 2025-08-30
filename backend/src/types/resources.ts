// Resource and global parameter type definitions for Terraforming Mars

export enum ResourceType {
  CREDITS = 'credits',
  STEEL = 'steel',
  TITANIUM = 'titanium',
  PLANTS = 'plants',
  ENERGY = 'energy',
  HEAT = 'heat'
}

export enum GlobalParameter {
  TEMPERATURE = 'temperature',
  OXYGEN = 'oxygen',
  OCEAN = 'oceans'
}

// Resource interfaces
export interface PlayerResources {
  credits: number;
  steel: number;
  titanium: number;
  plants: number;
  energy: number;
  heat: number;
}

export interface PlayerProduction {
  credits: number;
  steel: number;
  titanium: number;
  plants: number;
  energy: number;
  heat: number;
}

export interface GlobalParameters {
  temperature: number; // -30 to +8
  oxygen: number; // 0 to 14
  oceans: number; // 0 to 9
}

// Resource cost interfaces
export interface ResourceCost {
  credits?: number;
  steel?: number;
  titanium?: number;
  plants?: number;
  energy?: number;
  heat?: number;
}

export interface ResourceGain {
  credits?: number;
  steel?: number;
  titanium?: number;
  plants?: number;
  energy?: number;
  heat?: number;
}

export interface ProductionChange {
  credits?: number;
  steel?: number;
  titanium?: number;
  plants?: number;
  energy?: number;
  heat?: number;
}