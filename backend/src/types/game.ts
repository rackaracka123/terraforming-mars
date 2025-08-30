// Game state and flow type definitions for Terraforming Mars

import { Player } from './player';
import { GlobalParameters } from './resources';
import { Milestone } from './milestones';
import { Award } from './awards';

export interface GameState {
  id: string;
  players: Player[];
  currentPlayer: string;
  generation: number;
  phase: GamePhase;
  globalParameters: GlobalParameters;
  milestones: Milestone[];
  awards: Award[];
  firstPlayer: string;
  corporationDraft?: boolean;
  deck: string[]; // Card IDs
  discardPile: string[];
  soloMode: boolean;
  turn: number;
  draftDirection?: DraftDirection;
  gameSettings: GameSettings;
  currentActionCount?: number; // Current action number for the current player
  maxActionsPerTurn?: number; // Maximum actions per turn (usually 2)
}

export enum GamePhase {
  SETUP = 'setup',
  CORPORATION_SELECTION = 'corporation_selection',
  INITIAL_RESEARCH = 'initial_research',
  PRELUDE = 'prelude',
  RESEARCH = 'research',
  ACTION = 'action',
  PRODUCTION = 'production',
  DRAFT = 'draft',
  GAME_END = 'game_end'
}

export enum DraftDirection {
  CLOCKWISE = 'clockwise',
  COUNTER_CLOCKWISE = 'counter_clockwise'
}

export interface GameSettings {
  expansions: GameExpansion[];
  corporateEra: boolean;
  draftVariant: boolean;
  initialDraft: boolean;
  preludeExtension: boolean;
  venusNextExtension: boolean;
  coloniesExtension: boolean;
  turmoilExtension: boolean;
  removeNegativeAttackCards: boolean;
  includeVenusMA: boolean;
  moonExpansion: boolean;
  pathfindersExpansion: boolean;
  underworldExpansion: boolean;
  escapeVelocityExpansion: boolean;
  fast: boolean;
  showOtherPlayersVP: boolean;
  customCorporationsList?: string[];
  bannedCards?: string[];
  includedCards?: string[];
  soloTR: boolean;
  randomFirstPlayer: boolean;
  requiresVenusTrackCompletion: boolean;
  requiresMoonTrackCompletion: boolean;
  moonStandardProjectVariant: boolean;
  altVenusBoard: boolean;
  escapeVelocityMode: boolean;
  escapeVelocityThreshold: number;
  escapeVelocityPeriod: number;
  escapeVelocityPenalty: number;
  twoTempTerraformingThreshold: boolean;
  heatFor: boolean;
  breakthrough: boolean;
}

export enum GameExpansion {
  PRELUDE = 'prelude',
  VENUS = 'venus',
  COLONIES = 'colonies',
  TURMOIL = 'turmoil',
  BIG_BOX = 'big_box',
  ARES = 'ares',
  MOON = 'moon',
  PATHFINDERS = 'pathfinders',
  PRELUDE2 = 'prelude2',
  CEO = 'ceo',
  PROMO = 'promo',
  COMMUNITY = 'community',
  UNDERWORLD = 'underworld',
  ESCAPE_VELOCITY = 'escape_velocity',
  STAR_WARS = 'star_wars'
}

// Game events and triggers
export interface GameEvent {
  id: string;
  type: GameEventType;
  triggeredBy?: string;
  data?: any;
  timestamp: number;
}

export enum GameEventType {
  GAME_STARTED = 'game_started',
  PLAYER_JOINED = 'player_joined',
  PLAYER_LEFT = 'player_left',
  CARD_PLAYED = 'card_played',
  TILE_PLACED = 'tile_placed',
  PARAMETER_INCREASED = 'parameter_increased',
  MILESTONE_CLAIMED = 'milestone_claimed',
  AWARD_FUNDED = 'award_funded',
  GENERATION_END = 'generation_end',
  GAME_END = 'game_end',
  PRODUCTION_PHASE = 'production_phase',
  RESEARCH_PHASE = 'research_phase'
}

// Standard projects
export interface StandardProject {
  id: string;
  name: string;
  cost: number;
  description: string;
  effect: () => void;
  available: (gameState: GameState, playerId: string) => boolean;
}

export enum StandardProjectType {
  SELL_PATENTS = 'sell_patents',
  POWER_PLANT = 'power_plant',
  ASTEROID = 'asteroid',
  AQUIFER = 'aquifer',
  GREENERY = 'greenery',
  CITY = 'city',
  AIR_SCRAPPING = 'air_scrapping',
  BUFFER_GAS = 'buffer_gas'
}