import { useMemo } from "react";
import * as THREE from "three";
import { HexGrid2D } from "../../../utils/hex-grid-2d";
import ProjectedHexTile from "./ProjectedHexTile";
import {
  GameDto,
  TileDto,
  TileBonusDto,
} from "../../../types/generated/api-types";

interface ProjectedHexGridProps {
  gameState?: GameDto;
  onHexClick?: (hexCoordinate: string) => void;
}

// Local type for tiles with projected positions
interface ProjectedTile {
  backendTile?: TileDto;
  coordinate: { q: number; r: number; s: number };
  position: { x: number; y: number };
  spherePosition: THREE.Vector3;
  normal: THREE.Vector3;
  isOceanSpace: boolean;
  bonuses: { [key: string]: number };
}

// Type for the tile data returned by getTileData
type TileType = "city" | "empty" | "ocean" | "greenery" | "special";

interface TileData {
  type: TileType;
  ownerId: string | null;
  specialType: null;
}

export default function ProjectedHexGrid({
  gameState,
  onHexClick,
}: ProjectedHexGridProps) {
  const SPHERE_RADIUS = 2.02;

  // Convert hex coordinates to 2D pixel position (same as backend logic)
  const hexToPixel = (coord: { q: number; r: number; s: number }) => {
    const size = 0.3; // Same as HEX_SIZE in HexGrid2D
    const x = size * Math.sqrt(3) * (coord.q + coord.r / 2);
    const y = ((size * 3) / 2) * coord.r;
    return { x, y };
  };

  // Convert backend tile bonuses to legacy format
  const convertBackendBonuses = (bonuses: TileBonusDto[] | undefined) => {
    const converted: { [key: string]: number } = {};
    if (bonuses) {
      bonuses.forEach((bonus) => {
        converted[bonus.type] = bonus.amount;
      });
    }
    return converted;
  };

  // Use backend board tiles or fallback to hardcoded generation
  const projectedHexGrid = useMemo((): ProjectedTile[] => {
    // Use backend tiles if available
    if (gameState?.board?.tiles) {
      return gameState.board.tiles.map((tile: TileDto): ProjectedTile => {
        // Convert hex coordinate to 2D position for projection
        const position2D = hexToPixel(tile.coordinates);
        const spherePosition = projectToSphere(position2D, SPHERE_RADIUS);

        return {
          backendTile: tile,
          coordinate: tile.coordinates,
          position: position2D,
          spherePosition,
          normal: spherePosition.clone().normalize(),
          // Convert backend tile data to legacy interface for compatibility
          isOceanSpace: tile.type === "ocean-tile",
          bonuses: convertBackendBonuses(tile.bonuses),
        };
      });
    }

    // Fallback to hardcoded generation if backend data not available
    const hexGrid = HexGrid2D.generateGrid();
    return hexGrid.map((tile) => {
      const spherePosition = projectToSphere(tile.position, SPHERE_RADIUS);
      return {
        ...tile,
        spherePosition,
        normal: spherePosition.clone().normalize(),
      };
    });
  }, [gameState?.board?.tiles, SPHERE_RADIUS]);

  // Get tile type and occupancy data
  const getTileData = (tile: ProjectedTile): TileData => {
    if (tile.backendTile) {
      const backendTile: TileDto = tile.backendTile;

      // Determine tile type based on occupancy
      if (backendTile.occupiedBy) {
        switch (backendTile.occupiedBy.type) {
          case "ocean-tile":
            return {
              type: "ocean",
              ownerId: backendTile.ownerId || null,
              specialType: null,
            };
          case "city-tile":
            return {
              type: "city",
              ownerId: backendTile.ownerId || null,
              specialType: null,
            };
          case "greenery-tile":
            return {
              type: "greenery",
              ownerId: backendTile.ownerId || null,
              specialType: null,
            };
          default:
            return {
              type: "special",
              ownerId: backendTile.ownerId || null,
              specialType: null,
            };
        }
      }

      // Empty tile
      return {
        type: "empty",
        ownerId: backendTile.ownerId || null,
        specialType: null,
      };
    }

    // Fallback for hardcoded tiles
    return { type: "empty", ownerId: null, specialType: null };
  };

  return (
    <>
      {projectedHexGrid.map((tile) => {
        const hexKey = HexGrid2D.coordinateToKey(tile.coordinate);
        const tileData = getTileData(tile);

        return (
          <ProjectedHexTile
            key={hexKey}
            tileData={tile}
            tileType={tileData.type}
            ownerId={tileData.ownerId}
            displayName={tile.backendTile?.displayName}
            onClick={() => onHexClick?.(hexKey)}
          />
        );
      })}
    </>
  );
}

/**
 * Project a 2D point onto the surface of a sphere
 * This simulates "wrapping" the flat hex grid around the sphere
 */
function projectToSphere(
  position2D: { x: number; y: number },
  radius: number,
): THREE.Vector3 {
  // Scale the 2D coordinates to fit nicely on the sphere
  const scale = 0.4; // Reduced scale to bring hexagons closer together
  const x = position2D.x * scale;
  const y = position2D.y * scale;

  // Project onto sphere using azimuthal projection
  // This simulates "wrapping" the flat grid around the front hemisphere
  const r = Math.sqrt(x * x + y * y);

  if (r === 0) {
    // Center point
    return new THREE.Vector3(0, 0, radius);
  }

  // Convert to spherical coordinates
  const theta = Math.atan2(y, x); // Azimuth angle
  const phi = (r / radius) * (Math.PI / 2); // Polar angle (scaled to hemisphere)

  // Convert back to Cartesian coordinates on sphere
  const sphereX = radius * Math.sin(phi) * Math.cos(theta);
  const sphereY = radius * Math.sin(phi) * Math.sin(theta);
  const sphereZ = radius * Math.cos(phi);

  return new THREE.Vector3(sphereX, sphereY, sphereZ);
}
