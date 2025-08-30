// Player-related type definitions for Terraforming Mars

import { PlayerResources, PlayerProduction } from './resources';
import { Card, CardType, CardTag } from './cards';

export interface Player {
  id: string;
  name: string;
  resources: PlayerResources;
  production: PlayerProduction;
  terraformRating: number;
  victoryPoints: number;
  corporation?: string;
  passed?: boolean;
  playedCards: Card[];
  hand: Card[];
  availableActions: number;
  tags: CardTag[];
}

// Player actions and states
export enum PlayerPhase {
  WAITING = 'waiting',
  DRAFTING = 'drafting',
  RESEARCH = 'research',
  ACTION = 'action',
  PRODUCTION = 'production',
  PASSED = 'passed'
}

export interface PlayerAction {
  type: PlayerActionType;
  playerId: string;
  data?: any;
}

export enum PlayerActionType {
  PLAY_CARD = 'play_card',
  STANDARD_PROJECT = 'standard_project',
  PASS = 'pass',
  CLAIM_MILESTONE = 'claim_milestone',
  FUND_AWARD = 'fund_award',
  USE_EFFECT = 'use_effect',
  PLACE_TILE = 'place_tile',
  RAISE_PARAMETER = 'raise_parameter'
}

