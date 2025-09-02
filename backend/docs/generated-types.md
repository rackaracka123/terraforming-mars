# Generated TypeScript Types

This file documents the automatically generated TypeScript interfaces from Go structs.

**DO NOT EDIT** - This file is auto-generated from Go domain models.

## Generated Interfaces

```typescript
// Generated TypeScript interfaces from Go structs
// DO NOT EDIT - This file is auto-generated

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
}

export interface Award {
  id: string;
  name: string;
  description: string;
  cost: number;
  fundedBy?: string;
}

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
  deck: string[];
  discardPile: string[];
  soloMode: boolean;
  turn: number;
  draftDirection?: DraftDirection;
  gameSettings: GameSettings;
  currentActionCount?: number;
  maxActionsPerTurn?: number;
  createdAt: string;
  updatedAt: string;
}

export type GamePhase = string;

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

export interface StandardProject {
  id: string;
  name: string;
  cost: number;
  description: string;
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
  playedCards: string[];
  hand: string[];
  availableActions: number;
  tags: Tag[];
  actionsTaken: number;
  actionsRemaining: number;
  passed?: boolean;
  tilePositions: HexCoordinate[];
  reserved: ResourcesMap;
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
