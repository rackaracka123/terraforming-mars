// Card effects system for Terraforming Mars

import { ResourceCost, ResourceGain, ProductionChange } from '../resources';
import { CardTag, CardType } from './base';

// Card effects system
export interface Effect {
  trigger: EffectTrigger;
  condition?: EffectCondition;
  action: EffectAction;
  optional?: boolean;
  repeat?: number;
}

export enum EffectTrigger {
  IMMEDIATE = 'immediate',
  ONGOING = 'ongoing',
  ACTIVATED = 'activated',
  ON_CARD_PLAYED = 'on_card_played',
  ON_TILE_PLACED = 'on_tile_placed',
  ON_PARAMETER_INCREASE = 'on_parameter_increase',
  ON_CITY_PLACED = 'on_city_placed',
  ON_GREENERY_PLACED = 'on_greenery_placed',
  ON_OCEAN_PLACED = 'on_ocean_placed',
  ON_PRODUCTION_PHASE = 'on_production_phase',
  ON_RESEARCH_PHASE = 'on_research_phase'
}

export interface EffectCondition {
  type: ConditionType;
  count?: number;
  tag?: CardTag;
  cardType?: CardType;
  parameter?: string;
  comparison?: ComparisonOperator;
  value?: number;
}

export enum ConditionType {
  TAG_COUNT = 'tag_count',
  CARD_TYPE = 'card_type',
  PARAMETER_VALUE = 'parameter_value',
  RESOURCE_COUNT = 'resource_count',
  PLAYER_HAS = 'player_has',
  TILE_TYPE = 'tile_type'
}

export enum ComparisonOperator {
  EQUALS = 'equals',
  GREATER_THAN = 'greater_than',
  LESS_THAN = 'less_than',
  GREATER_EQUAL = 'greater_equal',
  LESS_EQUAL = 'less_equal'
}

export interface EffectAction {
  type: ActionType;
  resourceCost?: ResourceCost;
  resourceGain?: ResourceGain;
  productionChange?: ProductionChange;
  parameterIncrease?: ParameterIncrease;
  drawCards?: number;
  victoryPoints?: number;
  tileType?: CardTileType;
  customFunction?: string;
  choice?: EffectChoice;
}

export enum ActionType {
  GAIN_RESOURCES = 'gain_resources',
  LOSE_RESOURCES = 'lose_resources',
  INCREASE_PRODUCTION = 'increase_production',
  DECREASE_PRODUCTION = 'decrease_production',
  INCREASE_PARAMETER = 'increase_parameter',
  PLACE_TILE = 'place_tile',
  DRAW_CARDS = 'draw_cards',
  GAIN_VP = 'gain_vp',
  CUSTOM = 'custom',
  CHOICE = 'choice',
  ADD_RESOURCE_TO_CARD = 'add_resource_to_card',
  REMOVE_RESOURCE_FROM_CARD = 'remove_resource_from_card'
}

export interface ParameterIncrease {
  temperature?: number;
  oxygen?: number;
  venus?: number;
}

export enum CardTileType {
  GREENERY = 'greenery',
  CITY = 'city',
  OCEAN = 'ocean',
  SPECIAL = 'special',
  RESTRICTED_AREA = 'restricted_area',
  LAVA_FLOWS = 'lava_flows',
  MOHOLE_AREA = 'mohole_area',
  NATURAL_PRESERVE = 'natural_preserve',
  NUCLEAR_ZONE = 'nuclear_zone',
  COMMERCIAL_DISTRICT = 'commercial_district',
  ECOLOGICAL_ZONE = 'ecological_zone',
  INDUSTRIAL_CENTER = 'industrial_center',
  MINING_AREA = 'mining_area',
  DEIMOS_DOWN = 'deimos_down',
  GREAT_DAM = 'great_dam',
  MAGNETIC_FIELD_GENERATORS = 'magnetic_field_generators'
}

export interface EffectChoice {
  options: EffectAction[];
  min?: number;
  max?: number;
}