import { useMemo } from "react";
import * as THREE from "three";
import { HexGrid2D } from "../../../utils/hex-grid-2d";
import ProjectedHexTile, { TileHighlightMode } from "./ProjectedHexTile";
import { TileVPIndicator } from "../../ui/overlay/EndGameOverlay";
import { GameDto, TileDto, TileBonusDto } from "../../../types/generated/api-types";
import { usePreviousTiles } from "../../../hooks/usePreviousTiles";

interface ProjectedHexGridProps {
  gameState?: GameDto;
  onHexClick?: (hexCoordinate: string) => void;
  tileHighlightMode?: TileHighlightMode;
  vpIndicators?: TileVPIndicator[];
  animateHexEntrance?: boolean;
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
  tileHighlightMode,
  vpIndicators = [],
  animateHexEntrance = false,
}: ProjectedHexGridProps) {
  const SPHERE_RADIUS = 2.02;

  const newlyPlacedCities = usePreviousTiles(gameState?.board?.tiles);

  // Create lookup map for VP indicators by coordinate
  const vpIndicatorMap = useMemo(() => {
    const map = new Map<string, TileVPIndicator>();
    for (const indicator of vpIndicators) {
      map.set(indicator.coordinate, indicator);
    }
    return map;
  }, [vpIndicators]);

  // Determine highlight mode for individual tiles based on VP indicators or global mode
  const getHighlightModeForTile = (
    tileType: TileType,
    globalHighlightMode: TileHighlightMode | undefined,
    vpIndicator: TileVPIndicator | undefined,
  ): TileHighlightMode => {
    // Priority 1: Use VP indicator if present (for end-game VP counting)
    if (vpIndicator) {
      if (vpIndicator.type === "greenery") return "greenery";
      if (vpIndicator.type === "city-adjacency") return "adjacent";
    }

    // Priority 2: Use global highlight mode
    if (!globalHighlightMode) return null;
    if (globalHighlightMode === "greenery" && tileType === "greenery") return "greenery";
    if (globalHighlightMode === "city" && tileType === "city") return "city";
    // "adjacent" mode could be used for highlighting greeneries adjacent to cities
    return null;
  };

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
          isOceanSpace: tile.type === "ocean-space",
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

  // Get available hexes from current player's pending tile selection
  const availableHexes = gameState?.currentPlayer?.pendingTileSelection?.availableHexes || [];

  return (
    <>
      {projectedHexGrid.map((tile, index) => {
        const hexKey = HexGrid2D.coordinateToKey(tile.coordinate);
        const tileData = getTileData(tile);
        const isAvailable = availableHexes.includes(hexKey);
        const vpIndicator = vpIndicatorMap.get(hexKey);

        return (
          <ProjectedHexTile
            key={hexKey}
            tileData={tile}
            tileType={tileData.type}
            ownerId={tileData.ownerId}
            reservedById={tile.backendTile?.reservedBy || null}
            displayName={tile.backendTile?.displayName}
            onClick={() => {
              onHexClick?.(hexKey);
            }}
            isAvailableForPlacement={isAvailable}
            highlightMode={getHighlightModeForTile(tileData.type, tileHighlightMode, vpIndicator)}
            vpAmount={vpIndicator?.showVPText ? vpIndicator.amount : undefined}
            vpAnimating={vpIndicator?.isAnimating}
            animateEntrance={animateHexEntrance}
            entranceDelay={index * 15}
            isNewlyPlaced={newlyPlacedCities.has(hexKey)}
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
function projectToSphere(position2D: { x: number; y: number }, radius: number): THREE.Vector3 {
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
