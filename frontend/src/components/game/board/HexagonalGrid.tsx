import React, { useMemo } from "react";
import { GeodesicGrid } from "../../../utils/geodesic.ts";
import HexTile from "./HexTile.tsx";

// Backend tile types (should match backend/src/types/tiles.ts)
enum TileType {
  EMPTY = "empty",
  OCEAN = "ocean",
  GREENERY = "greenery",
  CITY = "city",
  SPECIAL = "special",
}

interface HexagonalGridProps {
  gameState?: any;
  onHexClick?: (hexCoordinate: string) => void;
}

export default function HexagonalGrid({
  gameState,
  onHexClick,
}: HexagonalGridProps) {
  // Generate the hexagonal grid positions
  const hexGrid = useMemo(() => {
    return GeodesicGrid.generateHexGrid();
  }, []);

  // Get tile data from game state or use default
  const getTileData = (hexCoordinate: string) => {
    // Try to get from game state board
    if (gameState?.board?.tiles) {
      const tile =
        gameState.board.tiles.get?.(hexCoordinate) ||
        gameState.board.tiles[hexCoordinate];
      if (tile) {
        return {
          type: tile.type || TileType.EMPTY,
          ownerId: tile.ownerId || null,
          specialType: tile.specialType || null,
        };
      }
    }

    // Default to empty tile
    return {
      type: TileType.EMPTY,
      ownerId: null,
      specialType: null,
    };
  };

  return (
    <>
      {hexGrid.map((hex, index) => {
        const hexKey = GeodesicGrid.coordinateToKey(hex.coordinate);
        const tileData = getTileData(hexKey);

        return (
          <HexTile
            key={hexKey}
            position={[
              hex.cartesianPosition.x,
              hex.cartesianPosition.y,
              hex.cartesianPosition.z,
            ]}
            coordinate={hex.coordinate}
            tileType={tileData.type}
            isOceanSpace={hex.isOceanSpace}
            bonuses={hex.bonuses}
            ownerId={tileData.ownerId}
            specialType={tileData.specialType}
            onClick={() => onHexClick?.(hexKey)}
          />
        );
      })}
    </>
  );
}
