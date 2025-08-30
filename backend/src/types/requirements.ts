// Card requirement types for Terraforming Mars

import { CardTag } from './cards';

export enum RequirementType {
  TEMPERATURE = 'temperature',
  OXYGEN = 'oxygen',
  OCEANS = 'oceans',
  VENUS_SCALE = 'venus_scale',
  TAG_COUNT = 'tag_count',
  PRODUCTION = 'production',
  RESOURCE = 'resource',
  PARTY_LEADER = 'party_leader',
  CHAIRMAN = 'chairman',
  CITIES = 'cities',
  GREENERIES = 'greeneries',
  FLOATERS = 'floaters',
  COLONIES = 'colonies',
  TR = 'tr'
}

// Card requirements
export interface CardRequirement {
  type: RequirementType;
  min?: number;
  max?: number;
  tag?: CardTag;
  resource?: string;
  party?: string;
  colony?: string;
}