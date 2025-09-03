import * as THREE from "three";

export interface HexCoordinate {
  q: number;
  r: number;
  s: number;
}

export interface SphericalPosition {
  theta: number; // longitude
  phi: number; // latitude
}

export interface CartesianPosition {
  x: number;
  y: number;
  z: number;
}

/**
 * Geodesic hexagonal grid system based on icosahedron subdivision
 * Creates approximately 61 hexagonal tiles similar to Terraforming Mars board
 */
export class GeodesicGrid {
  private static readonly SPHERE_RADIUS = 2;
  private static readonly GOLDEN_RATIO = (1 + Math.sqrt(5)) / 2;

  /**
   * Generate icosahedron vertices in 3D space
   */
  private static generateIcosahedronVertices(): THREE.Vector3[] {
    const t = GeodesicGrid.GOLDEN_RATIO;
    const vertices = [
      // 12 vertices of icosahedron
      new THREE.Vector3(-1, t, 0),
      new THREE.Vector3(1, t, 0),
      new THREE.Vector3(-1, -t, 0),
      new THREE.Vector3(1, -t, 0),
      new THREE.Vector3(0, -1, t),
      new THREE.Vector3(0, 1, t),
      new THREE.Vector3(0, -1, -t),
      new THREE.Vector3(0, 1, -t),
      new THREE.Vector3(t, 0, -1),
      new THREE.Vector3(t, 0, 1),
      new THREE.Vector3(-t, 0, -1),
      new THREE.Vector3(-t, 0, 1),
    ];

    // Normalize to unit sphere
    return vertices.map((v) => v.normalize());
  }

  /**
   * Generate icosahedron faces (triangles)
   */
  private static generateIcosahedronFaces(): number[][] {
    return [
      // 20 faces of icosahedron (each array contains 3 vertex indices)
      [0, 11, 5],
      [0, 5, 1],
      [0, 1, 7],
      [0, 7, 10],
      [0, 10, 11],
      [1, 5, 9],
      [5, 11, 4],
      [11, 10, 2],
      [10, 7, 6],
      [7, 1, 8],
      [3, 9, 4],
      [3, 4, 2],
      [3, 2, 6],
      [3, 6, 8],
      [3, 8, 9],
      [4, 9, 5],
      [2, 4, 11],
      [6, 2, 10],
      [8, 6, 7],
      [9, 8, 1],
    ];
  }

  /**
   * Determine if a row should be offset for proper hex grid alignment
   */
  private static shouldOffsetRow(
    row: number,
    hexesInRow: number,
    pattern: number[],
  ): boolean {
    // For Terraforming Mars pattern [5, 6, 6, 8, 8, 7, 6, 5]
    // The middle rows with 8 hexes (rows 3 and 4) should be offset 0.5 tile width to the right
    const centerRowIndex = Math.floor(pattern.length / 2);

    // Offset the middle rows with maximum hex count (8 hexes)
    return hexesInRow === 8;
  }

  /**
   * Create hexagonal grid positions on sphere surface
   * Based on the Terraforming Mars board layout - covers ~80% of one hemisphere
   */
  static generateHexGrid(): Array<{
    coordinate: HexCoordinate;
    sphericalPosition: SphericalPosition;
    cartesianPosition: CartesianPosition;
    face: number;
    isOceanSpace: boolean;
    bonuses: { [key: string]: number };
  }> {
    const hexes: Array<{
      coordinate: HexCoordinate;
      sphericalPosition: SphericalPosition;
      cartesianPosition: CartesianPosition;
      face: number;
      isOceanSpace: boolean;
      bonuses: { [key: string]: number };
    }> = [];

    // Generate a concentrated hex grid on one side of Mars
    // Create a more systematic approach like the actual Terraforming Mars board

    const baseRadius = this.SPHERE_RADIUS;

    // Focus on the front-facing hemisphere (positive Z direction)
    // Create a roughly hexagonal pattern centered around (0, 0, positive Z)

    // Define the hex pattern for Terraforming Mars board
    const hexPattern = [5, 6, 7, 8, 9, 8, 7, 6, 5]; // tiles per row

    for (let row = 0; row < hexPattern.length; row++) {
      const hexesInRow = hexPattern[row];

      for (let col = 0; col < hexesInRow; col++) {
        // Create hex coordinate using axial coordinates
        const q = col - Math.floor(hexesInRow / 2);
        const r = row - Math.floor(hexPattern.length / 2);
        const s = -q - r;

        // Proper hexagonal grid positioning
        const hexWidth = 0.46;
        const hexHeight = 0.4;

        // Calculate position using proper hex grid math
        // In a hex grid, every other row is offset by half a hex width
        // Alternate offset pattern for proper hex alignment
        const xOffset = row % 2 === 1 ? hexWidth / 2 : 0.25;

        const normalizedCol = ((col - hexesInRow / 2) * hexWidth + xOffset) / 2;
        const normalizedRow = ((row - hexPattern.length / 2) * hexHeight) / 2;

        // Scale to cover ~70% of front hemisphere (slightly more spread)
        const scale = 0.7;
        const flatX = normalizedCol * scale;
        const flatY = normalizedRow * scale;

        // Project flat coordinates onto front of sphere (positive Z hemisphere)
        // Use inverse projection to map from flat disc to sphere front
        const distFromCenter = Math.sqrt(flatX * flatX + flatY * flatY);

        if (distFromCenter <= 1) {
          // Project onto sphere front
          const z =
            Math.sqrt(Math.max(0, 1 - distFromCenter * distFromCenter)) *
            baseRadius;
          const x = flatX * baseRadius;
          const y = flatY * baseRadius;

          // Convert to spherical coordinates for completeness
          const phi = Math.acos(z / baseRadius);
          const theta = Math.atan2(y, x);

          // Determine if this is an ocean space (roughly 9 out of 61 tiles)
          const isOceanSpace = (row + col) % 7 === 0;

          // Add resource bonuses to specific tiles based on position
          const bonuses: { [key: string]: number } = {};
          if ((row * hexesInRow + col) % 8 === 0) bonuses.steel = 2;
          if ((row * hexesInRow + col) % 9 === 0) bonuses.titanium = 1;
          if ((row * hexesInRow + col) % 11 === 0) bonuses.plants = 1;
          if ((row * hexesInRow + col) % 13 === 0) bonuses.cards = 1;

          hexes.push({
            coordinate: { q, r, s },
            sphericalPosition: { theta, phi },
            cartesianPosition: { x, y, z },
            face: row, // Use row as face identifier
            isOceanSpace,
            bonuses,
          });
        }
      }
    }

    return hexes;
  }

  /**
   * Convert hex coordinate to unique string key
   */
  static coordinateToKey(coord: HexCoordinate): string {
    return `${coord.q},${coord.r},${coord.s}`;
  }

  /**
   * Calculate distance between two hex coordinates
   */
  static distance(a: HexCoordinate, b: HexCoordinate): number {
    return Math.max(
      Math.abs(a.q - b.q),
      Math.abs(a.r - b.r),
      Math.abs(a.s - b.s),
    );
  }

  /**
   * Get neighboring hex coordinates
   */
  static neighbors(coord: HexCoordinate): HexCoordinate[] {
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
