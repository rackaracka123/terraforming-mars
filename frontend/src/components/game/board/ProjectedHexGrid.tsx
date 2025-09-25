import { useMemo } from "react";
import * as THREE from "three";
import { HexGrid2D } from "../../../utils/hex-grid-2d";
import ProjectedHexTile from "./ProjectedHexTile";
import { GameDto } from "../../../types/generated/api-types";

enum TileType {
  EMPTY = "empty",
  OCEAN = "ocean",
  GREENERY = "greenery",
  CITY = "city",
  SPECIAL = "special",
}

interface ProjectedHexGridProps {
  gameState?: GameDto;
  onHexClick?: (hexCoordinate: string) => void;
}

export default function ProjectedHexGrid({
  gameState: _gameState,
  onHexClick,
}: ProjectedHexGridProps) {
  const SPHERE_RADIUS = 2.02;

  // Generate the 2D hexagonal grid and project onto sphere
  const projectedHexGrid = useMemo(() => {
    const hexGrid = HexGrid2D.generateGrid();

    return hexGrid.map((tile) => {
      // Project 2D position onto sphere surface
      const spherePosition = projectToSphere(tile.position, SPHERE_RADIUS);

      return {
        ...tile,
        spherePosition,
        normal: spherePosition.clone().normalize(),
      };
    });
  }, [SPHERE_RADIUS]);

  // Get tile data from game state
  const getTileData = (_hexCoordinate: string) => {
    return {
      type: TileType.EMPTY,
      ownerId: null,
      specialType: null,
    };
  };

  return (
    <>
      {projectedHexGrid.map((tile) => {
        const hexKey = HexGrid2D.coordinateToKey(tile.coordinate);
        const tileData = getTileData(hexKey);

        return (
          <ProjectedHexTile
            key={hexKey}
            tileData={tile}
            tileType={tileData.type}
            ownerId={tileData.ownerId}
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
