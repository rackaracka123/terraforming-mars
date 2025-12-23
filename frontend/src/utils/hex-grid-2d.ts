export interface HexCoordinate {
  q: number;
  r: number;
  s: number;
}

export interface HexTile2D {
  coordinate: HexCoordinate;
  position: { x: number; y: number };
  isOceanSpace: boolean;
  bonuses: { [key: string]: number };
}

/**
 * Simple 2D hexagonal grid generator for testing
 * Creates a honeycomb pattern with pointy-top hexagons
 */
export class HexGrid2D {
  private static readonly HEX_SIZE = 0.3; // Size of hexagons

  /**
   * Generate a 2D hexagonal grid with the pattern: 5-6-7-8-9-8-7-6-5
   */
  static generateGrid(): HexTile2D[] {
    const tiles: HexTile2D[] = [];

    // Row pattern: 5, 6, 7, 8, 9, 8, 7, 6, 5
    const rowPattern = [5, 6, 7, 8, 9, 8, 7, 6, 5];

    for (let rowIdx = 0; rowIdx < rowPattern.length; rowIdx++) {
      const hexCount = rowPattern[rowIdx];
      const r = rowIdx - Math.floor(rowPattern.length / 2); // Center the rows: -4 to +4

      for (let colIdx = 0; colIdx < hexCount; colIdx++) {
        // Calculate axial coordinates for honeycomb pattern
        const q = colIdx - Math.floor(hexCount / 2) - Math.floor(r / 2);
        const s = -q - r;

        const coordinate: HexCoordinate = { q, r, s };

        // Convert to 2D position (pointy-top orientation)
        const position = this.hexToPixel(coordinate);

        // Determine ocean spaces and bonuses
        const isOceanSpace = this.isOceanPosition(rowIdx, colIdx);
        const bonuses = this.calculateBonuses(rowIdx, colIdx);

        tiles.push({
          coordinate,
          position,
          isOceanSpace,
          bonuses,
        });
      }
    }

    return tiles;
  }

  /**
   * Convert hex coordinate to 2D pixel position
   * Uses pointy-top orientation (flat sides left/right, pointy top/bottom)
   */
  private static hexToPixel(coord: HexCoordinate): { x: number; y: number } {
    const size = this.HEX_SIZE;

    // Pointy-top hex positioning
    const x = size * Math.sqrt(3) * (coord.q + coord.r / 2);
    const y = ((size * 3) / 2) * coord.r;

    return { x, y };
  }

  /**
   * Determine if a tile should be an ocean space
   */
  private static isOceanPosition(row: number, col: number): boolean {
    // Simple pattern for ocean spaces
    const oceanPositions = [
      { row: 1, col: 2 },
      { row: 2, col: 1 },
      { row: 2, col: 5 },
      { row: 3, col: 3 },
      { row: 4, col: 1 },
      { row: 4, col: 7 },
      { row: 5, col: 4 },
      { row: 6, col: 2 },
      { row: 7, col: 3 },
    ];

    return oceanPositions.some((pos) => pos.row === row && pos.col === col);
  }

  /**
   * Calculate resource bonuses for specific tiles
   */
  private static calculateBonuses(row: number, col: number): { [key: string]: number } {
    const bonuses: { [key: string]: number } = {};
    const tileIndex = row * 10 + col;

    if (tileIndex % 8 === 0) bonuses.steel = 2;
    if (tileIndex % 9 === 0) bonuses.titanium = 1;
    if (tileIndex % 11 === 0) bonuses.plants = 1;
    if (tileIndex % 13 === 0) bonuses.cards = 1;

    return bonuses;
  }

  /**
   * Convert coordinate to unique string key
   */
  static coordinateToKey(coord: HexCoordinate): string {
    return `${coord.q},${coord.r},${coord.s}`;
  }

  /**
   * Get neighboring hex coordinates
   */
  static getNeighbors(coord: HexCoordinate): HexCoordinate[] {
    const directions = [
      { q: 1, r: 0, s: -1 },
      { q: 1, r: -1, s: 0 },
      { q: 0, r: -1, s: 1 },
      { q: -1, r: 0, s: 1 },
      { q: -1, r: 1, s: 0 },
      { q: 0, r: 1, s: -1 },
    ];

    return directions.map((dir) => ({
      q: coord.q + dir.q,
      r: coord.r + dir.r,
      s: coord.s + dir.s,
    }));
  }
}
