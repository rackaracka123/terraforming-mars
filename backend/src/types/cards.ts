// Core game type definitions for Terraforming Mars

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


// Player resource and production tracking
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

// Game state interfaces
export interface GlobalParameters {
  temperature: number; // -30 to +8
  oxygen: number; // 0 to 14
  oceans: number; // 0 to 9
}

export interface GameState {
  id: string;
  players: Player[];
  currentPlayer: string;
  generation: number;
  phase: 'research' | 'action' | 'production';
  globalParameters: GlobalParameters;
  milestones: any[]; // TODO: Define milestone system
  awards: any[]; // TODO: Define award system
}

export interface Player {
  id: string;
  name: string;
  resources: PlayerResources;
  production: PlayerProduction;
  terraformRating: number;
  victoryPoints: number;
}