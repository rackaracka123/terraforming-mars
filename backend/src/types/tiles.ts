// Mars board and tile system types

export enum TileType {
  EMPTY = 'empty',
  OCEAN = 'ocean',
  GREENERY = 'greenery',
  CITY = 'city',
  SPECIAL = 'special'
}

export enum SpecialTileType {
  NOCTIS_CITY = 'noctis_city',
  PHOBOS_SPACE_HAVEN = 'phobos_space_haven',
  GANYMEDE_COLONY = 'ganymede_colony',
  SPACE_ELEVATOR = 'space_elevator',
  LUNA_TRADE_STATION = 'luna_trade_station',
  MINING_AREA = 'mining_area',
  MINING_RIGHTS = 'mining_rights',
  RESTRICTED_AREA = 'restricted_area',
  ECOLOGICAL_ZONE = 'ecological_zone',
  NATURAL_PRESERVE = 'natural_preserve'
}

export interface HexCoordinate {
  q: number; // Column
  r: number; // Row
  s: number; // Diagonal (q + r + s = 0)
}

export interface TileBonuses {
  steel?: number;
  titanium?: number;
  plants?: number;
  cards?: number;
  credits?: number;
  heat?: number;
}

export interface MarsTile {
  coordinate: HexCoordinate;
  type: TileType;
  specialType?: SpecialTileType;
  ownerId?: string; // Player who placed the tile
  bonuses: TileBonuses;
  isOceanSpace: boolean; // Pre-designated ocean placement areas
  isRestricted: boolean; // Cannot place tiles here
  adjacentToOcean?: boolean; // Updated when oceans are placed
}

export interface MarsBoard {
  tiles: Map<string, MarsTile>; // Key: "q,r" coordinate string
  oceanSpaces: HexCoordinate[]; // Pre-designated ocean placement spots
  citySpaces: HexCoordinate[]; // Valid city placement spots
}

// Mars board layout based on the official game
export const MARS_BOARD_LAYOUT: { coordinate: HexCoordinate; bonuses: TileBonuses; isOceanSpace: boolean; isRestricted: boolean }[] = [
  // Row 0 (top)
  { coordinate: { q: 4, r: -4, s: 0 }, bonuses: { steel: 2 }, isOceanSpace: false, isRestricted: false },
  
  // Row 1
  { coordinate: { q: 3, r: -3, s: 0 }, bonuses: {}, isOceanSpace: true, isRestricted: false },
  { coordinate: { q: 4, r: -3, s: -1 }, bonuses: { steel: 2 }, isOceanSpace: false, isRestricted: false },
  { coordinate: { q: 5, r: -3, s: -2 }, bonuses: {}, isOceanSpace: true, isRestricted: false },
  
  // Row 2 
  { coordinate: { q: 2, r: -2, s: 0 }, bonuses: { cards: 1 }, isOceanSpace: false, isRestricted: false },
  { coordinate: { q: 3, r: -2, s: -1 }, bonuses: {}, isOceanSpace: false, isRestricted: false },
  { coordinate: { q: 4, r: -2, s: -2 }, bonuses: { steel: 1 }, isOceanSpace: false, isRestricted: false },
  { coordinate: { q: 5, r: -2, s: -3 }, bonuses: {}, isOceanSpace: false, isRestricted: false },
  { coordinate: { q: 6, r: -2, s: -4 }, bonuses: { titanium: 1 }, isOceanSpace: false, isRestricted: false },
  
  // Row 3
  { coordinate: { q: 1, r: -1, s: 0 }, bonuses: { plants: 1 }, isOceanSpace: false, isRestricted: false },
  { coordinate: { q: 2, r: -1, s: -1 }, bonuses: {}, isOceanSpace: false, isRestricted: false },
  { coordinate: { q: 3, r: -1, s: -2 }, bonuses: { steel: 1 }, isOceanSpace: false, isRestricted: false },
  { coordinate: { q: 4, r: -1, s: -3 }, bonuses: {}, isOceanSpace: false, isRestricted: false },
  { coordinate: { q: 5, r: -1, s: -4 }, bonuses: {}, isOceanSpace: false, isRestricted: false },
  { coordinate: { q: 6, r: -1, s: -5 }, bonuses: { titanium: 1 }, isOceanSpace: false, isRestricted: false },
  { coordinate: { q: 7, r: -1, s: -6 }, bonuses: {}, isOceanSpace: true, isRestricted: false },
  
  // Row 4 (middle)
  { coordinate: { q: 0, r: 0, s: 0 }, bonuses: { plants: 1 }, isOceanSpace: false, isRestricted: false },
  { coordinate: { q: 1, r: 0, s: -1 }, bonuses: { plants: 1 }, isOceanSpace: false, isRestricted: false },
  { coordinate: { q: 2, r: 0, s: -2 }, bonuses: {}, isOceanSpace: false, isRestricted: false },
  { coordinate: { q: 3, r: 0, s: -3 }, bonuses: {}, isOceanSpace: true, isRestricted: false },
  { coordinate: { q: 4, r: 0, s: -4 }, bonuses: {}, isOceanSpace: false, isRestricted: true }, // Tharsis Tholus
  { coordinate: { q: 5, r: 0, s: -5 }, bonuses: {}, isOceanSpace: false, isRestricted: false },
  { coordinate: { q: 6, r: 0, s: -6 }, bonuses: {}, isOceanSpace: true, isRestricted: false },
  { coordinate: { q: 7, r: 0, s: -7 }, bonuses: { titanium: 2 }, isOceanSpace: false, isRestricted: false },
  { coordinate: { q: 8, r: 0, s: -8 }, bonuses: {}, isOceanSpace: true, isRestricted: false },
  
  // Additional rows following the same pattern...
  // For now, I'll include a representative sample. Full board has 61 hexes.
  
  // Row 5
  { coordinate: { q: 0, r: 1, s: -1 }, bonuses: { plants: 2 }, isOceanSpace: false, isRestricted: false },
  { coordinate: { q: 1, r: 1, s: -2 }, bonuses: { plants: 1 }, isOceanSpace: false, isRestricted: false },
  { coordinate: { q: 2, r: 1, s: -3 }, bonuses: {}, isOceanSpace: false, isRestricted: false },
  { coordinate: { q: 3, r: 1, s: -4 }, bonuses: {}, isOceanSpace: false, isRestricted: false },
  { coordinate: { q: 4, r: 1, s: -5 }, bonuses: { steel: 1 }, isOceanSpace: false, isRestricted: false },
  { coordinate: { q: 5, r: 1, s: -6 }, bonuses: {}, isOceanSpace: true, isRestricted: false },
  { coordinate: { q: 6, r: 1, s: -7 }, bonuses: {}, isOceanSpace: false, isRestricted: false },
  { coordinate: { q: 7, r: 1, s: -8 }, bonuses: { titanium: 1 }, isOceanSpace: false, isRestricted: false },
];

