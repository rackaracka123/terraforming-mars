# Generated TypeScript Types

This file documents the automatically generated TypeScript interfaces from Go structs.

**DO NOT EDIT** - This file is auto-generated from Go domain models.

## Generated Interfaces

```typescript
// Generated TypeScript interfaces from Go structs
// DO NOT EDIT - This file is auto-generated

export interface GameAggregate {
}

export interface BoardSpace {
  position: HexCoordinate;
  type: SpaceType;
  bonus: ResourceType[];
  isOceanSpace: boolean;
  isReserved: boolean;
  reservedFor?: string;
  tile?: Tile;
  adjacentTo: HexCoordinate[];
}

export type SpaceType = string;

export interface PlacementRule {
  tileType: TileType;
  requirements: PlacementReq[];
  restrictions: PlacementReq[];
  bonusRules: PlacementBonus[];
}

export interface PlacementReq {
  type: PlacementReqType;
  target: PlacementTarget;
  distance?: number;
  count?: number;
  tileType?: TileType;
  spaceType?: SpaceType;
  playerId?: string;
}

export type PlacementReqType = string;

export type PlacementTarget = string;

export interface PlacementBonus {
  condition: PlacementCondition;
  effect: CardEffect;
  description: string;
}

export interface PlacementCondition {
  type: PlacementCondType;
  target: PlacementTarget;
  count?: number;
  tileType?: TileType;
  minDistance?: number;
  maxDistance?: number;
}

export type PlacementCondType = string;

export interface AdjacencyBonus {
  sourceTile: TileType;
  targetTile: TileType;
  bonus: CardEffect;
  description: string;
}

export interface ProjectCard {
  id: string;
  name: string;
  type: CardType;
  cost: number;
  tags: Tag[];
  requirements: Requirement[];
  effects: CardEffect[];
  victoryPoints?: number;
  description: string;
  flavorText?: string;
  imagePath: string;
}

export type CardType = string;

export interface CardEffect {
  type: EffectType;
  target: EffectTarget;
  amount?: number;
  resourceType?: ResourceType;
  tileType?: TileType;
  condition?: EffectCondition;
  trigger?: EffectTrigger;
}

export type EffectType = string;

export type EffectTarget = string;

export interface EffectCondition {
  type: ConditionType;
  parameter?: GlobalParam;
  minValue?: number;
  maxValue?: number;
  resourceType?: ResourceType;
  tag?: Tag;
  tileType?: TileType;
}

export type ConditionType = string;

export interface EffectTrigger {
  event: TriggerEvent;
  tag?: Tag;
  tileType?: TileType;
  parameter?: GlobalParam;
}

export type TriggerEvent = string;

export interface Requirement {
  type: RequirementType;
  parameter?: GlobalParam;
  minValue?: number;
  maxValue?: number;
  tag?: Tag;
  count?: number;
}

export type RequirementType = string;

export type GlobalParam = string;

export interface Corporation {
  id: string;
  name: string;
  description: string;
  startingMegaCredits: number;
  startingProduction?: ResourcesMap;
  startingResources?: ResourcesMap;
  startingTR?: number;
  startingCards: string[];
  tags: Tag[];
  logoPath: string;
  color: string;
  ability: string;
}

export type CorporationType = string;

export interface Milestone {
  id: string;
  name: string;
  description: string;
  cost: number;
  claimedBy?: string;
  achievementType: AchievementType;
  requiredValue: number;
}

export interface Award {
  id: string;
  name: string;
  description: string;
  cost: number;
  fundedBy?: string;
  competitionType: CompetitionType;
  ranking: AwardRanking[];
}

export type AchievementType = string;

export type CompetitionType = string;

export interface AwardRanking {
  playerId: string;
  value: number;
  rank: number;
}

export interface GameEventStream {
  gameId: string;
  events: GameEvent[];
  version: number;
  createdAt: string;
  updatedAt: string;
}

export interface GameEvent {
  id: string;
  gameId: string;
  type: GameEventType;
  version: number;
  playerId?: string;
  timestamp: string;
  data: any;
  metadata: EventMetadata;
}

export interface EventMetadata {
  correlationId?: string;
  causedBy?: string;
  tags: string[];
  userAgent?: string;
  ipAddress?: string;
  extra: Record<string,;
}

export type GameEventType = string;

export interface GameCreatedData {
  settings: GameSettings;
  createdBy: string;
  maxPlayers: number;
}

export interface PlayerJoinedData {
  playerId: string;
  playerName: string;
  joinOrder: number;
}

export interface CorporationSelectedData {
  playerId: string;
  corporationId: string;
}

export interface CardPlayedData {
  playerId: string;
  cardId: string;
  cost: number;
  resourcesSpent: ResourcesMap;
  placement?: HexCoordinate;
  targetPlayer?: string;
  requirements: Requirement[];
  immediateEffects: CardEffect[];
}

export interface TilePlacedData {
  playerId: string;
  tileType: TileType;
  position: HexCoordinate;
  spaceBonuses: ResourceType[];
  adjacencyBonuses: ResourceType[];
  triggeredEffects: CardEffect[];
}

export interface ResourcesChangedData {
  playerId: string;
  changes: ResourcesMap;
  newTotals: ResourcesMap;
  reason: string;
  sourceCard?: string;
}

export interface ProductionChangedData {
  playerId: string;
  changes: ResourcesMap;
  newTotals: ResourcesMap;
  reason: string;
  sourceCard?: string;
}

export interface ParameterIncreasedData {
  playerId: string;
  parameter: GlobalParam;
  oldValue: number;
  newValue: number;
  steps: number;
  trIncrease: number;
  bonusRewards: CardEffect[];
}

export interface MilestoneClaimedData {
  playerId: string;
  milestoneId: string;
  cost: number;
  requirements: Requirement[];
}

export interface AwardFundedData {
  playerId: string;
  awardId: string;
  cost: number;
  position: number;
}

export interface PhaseChangedData {
  oldPhase: GamePhase;
  newPhase: GamePhase;
  oldTurnPhase?: TurnPhase;
  newTurnPhase?: TurnPhase;
  generation: number;
  trigger: string;
}

export interface GenerationStartedData {
  generation: number;
  playerOrder: string[];
  firstPlayer: string;
}

export interface TurnStartedData {
  playerId: string;
  actionsRemaining: number;
  turnIndex: number;
  generation: number;
}

export interface VictoryPointsAwardedData {
  playerId: string;
  points: number;
  source: VPSourceType;
  description: string;
  details?: string;
}

export interface GameEndedData {
  winnerId: string;
  finalScores: Record<string,;
  endCondition: string;
  duration: number;
  generations: number;
}

export interface EventFactory {
}

export interface GameState {
  id: string;
  players: Player[];
  currentPlayer: string;
  generation: number;
  phase: GamePhase;
  turnPhase?: TurnPhase;
  globalParameters: GlobalParameters;
  endGameConditions: EndGameCondition[];
  board: BoardSpace[];
  milestones: Milestone[];
  awards: Award[];
  availableStandardProjects: StandardProject[];
  firstPlayer: string;
  corporationDraft?: boolean;
  deck: string[];
  discardPile: string[];
  soloMode: boolean;
  turn: number;
  draftDirection?: DraftDirection;
  gameSettings: GameSettings;
  currentActionCount?: number;
  maxActionsPerTurn?: number;
  actionHistory: Action[];
  events: GameEvent[];
  isGameEnded: boolean;
  winnerId?: string;
  createdAt: string;
  updatedAt: string;
}

export type GamePhase = string;

export type TurnPhase = string;

export interface TurnOrder {
  generation: number;
  playerOrder: string[];
  currentIndex: number;
}

export interface GenerationPhaseConfig {
  phase: TurnPhase;
  description: string;
  isSimultaneous: boolean;
  requiresAllPassed: boolean;
  hasActionLimit: boolean;
  maxActionsPerPlayer?: number;
}

export type DraftDirection = string;

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

export type GameExpansion = string;

export interface GameEvent {
  id: string;
  type: GameEventType;
  triggeredBy?: string;
  data?: any;
  timestamp: number;
}

export type GameEventType = string;

export interface Action {
  id: string;
  type: ActionType;
  playerId: string;
  cardId?: string;
  projectId?: string;
  position?: HexCoordinate;
  resources?: ResourcesMap;
  target?: string;
  data?: any;
}

export type ActionType = string;

export interface StandardProject {
  id: string;
  name: string;
  type: StandardProjectType;
  cost: number;
  requirements: Requirement[];
  effects: CardEffect[];
  description: string;
  isRepeatable: boolean;
}

export type StandardProjectType = string;

export interface Player {
  id: string;
  name: string;
  corporation?: string;
  resources: ResourcesMap;
  production: ResourcesMap;
  terraformRating: number;
  victoryPoints: number;
  victoryPointSources: VictoryPointSource[];
  playedCards: string[];
  hand: string[];
  availableActions: number;
  tags: Tag[];
  tagCounts: Record<Tag,;
  actionsTaken: number;
  actionsRemaining: number;
  passed?: boolean;
  tilePositions: HexCoordinate[];
  tileCounts: Record<TileType,;
  reserved: ResourcesMap;
  claimedMilestones: string[];
  fundedAwards: string[];
  handLimit: number;
}

export interface ResourcesMap {
  credits: number;
  steel: number;
  titanium: number;
  plants: number;
  energy: number;
  heat: number;
}

export type ResourceType = string;

export interface GlobalParameters {
  temperature: number;
  oxygen: number;
  oceans: number;
}

export type Tag = string;

export interface HexCoordinate {
  q: number;
  r: number;
  s: number;
}

export type TileType = string;

export interface Tile {
  type: TileType;
  position: HexCoordinate;
  playerId?: string;
  bonus: ResourceType[];
  isReserved: boolean;
}

export interface ResourceConversion {
  from: ResourceType;
  to: ResourceType;
  rate: number;
  gain: number;
  effect?: string;
}

export interface ResourceMultiplier {
  resourceType: ResourceType;
  cardTag: Tag;
  value: number;
  description: string;
}

export interface VictoryPointSource {
  type: VPSourceType;
  points: number;
  description: string;
  details?: string;
}

export type VPSourceType = string;

export interface EndGameCondition {
  parameter: GlobalParam;
  targetValue: number;
  currentValue: number;
  isCompleted: boolean;
}


```

## Usage

Import the types in your TypeScript/React code:

```typescript
import { GameState, Player, Corporation } from '../types/generated/api-types';
```

## Regeneration

To regenerate these types, run:

```bash
cd backend && go run tools/generate-types.go
```