// Utility functions for hex coordinate math
export class HexMath {
  static coordinateToKey(coord: HexCoordinate): string {
    return `${coord.q},${coord.r}`;
  }
  
  static keyToCoordinate(key: string): HexCoordinate {
    const [q, r] = key.split(',').map(Number);
    return { q, r, s: -q - r };
  }
  
  static add(a: HexCoordinate, b: HexCoordinate): HexCoordinate {
    return { q: a.q + b.q, r: a.r + b.r, s: a.s + b.s };
  }
  
  static subtract(a: HexCoordinate, b: HexCoordinate): HexCoordinate {
    return { q: a.q - b.q, r: a.r - b.r, s: a.s - b.s };
  }
  
  static distance(a: HexCoordinate, b: HexCoordinate): number {
    const diff = this.subtract(a, b);
    return (Math.abs(diff.q) + Math.abs(diff.q + diff.r) + Math.abs(diff.r)) / 2;
  }
  
  // Get all 6 neighbors of a hex
  static neighbors(coord: HexCoordinate): HexCoordinate[] {
    const directions = [
      { q: 1, r: 0, s: -1 },   // East
      { q: 1, r: -1, s: 0 },   // Northeast  
      { q: 0, r: -1, s: 1 },   // Northwest
      { q: -1, r: 0, s: 1 },   // West
      { q: -1, r: 1, s: 0 },   // Southwest
      { q: 0, r: 1, s: -1 }    // Southeast
    ];
    
    return directions.map(dir => this.add(coord, dir));
  }
  
  // Convert hex coordinate to pixel position for rendering
  static hexToPixel(coord: HexCoordinate, hexSize: number): { x: number, y: number } {
    const x = hexSize * (Math.sqrt(3) * coord.q + Math.sqrt(3) / 2 * coord.r);
    const y = hexSize * (3 / 2 * coord.r);
    return { x, y };
  }
  
  // Convert pixel position back to hex coordinate
  static pixelToHex(x: number, y: number, hexSize: number): HexCoordinate {
    const q = (Math.sqrt(3) / 3 * x - 1 / 3 * y) / hexSize;
    const r = (2 / 3 * y) / hexSize;
    const s = -q - r;
    
    // Round to nearest integer coordinates
    const rq = Math.round(q);
    const rr = Math.round(r);
    const rs = Math.round(s);
    
    const q_diff = Math.abs(rq - q);
    const r_diff = Math.abs(rr - r);
    const s_diff = Math.abs(rs - s);
    
    if (q_diff > r_diff && q_diff > s_diff) {
      return { q: -rr - rs, r: rr, s: rs };
    } else if (r_diff > s_diff) {
      return { q: rq, r: -rq - rs, s: rs };
    } else {
      return { q: rq, r: rr, s: -rq - rr };
    }
  }
}

// Initialize Mars board
export function createMarsBoard(): MarsBoard {
  const tiles = new Map<string, MarsTile>();
  const oceanSpaces: HexCoordinate[] = [];
  const citySpaces: HexCoordinate[] = [];
  
  for (const layout of MARS_BOARD_LAYOUT) {
    const tile: MarsTile = {
      coordinate: layout.coordinate,
      type: TileType.EMPTY,
      bonuses: layout.bonuses,
      isOceanSpace: layout.isOceanSpace,
      isRestricted: layout.isRestricted,
      adjacentToOcean: false
    };
    
    tiles.set(HexMath.coordinateToKey(layout.coordinate), tile);
    
    if (layout.isOceanSpace) {
      oceanSpaces.push(layout.coordinate);
    }
    
    if (!layout.isOceanSpace && !layout.isRestricted) {
      citySpaces.push(layout.coordinate);
    }
  }
  
  return { tiles, oceanSpaces, citySpaces };
}